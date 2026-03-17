package tui

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// identity holds a secp256k1 keypair.
type identity struct {
	privKey    *secp256k1.PrivateKey
	PubKeyHex  string
	NpubKey    string
	PrivKeyHex string
}

// generateIdentity creates a new random keypair.
func generateIdentity() (identity, error) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return identity{}, fmt.Errorf("generate keypair: %w", err)
	}
	pubKey := privKey.PubKey()
	pubKeyHex := hex.EncodeToString(pubKey.SerializeCompressed())
	npub, _ := pubKeyHexToNpub(pubKeyHex) // empty string on error
	return identity{
		privKey:    privKey,
		PubKeyHex:  pubKeyHex,
		NpubKey:    npub,
		PrivKeyHex: hex.EncodeToString(privKey.Serialize()),
	}, nil
}

// identityFromPrivKey builds a full identity from an already-generated private key.
// Used on vanity match to avoid the hex round-trip in the hot path.
func identityFromPrivKey(privKey *secp256k1.PrivateKey) identity {
	pubKey := privKey.PubKey()
	pubKeyHex := hex.EncodeToString(pubKey.SerializeCompressed())
	npub, _ := pubKeyHexToNpub(pubKeyHex) // empty string on error
	return identity{
		privKey:    privKey,
		PubKeyHex:  pubKeyHex,
		NpubKey:    npub,
		PrivKeyHex: hex.EncodeToString(privKey.Serialize()),
	}
}

// identityFromHex restores an identity from a hex-encoded private key.
func identityFromHex(privKeyHex string) (identity, error) {
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return identity{}, fmt.Errorf("decode private key: %w", err)
	}
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	pubKey := privKey.PubKey()
	pubKeyHex := hex.EncodeToString(pubKey.SerializeCompressed())
	npub, _ := pubKeyHexToNpub(pubKeyHex) // empty string on error
	return identity{
		privKey:    privKey,
		PubKeyHex:  pubKeyHex,
		NpubKey:    npub,
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

// GenerateKeypair generates a random secp256k1 keypair and returns npub and private key hex.
func GenerateKeypair() (npub, privKeyHex string, err error) {
	id, err := generateIdentity()
	if err != nil {
		return "", "", err
	}
	return id.NpubKey, id.PrivKeyHex, nil
}

// CurrentIdentity reads the saved identity from ~/.config/microchat/config.json.
func CurrentIdentity() (npub, privKeyHex string, err error) {
	cfg, err := loadConfig()
	if err != nil {
		return "", "", fmt.Errorf("load config: %w", err)
	}
	if len(cfg.Identities) == 0 {
		return "", "", fmt.Errorf("no identity configured")
	}
	idx := cfg.ActiveIndex
	if idx < 0 || idx >= len(cfg.Identities) {
		idx = 0
	}
	id, err := identityFromHex(cfg.Identities[idx].PrivateKey)
	if err != nil {
		return "", "", err
	}
	return id.NpubKey, id.PrivKeyHex, nil
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
