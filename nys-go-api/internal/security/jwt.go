package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type JWT struct {
	secret []byte
	ttl    time.Duration
}

type claims struct {
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func NewJWT(secret string, ttl time.Duration) *JWT {
	return &JWT{secret: []byte(secret), ttl: ttl}
}

func (j *JWT) Generate(subject string) (string, error) {
	now := time.Now()
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(claims{Subject: subject, IssuedAt: now.Unix(), ExpiresAt: now.Add(j.ttl).Unix()})
	if err != nil {
		return "", err
	}
	unsigned := encodeSegment(header) + "." + encodeSegment(payload)
	return unsigned + "." + encodeSegment(j.sign(unsigned)), nil
}

func (j *JWT) Verify(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("Token 格式错误")
	}
	unsigned := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(signature, j.sign(unsigned)) {
		return "", fmt.Errorf("Token 签名无效")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("Token 内容无效")
	}
	var value claims
	if err := json.Unmarshal(payload, &value); err != nil {
		return "", fmt.Errorf("Token 内容无效")
	}
	if value.Subject == "" || time.Now().Unix() >= value.ExpiresAt {
		return "", fmt.Errorf("Token 无效或已过期")
	}
	return value.Subject, nil
}

func (j *JWT) sign(value string) []byte {
	hasher := hmac.New(sha256.New, j.secret)
	_, _ = hasher.Write([]byte(value))
	return hasher.Sum(nil)
}

func encodeSegment(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}
