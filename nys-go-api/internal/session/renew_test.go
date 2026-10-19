package session

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
	"nys-go-api/internal/model"
)

const day = 24 * time.Hour

func newTestManager(store cache.Store) *Manager {
	return NewManager(store, config.SecurityConfig{
		SessionCookieName: "NYS_SESSION",
		LoginIdleSeconds:  int64((7 * day).Seconds()),
		LoginMaxSeconds:   int64((30 * day).Seconds()),
	})
}

func createSession(t *testing.T, manager *Manager) *http.Cookie {
	t.Helper()
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)
	c.Request = httptest.NewRequest("POST", "/login", nil)
	if err := manager.Create(c, model.User{GUID: "user-guid", Code: "user-code"}); err != nil {
		t.Fatal(err)
	}
	return writer.Result().Cookies()[0]
}

func requestWith(cookie *http.Cookie) (*gin.Context, *httptest.ResponseRecorder) {
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)
	c.Request = httptest.NewRequest("GET", "/any", nil)
	c.Request.AddCookie(cookie)
	return c, writer
}

// overwrite 直接改写缓存中的会话，模拟“若干天前登录、即将过期”的状态
func overwrite(t *testing.T, store cache.Store, cookie *http.Cookie, loginAt, expiresAt time.Time) {
	t.Helper()
	value, _ := json.Marshal(record{User: model.User{GUID: "user-guid", Code: "user-code"}, LoginAt: loginAt.Unix(), ExpiresAt: expiresAt.Unix()})
	if err := store.Set(context.Background(), sessionKeyPrefix+cookie.Value, string(value), time.Until(expiresAt)); err != nil {
		t.Fatal(err)
	}
}

func TestTouchExtendsOnlyAfterHalf(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := cache.NewMemoryStore()
	manager := newTestManager(store)
	cookie := createSession(t, manager)

	c, _ := requestWith(cookie)
	if manager.Touch(c) {
		t.Fatal("刚登录的会话不应续期")
	}

	now := time.Now()
	overwrite(t, store, cookie, now.Add(-6*day), now.Add(day))
	c, writer := requestWith(cookie)
	if !manager.Touch(c) {
		t.Fatal("有效期过半的会话应该续期")
	}
	after, err := manager.read(context.Background(), cookie.Value)
	if err != nil {
		t.Fatal(err)
	}
	if got := time.Until(time.Unix(after.ExpiresAt, 0)); got < 7*day-time.Minute {
		t.Fatalf("续期后应再有约 7 天有效期，实际剩余 %s", got)
	}
	if len(writer.Result().Cookies()) == 0 {
		t.Fatal("续期后应重新下发 Cookie")
	}
}

func TestTouchNeverExceedsLoginMax(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := cache.NewMemoryStore()
	manager := newTestManager(store)
	cookie := createSession(t, manager)

	now := time.Now()
	loginAt := now.Add(-29 * day)
	overwrite(t, store, cookie, loginAt, now.Add(time.Hour))
	c, _ := requestWith(cookie)
	if !manager.Touch(c) {
		t.Fatal("未到上限的会话应该续期")
	}
	after, _ := manager.read(context.Background(), cookie.Value)
	if limit := loginAt.Add(30 * day).Unix(); after.ExpiresAt != limit {
		t.Fatalf("续期不能超过登录后 30 天: got %d want %d", after.ExpiresAt, limit)
	}
	c, _ = requestWith(cookie)
	if manager.Touch(c) {
		t.Fatal("到达上限后不应再续期")
	}
}

func TestRevokeUserInvalidatesExistingSessions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := cache.NewMemoryStore()
	manager := newTestManager(store)
	cookie := createSession(t, manager)

	// 模拟一个早于作废时间的登录
	now := time.Now()
	overwrite(t, store, cookie, now.Add(-time.Hour), now.Add(6*day))
	if err := manager.RevokeUser(context.Background(), "user-guid"); err != nil {
		t.Fatal(err)
	}
	c, _ := requestWith(cookie)
	if _, err := manager.Get(c); err == nil {
		t.Fatal("修改密码前的会话应该失效")
	}
	if !manager.IsRevoked(context.Background(), "user-guid", now.Add(-time.Hour).Unix()) {
		t.Fatal("作废时间之前签发的 Token 应视为失效")
	}

	// 作废之后重新登录不受影响
	fresh := createSession(t, manager)
	c, _ = requestWith(fresh)
	if _, err := manager.Get(c); err != nil {
		t.Fatalf("重新登录后的会话应该有效: %v", err)
	}
}
