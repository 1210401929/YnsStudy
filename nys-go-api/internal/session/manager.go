package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
	"nys-go-api/internal/model"
)

const sessionKeyPrefix = "session:"

type Manager struct {
	store      cache.Store
	cookieName string
	ttl        time.Duration
	secure     bool
}

func NewManager(store cache.Store, cfg config.SecurityConfig) *Manager {
	return &Manager{
		store:      store,
		cookieName: cfg.SessionCookieName,
		ttl:        time.Duration(cfg.SessionExpirationSecond) * time.Second,
		secure:     cfg.SessionCookieSecure,
	}
}

func (m *Manager) Create(c *gin.Context, user model.User) error {
	id, err := randomID(32)
	if err != nil {
		return err
	}
	value, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("序列化会话: %w", err)
	}
	if err := m.store.Set(c.Request.Context(), sessionKeyPrefix+id, string(value), m.ttl); err != nil {
		return fmt.Errorf("保存会话: %w", err)
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     m.cookieName,
		Value:    id,
		Path:     "/",
		MaxAge:   int(m.ttl.Seconds()),
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (m *Manager) Get(c *gin.Context) (*model.User, error) {
	cookie, err := c.Request.Cookie(m.cookieName)
	if err != nil {
		return nil, cache.ErrNotFound
	}
	value, err := m.store.Get(c.Request.Context(), sessionKeyPrefix+cookie.Value)
	if err != nil {
		return nil, err
	}
	var user model.User
	if err := json.Unmarshal([]byte(value), &user); err != nil {
		return nil, fmt.Errorf("解析会话: %w", err)
	}
	return &user, nil
}

func (m *Manager) Update(c *gin.Context, user model.User) error {
	cookie, err := c.Request.Cookie(m.cookieName)
	if err != nil {
		return m.Create(c, user)
	}
	value, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return m.store.Set(c.Request.Context(), sessionKeyPrefix+cookie.Value, string(value), m.ttl)
}

func (m *Manager) Destroy(c *gin.Context) error {
	cookie, err := c.Request.Cookie(m.cookieName)
	if err == nil {
		if deleteErr := m.store.Delete(context.Background(), sessionKeyPrefix+cookie.Value); deleteErr != nil {
			return deleteErr
		}
	} else if !errors.Is(err, http.ErrNoCookie) {
		return err
	}
	http.SetCookie(c.Writer, &http.Cookie{Name: m.cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: m.secure})
	return nil
}

func randomID(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
