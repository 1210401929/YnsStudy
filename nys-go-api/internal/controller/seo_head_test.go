package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
	"nys-go-api/internal/security"
	"nys-go-api/internal/service"
	"nys-go-api/internal/session"
)

// 链接检查工具用 HEAD 检测页面是否有效：服务端输出的页面要和 GET 一样返回 200，只是没有正文。
func TestSEOPagesAnswerHEAD(t *testing.T) {
	cfg, err := config.Load("../../config/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Server.Mode = "test"
	store := cache.NewMemoryStore()
	jwt := security.NewJWT(cfg.Security.JWTSecret, time.Hour, time.Hour)
	services := service.New(cfg, nil, store, session.NewManager(store, cfg.Security), jwt)
	cipher, err := security.NewAESCipher(cfg.Security.AESKey, cfg.Security.AESIV)
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(cfg, services, cipher)

	// robots.txt 不需要数据库，可以直接比较 GET 和 HEAD
	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
	head := httptest.NewRecorder()
	router.ServeHTTP(head, httptest.NewRequest(http.MethodHead, "/robots.txt", nil))
	if get.Code != http.StatusOK || head.Code != http.StatusOK {
		t.Fatalf("GET=%d HEAD=%d，都应为 200", get.Code, head.Code)
	}
	if got, want := head.Header().Get("Content-Type"), get.Header().Get("Content-Type"); got != want {
		t.Fatalf("HEAD 的 Content-Type = %q，应与 GET 一致 %q", got, want)
	}

	// 其他服务端页面也要注册 HEAD（这里只检查路由，不实际执行，避免依赖数据库）
	registered := make(map[string]bool)
	for _, route := range router.Routes() {
		if route.Method == http.MethodHead {
			registered[route.Path] = true
		}
	}
	for _, path := range []string{"/", "/oneBlog/:id", "/sitemap.xml", "/robots.txt", "/rss.xml", "/archive", "/archive/page/:page", "/user/:u", "/note/:id", "/ynsStudy/Notes"} {
		if !registered[path] {
			t.Errorf("%s 没有注册 HEAD", path)
		}
	}
}
