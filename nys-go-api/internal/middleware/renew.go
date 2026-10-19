package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/security"
)

// RefreshTokenHeader 是续期后新 Token 所在的响应头，前端收到后替换本地保存的 Token。
const RefreshTokenHeader = "X-Refresh-Token"

// LoginSessions 是续期和作废检查需要的会话能力，由 session.Manager 实现。
type LoginSessions interface {
	Touch(c *gin.Context) bool
	IsRevoked(ctx context.Context, userGUID string, loginAt int64) bool
}

// RenewLogin 让活跃用户保持登录：Token 或会话的有效期用掉一半以上时自动续期，
// 续期后的期限不超过从登录算起的绝对上限。新 Token 放在 X-Refresh-Token 响应头中，响应体不变。
// 必须放在 Auth 之后，被 Auth 拒绝的请求不会续期。
func RenewLogin(jwt *security.JWT, sessions LoginSessions) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 退出登录的请求不续期，避免刚退出又拿到新 Token
		if c.Request.Method == http.MethodOptions || strings.HasSuffix(c.Request.URL.Path, "/logout") {
			c.Next()
			return
		}
		sessions.Touch(c)
		if token := bearerToken(c); token != "" {
			if claims, err := jwt.Parse(token); err == nil &&
				!sessions.IsRevoked(c.Request.Context(), claims.Subject, claims.LoginAt) &&
				jwt.ShouldRenew(claims, time.Now()) {
				if renewed, err := jwt.Issue(claims.Subject, time.Unix(claims.LoginAt, 0)); err == nil {
					c.Header(RefreshTokenHeader, renewed)
				}
			}
		}
		c.Next()
	}
}

func bearerToken(c *gin.Context) string {
	token := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	return token
}
