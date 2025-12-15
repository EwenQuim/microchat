package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

func TestCreateEventHash(t *testing.T) {
	// Test that event hash creation is deterministic
	pubkey := "02abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	timestamp := int64(1234567890)
	content := "Hello, World!"
	room := "general"

	hash1 := createEventHash(pubkey, timestamp, content, room)
	hash2 := createEventHash(pubkey, timestamp, content, room)

	if hash1 != hash2 {
		t.Errorf("Event hash should be deterministic, got %s and %s", hash1, hash2)
	}

	// Verify hash is 64 characters (32 bytes in hex)
	if len(hash1) != 64 {
		t.Errorf("Event hash should be 64 hex characters, got %d", len(hash1))
	}
}

func TestVerifyMessageSignature(t *testing.T) {
	// Generate a test keypair
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	pubkey := privateKey.PubKey()
	pubkeyHex := hex.EncodeToString(pubkey.SerializeCompressed())

	// Create test message
	content := "Test message"
	room := "test-room"
	timestamp := int64(1234567890)

	// Create event hash
	eventHash := createEventHash(pubkeyHex, timestamp, content, room)
	eventHashBytes, err := hex.DecodeString(eventHash)
	if err != nil {
		t.Fatalf("Failed to decode event hash: %v", err)
	}

	// Sign the message
	signature := ecdsa.Sign(privateKey, eventHashBytes)
	signatureHex := hex.EncodeToString(signature.Serialize())

	// Verify the signature
	err = VerifyMessageSignature(pubkeyHex, signatureHex, content, room, timestamp)
	if err != nil {
		t.Errorf("Signature verification failed: %v", err)
	}
}

func TestVerifyMessageSignature_InvalidSignature(t *testing.T) {
	// Generate a test keypair
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	pubkey := privateKey.PubKey()
	pubkeyHex := hex.EncodeToString(pubkey.SerializeCompressed())

	// Create test message
	content := "Test message"
	room := "test-room"
	timestamp := int64(1234567890)

	// Sign a different message
	differentContent := "Different message"
	differentHash := createEventHash(pubkeyHex, timestamp, differentContent, room)
	differentHashBytes, _ := hex.DecodeString(differentHash)
	signature := ecdsa.Sign(privateKey, differentHashBytes)
	signatureHex := hex.EncodeToString(signature.Serialize())

	// Try to verify with original content (should fail)
	err = VerifyMessageSignature(pubkeyHex, signatureHex, content, room, timestamp)
	if err == nil {
		t.Error("Expected signature verification to fail for mismatched content")
	}
}

func TestVerifyMessageSignature_InvalidPubkey(t *testing.T) {
	err := VerifyMessageSignature("invalid_hex", "signature", "content", "room", 123)
	if err == nil {
		t.Error("Expected error for invalid pubkey hex")
	}
}

func TestVerifyMessageSignature_InvalidSignatureFormat(t *testing.T) {
	// Generate a valid pubkey
	privateKey, _ := secp256k1.GeneratePrivateKey()
	pubkey := privateKey.PubKey()
	pubkeyHex := hex.EncodeToString(pubkey.SerializeCompressed())

	// Use an invalid signature (wrong length)
	err := VerifyMessageSignature(pubkeyHex, "abcd", "content", "room", 123)
	if err == nil {
		t.Error("Expected error for invalid signature format")
	}
}

// Benchmark event hash creation
func BenchmarkCreateEventHash(b *testing.B) {
	pubkey := "02abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	timestamp := int64(1234567890)
	content := "Hello, World!"
	room := "general"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		createEventHash(pubkey, timestamp, content, room)
	}
}

// Benchmark signature verification
func BenchmarkVerifyMessageSignature(b *testing.B) {
	// Setup
	privateKey, _ := secp256k1.GeneratePrivateKey()
	pubkey := privateKey.PubKey()
	pubkeyHex := hex.EncodeToString(pubkey.SerializeCompressed())
	content := "Test message"
	room := "test-room"
	timestamp := int64(1234567890)
	eventHash := createEventHash(pubkeyHex, timestamp, content, room)
	eventHashBytes, _ := hex.DecodeString(eventHash)
	signature := ecdsa.Sign(privateKey, eventHashBytes)
	signatureHex := hex.EncodeToString(signature.Serialize())

	for b.Loop() {
		_ = VerifyMessageSignature(pubkeyHex, signatureHex, content, room, timestamp)
	}
}
