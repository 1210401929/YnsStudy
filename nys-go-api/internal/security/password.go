package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

func NewSalt() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("生成密码盐: %w", err)
	}
	return base64.StdEncoding.EncodeToString(value), nil
}

func HashPassword(password, salt string) (string, error) {
	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return "", fmt.Errorf("密码盐无效: %w", err)
	}
	hasher := sha256.New()
	_, _ = hasher.Write(saltBytes)
	_, _ = hasher.Write([]byte(password))
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil)), nil
}

func VerifyPassword(password, expected, salt string) bool {
	actual, err := HashPassword(password, salt)
	return err == nil && hmacEqual(actual, expected)
}

func hmacEqual(left, right string) bool {
	if len(left) != len(right) {
		return false
	}
	var different byte
	for i := range left {
		different |= left[i] ^ right[i]
	}
	return different == 0
}
