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

// JWT 签发和校验登录 Token。
// ttl 是空闲有效期：Token 只在这段时间内有效，期间有访问会由续期中间件换发新 Token；
// maxTTL 是从登录时刻算起的绝对上限，无论怎样续期都不会超过它。
type JWT struct {
	secret []byte
	ttl    time.Duration
	maxTTL time.Duration
}

// Claims 是 Token 中保存的信息。LoginAt 是最初登录的时间，续期时保持不变。
type Claims struct {
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	LoginAt   int64  `json:"lat,omitempty"`
}

// NewJWT 创建签发器。maxTTL 小于 ttl 时按 ttl 处理（即不续期）。
func NewJWT(secret string, ttl, maxTTL time.Duration) *JWT {
	if maxTTL < ttl {
		maxTTL = ttl
	}
	return &JWT{secret: []byte(secret), ttl: ttl, maxTTL: maxTTL}
}

// TTL 返回空闲有效期。
func (j *JWT) TTL() time.Duration { return j.ttl }

// Generate 为一次新的登录签发 Token。
func (j *JWT) Generate(subject string) (string, error) {
	return j.Issue(subject, time.Now())
}

// Issue 签发 Token：过期时间为“现在 + 空闲有效期”，但不超过“登录时间 + 绝对上限”。
func (j *JWT) Issue(subject string, loginAt time.Time) (string, error) {
	now := time.Now()
	expiresAt := now.Add(j.ttl)
	if limit := loginAt.Add(j.maxTTL); expiresAt.After(limit) {
		expiresAt = limit
	}
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(Claims{Subject: subject, IssuedAt: now.Unix(), ExpiresAt: expiresAt.Unix(), LoginAt: loginAt.Unix()})
	if err != nil {
		return "", err
	}
	unsigned := encodeSegment(header) + "." + encodeSegment(payload)
	return unsigned + "." + encodeSegment(j.sign(unsigned)), nil
}

// Verify 校验 Token 并返回用户 ID。
func (j *JWT) Verify(token string) (string, error) {
	claims, err := j.Parse(token)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

// Parse 校验签名和有效期并返回完整信息。旧版本签发的 Token 没有 lat，以签发时间作为登录时间。
func (j *JWT) Parse(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, fmt.Errorf("Token 格式错误")
	}
	unsigned := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(signature, j.sign(unsigned)) {
		return Claims{}, fmt.Errorf("Token 签名无效")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, fmt.Errorf("Token 内容无效")
	}
	var value Claims
	if err := json.Unmarshal(payload, &value); err != nil {
		return Claims{}, fmt.Errorf("Token 内容无效")
	}
	if value.Subject == "" || time.Now().Unix() >= value.ExpiresAt {
		return Claims{}, fmt.Errorf("Token 无效或已过期")
	}
	if value.LoginAt == 0 {
		value.LoginAt = value.IssuedAt
	}
	return value, nil
}

// ShouldRenew 判断 Token 是否已用掉一半以上的有效期，并且还没到绝对上限。
func (j *JWT) ShouldRenew(claims Claims, now time.Time) bool {
	if j.maxTTL <= j.ttl {
		return false
	}
	remaining := time.Unix(claims.ExpiresAt, 0).Sub(now)
	limit := time.Unix(claims.LoginAt, 0).Add(j.maxTTL)
	return remaining < j.ttl/2 && time.Unix(claims.ExpiresAt, 0).Before(limit)
}

func (j *JWT) sign(value string) []byte {
	hasher := hmac.New(sha256.New, j.secret)
	_, _ = hasher.Write([]byte(value))
	return hasher.Sum(nil)
}

func encodeSegment(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}
