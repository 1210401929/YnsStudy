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

// record 把绝对过期时间和用户一起保存。即使 Cookie 或 Redis TTL 配置异常，
// 服务端也不会接受超过 expiresAt 的会话。
type record struct {
	User      model.User `json:"user"`
	ExpiresAt int64      `json:"expiresAt"`
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
	expiresAt := time.Now().Add(m.ttl)
	value, err := json.Marshal(record{User: user, ExpiresAt: expiresAt.Unix()})
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
		Expires:  expiresAt,
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
	sessionRecord, err := m.read(c.Request.Context(), cookie.Value)
	if err != nil {
		m.expireCookie(c)
		return nil, err
	}
	if sessionRecord.ExpiresAt <= time.Now().Unix() {
		_ = m.store.Delete(c.Request.Context(), sessionKeyPrefix+cookie.Value)
		m.expireCookie(c)
		return nil, cache.ErrNotFound
	}
	return &sessionRecord.User, nil
}

func (m *Manager) Update(c *gin.Context, user model.User) error {
	cookie, err := c.Request.Cookie(m.cookieName)
	if err != nil {
		return m.Create(c, user)
	}
	sessionRecord, err := m.read(c.Request.Context(), cookie.Value)
	if err != nil || sessionRecord.ExpiresAt <= time.Now().Unix() {
		m.expireCookie(c)
		return cache.ErrNotFound
	}
	sessionRecord.User = user
	value, err := json.Marshal(sessionRecord)
	if err != nil {
		return err
	}
	remaining := time.Until(time.Unix(sessionRecord.ExpiresAt, 0))
	return m.store.Set(c.Request.Context(), sessionKeyPrefix+cookie.Value, string(value), remaining)
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
	m.expireCookie(c)
	return nil
}

func (m *Manager) read(ctx context.Context, id string) (record, error) {
	value, err := m.store.Get(ctx, sessionKeyPrefix+id)
	if err != nil {
		return record{}, err
	}
	var sessionRecord record
	if err := json.Unmarshal([]byte(value), &sessionRecord); err != nil {
		return record{}, fmt.Errorf("解析会话: %w", err)
	}
	// 旧版本只保存 User，没有绝对过期时间。升级后将它视为失效，避免旧会话无限存活。
	if sessionRecord.ExpiresAt <= 0 || sessionRecord.User.GUID == "" {
		_ = m.store.Delete(ctx, sessionKeyPrefix+id)
		return record{}, cache.ErrNotFound
	}
	return sessionRecord, nil
}

func (m *Manager) expireCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: m.cookieName, Value: "", Path: "/", MaxAge: -1,
		Expires: time.Unix(1, 0), HttpOnly: true, Secure: m.secure, SameSite: http.SameSiteLaxMode,
	})
}

func randomID(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
