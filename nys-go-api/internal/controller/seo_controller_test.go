package controller

import (
	"strings"
	"testing"
	"time"

	"nys-go-api/internal/config"
)

func TestRenderSEOHomeHTMLContainsIntroductionAndRealArticleLinks(t *testing.T) {
	cfg := &config.Config{
		External: config.ExternalConfig{DomainName: "https://ynsstudy.cn"},
		SEO: config.SEOConfig{
			SiteName:           "YnsStudy",
			DefaultDescription: "记录编程学习、技术实践与生活思考。",
			DefaultImage:       "https://ynsstudy.cn/finder.png",
		},
	}
	articles := []map[string]any{{
		"GUID":        "435",
		"BLOG_TITLE":  "我把博客从 Spring Cloud 改成了 Go",
		"USERNAME":    "YuNanSong",
		"CREATE_TIME": time.Date(2026, 8, 18, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60)),
	}}
	shell := `<!doctype html><html><head><title>旧标题</title><meta data-yns-seo="default" name="description" content="旧摘要"></head><body><noscript>YnsStudy 需要启用 JavaScript。</noscript><div id="app"></div><script type="module" src="/assets/app.js"></script></body></html>`

	page := renderSEOHomeHTML(shell, articles, cfg)
	for _, expected := range []string{
		`<title>YnsStudy - 技术分享与个人博客</title>`,
		`<link rel="canonical" href="https://ynsstudy.cn/">`,
		`"@type":"WebSite"`,
		`<h1 id="seo-home-title">YnsStudy</h1>`,
		`记录编程学习、技术实践与生活思考。`,
		`<a href="https://ynsstudy.cn/oneBlog/435">我把博客从 Spring Cloud 改成了 Go</a>`,
		`src="/assets/app.js"`,
	} {
		if !strings.Contains(page, expected) {
			t.Errorf("生成的首页 HTML 缺少 %q", expected)
		}
	}
	if strings.Contains(page, "需要启用 JavaScript") {
		t.Fatal("首页 SEO HTML 不应保留 JavaScript 提示")
	}
}

func TestRenderSEOArticleHTMLContainsArticleSignalsAndBody(t *testing.T) {
	cfg := &config.Config{
		External: config.ExternalConfig{DomainName: "https://ynsstudy.cn"},
		SEO: config.SEOConfig{
			SiteName:           "YnsStudy",
			DefaultDescription: "默认摘要",
			DefaultImage:       "https://ynsstudy.cn/finder.png",
		},
	}
	article := map[string]any{
		"GUID":        "411",
		"BLOG_TITLE":  "Go $1 SEO 实践",
		"USERNAME":    "YuNanSong",
		"CREATE_TIME": time.Date(2026, 8, 11, 10, 30, 0, 0, time.FixedZone("CST", 8*60*60)),
		"UPDATE_TIME": time.Date(2026, 8, 18, 9, 0, 0, 0, time.FixedZone("CST", 8*60*60)),
		"MAINTEXT": `<h1>正文一级标题</h1><p>这是真实文章摘要。</p>
<img src="/uploadFile/editorImage/a.png" alt="image.png" onerror="alert(1)">
<script>alert('xss')</script>`,
	}
	shell := `<!doctype html><html lang="en"><head><title>YnsStudy</title></head><body><div id="app"></div><script type="module" src="/assets/app.js"></script></body></html>`

	page := renderSEOArticleHTML(shell, article, cfg)
	for _, expected := range []string{
		"<title>Go $1 SEO 实践 - YnsStudy</title>",
		`<meta name="description" content="正文一级标题 这是真实文章摘要。">`,
		`<link rel="canonical" href="https://ynsstudy.cn/oneBlog/411">`,
		`<meta property="og:type" content="article">`,
		`<meta name="twitter:card" content="summary_large_image">`,
		`<script type="application/ld+json">`,
		`"@type":"BlogPosting"`,
		`<h1 itemprop="headline">Go $1 SEO 实践</h1>`,
		`<h2>正文一级标题</h2>`,
		`alt="Go $1 SEO 实践 配图 1"`,
		`loading="eager"`,
		`src="/assets/app.js"`,
	} {
		if !strings.Contains(page, expected) {
			t.Errorf("生成的文章 HTML 缺少 %q", expected)
		}
	}
	for _, forbidden := range []string{"<script>alert", "onerror=", "javascript:"} {
		if strings.Contains(page, forbidden) {
			t.Errorf("生成的文章 HTML 不应包含 %q", forbidden)
		}
	}
}

func TestSanitizeArticleHTMLKeepsContentAndRejectsDangerousURLs(t *testing.T) {
	content, firstImage := sanitizeArticleHTML(
		`<p onclick="bad()">正文<a href="javascript:bad()">危险链接</a></p>
<iframe src="https://example.com"></iframe><img src="https://ynsstudy.cn/a.jpg" alt="架构图">`,
		"测试文章",
		"https://ynsstudy.cn",
	)
	if !strings.Contains(content, "正文") || !strings.Contains(content, `alt="架构图"`) {
		t.Fatalf("白名单误删了正常内容: %s", content)
	}
	if strings.Contains(content, "onclick") || strings.Contains(content, "javascript:") || strings.Contains(content, "iframe") {
		t.Fatalf("白名单未清理危险内容: %s", content)
	}
	if firstImage != "https://ynsstudy.cn/a.jpg" {
		t.Fatalf("首图地址不正确: %s", firstImage)
	}
}
