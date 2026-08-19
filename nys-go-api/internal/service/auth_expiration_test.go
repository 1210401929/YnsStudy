package service

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
	"nys-go-api/internal/model"
	"nys-go-api/internal/security"
	"nys-go-api/internal/session"
)

func TestCheckUserLoginKeepsOriginalToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := cache.NewMemoryStore()
	cfg := &config.Config{Security: config.SecurityConfig{
		SessionCookieName:       "NYS_SESSION",
		SessionExpirationSecond: 7 * 24 * 60 * 60,
	}}
	jwt := security.NewJWT("test-secret", 7*24*time.Hour)
	manager := session.NewManager(store, cfg.Security)
	service := &Service{Config: cfg, Cache: store, Sessions: manager, JWT: jwt}
	user := model.User{GUID: "user-guid", Code: "user-code", Name: "测试用户"}

	loginWriter := httptest.NewRecorder()
	loginContext, _ := gin.CreateTestContext(loginWriter)
	loginContext.Request = httptest.NewRequest("POST", "/login", nil)
	if err := manager.Create(loginContext, user); err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Generate(user.GUID)
	if err != nil {
		t.Fatal(err)
	}

	checkWriter := httptest.NewRecorder()
	checkContext, _ := gin.CreateTestContext(checkWriter)
	checkContext.Request = httptest.NewRequest("POST", "/check", nil)
	checkContext.Request.AddCookie(loginWriter.Result().Cookies()[0])
	checkContext.Request.Header.Set("Authorization", token)
	result := service.CheckUserLogin(checkContext)
	if result.IsError {
		t.Fatalf("有效登录被拒绝: %s", result.ErrMsg)
	}
	payload, ok := result.Result.(map[string]any)
	if !ok || payload["userToken"] != token {
		t.Fatal("检查登录状态时换发了新 Token")
	}
}

func TestCheckUserLoginRejectsExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := cache.NewMemoryStore()
	cfg := &config.Config{Security: config.SecurityConfig{
		SessionCookieName:       "NYS_SESSION",
		SessionExpirationSecond: 7 * 24 * 60 * 60,
	}}
	expiredJWT := security.NewJWT("test-secret", -time.Second)
	manager := session.NewManager(store, cfg.Security)
	service := &Service{Config: cfg, Cache: store, Sessions: manager, JWT: expiredJWT}
	user := model.User{GUID: "user-guid", Code: "user-code"}

	loginWriter := httptest.NewRecorder()
	loginContext, _ := gin.CreateTestContext(loginWriter)
	loginContext.Request = httptest.NewRequest("POST", "/login", nil)
	if err := manager.Create(loginContext, user); err != nil {
		t.Fatal(err)
	}
	token, err := expiredJWT.Generate(user.GUID)
	if err != nil {
		t.Fatal(err)
	}

	checkWriter := httptest.NewRecorder()
	checkContext, _ := gin.CreateTestContext(checkWriter)
	checkContext.Request = httptest.NewRequest("POST", "/check", nil)
	checkContext.Request.AddCookie(loginWriter.Result().Cookies()[0])
	checkContext.Request.Header.Set("Authorization", token)
	result := service.CheckUserLogin(checkContext)
	if !result.IsError {
		t.Fatal("过期 Token 仍然通过登录检查")
	}
	if _, err := manager.Get(checkContext); err == nil {
		t.Fatal("过期 Token 对应的会话没有被销毁")
	}
}
