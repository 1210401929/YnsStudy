package controller

import (
	"strings"
	"testing"
	"time"

	"nys-go-api/internal/config"
	"nys-go-api/internal/service"
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
		"USERNUM":     "10001",
		"CREATE_TIME": time.Date(2026, 8, 11, 10, 30, 0, 0, time.FixedZone("CST", 8*60*60)),
		"UPDATE_TIME": time.Date(2026, 8, 18, 9, 0, 0, 0, time.FixedZone("CST", 8*60*60)),
		"MAINTEXT": `<h1>正文一级标题</h1><p>这是真实文章摘要。</p>
<img src="/uploadFile/editorImage/a.png" alt="image.png" onerror="alert(1)">
<script>alert('xss')</script>`,
	}
	shell := `<!doctype html><html lang="en"><head><title>YnsStudy</title></head><body><div id="app"></div><script type="module" src="/assets/app.js"></script></body></html>`

	links := service.ArticleLinks{
		Previous: map[string]any{"GUID": "410", "BLOG_TITLE": "上一篇文章"},
		Next:     map[string]any{"GUID": "412", "BLOG_TITLE": "下一篇文章"},
		Related: []map[string]any{
			{"GUID": "410", "BLOG_TITLE": "上一篇文章"},
			{"GUID": "300", "BLOG_TITLE": "同作者的旧文章"},
		},
	}
	page := renderSEOArticleHTML(shell, article, links, cfg)
	for _, expected := range []string{
		"<title>Go $1 SEO 实践 - YnsStudy</title>",
		`<meta name="description" content="正文一级标题 这是真实文章摘要。">`,
		`<link rel="canonical" href="https://ynsstudy.cn/oneBlog/411">`,
		`<meta property="og:type" content="article">`,
		`<meta name="twitter:card" content="summary_large_image">`,
		`<script type="application/ld+json">`,
		`"@type":"BlogPosting"`,
		`"logo":"https://ynsstudy.cn/icon-512.png"`,
		`<h1 itemprop="headline">Go $1 SEO 实践</h1>`,
		`<h2>正文一级标题</h2>`,
		`alt="Go $1 SEO 实践 配图 1"`,
		`loading="eager"`,
		`src="/assets/app.js"`,
		`"@type":"BreadcrumbList"`,
		`<a href="https://ynsstudy.cn/archive">文章归档</a>`,
		`<a href="https://ynsstudy.cn/oneBlog/410">← 上一篇：上一篇文章</a>`,
		`<a href="https://ynsstudy.cn/oneBlog/412">下一篇：下一篇文章 →</a>`,
		`<h2>YuNanSong 的其他文章</h2>`,
		`<a href="https://ynsstudy.cn/user/10001" itemprop="author">YuNanSong</a>`,
		`"author":{"@type":"Person","name":"YuNanSong","url":"https://ynsstudy.cn/user/10001"}`,
		`<a href="https://ynsstudy.cn/oneBlog/300">同作者的旧文章</a>`,
	} {
		if !strings.Contains(page, expected) {
			t.Errorf("生成的文章 HTML 缺少 %q", expected)
		}
	}
	if strings.Count(page, `href="https://ynsstudy.cn/oneBlog/410"`) != 1 {
		t.Error("同作者文章列表应排除已作为上一篇展示的文章")
	}
	if strings.Contains(page, "%!") {
		t.Fatal("文章页格式化参数数量不匹配")
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

func TestRenderSEOArchiveHTMLListsArticlesAndPagination(t *testing.T) {
	cfg := &config.Config{
		External: config.ExternalConfig{DomainName: "https://ynsstudy.cn"},
		SEO:      config.SEOConfig{SiteName: "YnsStudy", DefaultImage: "https://ynsstudy.cn/finder.png"},
	}
	rows := []map[string]any{{
		"GUID":         "435",
		"BLOG_TITLE":   "Go <SEO> 实践",
		"USERNAME":     "YuNanSong",
		"CREATE_TIME":  time.Date(2026, 8, 18, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60)),
		"EXCERPT_HTML": `<p>这是摘要<script>alert(1)</script></p><img src="x.png"><p>被截断的`,
	}}

	page := renderSEOArchiveHTML(rows, 2, 5, 90, cfg)
	for _, expected := range []string{
		`<title>文章归档（第 2 页） - YnsStudy</title>`,
		`<link rel="canonical" href="https://ynsstudy.cn/archive/page/2">`,
		`<link rel="prev" href="https://ynsstudy.cn/archive">`,
		`<link rel="next" href="https://ynsstudy.cn/archive/page/3">`,
		`<a href="https://ynsstudy.cn/oneBlog/435">Go &lt;SEO&gt; 实践</a>`,
		`<time datetime="2026-08-18T10:00:00+08:00">2026-08-18</time>`,
		`共 90 篇公开文章 · 第 2 / 5 页`,
		`<span class="current" aria-current="page">2</span>`,
		`"@type":"BreadcrumbList"`,
		`"@type":"CollectionPage"`,
		`<link rel="icon" href="/favicon.ico" sizes="48x48">`,
	} {
		if !strings.Contains(page, expected) {
			t.Errorf("生成的归档 HTML 缺少 %q", expected)
		}
	}
	if strings.Contains(page, "<script>alert") || strings.Contains(page, "<SEO>") {
		t.Fatal("归档页必须转义标题和摘要")
	}
	if strings.Contains(page, "%!") {
		t.Fatalf("归档页格式化参数数量不匹配: %s", page)
	}
}

func TestArchivePagerAndPageCount(t *testing.T) {
	if got := archiveTotalPages(0); got != 1 {
		t.Fatalf("没有文章时也应有 1 页，得到 %d", got)
	}
	if got := archiveTotalPages(service.ArchivePageSize + 1); got != 2 {
		t.Fatalf("超过一页的文章应分为 2 页，得到 %d", got)
	}
	if pager := renderArchivePager("https://ynsstudy.cn", 1, 1); pager != "" {
		t.Fatalf("只有一页时不应输出分页: %s", pager)
	}
	pager := renderArchivePager("https://ynsstudy.cn", 10, 20)
	for _, expected := range []string{
		`<a href="https://ynsstudy.cn/archive">1</a>`,
		`<a href="https://ynsstudy.cn/archive/page/20">20</a>`,
		`<a href="https://ynsstudy.cn/archive/page/9">← 上一页</a>`,
		`<a href="https://ynsstudy.cn/archive/page/11">下一页 →</a>`,
	} {
		if !strings.Contains(pager, expected) {
			t.Errorf("分页缺少 %q: %s", expected, pager)
		}
	}
}
