package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// TestCrossVerify tests that we can verify a signature created by noble-secp256k1
func TestCrossVerify(t *testing.T) {
	// Actual signature from the frontend
	content := "hello"
	room := "general"
	timestamp := int64(1765796171)
	pubkeyHex := "036903c174e82ef03e7fd5d721f233fa7b86eea298fda2e27372015b32d2bc7a29"
	signatureHex := "18b5d24af7cf955e68cbbdfa111cd75ff7f3290eee1e6e73370a60d2591976464312bd757fd8a9fa6b915361bf6727acc62de7fc2f920ebab00a3465d9fe2ce7"

	// Verify event hash matches
	eventHash := createEventHash(pubkeyHex, timestamp, content, room)
	expectedHash := "1b8c1e93eea9e9f8307f954ee1b9f134d2515b743fb2932eda7079763957b718"
	if eventHash != expectedHash {
		t.Fatalf("Event hash mismatch: got %s, want %s", eventHash, expectedHash)
	}
	t.Logf("Event hash matches: %s", eventHash)

	// Decode public key
	pubkeyBytes, err := hex.DecodeString(pubkeyHex)
	if err != nil {
		t.Fatalf("Failed to decode pubkey: %v", err)
	}
	t.Logf("Pubkey bytes length: %d", len(pubkeyBytes))

	pubkey, err := secp256k1.ParsePubKey(pubkeyBytes)
	if err != nil {
		t.Fatalf("Failed to parse pubkey: %v", err)
	}
	t.Logf("Pubkey parsed successfully")

	// Decode signature
	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		t.Fatalf("Failed to decode signature: %v", err)
	}
	t.Logf("Signature bytes length: %d", len(signatureBytes))

	// Parse as compact signature
	if len(signatureBytes) != 64 {
		t.Fatalf("Expected 64-byte signature, got %d", len(signatureBytes))
	}

	// Try different ways of parsing the signature
	t.Run("Method1_ModNScalar", func(t *testing.T) {
		var r, s secp256k1.ModNScalar
		r.SetByteSlice(signatureBytes[:32])
		s.SetByteSlice(signatureBytes[32:])
		signature := ecdsa.NewSignature(&r, &s)

		eventHashBytes, _ := hex.DecodeString(eventHash)
		verified := signature.Verify(eventHashBytes, pubkey)
		t.Logf("Verification result: %v", verified)
		if !verified {
			t.Logf("Method 1 failed")
		} else {
			t.Logf("Method 1 SUCCESS!")
		}
	})

	t.Run("Method2_BigEndian", func(t *testing.T) {
		// Check if the bytes are in the correct order
		t.Logf("R bytes (first 32): %x", signatureBytes[:32])
		t.Logf("S bytes (last 32): %x", signatureBytes[32:])
	})

	// Try signing with the same data to compare
	t.Run("CompareWithBackendSignature", func(t *testing.T) {
		// Generate a test signature with backend
		privateKey, _ := secp256k1.GeneratePrivateKey()
		testPubkey := privateKey.PubKey()
		testPubkeyHex := hex.EncodeToString(testPubkey.SerializeCompressed())

		testEventHash := createEventHash(testPubkeyHex, timestamp, content, room)
		testEventHashBytes, _ := hex.DecodeString(testEventHash)

		testSignature := ecdsa.Sign(privateKey, testEventHashBytes)
		testSignatureBytes := testSignature.Serialize()

		t.Logf("Backend signature length: %d bytes", len(testSignatureBytes))
		t.Logf("Frontend signature length: %d bytes", len(signatureBytes))

		// Try to verify backend signature
		verified := testSignature.Verify(testEventHashBytes, testPubkey)
		t.Logf("Backend self-verification: %v", verified)
	})
}
