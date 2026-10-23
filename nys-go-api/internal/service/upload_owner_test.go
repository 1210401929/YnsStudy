package service

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/config"
)

func uploadTestService(t *testing.T) *Service {
	t.Helper()
	cfg := &config.Config{}
	cfg.Server.MaxUploadMB = 1
	cfg.Upload.Directory = t.TempDir()
	cfg.Upload.PublicPrefix = "/uploadFile/"
	cfg.Upload.AllowedExtensions = []string{".png", ".svg", ".html"}
	cfg.External.DomainName = "https://ynsstudy.cn"
	return &Service{Config: cfg}
}

func multipartFile(t *testing.T, name, content string) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte(content))
	_ = writer.Close()
	request := httptest.NewRequest("POST", "/", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if err := request.ParseMultipartForm(1 << 20); err != nil {
		t.Fatal(err)
	}
	return request.MultipartForm.File["file"][0]
}

func TestSaveUploadedFileBlocksScriptableTypes(t *testing.T) {
	s := uploadTestService(t)
	for _, name := range []string{"a.svg", "b.SVG", "c.html"} {
		result := s.SaveUploadedFile(context.Background(), multipartFile(t, name, "<svg onload=alert(1)>"), "editorImage", "alice")
		if !result.IsError {
			t.Errorf("%s 即使在允许列表里也应被拒绝", name)
		}
	}
	result := s.SaveUploadedFile(context.Background(), multipartFile(t, "ok.png", "png"), "editorImage", "alice")
	if result.IsError {
		t.Fatalf("png 应能上传: %s", result.ErrMsg)
	}
	values := result.Result.(map[string]string)
	if _, leaked := values["fileRealUrl"]; leaked {
		t.Fatal("不应返回服务器上的绝对路径")
	}
}

func TestUploadKeyNormalizesSiteFiles(t *testing.T) {
	s := uploadTestService(t)
	for input, want := range map[string]string{
		"/uploadFile/editorImage/a.png":                     "/uploadFile/editorImage/a.png",
		"https://ynsstudy.cn/uploadFile/editorImage/a.png?x": "/uploadFile/editorImage/a.png",
		"/uploadFile/../config/config.yaml":                 "",
		"https://example.com/a.png":                         "",
		"/uploadFile/":                                      "",
	} {
		if got := s.uploadKey(input); got != want {
			t.Errorf("%q: 期望 %q, 实际 %q", input, want, got)
		}
	}
}

// 没有上传记录表（迁移 014 之前、或查询不到归属）时一律不删除，宁可留下文件也不误删别人的
func TestDeleteOwnedUploadRefusesWithoutRecord(t *testing.T) {
	s := uploadTestService(t)
	target := filepath.Join(s.Config.Upload.Directory, "editorImage", "victim.png")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("img"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := s.DeleteOwnedUpload(context.Background(), "/uploadFile/editorImage/victim.png", "attacker")
	if !result.IsError {
		t.Fatal("没有归属记录时不应删除")
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("文件不应被删除: %v", err)
	}
}

func TestNormalizeFriendLinkURLs(t *testing.T) {
	s := uploadTestService(t)
	for _, link := range []string{"javascript:alert(1)", " JavaScript:alert(document.cookie)", "data:text/html,<script>", "vbscript:x", ""} {
		if failure := s.normalizeFriendLinkURLs(map[string]any{"LINK": link}); failure == nil {
			t.Errorf("%q 应被拒绝", link)
		}
	}
	data := map[string]any{"LINK": "www.example.com", "AVATAR": "/uploadFile/friend/a.png"}
	if failure := s.normalizeFriendLinkURLs(data); failure != nil {
		t.Fatalf("正常友链被拒绝: %s", failure.ErrMsg)
	}
	if data["LINK"] != "https://www.example.com" || data["AVATAR"] != "/uploadFile/friend/a.png" {
		t.Fatalf("规范化结果不对: %v", data)
	}
	if failure := s.normalizeFriendLinkURLs(map[string]any{"AVATAR": "javascript:alert(1)"}); failure == nil {
		t.Fatal("javascript: 图标应被拒绝")
	}
}

func TestSiteNoticePath(t *testing.T) {
	s := uploadTestService(t)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "http://127.0.0.1:8080/pub-api/notice/addNotice", nil)
	for link, want := range map[string]string{
		"/oneBlog/12":                         "/oneBlog/12",
		"https://ynsstudy.cn/user/3?tab=blog": "/user/3?tab=blog",
		"http://127.0.0.1:8080/user/3":        "/user/3",
		"https://evil.example/login":          "",
		"//evil.example/x":                    "",
		"javascript:alert(1)":                 "",
	} {
		if got := s.siteNoticePath(c, link); got != want {
			t.Errorf("%q: 期望 %q, 实际 %q", link, want, got)
		}
	}
}
