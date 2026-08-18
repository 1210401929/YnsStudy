package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/config"
	"nys-go-api/internal/security"
)

const UserIDContextKey = "authenticated_user_id"

func Auth(cfg config.SecurityConfig, jwt *security.JWT) gin.HandlerFunc {
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

		token := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = strings.TrimSpace(token[7:])
		}
		if token == "" {
			unauthorized(c, "未携带Token，请先登录")
			return
		}
		userID, err := jwt.Verify(token)
		if err != nil {
			unauthorized(c, "Token无效或已过期")
			return
		}
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
