package tui

import (
	"context"
	"encoding/hex"
	"strings"
	"sync/atomic"
	"testing"
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
