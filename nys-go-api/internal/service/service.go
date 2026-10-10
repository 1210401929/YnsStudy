package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
	"nys-go-api/internal/middleware"
	"nys-go-api/internal/model"
	"nys-go-api/internal/repository"
	"nys-go-api/internal/security"
	"nys-go-api/internal/session"
)

type Service struct {
	Config   *config.Config
	Repo     *repository.SQLRepository
	Cache    cache.Store
	Sessions *session.Manager
	JWT      *security.JWT
	HTTP     *http.Client
}

func New(cfg *config.Config, repo *repository.SQLRepository, store cache.Store, sessions *session.Manager, jwt *security.JWT) *Service {
	return &Service{
		Config:   cfg,
		Repo:     repo,
		Cache:    store,
		Sessions: sessions,
		JWT:      jwt,
		HTTP:     &http.Client{Timeout: time.Duration(cfg.External.HTTPTimeoutSeconds) * time.Second},
	}
}

func (s *Service) CurrentUser(c *gin.Context) (*model.User, error) {
	userID, ok := c.Get(middleware.UserIDContextKey)
	if !ok || fmt.Sprint(userID) == "" {
		// 白名单和 get 开头的接口不经过 JWT 中间件校验，这里补充解析，保证携带 Token 时也能识别当前用户
		userID = s.userIDFromToken(c)
	}
	if fmt.Sprint(userID) != "" && userID != nil {
		rows, err := s.Repo.Query(c.Request.Context(), "SELECT * FROM userInfo WHERE GUID = ? LIMIT 1", fmt.Sprint(userID))
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return nil, fmt.Errorf("用户不存在")
		}
		user := model.UserFromRow(rows[0])
		return &user, nil
	}
	if user, err := s.Sessions.Get(c); err == nil {
		return user, nil
	}
	return nil, fmt.Errorf("用户未登录!")
}

func (s *Service) userIDFromToken(c *gin.Context) any {
	if s.JWT == nil {
		return nil
	}
	token := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	if token == "" {
		return nil
	}
	userID, err := s.JWT.Verify(token)
	if err != nil {
		return nil
	}
	return userID
}

// ClientIP 优先使用 nginx 设置的 X-Real-IP（$remote_addr，客户端无法伪造）。
// X-Forwarded-For 经 $proxy_add_x_forwarded_for 转发时会保留客户端自带的值，第一个地址可以伪造，
// 只在没有 X-Real-IP 时使用。
func ClientIP(c *gin.Context) string {
	for _, name := range []string{"X-Real-IP", "X-Forwarded-For", "Proxy-Client-IP", "WL-Proxy-Client-IP"} {
		value := strings.TrimSpace(c.GetHeader(name))
		if value != "" && !strings.EqualFold(value, "unknown") {
			return strings.TrimSpace(strings.Split(value, ",")[0])
		}
	}
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err == nil {
		return host
	}
	return c.Request.RemoteAddr
}

func dbFailure(action string, err error) model.Result {
	return model.Failure(fmt.Sprintf("%s失败: %v", action, err))
}

func contextOf(c *gin.Context) context.Context {
	return c.Request.Context()
}
