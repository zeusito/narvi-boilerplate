package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type HmacSHA256 struct {
	secret []byte
}

func (s *HmacSHA256) Hash(data string) (string, error) {
	if len(data) == 0 {
		return "", errors.New("empty data")
	}

	h := hmac.New(sha256.New, s.secret)

	h.Write([]byte(data))
	hash := h.Sum(nil)

	return hex.EncodeToString(hash), nil
}

func (s *HmacSHA256) Verify(data, hashedData string) bool {
	h := hmac.New(sha256.New, s.secret)

	h.Write([]byte(data))
	computedHash := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(computedHash), []byte(hashedData))
}
