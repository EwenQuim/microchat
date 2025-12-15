package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

func TestVerifyBTCDWithGeneratedSignature(t *testing.T) {
	// Generate a signature using btcd and verify it works
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	pubkey := privateKey.PubKey()
	pubkeyHex := hex.EncodeToString(pubkey.SerializeCompressed())

	content := "hello"
	room := "general"
	timestamp := int64(1765796171)

	// Create event hash
	eventHash := createEventHashBTCD(pubkeyHex, timestamp, content, room)
	eventHashBytes, err := hex.DecodeString(eventHash)
	if err != nil {
		t.Fatalf("Failed to decode event hash: %v", err)
	}

	// Sign with btcd
	signature := ecdsa.Sign(privateKey, eventHashBytes)

	// Convert to compact format (64 bytes: R || S)
	signatureBytes := make([]byte, 64)
	r, s := signature.R(), signature.S()
	r.PutBytesUnchecked(signatureBytes[:32])
	s.PutBytesUnchecked(signatureBytes[32:])
	signatureHex := hex.EncodeToString(signatureBytes)

	t.Logf("Generated signature (compact 64 bytes): %s", signatureHex)

	// Verify with our implementation
	err = VerifyMessageSignatureBTCD(pubkeyHex, signatureHex, content, room, timestamp)
	if err != nil {
		t.Errorf("BTCD verification failed: %v", err)
	} else {
		t.Log("BTCD verification SUCCESS!")
	}
}

func TestVerifyBTCD(t *testing.T) {
	// Test with valid signature data generated using a known keypair
	// The previous test data had an invalid signature (signed with different key)
	// So we generate valid test data here

	// Fixed private key for reproducible tests (32 bytes)
	privateKeyHex := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)
	pubkey := privateKey.PubKey()
	pubkeyHex := hex.EncodeToString(pubkey.SerializeCompressed())

	content := "hello"
	room := "general"
	timestamp := int64(1765796171)

	// Create event hash
	eventHash := createEventHashBTCD(pubkeyHex, timestamp, content, room)
	eventHashBytes, err := hex.DecodeString(eventHash)
	if err != nil {
		t.Fatalf("Failed to decode event hash: %v", err)
	}

	// Sign with btcd
	signature := ecdsa.Sign(privateKey, eventHashBytes)

	// Convert to compact format (64 bytes: R || S) matching noble-secp256k1 format
	signatureBytes := make([]byte, 64)
	r, s := signature.R(), signature.S()
	r.PutBytesUnchecked(signatureBytes[:32])
	s.PutBytesUnchecked(signatureBytes[32:])
	signatureHex := hex.EncodeToString(signatureBytes)

	t.Logf("Testing with valid signature data:")
	t.Logf("  Pubkey: %s", pubkeyHex)
	t.Logf("  Signature: %s", signatureHex)
	t.Logf("  Event hash: %s", eventHash)

	// Verify with our implementation
	err = VerifyMessageSignatureBTCD(pubkeyHex, signatureHex, content, room, timestamp)
	if err != nil {
		t.Errorf("BTCD verification failed: %v", err)
	} else {
		t.Log("BTCD verification SUCCESS!")
	}
}
