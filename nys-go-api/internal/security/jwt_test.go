package security

import (
	"strings"
	"testing"
	"time"
)

func TestJWTRoundTrip(t *testing.T) {
	jwt := NewJWT("test-secret", time.Hour)
	token, err := jwt.Generate("user-guid")
	if err != nil {
		t.Fatal(err)
	}
	subject, err := jwt.Verify(token)
	if err != nil {
		t.Fatal(err)
	}
	if subject != "user-guid" {
		t.Fatalf("subject = %q", subject)
	}

	parts := strings.Split(token, ".")
	parts[1] += "x"
	if _, err := jwt.Verify(strings.Join(parts, ".")); err == nil {
		t.Fatal("篡改后的 Token 应校验失败")
	}
}

func TestJWTExpires(t *testing.T) {
	jwt := NewJWT("test-secret", -time.Second)
	token, err := jwt.Generate("user-guid")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := jwt.Verify(token); err == nil {
		t.Fatal("过期 Token 应校验失败")
	}
}
