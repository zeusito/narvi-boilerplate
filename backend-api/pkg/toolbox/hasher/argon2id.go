package hasher

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Argon2IdHasher struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

// Hash generates a secure Argon2id hash of the given password.
func (h *Argon2IdHasher) Hash(data string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("password must be at least 8 characters")
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	hash := argon2.IDKey([]byte(data), salt, h.time, h.memory, h.threads, h.keyLen)

	// Base64 encode the salt and the hash
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return the encoded hash in the form of "$argon2id$v=19$m=65536,t=1,p=4$c29tZXNhbHQ$aGFzaA"
	encodedHash := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", h.memory, h.time, h.threads, b64Salt, b64Hash)

	return encodedHash, nil
}

// Verify verifies if the given data matches the stored hash.
func (h *Argon2IdHasher) Verify(data, hashedData string) bool {
	var (
		version int
		memory  uint32
		time    uint32
		threads uint8
	)

	vals := strings.Split(hashedData, "$")
	if len(vals) != 6 {
		return false
	}

	// Check if the hash is argon2id
	if vals[1] != "argon2id" {
		return false
	}

	// Check if the version is 19
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil || version != 19 {
		return false
	}

	// Parse memory, time, and threads
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false
	}

	// Decode the salt
	salt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return false
	}

	// Decode the hash
	hash, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return false
	}

	//nolint:gosec
	calculatedHash := argon2.IDKey([]byte(data), salt, time, memory, threads, uint32(len(hash)))

	if len(hash) != len(calculatedHash) {
		return false
	}

	for i := 0; i < len(hash); i++ {
		if hash[i] != calculatedHash[i] {
			return false
		}
	}

	return true
}
