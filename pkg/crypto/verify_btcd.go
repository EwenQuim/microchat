package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

// VerifyMessageSignatureBTCD verifies a signature using btcd library
// This is more compatible with noble-secp256k1
func VerifyMessageSignatureBTCD(pubkeyHex, signatureHex, content, room string, timestamp int64) error {
	// Decode public key from hex
	pubkeyBytes, err := hex.DecodeString(pubkeyHex)
	if err != nil {
		return fmt.Errorf("invalid public key hex: %w", err)
	}

	pubkey, err := btcec.ParsePubKey(pubkeyBytes)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	// Decode signature from hex
	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}

	if len(signatureBytes) != 64 {
		return fmt.Errorf("invalid signature length: expected 64 bytes, got %d", len(signatureBytes))
	}

	// Parse compact signature: first 32 bytes are R, last 32 bytes are S
	// btcd's ecdsa.ParseSignature expects DER, so we need to use ModNScalar
	var r, s btcec.ModNScalar
	r.SetByteSlice(signatureBytes[:32])
	s.SetByteSlice(signatureBytes[32:])
	signature := ecdsa.NewSignature(&r, &s)

	// Create the event hash
	eventHash := createEventHashBTCD(pubkeyHex, timestamp, content, room)
	eventHashBytes, err := hex.DecodeString(eventHash)
	if err != nil {
		return fmt.Errorf("failed to decode event hash: %w", err)
	}

	// Verify the signature
	if !signature.Verify(eventHashBytes, pubkey) {
		return fmt.Errorf("signature verification failed: signature does not match")
	}

	return nil
}

// createEventHashBTCD creates the same event hash as the frontend
func createEventHashBTCD(pubkey string, timestamp int64, content, room string) string {
	event := []interface{}{
		0,
		pubkey,
		timestamp,
		content,
		room,
	}

	serialized, _ := json.Marshal(event)
	hash := sha256.Sum256(serialized)
	return hex.EncodeToString(hash[:])
}
