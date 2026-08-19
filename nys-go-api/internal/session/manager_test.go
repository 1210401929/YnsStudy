package session

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
	"nys-go-api/internal/model"
)

func TestSessionRejectsAbsoluteExpiration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := cache.NewMemoryStore()
	manager := NewManager(store, config.SecurityConfig{
		SessionCookieName:       "NYS_SESSION",
		SessionExpirationSecond: 7 * 24 * 60 * 60,
	})

	loginWriter := httptest.NewRecorder()
	loginContext, _ := gin.CreateTestContext(loginWriter)
	loginContext.Request = httptest.NewRequest("POST", "/login", nil)
	if err := manager.Create(loginContext, model.User{GUID: "user-guid", Code: "user-code"}); err != nil {
		t.Fatal(err)
	}
	cookies := loginWriter.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("登录后没有写入会话 Cookie")
	}
	cookie := cookies[0]

	expiredValue, _ := json.Marshal(record{
		User:      model.User{GUID: "user-guid", Code: "user-code"},
		ExpiresAt: time.Now().Add(-time.Minute).Unix(),
	})
	if err := store.Set(context.Background(), sessionKeyPrefix+cookie.Value, string(expiredValue), time.Hour); err != nil {
		t.Fatal(err)
	}

	checkWriter := httptest.NewRecorder()
	checkContext, _ := gin.CreateTestContext(checkWriter)
	checkContext.Request = httptest.NewRequest("GET", "/check", nil)
	checkContext.Request.AddCookie(cookie)
	if _, err := manager.Get(checkContext); err == nil {
		t.Fatal("超过绝对过期时间的会话仍然被接受")
	}
	if _, err := store.Get(context.Background(), sessionKeyPrefix+cookie.Value); err == nil {
		t.Fatal("过期会话没有从缓存删除")
	}
}

func TestSessionUpdateDoesNotExtendExpiration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := cache.NewMemoryStore()
	manager := NewManager(store, config.SecurityConfig{
		SessionCookieName:       "NYS_SESSION",
		SessionExpirationSecond: 7 * 24 * 60 * 60,
	})
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)
	c.Request = httptest.NewRequest("POST", "/login", nil)
	if err := manager.Create(c, model.User{GUID: "user-guid", Name: "旧名称"}); err != nil {
		t.Fatal(err)
	}
	cookie := writer.Result().Cookies()[0]
	before, err := manager.read(context.Background(), cookie.Value)
	if err != nil {
		t.Fatal(err)
	}

	updateWriter := httptest.NewRecorder()
	updateContext, _ := gin.CreateTestContext(updateWriter)
	updateContext.Request = httptest.NewRequest("POST", "/update", nil)
	updateContext.Request.AddCookie(cookie)
	if err := manager.Update(updateContext, model.User{GUID: "user-guid", Name: "新名称"}); err != nil {
		t.Fatal(err)
	}
	after, err := manager.read(context.Background(), cookie.Value)
	if err != nil {
		t.Fatal(err)
	}
	if after.ExpiresAt != before.ExpiresAt {
		t.Fatalf("更新资料延长了会话: before=%d after=%d", before.ExpiresAt, after.ExpiresAt)
	}
}
