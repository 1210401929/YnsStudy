package controller

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
	"nys-go-api/internal/security"
	"nys-go-api/internal/service"
	"nys-go-api/internal/session"
)

func TestUploadedFilesAreCachedAndNotListed(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "editorImage"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "editorImage", "a.webp"), []byte("img"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{}
	cfg.Server.Mode = "test"
	cfg.Upload.PublicPrefix = "/uploadFile/"
	cfg.Upload.Directory = dir
	cfg.Security.Whitelist = []string{"/uploadFile/**"}
	cfg.Security.InnerCallHeader = "X-Inner-Test"
	cfg.Security.InnerCallSecret = "secret"
	store := cache.NewMemoryStore()
	jwt := security.NewJWT("test", time.Hour)
	services := service.New(cfg, nil, store, session.NewManager(store, cfg.Security), jwt)
	cipher, err := security.NewAESCipher("7546455674406856", "7513956994549178")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(cfg, services, cipher)

	file := httptest.NewRecorder()
	router.ServeHTTP(file, httptest.NewRequest("GET", "/uploadFile/editorImage/a.webp", nil))
	if file.Code != 200 || file.Body.String() != "img" {
		t.Fatalf("上传文件应可访问, 状态 %d", file.Code)
	}
	if !strings.Contains(file.Header().Get("Cache-Control"), "max-age=31536000") {
		t.Fatalf("上传文件应长期缓存, 实际 %q", file.Header().Get("Cache-Control"))
	}

	listing := httptest.NewRecorder()
	router.ServeHTTP(listing, httptest.NewRequest("GET", "/uploadFile/editorImage/", nil))
	if strings.Contains(listing.Body.String(), "a.webp") {
		t.Fatalf("不应列出目录内容, 状态 %d: %s", listing.Code, listing.Body.String())
	}
}
