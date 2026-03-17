package tui

import (
	"context"
	"fmt"
	"runtime"
	"sync/atomic"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// FormatCount formats n with underscore thousands separators (e.g. 1_234_567).
func FormatCount(n int64) string {
	s := fmt.Sprintf("%d", n)
	out := make([]byte, 0, len(s)+(len(s)-1)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '_')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

// isValidVanitySuffix returns true iff suffix has 1–5 bech32 chars.
func isValidVanitySuffix(suffix string) bool {
	if len(suffix) < 1 || len(suffix) > 5 {
		return false
	}
	for i := range len(suffix) {
		if !isValidBech32Char(suffix[i]) {
			return false
		}
	}
	return true
}

type vanityResult struct {
	id  identity
	err error
}

// generateVanityIdentity spawns NumCPU goroutines to find a keypair whose
// compressed public key hex ends with suffix. counter is incremented for
// every attempt. Returns context.Err() if the context is cancelled before a
// match is found.
func generateVanityIdentity(ctx context.Context, suffix string, counter *atomic.Int64) (identity, error) {
	target, err := bech32SuffixToVals(suffix)
	if err != nil {
		return identity{}, fmt.Errorf("invalid suffix: %w", err)
	}

	ch := make(chan vanityResult, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for range runtime.NumCPU() {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				privKey, err := secp256k1.GeneratePrivateKey()
				if err != nil {
					select {
					case ch <- vanityResult{err: err}:
						cancel()
					default:
					}
					return
				}
				counter.Add(1)
				compressed := privKey.PubKey().SerializeCompressed()
				if !npubSuffixMatch(compressed[1:], target) {
					continue
				}
				select {
				case ch <- vanityResult{id: identityFromPrivKey(privKey)}:
					cancel()
				default:
				}
				return
			}
		}()
	}

	select {
	case res := <-ch:
		return res.id, res.err
	case <-ctx.Done():
		// drain in case a result arrived at the same time
		select {
		case res := <-ch:
			return res.id, res.err
		default:
		}
		return identity{}, ctx.Err()
	}
}

// GenerateVanityKeypair finds a keypair whose npub ends with suffix.
func GenerateVanityKeypair(ctx context.Context, suffix string, counter *atomic.Int64) (npub, privKeyHex string, err error) {
	id, err := generateVanityIdentity(ctx, suffix, counter)
	if err != nil {
		return "", "", err
	}
	return id.NpubKey, id.PrivKeyHex, nil
}

// ValidateVanitySuffix returns a descriptive error if suffix is invalid, or nil.
func ValidateVanitySuffix(suffix string) error {
	if len(suffix) == 0 {
		return fmt.Errorf("vanity suffix must be 1–5 bech32 characters")
	}
	if len(suffix) > 5 {
		return fmt.Errorf("vanity suffix too long: max 5 characters, got %d (use --unsafe-cpu-usage to bypass)", len(suffix))
	}
	for i := range len(suffix) {
		if !isValidBech32Char(suffix[i]) {
			return fmt.Errorf("vanity suffix %q contains invalid character %q (only bech32 charset allowed: %s)", suffix, suffix[i], bech32Charset)
		}
	}
	return nil
}

// ValidateVanitySuffixUnsafe validates only charset validity (no length cap).
// Intended for CLI use with --unsafe-cpu-usage.
func ValidateVanitySuffixUnsafe(suffix string) error {
	if len(suffix) == 0 {
		return fmt.Errorf("vanity suffix must be at least 1 bech32 character")
	}
	for i := range len(suffix) {
		if !isValidBech32Char(suffix[i]) {
			return fmt.Errorf("vanity suffix %q contains invalid character %q (only bech32 charset allowed: %s)", suffix, suffix[i], bech32Charset)
		}
	}
	return nil
}
