package tui

import (
	"encoding/hex"
	"fmt"
)

const bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

var (
	// npubHRPPolymodState is the bech32 polymod state after processing bech32HRPExpand("npub"),
	// precomputed once to avoid 9 polymod iterations per vanity attempt.
	npubHRPPolymodState uint32
	// bech32CharsetReverse maps an ASCII byte to its 5-bit bech32 value (255 = invalid).
	bech32CharsetReverse [256]byte
)

func init() {
	for i := range bech32CharsetReverse {
		bech32CharsetReverse[i] = 255
	}
	for i := 0; i < len(bech32Charset); i++ {
		bech32CharsetReverse[bech32Charset[i]] = byte(i)
	}

	chk := uint32(1)
	for _, v := range bech32HRPExpand("npub") {
		chk = polymodStep(chk, v)
	}
	npubHRPPolymodState = chk
}

// polymodStep advances the bech32 polymod checksum by one 5-bit value.
func polymodStep(chk uint32, v byte) uint32 {
	gen := [5]uint32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	b := chk >> 25
	chk = (chk&0x1ffffff)<<5 ^ uint32(v)
	for i := 0; i < 5; i++ {
		if (b>>uint(i))&1 == 1 {
			chk ^= gen[i]
		}
	}
	return chk
}

// bech32SuffixToVals converts a bech32 suffix string to its 5-bit values.
// Call once before the hot loop to avoid per-attempt string scanning.
func bech32SuffixToVals(suffix string) ([]byte, error) {
	out := make([]byte, len(suffix))
	for i := range len(suffix) {
		v := bech32CharsetReverse[suffix[i]]
		if v == 255 {
			return nil, fmt.Errorf("invalid bech32 character %q", suffix[i])
		}
		out[i] = v
	}
	return out, nil
}

// npubSuffixMatch checks whether the npub encoding of xCoord (32-byte secp256k1
// x-coordinate) ends with the given target 5-bit values. Zero allocations: all
// intermediate state lives on the stack.
func npubSuffixMatch(xCoord []byte, target []byte) bool {
	// Convert 32 bytes → 52 5-bit values (padded), same as convertBits(xCoord, 8, 5, true).
	var data [52]byte
	acc := 0
	bits := 0
	pos := 0
	for _, byt := range xCoord {
		acc = (acc << 8) | int(byt)
		bits += 8
		for bits >= 5 {
			bits -= 5
			data[pos] = byte((acc >> bits) & 31)
			pos++
		}
	}
	if bits > 0 {
		data[pos] = byte((acc << (5 - bits)) & 31)
	}
	// pos == 52

	// Compute checksum starting from precomputed HRP state.
	chk := npubHRPPolymodState
	for i := 0; i < 52; i++ {
		chk = polymodStep(chk, data[i])
	}
	for i := 0; i < 6; i++ {
		chk = polymodStep(chk, 0)
	}
	chk ^= 1

	var checksum [6]byte
	for i := 0; i < 6; i++ {
		checksum[i] = byte((chk >> uint(5*(5-i))) & 31)
	}

	// Full encoded payload is 58 values: data[0..51] + checksum[0..5].
	// Compare the last len(target) values.
	n := len(target)
	for i := 0; i < n; i++ {
		p := 58 - n + i
		var val byte
		if p < 52 {
			val = data[p]
		} else {
			val = checksum[p-52]
		}
		if val != target[i] {
			return false
		}
	}
	return true
}

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
