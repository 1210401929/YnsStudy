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
	if userID, ok := c.Get(middleware.UserIDContextKey); ok && fmt.Sprint(userID) != "" {
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

func ClientIP(c *gin.Context) string {
	for _, name := range []string{"X-Forwarded-For", "X-Real-IP", "Proxy-Client-IP", "WL-Proxy-Client-IP"} {
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
