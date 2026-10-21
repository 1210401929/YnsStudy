package controller

import (
	"strings"
	"testing"
	"time"

	"nys-go-api/internal/config"
	"nys-go-api/internal/service"
)

func noteTestConfig() *config.Config {
	return &config.Config{
		External: config.ExternalConfig{DomainName: "https://ynsstudy.cn"},
		SEO:      config.SEOConfig{SiteName: "YnsStudy", DefaultImage: "https://ynsstudy.cn/og-image.png"},
	}
}

const noteTestShell = `<!doctype html><html lang="en"><head><title>YnsStudy</title></head><body><div id="app"></div><script type="module" src="/assets/app.js"></script></body></html>`

func TestRenderSEONoteHTML(t *testing.T) {
	note := map[string]any{
		"GUID":        "12",
		"TYPE":        "tip",
		"TEXT":        `<p>Go 里用 <code>errors.Is</code> 判断错误</p><script>alert(1)</script>`,
		"IMAGES":      []string{"/uploadFile/editorImage/a.webp"},
		"TAGS":        []string{"Go", "踩坑"},
		"LINK":        service.LinkPreview{URL: "https://go.dev/doc", Title: "Go 文档", Description: "官方文档", SiteName: "go.dev"},
		"USERNAME":    "余南松",
		"USERNUM":     "101",
		"CREATE_TIME": time.Date(2026, 10, 1, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60)),
	}
	page := renderSEONoteHTML(noteTestShell, note,
		map[string]any{"GUID": "11", "TITLE": "更早的一条"}, nil, testOwner(), noteTestConfig())
	for _, expected := range []string{
		"<title>Go 里用 errors.Is 判断错误 - 便签墙 - YnsStudy</title>",
		`<link rel="canonical" href="https://ynsstudy.cn/note/12">`,
		`"@type":"SocialMediaPosting"`,
		`"image":["https://ynsstudy.cn/uploadFile/editorImage/a.webp"]`,
		`<h1 itemprop="headline">Go 里用 errors.Is 判断错误</h1>`,
		`技巧 · <a href="https://ynsstudy.cn/user/101"`,
		`<img src="/uploadFile/editorImage/a.webp" alt="Go 里用 errors.Is 判断错误 配图 1" loading="lazy">`,
		`<a href="https://ynsstudy.cn/note/11">← 上一条：更早的一条</a>`,
		`<a href="https://ynsstudy.cn/ynsStudy/Notes">回到便签墙 →</a>`,
		`"@type":"BreadcrumbList"`,
		`src="/assets/app.js"`,
		`<meta property="article:tag" content="Go">`,
		`"keywords":"Go,踩坑"`,
		`<a href="https://go.dev/doc" rel="nofollow ugc noopener" target="_blank">Go 文档</a> · go.dev<br>官方文档`,
		`<a href="https://ynsstudy.cn/ynsStudy/Notes?tag=%E8%B8%A9%E5%9D%91" rel="tag">#踩坑</a>`,
	} {
		if !strings.Contains(page, expected) {
			t.Errorf("随记页缺少 %q", expected)
		}
	}
	if strings.Contains(page, "alert(1)") {
		t.Fatal("随记页不应输出脚本")
	}
	if strings.Contains(page, "下一条") {
		t.Fatal("没有下一条时不应输出下一条链接")
	}
	if strings.Contains(page, "%!") {
		t.Fatal("随记页格式化参数数量不匹配")
	}
}

func TestRenderSEONoteWallHTML(t *testing.T) {
	notes := []map[string]any{
		{"GUID": "12", "TEXT": "<p>第一条</p>", "TYPE": "note", "USERNAME": "余南松"},
		{"GUID": "a-b", "TITLE": "旧帖子", "TYPE": "share", "USERNAME": "访客"},
	}
	page := renderSEONoteWallHTML(noteTestShell, notes, noteTestConfig())
	for _, expected := range []string{
		"<title>便签墙 - YnsStudy</title>",
		`<link rel="canonical" href="https://ynsstudy.cn/ynsStudy/Notes">`,
		`<a href="https://ynsstudy.cn/note/12">第一条</a>`,
		`<a href="https://ynsstudy.cn/note/a-b">旧帖子</a>`,
		`分享 · 访客`,
	} {
		if !strings.Contains(page, expected) {
			t.Errorf("便签墙缺少 %q", expected)
		}
	}
	if strings.Contains(page, "%!") {
		t.Fatal("便签墙格式化参数数量不匹配")
	}
}
