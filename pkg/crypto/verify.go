package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// VerifyMessageSignature verifies a Nostr-style message signature
// This matches the signing logic in the frontend
func VerifyMessageSignature(pubkeyHex, signatureHex, content, room string, timestamp int64) error {
	// Decode public key from hex
	pubkeyBytes, err := hex.DecodeString(pubkeyHex)
	if err != nil {
		return fmt.Errorf("invalid public key hex: %w", err)
	}

	pubkey, err := secp256k1.ParsePubKey(pubkeyBytes)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	// Decode signature from hex
	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}

	// @noble/secp256k1 v3 uses compact signatures (64 bytes: 32-byte R + 32-byte S)
	var signature *ecdsa.Signature
	if len(signatureBytes) == 64 {
		// Compact signature format: parse R and S as big-endian integers
		// noble-secp256k1 uses standard compact format (r || s)
		var r, s secp256k1.ModNScalar

		// SetByteSlice interprets bytes as big-endian
		overflow := r.SetByteSlice(signatureBytes[:32])
		if overflow {
			return fmt.Errorf("signature R value overflows")
		}

		overflow = s.SetByteSlice(signatureBytes[32:])
		if overflow {
			return fmt.Errorf("signature S value overflows")
		}

		signature = ecdsa.NewSignature(&r, &s)
	} else {
		// Try parsing as DER signature as fallback
		signature, err = ecdsa.ParseDERSignature(signatureBytes)
		if err != nil {
			return fmt.Errorf("invalid signature format (expected 64 bytes compact or DER): got %d bytes", len(signatureBytes))
		}
	}

	// Create the event hash using the same serialization as the frontend
	eventHash := createEventHash(pubkeyHex, timestamp, content, room)
	eventHashBytes, err := hex.DecodeString(eventHash)
	if err != nil {
		return fmt.Errorf("failed to decode event hash: %w", err)
	}

	// Verify the signature
	if !signature.Verify(eventHashBytes, pubkey) {
		// Debug: print what we're verifying
		fmt.Printf("DEBUG Signature Verification Failed:\n")
		fmt.Printf("  Pubkey: %s\n", pubkeyHex)
		fmt.Printf("  Signature length: %d bytes\n", len(signatureBytes))
		fmt.Printf("  Event hash: %s\n", eventHash)
		fmt.Printf("  Content: %s\n", content)
		fmt.Printf("  Room: %s\n", room)
		fmt.Printf("  Timestamp: %d\n", timestamp)
		return fmt.Errorf("signature verification failed: signature does not match")
	}

	return nil
}

// createEventHash creates a canonical event hash following Nostr's format
// This must match the frontend implementation exactly
func createEventHash(pubkey string, timestamp int64, content, room string) string {
	// Nostr event format: [0, pubkey, created_at, content, room]
	event := []any{
		0,         // reserved for future use
		pubkey,    // public key
		timestamp, // unix timestamp
		content,   // message content
		room,      // room name
	}

	// Serialize to JSON
	serialized, _ := json.Marshal(event)

	// Hash with SHA-256
	hash := sha256.Sum256(serialized)

	return hex.EncodeToString(hash[:])
}

// VerifyRoomVisibilitySignature verifies a signature for room visibility update
// The signature covers the action, room name, hidden status, and timestamp
func VerifyRoomVisibilitySignature(pubkeyHex, signatureHex, roomName string, hidden bool, timestamp int64) error {
	// Decode public key from hex
	pubkeyBytes, err := hex.DecodeString(pubkeyHex)
	if err != nil {
		return fmt.Errorf("invalid public key hex: %w", err)
	}

	pubkey, err := secp256k1.ParsePubKey(pubkeyBytes)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	// Decode signature from hex
	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}

	// @noble/secp256k1 v3 uses compact signatures (64 bytes: 32-byte R + 32-byte S)
	var signature *ecdsa.Signature
	if len(signatureBytes) == 64 {
		// Compact signature format: parse R and S as big-endian integers
		var r, s secp256k1.ModNScalar

		overflow := r.SetByteSlice(signatureBytes[:32])
		if overflow {
			return fmt.Errorf("signature R value overflows")
		}

		overflow = s.SetByteSlice(signatureBytes[32:])
		if overflow {
			return fmt.Errorf("signature S value overflows")
		}

		signature = ecdsa.NewSignature(&r, &s)
	} else {
		// Try parsing as DER signature as fallback
		signature, err = ecdsa.ParseDERSignature(signatureBytes)
		if err != nil {
			return fmt.Errorf("invalid signature format (expected 64 bytes compact or DER): got %d bytes", len(signatureBytes))
		}
	}

	// Create the event hash for room visibility update
	eventHash := createRoomVisibilityHash(pubkeyHex, roomName, hidden, timestamp)
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

// createRoomVisibilityHash creates a canonical hash for room visibility updates
// Format: [0, pubkey, timestamp, "update_room_visibility", roomName, hidden]
func createRoomVisibilityHash(pubkey, roomName string, hidden bool, timestamp int64) string {
	event := []any{
		0,                        // reserved for future use
		pubkey,                   // public key
		timestamp,                // unix timestamp
		"update_room_visibility", // action type
		roomName,                 // room name
		hidden,                   // hidden status
	}

	// Serialize to JSON
	serialized, _ := json.Marshal(event)

	// Hash with SHA-256
	hash := sha256.Sum256(serialized)

	return hex.EncodeToString(hash[:])
}
