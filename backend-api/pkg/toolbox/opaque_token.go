package toolbox

import (
	"backend-api/pkg/toolbox/hasher"
	"crypto/rand"
)

// GenerateOpaqueToken generates a new opaque token following the spec:
// - Generate 256 bits of random data
// - Apply Crockford Base32 encoding
// - Add a descriptive prefix
// - Hash the full prefixed string for storage
func GenerateOpaqueToken(th hasher.Hasher, prefix string) (token string, hashedToken string, err error) {
	// 1. Generate 32 bytes of raw entropy (256 bits)
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", "", err
	}

	// 2. Apply Crockford Base32 encoding
	encoded := encodeCrockfordBase32(rawBytes)

	// 3. Add a descriptive prefix (if any)
	token = encoded
	if prefix != "" {
		token = prefix + "_" + encoded
	}

	// 4. Hash the full prefixed string for storage
	hashedToken, err = th.Hash(token)

	return token, hashedToken, err
}

func encodeCrockfordBase32(src []byte) string {
	if len(src) == 0 {
		return ""
	}

	// Crockford Base32 alphabet (excludes I, L, O, U)
	const crockfordAlphabet = "0123456789abcdefghjkmnpqrstvwxyz"

	// Calculate output length: ceil(len(src) * 8 / 5)
	outLen := (len(src)*8 + 4) / 5
	out := make([]byte, outLen)

	var buffer uint64
	var bitsLeft uint
	outIdx := 0

	for _, b := range src {
		buffer = (buffer << 8) | uint64(b)
		bitsLeft += 8

		for bitsLeft >= 5 {
			bitsLeft -= 5
			val := (buffer >> bitsLeft) & 0x1F
			out[outIdx] = crockfordAlphabet[val]
			outIdx++
		}
	}

	// Handle leftover bits (if any)
	if bitsLeft > 0 {
		val := (buffer << (5 - bitsLeft)) & 0x1F
		out[outIdx] = crockfordAlphabet[val]
		outIdx++
	}

	return string(out[:outIdx])
}
