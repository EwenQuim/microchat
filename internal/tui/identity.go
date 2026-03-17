package tui

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// identity holds a secp256k1 keypair.
type identity struct {
	privKey    *secp256k1.PrivateKey
	PubKeyHex  string
	PrivKeyHex string
}

// generateIdentity creates a new random keypair.
func generateIdentity() (identity, error) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return identity{}, fmt.Errorf("generate keypair: %w", err)
	}
	pubKey := privKey.PubKey()
	return identity{
		privKey:    privKey,
		PubKeyHex:  hex.EncodeToString(pubKey.SerializeCompressed()),
		PrivKeyHex: hex.EncodeToString(privKey.Serialize()),
	}, nil
}

// identityFromHex restores an identity from a hex-encoded private key.
func identityFromHex(privKeyHex string) (identity, error) {
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return identity{}, fmt.Errorf("decode private key: %w", err)
	}
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	pubKey := privKey.PubKey()
	return identity{
		privKey:    privKey,
		PubKeyHex:  hex.EncodeToString(pubKey.SerializeCompressed()),
		PrivKeyHex: privKeyHex,
	}, nil
}

// SignMessage signs a chat message using the Nostr event format expected by the backend.
// It returns a hex-encoded 64-byte compact ECDSA signature (R || S).
func (id identity) SignMessage(content, room string, timestamp int64) (string, error) {
	event := []any{0, id.PubKeyHex, timestamp, content, room}
	serialized, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("marshal event: %w", err)
	}
	hash := sha256.Sum256(serialized)
	sig := ecdsa.Sign(id.privKey, hash[:])
	compact := derToCompact(sig.Serialize())
	return hex.EncodeToString(compact), nil
}

// isHexChar reports whether c is a lowercase hex character [0-9a-f].
func isHexChar(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')
}

// isValidVanitySuffix returns true iff suffix has 1–4 lowercase hex chars.
func isValidVanitySuffix(suffix string) bool {
	if len(suffix) < 1 || len(suffix) > 4 {
		return false
	}
	for i := range len(suffix) {
		if !isHexChar(suffix[i]) {
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
				id, err := generateIdentity()
				if err != nil {
					select {
					case ch <- vanityResult{err: err}:
						cancel()
					default:
					}
					return
				}
				counter.Add(1)
				if strings.HasSuffix(id.PubKeyHex, suffix) {
					select {
					case ch <- vanityResult{id: id}:
						cancel()
					default:
					}
					return
				}
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

// GenerateKeypair generates a random secp256k1 keypair and returns hex strings.
func GenerateKeypair() (pubKeyHex, privKeyHex string, err error) {
	id, err := generateIdentity()
	if err != nil {
		return "", "", err
	}
	return id.PubKeyHex, id.PrivKeyHex, nil
}

// GenerateVanityKeypair finds a keypair whose public key hex ends with suffix.
func GenerateVanityKeypair(ctx context.Context, suffix string, counter *atomic.Int64) (pubKeyHex, privKeyHex string, err error) {
	id, err := generateVanityIdentity(ctx, suffix, counter)
	if err != nil {
		return "", "", err
	}
	return id.PubKeyHex, id.PrivKeyHex, nil
}

// ValidateVanitySuffix returns a descriptive error if suffix is invalid, or nil.
func ValidateVanitySuffix(suffix string) error {
	if len(suffix) == 0 {
		return fmt.Errorf("vanity suffix must be 1–4 lowercase hex characters")
	}
	if len(suffix) > 4 {
		return fmt.Errorf("vanity suffix too long: max 4 characters, got %d", len(suffix))
	}
	for i := range len(suffix) {
		if !isHexChar(suffix[i]) {
			return fmt.Errorf("vanity suffix %q contains invalid character %q (only 0-9, a-f allowed)", suffix, suffix[i])
		}
	}
	return nil
}

// CurrentIdentity reads the saved identity from ~/.config/microchat/config.json.
func CurrentIdentity() (pubKeyHex, privKeyHex string, err error) {
	cfg, err := loadConfig()
	if err != nil {
		return "", "", fmt.Errorf("load config: %w", err)
	}
	if cfg.Identity == nil {
		return "", "", fmt.Errorf("no identity configured")
	}
	return cfg.Identity.PublicKey, cfg.Identity.PrivateKey, nil
}

// derToCompact converts a DER-encoded ECDSA signature to a 64-byte R || S compact form.
// DER format: 0x30 [total_len] 0x02 [r_len] [r...] 0x02 [s_len] [s...]
func derToCompact(der []byte) []byte {
	rLen := int(der[3])
	r := der[4 : 4+rLen]
	sOffset := 4 + rLen
	sLen := int(der[sOffset+1])
	s := der[sOffset+2 : sOffset+2+sLen]

	out := make([]byte, 64)
	if len(r) > 32 {
		r = r[len(r)-32:] // trim leading zero padding
	}
	if len(s) > 32 {
		s = s[len(s)-32:]
	}
	copy(out[32-len(r):32], r)
	copy(out[64-len(s):64], s)
	return out
}
