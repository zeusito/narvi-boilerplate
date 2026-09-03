package hasher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:gosec
func TestNewHmacSHA256(t *testing.T) {
	// Valid base64 secret
	validSecret := "dGVzdC1zZWNyZXQ=" // base64 for "test-secret"

	theHasher, err := NewHmacSHA256(validSecret)
	assert.NoError(t, err, "NewHmacSHA256 should not return an error for valid secret")
	assert.NotNil(t, theHasher, "NewHmacSHA256 should return a non-nil hasher instance")

	invalidSecret := "invalid-base64!"
	_, err = NewHmacSHA256(invalidSecret)
	assert.Error(t, err, "NewHmacSHA256 should return an error for invalid secret")
}

//nolint:gosec
func TestHmacSHA256_Hash(t *testing.T) {
	validSecret := "dGVzdC1zZWNyZXQ=" // base64 for "test-secret"
	theHasher, err := NewHmacSHA256(validSecret)

	assert.NoError(t, err, "NewHmacSHA256 should not return an error for valid secret")
	assert.NotNil(t, theHasher, "NewHmacSHA256 should return a non-nil hasher instance")

	// Valid data
	hash, err := theHasher.Hash("test-data")
	assert.NoError(t, err, "Hash should not return an error for valid data")
	assert.NotEmpty(t, hash, "Hash should return a non-empty hash for valid data")

	// Empty data
	_, err = theHasher.Hash("")
	assert.Error(t, err, "Hash should return an error for empty data")
}

//nolint:gosec
func TestHmacSHA256_Verify(t *testing.T) {
	validSecret := "dGVzdC1zZWNyZXQ=" // base64 for "test-secret"
	theHasher, err := NewHmacSHA256(validSecret)

	assert.NoError(t, err, "NewHmacSHA256 should not return an error for valid secret")
	assert.NotNil(t, theHasher, "NewHmacSHA256 should return a non-nil hasher instance")

	// Compute a hash for verification
	hash, _ := theHasher.Hash("test-data")

	// Matching hash
	assert.True(t, theHasher.Verify("test-data", hash),
		"Verify should return true for matching hash")

	// Non-matching hash
	assert.False(t, theHasher.Verify("test-data", "invalid-hash"),
		"Verify should return false for non-matching hash")

	// Different data
	assert.False(t, theHasher.Verify("different-data", hash),
		"Verify should return false for different data")
}

//nolint:gosec
func TestArgon2IdHash(t *testing.T) {
	hasher := NewArgon2IdHasherWithSaneDefaults()
	password := "password"
	hash, err := hasher.Hash(password)

	assert.NoError(t, err)
	assert.True(t, len(hash) > 0)
	assert.Containsf(t, hash, "$argon2id$v=19$m=65536,t=1,p=4$", "hash should contain $argon2id$")
}

//nolint:gosec
func TestVerifyArgon2IdHash(t *testing.T) {
	hasher := NewArgon2IdHasherWithSaneDefaults()
	hashedPwd := "$argon2id$v=19$m=16,t=2,p=1$N1FkeWl6S0RaTzZPVkpTdQ$NZwZ/YEeSD+oKh3TOMqDDg"
	result := hasher.Verify("1234567890", hashedPwd)

	assert.True(t, result)
}
