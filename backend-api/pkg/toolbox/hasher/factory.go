package hasher

import "encoding/base64"

type Hasher interface {
	Hash(data string) (string, error)
	Verify(data, hashedData string) bool
}

// NewHmacSHA256 Creates a new HMAC-SHA256 hasher instance based on the given base64 encoded secret
func NewHmacSHA256(encodedSecret string) (Hasher, error) {
	decoded, err := base64.StdEncoding.DecodeString(encodedSecret)
	if err != nil {
		return nil, err
	}

	return &HmacSHA256{secret: decoded}, nil
}

func NewArgon2IdHasherWithSaneDefaults() *Argon2IdHasher {
	return &Argon2IdHasher{
		time:    1,
		memory:  64 * 1024,
		threads: 4,
		keyLen:  32,
	}
}
