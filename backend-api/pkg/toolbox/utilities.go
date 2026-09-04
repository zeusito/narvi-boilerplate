package toolbox

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/labstack/echo/v5"
)

func GetTraceId(ctx *echo.Context) string {
	return ctx.Response().Header().Get(echo.HeaderXRequestID)
}

// SanitizeEmail sanitizes an email address by converting it to lowercase and trimming whitespace
func SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// SecureRandomInt generates a random integer in the range [min, max]
func SecureRandomInt(min, max int) int {
	if min >= max {
		return min
	}

	rangeSize := max - min + 1
	n, err := rand.Int(rand.Reader, big.NewInt(int64(rangeSize)))
	if err != nil {
		return min // fallback
	}

	return int(n.Int64()) + min
}

// SecureRandomIntOfLen generates a random integer of the specified length
func SecureRandomOTP(length int) (string, error) {
	max := big.NewInt(1000000) // upper bound (exclusive): 0 - 999999
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// SecureRandomString generates a random string of the specified length using crypto/rand
func SecureRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[SecureRandomInt(0, len(charset)-1)]
	}
	return string(b)
}

// InflateCompressedData decompresses a zlib compressed string
func InflateCompressedData(data string) (string, error) {
	// Decode base64 string
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	// Create zlib reader
	reader, err := zlib.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return "", err
	}

	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

	// Read decompressed data
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(decompressed), nil
}

// DeflateData compresses a string using zlib
func DeflateData(data string) (string, error) {
	// Create zlib writer
	writer := bytes.NewBuffer(nil)
	zlibWriter := zlib.NewWriter(writer)

	// Write data to zlib writer
	_, err := zlibWriter.Write([]byte(data))
	if err != nil {
		return "", err
	}

	// Close zlib writer
	if err := zlibWriter.Close(); err != nil {
		return "", err
	}

	// Encode zlib compressed data to base64
	return base64.StdEncoding.EncodeToString(writer.Bytes()), nil
}
