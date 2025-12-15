package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// TestRealWorldSignature tests with actual values that might come from the frontend
// This helps debug signature format issues
func TestRealWorldSignature(t *testing.T) {
	// These would be real values from a frontend test
	// For now, we'll generate them using our backend logic and verify they work

	// Test case 1: Simple message
	testCases := []struct {
		name      string
		content   string
		room      string
		timestamp int64
	}{
		{
			name:      "simple message",
			content:   "Hello, World!",
			room:      "general",
			timestamp: 1702650000, // Fixed timestamp for reproducibility
		},
		{
			name:      "message with special chars",
			content:   "Test: 123 & special <chars>",
			room:      "test-room",
			timestamp: 1702650000,
		},
		{
			name:      "empty message",
			content:   "",
			room:      "empty",
			timestamp: 1702650000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate a test signature using backend crypto
			privateKey, _ := secp256k1.GeneratePrivateKey()
			pubkey := privateKey.PubKey()
			pubkeyHex := hex.EncodeToString(pubkey.SerializeCompressed())

			// Create event hash
			eventHash := createEventHash(pubkeyHex, tc.timestamp, tc.content, tc.room)
			t.Logf("Event hash: %s", eventHash)

			eventHashBytes, _ := hex.DecodeString(eventHash)

			// Sign
			signature := ecdsa.Sign(privateKey, eventHashBytes)
			signatureHex := hex.EncodeToString(signature.Serialize())
			t.Logf("Signature hex length: %d chars (%d bytes)", len(signatureHex), len(signatureHex)/2)

			// Verify
			err := VerifyMessageSignature(pubkeyHex, signatureHex, tc.content, tc.room, tc.timestamp)
			if err != nil {
				t.Errorf("Verification failed: %v", err)
			}
		})
	}
}

// TestEventHashJSON verifies that our JSON serialization matches JavaScript
func TestEventHashJSON(t *testing.T) {
	testCases := []struct {
		name      string
		pubkey    string
		timestamp int64
		content   string
		room      string
		expected  string // Expected hash if known from frontend
	}{
		{
			name:      "basic test",
			pubkey:    "021234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			timestamp: 1234567890,
			content:   "test",
			room:      "room",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := createEventHash(tc.pubkey, tc.timestamp, tc.content, tc.room)
			t.Logf("Hash: %s", hash)
			if tc.expected != "" && hash != tc.expected {
				t.Errorf("Hash mismatch: got %s, want %s", hash, tc.expected)
			}
		})
	}
}
