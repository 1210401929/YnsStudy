package controller

import (
	"encoding/json"
	"fmt"
	htmlstd "html"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"

	"nys-go-api/internal/config"
	"nys-go-api/internal/model"
)

var (
	titleElementPattern = regexp.MustCompile(`(?is)<title\b[^>]*>.*?</title>`)
	appElementPattern   = regexp.MustCompile(`(?is)<div\s+id=["']app["']\s*>\s*</div>`)
	defaultSEOElement   = regexp.MustCompile(`(?is)\s*<(?:meta|link)\b[^>]*data-yns-seo=["']default["'][^>]*>\s*`)
	javascriptNotice    = regexp.MustCompile(`(?is)\s*<noscript\b[^>]*>.*?需要启用\s*JavaScript.*?</noscript>\s*`)
	numericHTMLAttr     = regexp.MustCompile(`^[0-9]{1,5}$`)
)

func (h *Controller) registerSEORoutes(router *gin.Engine) {
	// 这些地址由 Nginx 直接转发到 Go。API 原地址继续保留，避免影响现有前端和订阅者。
	router.GET("/", h.seoHome)
	router.GET("/oneBlog/:id", h.seoArticle)
	router.GET("/sitemap.xml", h.sitemap)
	router.GET("/robots.txt", h.robots)
	router.GET("/rss.xml", h.rss)
}

// seoHome 为根路径输出站点介绍和最新公开文章链接。
// Vue 启动后仍会接管页面；不执行 JavaScript 的抓取器也能理解站点主题并发现新文章。
func (h *Controller) seoHome(c *gin.Context) {
	result := h.service.GetLatestBlogsForSEO(c.Request.Context(), 10)
	if result.IsError {
		h.renderSEOStatusPage(c, http.StatusInternalServerError, "首页暂时无法访问", "服务器读取最新文章时出现异常，请稍后重试。")
		return
	}
	articles, _ := result.Result.([]map[string]any)
	page := renderSEOHomeHTML(h.readFrontendShell(), articles, h.service.Config)
	cacheSeconds := h.service.Config.SEO.ResponseCacheSeconds
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=0, s-maxage=%d, stale-while-revalidate=60", cacheSeconds))
	c.Header("X-Robots-Tag", "index, follow, max-image-preview:large, max-snippet:-1")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}

// seoArticle 返回“真实文章 HTML + 原 Vue 入口脚本”。
// 搜索引擎无需执行 JavaScript 就能看到标题和正文；普通浏览器加载完成后仍由 Vue 接管，
// 所以评论、登录、点赞等现有功能不需要重写。
func (h *Controller) seoArticle(c *gin.Context) {
	articleID := strings.TrimSpace(c.Param("id"))
	article, found, err := h.service.GetPublicBlogForSEO(c.Request.Context(), articleID)
	if err != nil {
		h.renderSEOStatusPage(c, http.StatusInternalServerError, "文章暂时无法访问", "服务器读取文章时出现异常，请稍后重试。")
		return
	}
	if !found {
		h.renderSEOStatusPage(c, http.StatusNotFound, "文章不存在", "这篇文章可能已删除、设为私密或地址有误。")
		return
	}

	page := renderSEOArticleHTML(h.readFrontendShell(), article, h.service.Config)
	cacheSeconds := h.service.Config.SEO.ResponseCacheSeconds
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=0, s-maxage=%d, stale-while-revalidate=60", cacheSeconds))
	c.Header("X-Robots-Tag", "index, follow, max-image-preview:large, max-snippet:-1")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}

// readFrontendShell 读取正在部署的 Vue 入口。路径配置错误时仍返回一个最小 HTML 外壳，
// 让首页和文章页保持可读，而不是因为静态文件问题直接返回 500。
func (h *Controller) readFrontendShell() string {
	shell, err := os.ReadFile(h.service.Config.SEO.FrontendIndexFile)
	if err != nil {
		return defaultFrontendShell(h.service.Config.SEO.SiteName)
	}
	return string(shell)
}

func (h *Controller) sitemap(c *gin.Context) {
	result := h.service.GetAllPublicBlogIDs(c.Request.Context())
	if result.IsError {
		writeResult(c, result)
		return
	}
	rows, _ := result.Result.([]map[string]any)
	domain := strings.TrimRight(h.service.Config.External.DomainName, "/")
	urls := []sitemapURL{
		{Location: domain + "/", Frequency: "daily", Priority: "1.0"},
		{Location: domain + "/ynsStudy/Home", Frequency: "daily", Priority: "0.8"},
		{Location: domain + "/ynsStudy/MyBlog", Frequency: "daily", Priority: "0.9"},
		{Location: domain + "/ynsStudy/Resources", Frequency: "weekly", Priority: "0.6"},
		{Location: domain + "/ynsStudy/Community", Frequency: "daily", Priority: "0.6"},
		{Location: domain + "/ynsStudy/FriendLink", Frequency: "monthly", Priority: "0.4"},
		{Location: domain + "/ynsStudy/About", Frequency: "monthly", Priority: "0.5"},
	}
	for _, row := range rows {
		urls = append(urls, sitemapURL{
			Location:   domain + "/oneBlog/" + url.PathEscape(model.StringValue(row, "GUID")),
			LastModify: formatDate(model.Lookup(row, "LAST_MODIFIED")),
			Frequency:  "weekly",
			Priority:   "0.8",
		})
	}
	c.Header("Cache-Control", "public, max-age=300")
	writeXML(c, sitemapDocument{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: urls})
}

func (h *Controller) robots(c *gin.Context) {
	domain := strings.TrimRight(h.service.Config.External.DomainName, "/")
	body := "User-agent: *\n" +
		"Allow: /\n" +
		"Disallow: /api/\n" +
		"Disallow: /blog-api/\n" +
		"Disallow: /sso\n" +
		"Disallow: /personalCenter\n" +
		"Disallow: /ynsStudy/MyBlog/content\n\n" +
		"Sitemap: " + domain + "/sitemap.xml\n"
	c.Header("Cache-Control", "public, max-age=3600")
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(body))
}

func (h *Controller) renderSEOStatusPage(c *gin.Context, status int, title, message string) {
	domain := strings.TrimRight(h.service.Config.External.DomainName, "/")
	c.Header("Cache-Control", "no-store")
	c.Header("X-Robots-Tag", "noindex, follow")
	page := fmt.Sprintf(`<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="robots" content="noindex,follow"><title>%s - %s</title>
<style>body{margin:0;background:#f5f7fa;color:#26384a;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC",sans-serif}.box{max-width:680px;margin:12vh auto;padding:40px 28px;background:#fff;border-radius:16px;box-shadow:0 16px 45px rgba(31,45,61,.1);text-align:center}h1{font-size:28px}p{color:#68788a;line-height:1.8}a{display:inline-block;margin-top:14px;color:#087cad;text-decoration:none}</style></head>
<body><main class="box"><h1>%s</h1><p>%s</p><a href="%s/">返回 YnsStudy 首页</a></main></body></html>`,
		htmlstd.EscapeString(title), htmlstd.EscapeString(h.service.Config.SEO.SiteName),
		htmlstd.EscapeString(title), htmlstd.EscapeString(message), htmlstd.EscapeString(domain))
	c.Data(status, "text/html; charset=utf-8", []byte(page))
}

type articleStructuredData struct {
	Context          string                 `json:"@context"`
	Type             string                 `json:"@type"`
	Headline         string                 `json:"headline"`
	Description      string                 `json:"description"`
	URL              string                 `json:"url"`
	DatePublished    string                 `json:"datePublished,omitempty"`
	DateModified     string                 `json:"dateModified,omitempty"`
	MainEntityOfPage structuredMainEntity   `json:"mainEntityOfPage"`
	Author           structuredPerson       `json:"author"`
	Publisher        structuredOrganization `json:"publisher"`
	Image            []string               `json:"image,omitempty"`
}

type structuredMainEntity struct {
	Type string `json:"@type"`
	ID   string `json:"@id"`
}

type structuredPerson struct {
	Type string `json:"@type"`
	Name string `json:"name"`
}

type structuredOrganization struct {
	Type string `json:"@type"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

func renderSEOHomeHTML(shell string, articles []map[string]any, cfg *config.Config) string {
	// 首页使用服务端生成的完整站点 meta，移除 Vue 外壳中的默认副本，避免重复 description/OG。
	shell = defaultSEOElement.ReplaceAllString(shell, "\n")
	shell = javascriptNotice.ReplaceAllString(shell, "\n")
	domain := strings.TrimRight(cfg.External.DomainName, "/")
	canonical := domain + "/"
	title := cfg.SEO.SiteName + " - 技术分享与个人博客"
	description := strings.TrimSpace(cfg.SEO.DefaultDescription)
	if description == "" {
		description = cfg.SEO.SiteName + " 个人博客"
	}

	websiteJSON, _ := json.Marshal(map[string]any{
		"@context":    "https://schema.org",
		"@type":       "WebSite",
		"name":        cfg.SEO.SiteName,
		"url":         canonical,
		"description": description,
	})
	escapedTitle := htmlstd.EscapeString(title)
	escapedDescription := htmlstd.EscapeString(description)
	escapedCanonical := htmlstd.EscapeString(canonical)
	escapedImage := htmlstd.EscapeString(cfg.SEO.DefaultImage)

	head := fmt.Sprintf(`
<meta name="description" content="%s">
<meta name="robots" content="index,follow,max-image-preview:large,max-snippet:-1,max-video-preview:-1">
<link rel="canonical" href="%s">
<meta property="og:type" content="website">
<meta property="og:site_name" content="%s">
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
<meta property="og:url" content="%s">
<meta property="og:image" content="%s">
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="%s">
<meta name="twitter:description" content="%s">
<meta name="twitter:image" content="%s">
<script type="application/ld+json">%s</script>
<style id="yns-seo-home-first-paint">.seo-home-fallback{max-width:1040px;margin:36px auto;padding:0 20px;color:#26384a;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC",sans-serif}.seo-home-intro,.seo-home-latest{background:#fff;border-radius:14px;padding:28px 30px;box-shadow:0 10px 32px rgba(31,45,61,.08)}.seo-home-intro h1{margin:0 0 12px;font-size:34px}.seo-home-intro p{margin:0;color:#607286;line-height:1.8}.seo-home-nav{display:flex;flex-wrap:wrap;gap:16px;margin-top:18px}.seo-home-nav a,.seo-home-latest a{color:#087cad;text-decoration:none}.seo-home-latest{margin-top:22px}.seo-home-latest h2{margin:0 0 16px;font-size:24px}.seo-home-latest ul{list-style:none;margin:0;padding:0}.seo-home-latest li{display:flex;justify-content:space-between;gap:20px;padding:12px 0;border-bottom:1px solid #edf0f3}.seo-home-latest li:last-child{border-bottom:0}.seo-home-meta{flex:none;color:#8592a2;font-size:13px}@media(max-width:640px){.seo-home-fallback{margin:14px auto;padding:0 10px}.seo-home-intro,.seo-home-latest{padding:22px 18px}.seo-home-latest li{display:block}.seo-home-meta{display:block;margin-top:6px}}</style>
`, escapedDescription, escapedCanonical, htmlstd.EscapeString(cfg.SEO.SiteName), escapedTitle,
		escapedDescription, escapedCanonical, escapedImage, escapedTitle, escapedDescription, escapedImage, websiteJSON)

	var latest strings.Builder
	for _, article := range articles {
		articleID := strings.TrimSpace(model.StringValue(article, "GUID"))
		if articleID == "" {
			continue
		}
		articleTitle := strings.TrimSpace(model.StringValue(article, "BLOG_TITLE"))
		if articleTitle == "" {
			articleTitle = "未命名文章"
		}
		articleURL := domain + "/oneBlog/" + url.PathEscape(articleID)
		author := strings.TrimSpace(model.StringValue(article, "USERNAME"))
		dateText := formatDate(model.Lookup(article, "CREATE_TIME"))
		metaText := strings.Trim(strings.Join([]string{author, dateText}, " · "), " ·")
		fmt.Fprintf(&latest, `<li><a href="%s">%s</a><span class="seo-home-meta">%s</span></li>`,
			htmlstd.EscapeString(articleURL), htmlstd.EscapeString(articleTitle), htmlstd.EscapeString(metaText))
	}
	if latest.Len() == 0 {
		latest.WriteString(`<li>暂时还没有公开文章</li>`)
	}

	body := fmt.Sprintf(`<main class="seo-home-fallback">
<section class="seo-home-intro" aria-labelledby="seo-home-title">
<h1 id="seo-home-title">%s</h1>
<p>%s</p>
<nav class="seo-home-nav" aria-label="主要栏目"><a href="%s/ynsStudy/Home">首页</a><a href="%s/ynsStudy/MyBlog">博客</a><a href="%s/ynsStudy/Resources">资源</a><a href="%s/ynsStudy/Community">社区</a><a href="%s/ynsStudy/About">关于</a></nav>
</section>
<section class="seo-home-latest" aria-labelledby="seo-latest-title"><h2 id="seo-latest-title">最新文章</h2><ul>%s</ul></section>
</main>`, htmlstd.EscapeString(cfg.SEO.SiteName), escapedDescription,
		htmlstd.EscapeString(domain), htmlstd.EscapeString(domain), htmlstd.EscapeString(domain),
		htmlstd.EscapeString(domain), htmlstd.EscapeString(domain), latest.String())

	if titleElementPattern.MatchString(shell) {
		shell = titleElementPattern.ReplaceAllStringFunc(shell, func(string) string {
			return "<title>" + escapedTitle + "</title>"
		})
	} else {
		head = "<title>" + escapedTitle + "</title>\n" + head
	}
	if strings.Contains(shell, "</head>") {
		shell = strings.Replace(shell, "</head>", head+"</head>", 1)
	}
	if appElementPattern.MatchString(shell) {
		shell = appElementPattern.ReplaceAllStringFunc(shell, func(string) string {
			return `<div id="app">` + body + `</div>`
		})
	} else {
		shell = strings.Replace(shell, "</body>", body+"</body>", 1)
	}
	return strings.Replace(shell, "window.prerenderReady = false", "window.prerenderReady = true", 1)
}

func renderSEOArticleHTML(shell string, article map[string]any, cfg *config.Config) string {
	// index.html 的站点级默认 meta 适用于普通页面；文章页必须移除它们，避免出现两份 description/OG。
	shell = defaultSEOElement.ReplaceAllString(shell, "\n")
	shell = javascriptNotice.ReplaceAllString(shell, "\n")
	domain := strings.TrimRight(cfg.External.DomainName, "/")
	articleID := model.StringValue(article, "GUID")
	canonical := domain + "/oneBlog/" + url.PathEscape(articleID)
	title := strings.TrimSpace(model.StringValue(article, "BLOG_TITLE"))
	if title == "" {
		title = "博客文章"
	}
	author := strings.TrimSpace(model.StringValue(article, "USERNAME"))
	if author == "" {
		author = cfg.SEO.SiteName
	}

	safeContent, firstImage := sanitizeArticleHTML(model.StringValue(article, "MAINTEXT"), title, domain)
	description := truncateRunes(collapseWhitespace(plainTextFromHTML(safeContent)), 180)
	if description == "" {
		description = cfg.SEO.DefaultDescription
	}
	if firstImage == "" {
		firstImage = cfg.SEO.DefaultImage
	}
	published := formatSEOTime(model.Lookup(article, "CREATE_TIME"))
	modified := formatSEOTime(model.Lookup(article, "UPDATE_TIME"))
	if modified == "" {
		modified = published
	}

	structured := articleStructuredData{
		Context:       "https://schema.org",
		Type:          "BlogPosting",
		Headline:      title,
		Description:   description,
		URL:           canonical,
		DatePublished: published,
		DateModified:  modified,
		MainEntityOfPage: structuredMainEntity{
			Type: "WebPage",
			ID:   canonical,
		},
		Author:    structuredPerson{Type: "Person", Name: author},
		Publisher: structuredOrganization{Type: "Organization", Name: cfg.SEO.SiteName, URL: domain},
	}
	if firstImage != "" {
		structured.Image = []string{firstImage}
	}
	structuredJSON, _ := json.Marshal(structured)

	escapedTitle := htmlstd.EscapeString(title)
	escapedFullTitle := htmlstd.EscapeString(title + " - " + cfg.SEO.SiteName)
	escapedDescription := htmlstd.EscapeString(description)
	escapedCanonical := htmlstd.EscapeString(canonical)
	escapedImage := htmlstd.EscapeString(firstImage)
	escapedAuthor := htmlstd.EscapeString(author)
	escapedPublished := htmlstd.EscapeString(published)

	head := fmt.Sprintf(`
<meta name="description" content="%s">
<meta name="robots" content="index,follow,max-image-preview:large,max-snippet:-1,max-video-preview:-1">
<link rel="canonical" href="%s">
<meta property="og:type" content="article">
<meta property="og:site_name" content="%s">
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
<meta property="og:url" content="%s">
<meta property="og:image" content="%s">
<meta property="article:published_time" content="%s">
<meta property="article:modified_time" content="%s">
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="%s">
<meta name="twitter:description" content="%s">
<meta name="twitter:image" content="%s">
<script type="application/ld+json">%s</script>
<style id="yns-seo-first-paint">.seo-article-fallback{max-width:960px;margin:32px auto;padding:0 20px;color:#26384a;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC",sans-serif}.seo-article-fallback article{background:#fff;border-radius:14px;padding:32px;box-shadow:0 10px 32px rgba(31,45,61,.08)}.seo-breadcrumb{margin-bottom:22px;font-size:14px}.seo-breadcrumb a{color:#087cad;text-decoration:none}.seo-article-fallback h1{font-size:32px;line-height:1.35;margin:0 0 12px}.seo-article-meta{color:#778596;font-size:14px;margin-bottom:28px}.seo-article-content{font-size:16px;line-height:1.85;overflow-wrap:anywhere}.seo-article-content img{display:block;max-width:100%%;height:auto;margin:20px auto}.seo-article-content pre{overflow:auto;padding:16px;background:#f6f8fa;border-radius:8px}.seo-article-content table{display:block;max-width:100%%;overflow:auto;border-collapse:collapse}.seo-article-content td,.seo-article-content th{padding:8px;border:1px solid #dfe4ea}@media(max-width:640px){.seo-article-fallback{margin:12px auto;padding:0 10px}.seo-article-fallback article{padding:22px 16px}.seo-article-fallback h1{font-size:25px}}</style>
`, escapedDescription, escapedCanonical, htmlstd.EscapeString(cfg.SEO.SiteName), escapedTitle,
		escapedDescription, escapedCanonical, escapedImage, escapedPublished, htmlstd.EscapeString(modified),
		escapedTitle, escapedDescription, escapedImage, structuredJSON)

	body := fmt.Sprintf(`<main class="seo-article-fallback">
<article itemscope itemtype="https://schema.org/BlogPosting">
<nav class="seo-breadcrumb" aria-label="面包屑"><a href="%s/">YnsStudy</a> / <a href="%s/ynsStudy/MyBlog">博客</a></nav>
<header><h1 itemprop="headline">%s</h1><p class="seo-article-meta"><span itemprop="author">%s</span> · <time itemprop="datePublished" datetime="%s">%s</time></p></header>
<div class="seo-article-content" itemprop="articleBody">%s</div>
</article></main>`, htmlstd.EscapeString(domain), htmlstd.EscapeString(domain), escapedTitle, escapedAuthor,
		escapedPublished, htmlstd.EscapeString(formatDate(model.Lookup(article, "CREATE_TIME"))), safeContent)

	if titleElementPattern.MatchString(shell) {
		shell = titleElementPattern.ReplaceAllStringFunc(shell, func(string) string {
			return "<title>" + escapedFullTitle + "</title>"
		})
	} else {
		head = "<title>" + escapedFullTitle + "</title>\n" + head
	}
	shell = strings.Replace(shell, `lang="en"`, `lang="zh-CN"`, 1)
	if strings.Contains(shell, "</head>") {
		shell = strings.Replace(shell, "</head>", head+"</head>", 1)
	}
	if appElementPattern.MatchString(shell) {
		shell = appElementPattern.ReplaceAllStringFunc(shell, func(string) string {
			return `<div id="app">` + body + `</div>`
		})
	} else {
		shell = strings.Replace(shell, "</body>", body+"</body>", 1)
	}
	return strings.Replace(shell, "window.prerenderReady = false", "window.prerenderReady = true", 1)
}

func defaultFrontendShell(siteName string) string {
	return `<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>` +
		htmlstd.EscapeString(siteName) + `</title></head><body><div id="app"></div></body></html>`
}

type articleSanitizeState struct {
	articleTitle string
	domain       string
	imageCount   int
	firstImage   string
}

// sanitizeArticleHTML 为服务端首屏建立一层白名单。
// 数据库存的是编辑器 HTML，不能直接拼进 Go 响应，否则恶意 script/onerror 会变成存储型 XSS。
func sanitizeArticleHTML(raw, articleTitle, domain string) (string, string) {
	document, err := html.Parse(strings.NewReader("<!doctype html><html><body>" + raw + "</body></html>"))
	if err != nil {
		return htmlstd.EscapeString(plainTextFromHTML(raw)), ""
	}
	body := findHTMLElement(document, "body")
	if body == nil {
		return "", ""
	}
	state := &articleSanitizeState{articleTitle: articleTitle, domain: domain}
	var output strings.Builder
	for child := body.FirstChild; child != nil; child = child.NextSibling {
		for _, safeNode := range cloneSafeArticleNode(child, state) {
			_ = html.Render(&output, safeNode)
		}
	}
	return output.String(), state.firstImage
}

func cloneSafeArticleNode(node *html.Node, state *articleSanitizeState) []*html.Node {
	switch node.Type {
	case html.TextNode:
		return []*html.Node{{Type: html.TextNode, Data: node.Data}}
	case html.ElementNode:
		tag := strings.ToLower(node.Data)
		if blockedArticleTags[tag] {
			return nil
		}
		if tag == "h1" {
			tag = "h2" // 页面主标题已经使用唯一 H1，正文旧数据中的 H1 下调一级。
		}
		if !allowedArticleTags[tag] {
			return cloneSafeArticleChildren(node, state)
		}
		if tag == "img" {
			return cloneSafeArticleImage(node, state)
		}
		clone := &html.Node{Type: html.ElementNode, Data: tag}
		clone.Attr = safeArticleAttributes(tag, node.Attr)
		for _, child := range cloneSafeArticleChildren(node, state) {
			clone.AppendChild(child)
		}
		return []*html.Node{clone}
	default:
		return nil
	}
}

func cloneSafeArticleChildren(node *html.Node, state *articleSanitizeState) []*html.Node {
	children := make([]*html.Node, 0)
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, cloneSafeArticleNode(child, state)...)
	}
	return children
}

func cloneSafeArticleImage(node *html.Node, state *articleSanitizeState) []*html.Node {
	src := safeArticleURL(attributeValue(node.Attr, "src"))
	if src == "" {
		return nil
	}
	state.imageCount++
	alt := strings.TrimSpace(attributeValue(node.Attr, "alt"))
	if genericImageAlt(alt) {
		alt = fmt.Sprintf("%s 配图 %d", state.articleTitle, state.imageCount)
	}
	attributes := []html.Attribute{{Key: "src", Val: src}, {Key: "alt", Val: alt}, {Key: "decoding", Val: "async"}}
	if state.imageCount == 1 {
		attributes = append(attributes, html.Attribute{Key: "loading", Val: "eager"}, html.Attribute{Key: "fetchpriority", Val: "high"})
		state.firstImage = absoluteArticleURL(state.domain, src)
	} else {
		attributes = append(attributes, html.Attribute{Key: "loading", Val: "lazy"})
	}
	for _, name := range []string{"width", "height"} {
		if value := attributeValue(node.Attr, name); numericHTMLAttr.MatchString(value) {
			attributes = append(attributes, html.Attribute{Key: name, Val: value})
		}
	}
	return []*html.Node{{Type: html.ElementNode, Data: "img", Attr: attributes}}
}

func safeArticleAttributes(tag string, attributes []html.Attribute) []html.Attribute {
	result := make([]html.Attribute, 0, 4)
	switch tag {
	case "a":
		if href := safeArticleURL(attributeValue(attributes, "href")); href != "" {
			result = append(result, html.Attribute{Key: "href", Val: href})
		}
		if title := strings.TrimSpace(attributeValue(attributes, "title")); title != "" {
			result = append(result, html.Attribute{Key: "title", Val: title})
		}
		if strings.EqualFold(attributeValue(attributes, "target"), "_blank") {
			result = append(result, html.Attribute{Key: "target", Val: "_blank"}, html.Attribute{Key: "rel", Val: "noopener noreferrer"})
		}
	case "td", "th":
		for _, name := range []string{"colspan", "rowspan"} {
			if value := attributeValue(attributes, name); numericHTMLAttr.MatchString(value) {
				result = append(result, html.Attribute{Key: name, Val: value})
			}
		}
	case "code", "pre":
		if className := strings.TrimSpace(attributeValue(attributes, "class")); strings.HasPrefix(className, "language-") {
			result = append(result, html.Attribute{Key: "class", Val: className})
		}
	}
	return result
}

var allowedArticleTags = map[string]bool{
	"p": true, "br": true, "strong": true, "b": true, "em": true, "i": true, "u": true, "s": true,
	"blockquote": true, "pre": true, "code": true, "ul": true, "ol": true, "li": true,
	"h2": true, "h3": true, "h4": true, "h5": true, "h6": true, "a": true, "img": true,
	"figure": true, "figcaption": true, "table": true, "thead": true, "tbody": true, "tfoot": true,
	"tr": true, "th": true, "td": true, "hr": true, "div": true, "span": true, "sup": true, "sub": true,
}

var blockedArticleTags = map[string]bool{
	"script": true, "style": true, "iframe": true, "object": true, "embed": true, "form": true,
	"input": true, "button": true, "textarea": true, "select": true, "option": true, "link": true, "meta": true,
}

func findHTMLElement(node *html.Node, name string) *html.Node {
	if node.Type == html.ElementNode && strings.EqualFold(node.Data, name) {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findHTMLElement(child, name); found != nil {
			return found
		}
	}
	return nil
}

func attributeValue(attributes []html.Attribute, name string) string {
	for _, attribute := range attributes {
		if strings.EqualFold(attribute.Key, name) {
			return strings.TrimSpace(attribute.Val)
		}
	}
	return ""
}

func safeArticleURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "//") {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	return value
}

func absoluteArticleURL(domain, value string) string {
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if parsed.IsAbs() {
		return parsed.String()
	}
	base, err := url.Parse(strings.TrimRight(domain, "/") + "/")
	if err != nil {
		return ""
	}
	return base.ResolveReference(parsed).String()
}

func genericImageAlt(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return true
	}
	if index := strings.IndexAny(normalized, "?#"); index >= 0 {
		normalized = normalized[:index]
	}
	normalized = strings.TrimRight(normalized, "/\\")
	if index := strings.LastIndexAny(normalized, "/\\"); index >= 0 {
		normalized = normalized[index+1:]
	}
	for _, suffix := range []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp"} {
		if strings.HasSuffix(normalized, suffix) {
			normalized = strings.TrimSuffix(normalized, suffix)
			break
		}
	}
	switch strings.TrimSpace(normalized) {
	case "image", "img", "picture", "pic", "photo", "screenshot", "图片", "配图", "截图", "照片", "图":
		return true
	default:
		return false
	}
}

func plainTextFromHTML(value string) string {
	document, err := html.Parse(strings.NewReader(value))
	if err != nil {
		return value
	}
	var output strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			output.WriteString(node.Data)
			output.WriteByte(' ')
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(document)
	return output.String()
}

func collapseWhitespace(value string) string {
	return strings.Join(strings.FieldsFunc(value, unicode.IsSpace), " ")
}

func truncateRunes(value string, maximum int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= maximum {
		return string(runes)
	}
	return string(runes[:maximum]) + "…"
}

func formatSEOTime(value any) string {
	parsed := parseControllerTime(value)
	if parsed.IsZero() {
		return ""
	}
	return parsed.Format(time.RFC3339)
}
