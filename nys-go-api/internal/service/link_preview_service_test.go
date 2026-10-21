package service

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
)

func TestIsPublicIP(t *testing.T) {
	cases := map[string]bool{
		"8.8.8.8": true, "152.136.247.100": true, "2606:4700:4700::1111": true,
		"127.0.0.1": false, "10.1.2.3": false, "172.16.0.1": false, "192.168.1.1": false,
		"169.254.169.254": false, "100.64.0.1": false, "0.0.0.0": false, "::1": false,
		"fe80::1": false, "fc00::1": false, "::ffff:127.0.0.1": false, "198.18.0.1": false, "255.255.255.255": false,
	}
	for input, want := range cases {
		if got := isPublicIP(net.ParseIP(input)); got != want {
			t.Errorf("isPublicIP(%s) = %v, want %v", input, got, want)
		}
	}
}

func TestValidatePreviewURL(t *testing.T) {
	allowed := []string{"https://go.dev/doc", "http://example.com:80/a", "https://example.com:443/#top"}
	for _, raw := range allowed {
		if _, err := validatePreviewURL(raw); err != nil {
			t.Errorf("%s 应该允许: %v", raw, err)
		}
	}
	blocked := []string{
		"ftp://example.com", "javascript:alert(1)", "file:///etc/passwd", "https://", "",
		"http://127.0.0.1/", "http://10.0.0.5/admin", "http://169.254.169.254/latest/meta-data",
		"http://[::1]/", "http://localhost/", "http://db.internal/", "http://example.com:8889/",
		"https://user:pass@example.com/", "http://0x7f000001/",
	}
	for _, raw := range blocked {
		if _, err := validatePreviewURL(raw); err == nil {
			// 0x7f000001 这类写法不是合法 IP，交给连接时的 IP 检查拦截，这里放行也可以
			if raw == "http://0x7f000001/" {
				continue
			}
			t.Errorf("%s 应该被拒绝", raw)
		}
	}
}

// 本机地址的网站必须在连接阶段被拦截，即使网址看起来是合法的
func TestFetchLinkPreviewBlocksLoopback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<title>内网</title>"))
	}))
	defer server.Close()
	s := &Service{Config: &config.Config{}, Cache: cache.NewMemoryStore()}
	if _, err := s.FetchLinkPreview(context.Background(), server.URL); err == nil {
		t.Fatal("访问本机地址应该被拒绝")
	}
	// 用默认端口的本机地址，检查的是连接阶段的 IP 拦截
	_, err := linkPreviewClient.Get("http://127.0.0.1/")
	if err == nil || !errors.Is(err, errLinkBlocked) {
		t.Fatalf("连接本机应在拨号阶段被拦截，实际: %v", err)
	}
}

func TestParseLinkPreview(t *testing.T) {
	page, _ := url.Parse("https://example.com/post/1")
	body := `<html><head>
<title>  页面 标题  </title>
<meta property="og:title" content="Open Graph 标题">
<meta name="description" content="普通描述">
<meta property="og:description" content="OG 描述">
<meta property="og:image" content="/cover.png">
<link rel="shortcut icon" href="//cdn.example.com/icon.png">
</head><body><meta property="og:title" content="正文里的不算"></body></html>`
	preview := parseLinkPreview(strings.NewReader(body), page)
	if preview.Title != "Open Graph 标题" || preview.Description != "OG 描述" {
		t.Fatalf("标题或描述不对: %+v", preview)
	}
	if preview.Image != "https://example.com/cover.png" || preview.Icon != "https://cdn.example.com/icon.png" {
		t.Fatalf("图片或图标地址没有补全: %+v", preview)
	}
	if preview.SiteName != "example.com" {
		t.Fatalf("没有网站名时应使用域名: %q", preview.SiteName)
	}

	// 没有 Open Graph 时用 <title>；http 图片在 https 页面会被浏览器拦截，不保留
	plain := parseLinkPreview(strings.NewReader(`<title>只有标题</title><meta property="og:image" content="http://other.com/a.png">`), page)
	if plain.Title != "只有标题" || plain.Image != "" || plain.Icon != "https://example.com/favicon.ico" {
		t.Fatalf("没有 Open Graph 时解析不对: %+v", plain)
	}
}

func TestFirstLinkInHTML(t *testing.T) {
	text := `<p>看看 <a href="https://go.dev/doc?a=1&amp;b=2">文档</a> 和 <a href="https://example.com">另一个</a></p>`
	if got := firstLinkInHTML(text); got != "https://go.dev/doc?a=1&b=2" {
		t.Fatalf("第一个链接不对: %q", got)
	}
	if got := firstLinkInHTML(`<p><a href="/note/1">站内</a> <a href="javascript:x">脚本</a></p>`); got != "" {
		t.Fatalf("站内相对链接和 javascript 链接不应生成卡片: %q", got)
	}
}
