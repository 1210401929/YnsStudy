package controller

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
	"nys-go-api/internal/security"
	"nys-go-api/internal/service"
	"nys-go-api/internal/session"
)

func TestQQAuthorizeBindsStateToBrowserCookie(t *testing.T) {
	cfg, err := config.Load("../../config/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Server.Mode = "test"
	cfg.QQOAuth.Enabled = true
	cfg.QQOAuth.AppID = "test-app"
	cfg.QQOAuth.AppKey = "test-key"

	store := cache.NewMemoryStore()
	jwt := security.NewJWT(cfg.Security.JWTSecret, time.Hour)
	sessions := session.NewManager(store, cfg.Security)
	services := service.New(cfg, nil, store, sessions, jwt)
	cipher, err := security.NewAESCipher(cfg.Security.AESKey, cfg.Security.AESIV)
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(cfg, services, cipher)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/pub-api/login/qq/authorize", nil)
	router.ServeHTTP(response, request)
	if response.Code != http.StatusFound {
		t.Fatalf("期望跳转到QQ授权页，实际状态码 %d", response.Code)
	}

	authorizationURL, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	state := authorizationURL.Query().Get("state")
	if state == "" {
		t.Fatal("QQ授权地址缺少 state")
	}

	var stateCookie *http.Cookie
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == qqOAuthStateCookieName {
			stateCookie = cookie
			break
		}
	}
	if stateCookie == nil || stateCookie.Value != state {
		t.Fatal("未将QQ登录 state 绑定到当前浏览器")
	}
	if !stateCookie.HttpOnly || !stateCookie.Secure || stateCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("QQ state Cookie 安全属性不正确: %#v", stateCookie)
	}

	callbackResponse := httptest.NewRecorder()
	callbackRequest := httptest.NewRequest(http.MethodGet, "/pub-api/login/qq/callback?code=test&state="+url.QueryEscape(state), nil)
	router.ServeHTTP(callbackResponse, callbackRequest)
	if !strings.Contains(callbackResponse.Body.String(), "当前浏览器不匹配") {
		t.Fatal("缺少浏览器 state Cookie 时应拒绝QQ登录回调")
	}
}
