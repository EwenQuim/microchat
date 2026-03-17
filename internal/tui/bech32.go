package tui

import (
	"encoding/hex"
	"fmt"
)

const bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

// isValidBech32Char reports whether c is in the bech32 charset (lowercase).
func isValidBech32Char(c byte) bool {
	for i := 0; i < len(bech32Charset); i++ {
		if bech32Charset[i] == c {
			return true
		}
	}
	return false
}

func bech32Polymod(values []byte) uint32 {
	gen := []uint32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := uint32(1)
	for _, v := range values {
		b := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ uint32(v)
		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				chk ^= gen[i]
			}
		}
	}
	return chk
}

func bech32HRPExpand(hrp string) []byte {
	result := make([]byte, len(hrp)*2+1)
	for i := 0; i < len(hrp); i++ {
		result[i] = hrp[i] >> 5
		result[i+len(hrp)+1] = hrp[i] & 31
	}
	result[len(hrp)] = 0
	return result
}

func bech32CreateChecksum(hrp string, data []byte) []byte {
	values := append(bech32HRPExpand(hrp), data...)
	polymod := bech32Polymod(append(values, 0, 0, 0, 0, 0, 0)) ^ 1
	checksum := make([]byte, 6)
	for i := 0; i < 6; i++ {
		checksum[i] = byte((polymod >> uint(5*(5-i))) & 31)
	}
	return checksum
}

func bech32Encode(hrp string, data []byte) string {
	combined := append(data, bech32CreateChecksum(hrp, data)...)
	result := hrp + "1"
	for _, b := range combined {
		result += string(bech32Charset[b])
	}
	return result
}

func convertBits(data []byte, fromBits, toBits uint, pad bool) ([]byte, error) {
	acc := 0
	bits := uint(0)
	result := []byte{}
	maxv := (1 << toBits) - 1
	for _, value := range data {
		acc = (acc << fromBits) | int(value)
		bits += fromBits
		for bits >= toBits {
			bits -= toBits
			result = append(result, byte((acc>>bits)&maxv))
		}
	}
	if pad {
		if bits > 0 {
			result = append(result, byte((acc<<(toBits-bits))&maxv))
		}
	} else if bits >= fromBits || ((acc<<(toBits-bits))&maxv) != 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	return result, nil
}

// pubKeyHexToNpub converts a compressed secp256k1 public key hex (66 chars)
// to Nostr bech32 npub format.
func pubKeyHexToNpub(hexPubKey string) (string, error) {
	if len(hexPubKey) != 66 {
		return "", fmt.Errorf("expected 66-char hex pubkey, got %d", len(hexPubKey))
	}
	// Strip 2-char compression prefix (02/03) → 64-char hex = 32 bytes (x-coord)
	xBytes, err := hex.DecodeString(hexPubKey[2:])
	if err != nil {
		return "", fmt.Errorf("decode pubkey hex: %w", err)
	}
	words, err := convertBits(xBytes, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("convertBits: %w", err)
	}
	return bech32Encode("npub", words), nil
}
