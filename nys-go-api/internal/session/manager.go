package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
	"nys-go-api/internal/model"
)

const (
	sessionKeyPrefix = "session:"
	// revokedKeyPrefix 记录某个用户“在此时间之前的登录全部作废”，用于修改密码、封禁后让所有设备退出。
	revokedKeyPrefix = "session:revoked:"
)

// Manager 管理 Cookie 会话。会话有两个期限：
//   - ttl（空闲有效期）：超过这么久没有访问就失效，有访问时由 Touch 续期；
//   - maxTTL（绝对上限）：从登录时刻算起，无论怎样续期都不会超过。
type Manager struct {
	store      cache.Store
	cookieName string
	ttl        time.Duration
	maxTTL     time.Duration
	secure     bool
}

// record 把过期时间、登录时间和用户一起保存。即使 Cookie 或 Redis TTL 配置异常，
// 服务端也不会接受超过 ExpiresAt 的会话。
type record struct {
	User      model.User `json:"user"`
	ExpiresAt int64      `json:"expiresAt"`
	LoginAt   int64      `json:"loginAt,omitempty"`
}

func NewManager(store cache.Store, cfg config.SecurityConfig) *Manager {
	return &Manager{
		store:      store,
		cookieName: cfg.SessionCookieName,
		ttl:        cfg.LoginIdleTTL(),
		maxTTL:     cfg.LoginMaxTTL(),
		secure:     cfg.SessionCookieSecure,
	}
}

func (m *Manager) Create(c *gin.Context, user model.User) error {
	id, err := randomID(32)
	if err != nil {
		return err
	}
	now := time.Now()
	return m.save(c, id, record{User: user, ExpiresAt: now.Add(m.ttl).Unix(), LoginAt: now.Unix()})
}

func (m *Manager) Get(c *gin.Context) (*model.User, error) {
	_, sessionRecord, err := m.current(c)
	if err != nil {
		return nil, err
	}
	return &sessionRecord.User, nil
}

// Update 只替换会话中的用户信息（例如修改资料后），不延长有效期。
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

// Touch 在有效期过半后把会话续到“现在 + 空闲有效期”，但不超过“登录时间 + 绝对上限”。
// 没有会话、会话已失效或尚未过半时什么也不做。返回是否续期。
func (m *Manager) Touch(c *gin.Context) bool {
	if m.maxTTL <= m.ttl {
		return false
	}
	id, sessionRecord, err := m.current(c)
	if err != nil || sessionRecord.LoginAt <= 0 {
		// 旧版本会话没有登录时间，无法计算上限，到期后重新登录即可
		return false
	}
	now := time.Now()
	expiresAt := time.Unix(sessionRecord.ExpiresAt, 0)
	limit := time.Unix(sessionRecord.LoginAt, 0).Add(m.maxTTL)
	if expiresAt.Sub(now) >= m.ttl/2 || !expiresAt.Before(limit) {
		return false
	}
	next := now.Add(m.ttl)
	if next.After(limit) {
		next = limit
	}
	sessionRecord.ExpiresAt = next.Unix()
	return m.save(c, id, sessionRecord) == nil
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

// RevokeUser 让该用户此刻之前的所有登录（各设备的会话和 Token）失效，用于修改密码、封禁。
func (m *Manager) RevokeUser(ctx context.Context, userGUID string) error {
	if userGUID == "" {
		return nil
	}
	return m.store.Set(ctx, revokedKeyPrefix+userGUID, strconv.FormatInt(time.Now().Unix(), 10), m.maxTTL)
}

// IsRevoked 判断某次登录（loginAt 为登录时间）是否已被 RevokeUser 作废。
func (m *Manager) IsRevoked(ctx context.Context, userGUID string, loginAt int64) bool {
	if userGUID == "" {
		return false
	}
	value, err := m.store.Get(ctx, revokedKeyPrefix+userGUID)
	if err != nil {
		return false
	}
	revokedAt, err := strconv.ParseInt(value, 10, 64)
	return err == nil && loginAt < revokedAt
}

// current 读取当前请求的有效会话；会话过期或已被作废时清除 Cookie。
func (m *Manager) current(c *gin.Context) (string, record, error) {
	cookie, err := c.Request.Cookie(m.cookieName)
	if err != nil {
		return "", record{}, cache.ErrNotFound
	}
	sessionRecord, err := m.read(c.Request.Context(), cookie.Value)
	if err != nil {
		m.expireCookie(c)
		return "", record{}, err
	}
	if sessionRecord.ExpiresAt <= time.Now().Unix() || m.IsRevoked(c.Request.Context(), sessionRecord.User.GUID, sessionRecord.LoginAt) {
		_ = m.store.Delete(c.Request.Context(), sessionKeyPrefix+cookie.Value)
		m.expireCookie(c)
		return "", record{}, cache.ErrNotFound
	}
	return cookie.Value, sessionRecord, nil
}

// save 写入会话并下发与过期时间一致的 Cookie。
func (m *Manager) save(c *gin.Context, id string, sessionRecord record) error {
	value, err := json.Marshal(sessionRecord)
	if err != nil {
		return fmt.Errorf("序列化会话: %w", err)
	}
	expiresAt := time.Unix(sessionRecord.ExpiresAt, 0)
	remaining := time.Until(expiresAt)
	if err := m.store.Set(c.Request.Context(), sessionKeyPrefix+id, string(value), remaining); err != nil {
		return fmt.Errorf("保存会话: %w", err)
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     m.cookieName,
		Value:    id,
		Path:     "/",
		MaxAge:   int(remaining.Seconds()),
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	})
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
