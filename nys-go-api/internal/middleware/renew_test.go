package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/config"
	"nys-go-api/internal/security"
)

type fakeSessions struct {
	revoked bool
	touched int
}

func (f *fakeSessions) Touch(*gin.Context) bool { f.touched++; return false }
func (f *fakeSessions) IsRevoked(context.Context, string, int64) bool {
	return f.revoked
}

func newRenewRouter(jwt *security.JWT, sessions *fakeSessions) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Auth(config.SecurityConfig{InnerCallHeader: "X-Inner-Call", InnerCallSecret: "secret"}, jwt, sessions), RenewLogin(jwt, sessions))
	router.POST("/blog-api/blog/addBlog", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return router
}

func call(router *gin.Engine, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest("POST", "/blog-api/blog/addBlog", nil)
	request.Header.Set("Authorization", token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func loginAtOf(t *testing.T, token string) int64 {
	t.Helper()
	payload, err := base64.RawURLEncoding.DecodeString(strings.Split(token, ".")[1])
	if err != nil {
		t.Fatal(err)
	}
	var claims security.Claims
	_ = json.Unmarshal(payload, &claims)
	return claims.LoginAt
}

func TestRenewLoginIssuesNewTokenAfterHalf(t *testing.T) {
	// 空闲有效期 4 秒、上限 1 小时：签发 2.5 秒后已用掉一半以上（Token 时间精确到秒，留出余量）
	jwt := security.NewJWT("test-secret", 4*time.Second, time.Hour)
	sessions := &fakeSessions{}
	router := newRenewRouter(jwt, sessions)

	loginAt := time.Now()
	token, err := jwt.Issue("user-guid", loginAt)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2500 * time.Millisecond)
	response := call(router, token)
	if response.Code != http.StatusOK {
		t.Fatalf("有效 Token 被拒绝: %d", response.Code)
	}
	renewed := response.Header().Get(RefreshTokenHeader)
	if renewed == "" || renewed == token {
		t.Fatal("有效期过半时应下发新 Token")
	}
	if loginAtOf(t, renewed) != loginAt.Unix() {
		t.Fatal("新 Token 应保留原来的登录时间，绝对上限才能生效")
	}
	if sessions.touched != 1 {
		t.Fatal("请求时应尝试续期会话")
	}
}

func TestRenewLoginSkipsFreshAndRevokedTokens(t *testing.T) {
	jwt := security.NewJWT("test-secret", time.Hour, 30*time.Hour)
	sessions := &fakeSessions{}
	router := newRenewRouter(jwt, sessions)

	token, _ := jwt.Generate("user-guid")
	if response := call(router, token); response.Header().Get(RefreshTokenHeader) != "" {
		t.Fatal("刚签发的 Token 不应续期")
	}

	sessions.revoked = true
	if response := call(router, token); response.Code != http.StatusUnauthorized {
		t.Fatalf("已作废的登录应被拒绝，实际 %d", response.Code)
	}
}
