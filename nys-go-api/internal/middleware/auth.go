package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/config"
	"nys-go-api/internal/security"
)

const UserIDContextKey = "authenticated_user_id"

// Auth 校验 Token；sessions 用于识别已被作废（修改密码、封禁）的登录，传 nil 时不检查。
func Auth(cfg config.SecurityConfig, jwt *security.JWT, sessions LoginSessions) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if c.Request.Method == http.MethodOptions || path == "/health" {
			c.Next()
			return
		}
		if c.GetHeader(cfg.InnerCallHeader) == cfg.InnerCallSecret {
			c.Next()
			return
		}
		methodName := path[strings.LastIndex(path, "/")+1:]
		if strings.HasPrefix(methodName, "get") || matchesAny(path, cfg.Whitelist) {
			c.Next()
			return
		}

		token := bearerToken(c)
		if token == "" {
			unauthorized(c, "未携带Token，请先登录")
			return
		}
		claims, err := jwt.Parse(token)
		if err != nil || (sessions != nil && sessions.IsRevoked(c.Request.Context(), claims.Subject, claims.LoginAt)) {
			unauthorized(c, "Token无效或已过期")
			return
		}
		userID := claims.Subject
		c.Set(UserIDContextKey, userID)
		c.Request.Header.Set("X-User-Id", userID)
		c.Next()
	}
}

func matchesAny(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.HasSuffix(pattern, "/**") {
			if strings.HasPrefix(path, strings.TrimSuffix(pattern, "**")) {
				return true
			}
			continue
		}
		if path == pattern {
			return true
		}
	}
	return false
}

func unauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    http.StatusUnauthorized,
		"message": message,
		"data":    nil,
	})
}
