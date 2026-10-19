package security

import (
	"encoding/json"
	"testing"
	"time"
)

func TestIssueCapsExpirationAtLoginMax(t *testing.T) {
	jwt := NewJWT("test-secret", 7*24*time.Hour, 30*24*time.Hour)

	// 29 天前登录：续期后最多只能到登录后第 30 天，即还剩约 1 天
	loginAt := time.Now().Add(-29 * 24 * time.Hour)
	token, err := jwt.Issue("user-guid", loginAt)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := jwt.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.LoginAt != loginAt.Unix() {
		t.Fatalf("续期不应改变登录时间: got %d want %d", claims.LoginAt, loginAt.Unix())
	}
	if limit := loginAt.Add(30 * 24 * time.Hour).Unix(); claims.ExpiresAt != limit {
		t.Fatalf("过期时间应被限制在登录后 30 天: got %d want %d", claims.ExpiresAt, limit)
	}
}

func TestShouldRenewOnlyAfterHalfAndBeforeLimit(t *testing.T) {
	jwt := NewJWT("test-secret", 7*24*time.Hour, 30*24*time.Hour)
	now := time.Now()
	day := int64(24 * 60 * 60)

	fresh := Claims{Subject: "u", LoginAt: now.Unix(), ExpiresAt: now.Unix() + 7*day}
	if jwt.ShouldRenew(fresh, now) {
		t.Fatal("刚签发的 Token 不应续期")
	}
	pastHalf := Claims{Subject: "u", LoginAt: now.Unix() - 5*day, ExpiresAt: now.Unix() + 2*day}
	if !jwt.ShouldRenew(pastHalf, now) {
		t.Fatal("有效期过半的 Token 应该续期")
	}
	atLimit := Claims{Subject: "u", LoginAt: now.Unix() - 29*day, ExpiresAt: now.Unix() - 29*day + 30*day}
	if jwt.ShouldRenew(atLimit, now) {
		t.Fatal("已经到达绝对上限的 Token 不应再续期")
	}

	noRenew := NewJWT("test-secret", time.Hour, time.Hour)
	if noRenew.ShouldRenew(Claims{LoginAt: now.Unix(), ExpiresAt: now.Unix() + 60}, now) {
		t.Fatal("绝对上限等于空闲有效期时不应续期")
	}
}

func TestParseTreatsLegacyTokenIssuedAtAsLoginAt(t *testing.T) {
	jwt := NewJWT("test-secret", time.Hour, 30*24*time.Hour)
	now := time.Now()
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, _ := json.Marshal(map[string]any{"sub": "user-guid", "iat": now.Unix(), "exp": now.Add(time.Hour).Unix()})
	unsigned := encodeSegment(header) + "." + encodeSegment(payload)
	legacy := unsigned + "." + encodeSegment(jwt.sign(unsigned))

	claims, err := jwt.Parse(legacy)
	if err != nil {
		t.Fatalf("旧版本 Token 应仍然有效: %v", err)
	}
	if claims.LoginAt != now.Unix() {
		t.Fatalf("旧版本 Token 应以签发时间作为登录时间: got %d", claims.LoginAt)
	}
}
