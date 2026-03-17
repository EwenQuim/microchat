package tui

import (
	"context"
	"encoding/hex"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func TestGenerateIdentity(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity() error: %v", err)
	}

	// Compressed secp256k1 pubkey = 33 bytes = 66 hex chars
	if len(id.PubKeyHex) != 66 {
		t.Errorf("PubKeyHex length = %d, want 66", len(id.PubKeyHex))
	}
	// Private key = 32 bytes = 64 hex chars
	if len(id.PrivKeyHex) != 64 {
		t.Errorf("PrivKeyHex length = %d, want 64", len(id.PrivKeyHex))
	}
	if id.privKey == nil {
		t.Error("privKey is nil")
	}
}

func TestGenerateIdentityRoundtrip(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity() error: %v", err)
	}

	restored, err := identityFromHex(id.PrivKeyHex)
	if err != nil {
		t.Fatalf("identityFromHex() error: %v", err)
	}

	if restored.PubKeyHex != id.PubKeyHex {
		t.Errorf("roundtrip PubKeyHex mismatch: got %s, want %s", restored.PubKeyHex, id.PubKeyHex)
	}
	if restored.PrivKeyHex != id.PrivKeyHex {
		t.Errorf("roundtrip PrivKeyHex mismatch: got %s, want %s", restored.PrivKeyHex, id.PrivKeyHex)
	}
}

func TestIdentityFromHex_Valid(t *testing.T) {
	// Generate a known identity and restore it
	orig, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity() error: %v", err)
	}

	id, err := identityFromHex(orig.PrivKeyHex)
	if err != nil {
		t.Fatalf("identityFromHex() error: %v", err)
	}

	if id.PubKeyHex != orig.PubKeyHex {
		t.Errorf("PubKeyHex mismatch: got %s, want %s", id.PubKeyHex, orig.PubKeyHex)
	}
}

func TestIdentityFromHex_InvalidHex(t *testing.T) {
	_, err := identityFromHex("not-valid-hex!!")
	if err == nil {
		t.Error("expected error for invalid hex, got nil")
	}
}

func TestIdentityFromHex_WrongLength(t *testing.T) {
	// Too short — valid hex but wrong byte count
	_, err := identityFromHex("deadbeef")
	// secp256k1.PrivKeyFromBytes doesn't error, but the pubkey derivation
	// should still succeed (library is permissive). Verify we at least don't panic.
	_ = err
}

func TestSignMessage_Length(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity() error: %v", err)
	}

	sig, err := id.SignMessage("hello", "general", 1234567890)
	if err != nil {
		t.Fatalf("SignMessage() error: %v", err)
	}

	// Compact R||S = 64 bytes = 128 hex chars
	if len(sig) != 128 {
		t.Errorf("signature length = %d, want 128", len(sig))
	}
	if _, err := hex.DecodeString(sig); err != nil {
		t.Errorf("signature is not valid hex: %v", err)
	}
}

func TestSignMessage_DifferentContentDifferentSig(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity() error: %v", err)
	}

	sig1, err := id.SignMessage("hello", "general", 1234567890)
	if err != nil {
		t.Fatalf("SignMessage() error: %v", err)
	}
	sig2, err := id.SignMessage("world", "general", 1234567890)
	if err != nil {
		t.Fatalf("SignMessage() error: %v", err)
	}

	if sig1 == sig2 {
		t.Error("different content should produce different signatures")
	}
}

func TestSignMessage_Deterministic(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity() error: %v", err)
	}

	sig1, err := id.SignMessage("hello", "room", 42)
	if err != nil {
		t.Fatalf("SignMessage() error: %v", err)
	}
	sig2, err := id.SignMessage("hello", "room", 42)
	if err != nil {
		t.Fatalf("SignMessage() error: %v", err)
	}

	// secp256k1 deterministic signing (RFC 6979) — same inputs must produce same sig
	if sig1 != sig2 {
		t.Error("signing same inputs should be deterministic")
	}
}

func TestDerToCompact_Length(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity() error: %v", err)
	}

	// Sign something to produce a real DER signature
	sig, err := id.SignMessage("test", "room", 1)
	if err != nil {
		t.Fatalf("SignMessage() error: %v", err)
	}

	compact, err := hex.DecodeString(sig)
	if err != nil {
		t.Fatalf("decode compact sig: %v", err)
	}
	if len(compact) != 64 {
		t.Errorf("compact sig length = %d, want 64", len(compact))
	}
}

func TestDerToCompact_NotAllZero(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity() error: %v", err)
	}

	sig, err := id.SignMessage("content", "room", 999)
	if err != nil {
		t.Fatalf("SignMessage() error: %v", err)
	}

	if strings.Count(sig, "0") == len(sig) {
		t.Error("compact signature is all zeros")
	}
}

func TestIsValidVanitySuffix(t *testing.T) {
	// valid: chars in bech32 charset, 1–5 chars
	valid := []string{"a", "0f", "cafe", "3j5k", "ewen0"}
	for _, s := range valid {
		if !isValidVanitySuffix(s) {
			t.Errorf("isValidVanitySuffix(%q) = false, want true", s)
		}
	}
	// invalid: empty, too long (6+), contains '1'/'b'/'i'/'o' (excluded), uppercase
	invalid := []string{"", "cafe00", "1b", "AB"}
	for _, s := range invalid {
		if isValidVanitySuffix(s) {
			t.Errorf("isValidVanitySuffix(%q) = true, want false", s)
		}
	}
}

func TestGenerateVanityIdentity_Finds(t *testing.T) {
	ctx := context.Background()
	var counter atomic.Int64
	id, err := generateVanityIdentity(ctx, "q", &counter)
	if err != nil {
		t.Fatalf("generateVanityIdentity() error: %v", err)
	}
	if !strings.HasSuffix(id.NpubKey, "q") {
		t.Errorf("NpubKey %q does not end with 'q'", id.NpubKey)
	}
}

func TestPubKeyHexToNpub_Format(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity() error: %v", err)
	}
	npub, err := pubKeyHexToNpub(id.PubKeyHex)
	if err != nil {
		t.Fatalf("pubKeyHexToNpub() error: %v", err)
	}
	if !strings.HasPrefix(npub, "npub1") {
		t.Errorf("npub %q does not start with 'npub1'", npub)
	}
	if len(npub) != 63 {
		t.Errorf("npub length = %d, want 63", len(npub))
	}
}

func TestGenerateVanityIdentity_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var counter atomic.Int64
	_, err := generateVanityIdentity(ctx, "cafe", &counter)
	if err == nil {
		t.Error("expected error when context cancelled, got nil")
	}
	if ctx.Err() == nil {
		t.Error("expected ctx.Err() != nil")
	}
}

func TestGenerateVanityIdentity_CounterIncrements(t *testing.T) {
	ctx := context.Background()
	var counter atomic.Int64
	_, err := generateVanityIdentity(ctx, "0", &counter)
	if err != nil {
		t.Fatalf("generateVanityIdentity() error: %v", err)
	}
	if counter.Load() == 0 {
		t.Error("expected counter > 0 after finding a vanity key")
	}
}

func TestBech32SuffixToVals(t *testing.T) {
	// Round-trip: bech32Charset[val] == char for every charset character.
	for i := range len(bech32Charset) {
		c := bech32Charset[i]
		vals, err := bech32SuffixToVals(string(c))
		if err != nil {
			t.Errorf("bech32SuffixToVals(%q) error: %v", c, err)
			continue
		}
		if len(vals) != 1 {
			t.Errorf("bech32SuffixToVals(%q) len = %d, want 1", c, len(vals))
			continue
		}
		if bech32Charset[vals[0]] != c {
			t.Errorf("round-trip failed for %q: got index %d → %q", c, vals[0], bech32Charset[vals[0]])
		}
	}

	// Error on invalid character.
	_, err := bech32SuffixToVals("b") // 'b' is not in bech32 charset
	if err == nil {
		t.Error("expected error for invalid bech32 char 'b', got nil")
	}
}

func TestNpubSuffixMatch_AgainstReference(t *testing.T) {
	for range 100 {
		privKey, err := secp256k1.GeneratePrivateKey()
		if err != nil {
			t.Fatalf("GeneratePrivateKey: %v", err)
		}
		id := identityFromPrivKey(privKey)
		compressed := privKey.PubKey().SerializeCompressed()

		for suffixLen := 1; suffixLen <= 6; suffixLen++ {
			suffix := id.NpubKey[len(id.NpubKey)-suffixLen:]
			target, err := bech32SuffixToVals(suffix)
			if err != nil {
				t.Fatalf("bech32SuffixToVals(%q): %v", suffix, err)
			}
			got := npubSuffixMatch(compressed[1:], target)
			want := strings.HasSuffix(id.NpubKey, suffix)
			if got != want {
				t.Errorf("npubSuffixMatch mismatch for npub=%s suffix=%q: got %v want %v",
					id.NpubKey, suffix, got, want)
			}
		}
	}
}

func TestNpubSuffixMatch_LongSuffix(t *testing.T) {
	for range 20 {
		privKey, err := secp256k1.GeneratePrivateKey()
		if err != nil {
			t.Fatalf("GeneratePrivateKey: %v", err)
		}
		id := identityFromPrivKey(privKey)
		compressed := privKey.PubKey().SerializeCompressed()

		suffix := id.NpubKey[len(id.NpubKey)-10:]
		target, err := bech32SuffixToVals(suffix)
		if err != nil {
			t.Fatalf("bech32SuffixToVals(%q): %v", suffix, err)
		}
		if !npubSuffixMatch(compressed[1:], target) {
			t.Errorf("npubSuffixMatch returned false for own suffix: npub=%s suffix=%q",
				id.NpubKey, suffix)
		}

		// Negative check: a suffix with a different last char should not match.
		otherChar := bech32Charset[(strings.IndexByte(bech32Charset, suffix[len(suffix)-1])+1)%len(bech32Charset)]
		wrongSuffix := suffix[:len(suffix)-1] + string(otherChar)
		wrongTarget, _ := bech32SuffixToVals(wrongSuffix)
		if id.NpubKey[len(id.NpubKey)-10:] != wrongSuffix {
			if npubSuffixMatch(compressed[1:], wrongTarget) {
				t.Errorf("npubSuffixMatch returned true for wrong suffix: npub=%s wrongSuffix=%q",
					id.NpubKey, wrongSuffix)
			}
		}
	}
}

func BenchmarkGenerateVanityIteration(b *testing.B) {
	target, _ := bech32SuffixToVals("qqq")
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		privKey, _ := secp256k1.GeneratePrivateKey()
		compressed := privKey.PubKey().SerializeCompressed()
		npubSuffixMatch(compressed[1:], target)
	}
}

func BenchmarkGenerateVanityIteration_Baseline(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		vanityIterationOld("qqq")
	}
}
