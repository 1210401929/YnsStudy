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
	"nys-go-api/internal/service"
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
	// 同时注册 HEAD：链接检查工具、站长平台检测死链时常用 HEAD，只注册 GET 会被当成 404。
	// HEAD 请求复用 GET 的处理函数，Go 的 net/http 会自动丢掉响应正文，只返回状态码和响应头。
	page := func(path string, handler gin.HandlerFunc) {
		router.GET(path, handler)
		router.HEAD(path, handler)
	}
	page("/", h.seoHome)
	page("/oneBlog/:id", h.seoArticle)
	page("/sitemap.xml", h.sitemap)
	page("/robots.txt", h.robots)
	page("/rss.xml", h.rss)
	page("/archive", h.seoArchive)
	page("/user/:u", h.seoProfile)
	page("/archive/page/:page", h.seoArchive)
	page("/note/:id", h.seoNote)
	page("/tag/:name", h.seoTag)
	page(seriesListPath, h.seoSeriesList)
	page("/series/:id", h.seoSeries)
	page(noteWallPath, h.seoNoteWall)
	page(service.IndexNowKeyPath, h.indexNowKey)
}

// indexNowKey 返回 IndexNow 密钥文件，搜索引擎用它确认推送请求来自站点所有者。
func (h *Controller) indexNowKey(c *gin.Context) {
	if !h.service.IndexNowEnabled() {
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(strings.TrimSpace(h.service.Config.SEO.IndexNowKey)))
}

// homeTopics 是根地址首屏里的专题入口：热门标签和文章系列
type homeTopics struct {
	Tags   []map[string]any // 名字 NAME、公开文章数 TOTAL
	Series []map[string]any // 编号 ID、名称 TITLE、作者 USERNAME、公开文章数 ARTICLE_COUNT
}

const (
	homeTopicTagLimit    = 20
	homeTopicSeriesLimit = 6
)

// homeTopics 读取首屏的专题入口。只取达到收录标准（公开文章不少于 2 篇）的标签和系列，
// 首页不链接到 noindex 页面；查询失败时这一块不显示，不影响首页其他内容。
func (h *Controller) homeTopics(c *gin.Context) homeTopics {
	ctx := c.Request.Context()
	var topics homeTopics
	if tags, err := h.service.GetIndexableBlogTags(ctx); err == nil {
		if len(tags) > homeTopicTagLimit {
			tags = tags[:homeTopicTagLimit]
		}
		topics.Tags = tags
	}
	if series, err := h.service.PublicSeriesList(ctx, "", "hot", 50); err == nil {
		for _, row := range series {
			if len(topics.Series) == homeTopicSeriesLimit {
				break
			}
			if model.IntValue(row, "ARTICLE_COUNT") >= service.SeriesIndexableMin {
				topics.Series = append(topics.Series, row)
			}
		}
	}
	return topics
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
	owner := h.service.GetSiteOwner(c.Request.Context())
	page := renderSEOHomeHTML(h.readFrontendShell(), articles, h.homeTopics(c), owner, h.service.Config)
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

	links := h.service.GetArticleLinksForSEO(c.Request.Context(), article)
	owner := h.service.GetSiteOwner(c.Request.Context())
	page := renderSEOArticleHTML(h.readFrontendShell(), article, links, owner, h.service.Config)
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
		{Location: domain + "/archive", Frequency: "daily", Priority: "0.9"},
		{Location: domain + "/ynsStudy/Home", Frequency: "daily", Priority: "0.8"},
		{Location: domain + "/ynsStudy/Resources", Frequency: "weekly", Priority: "0.6"},
		{Location: domain + noteWallPath, Frequency: "daily", Priority: "0.6"},
		{Location: domain + "/ynsStudy/FriendLink", Frequency: "monthly", Priority: "0.4"},
		{Location: domain + "/ynsStudy/About", Frequency: "monthly", Priority: "0.5"},
	}
	// 只提交有公开文章的作者主页；查询失败时跳过，不影响文章地址。
	if authors, err := h.service.GetPublicAuthorNumbers(c.Request.Context()); err == nil {
		for _, author := range authors {
			urls = append(urls, sitemapURL{
				Location:   domain + "/user/" + url.PathEscape(model.StringValue(author, "USERNUM")),
				LastModify: formatDate(model.Lookup(author, "LAST_MODIFIED")),
				Frequency:  "weekly",
				Priority:   "0.5",
			})
		}
	}
	for _, row := range rows {
		urls = append(urls, sitemapURL{
			Location:   domain + "/oneBlog/" + url.PathEscape(model.StringValue(row, "GUID")),
			LastModify: formatDate(model.Lookup(row, "LAST_MODIFIED")),
			Frequency:  "weekly",
			Priority:   "0.8",
		})
	}
	// 公开文章数达到收录标准的标签页；查询失败时跳过。
	if tags, err := h.service.GetIndexableBlogTags(c.Request.Context()); err == nil {
		for _, tag := range tags {
			urls = append(urls, sitemapURL{
				Location:   tagPageURL(domain, model.StringValue(tag, "NAME"), 1),
				LastModify: formatDate(model.Lookup(tag, "LAST_MODIFIED")),
				Frequency:  "weekly",
				Priority:   "0.6",
			})
		}
	}
	// 文章系列：有系列时提交“全部系列”页，以及公开文章数达到收录标准的系列页；查询失败时跳过。
	if series, err := h.service.GetIndexableSeries(c.Request.Context()); err == nil && len(series) > 0 {
		urls = append(urls, sitemapURL{Location: domain + seriesListPath, Frequency: "weekly", Priority: "0.6"})
		for _, row := range series {
			urls = append(urls, sitemapURL{
				Location:   seriesURL(domain, model.StringValue(row, "ID")),
				LastModify: formatDate(model.Lookup(row, "LAST_MODIFIED")),
				Frequency:  "weekly",
				Priority:   "0.7",
			})
		}
	}
	// 便签墙的每一条随记；查询失败时跳过，不影响文章地址。
	if notes, err := h.service.GetAllNoteIDs(c.Request.Context()); err == nil {
		for _, note := range notes {
			urls = append(urls, sitemapURL{
				Location:   domain + "/note/" + url.PathEscape(model.StringValue(note, "GUID")),
				LastModify: formatDate(model.Lookup(note, "LAST_MODIFIED")),
				Frequency:  "monthly",
				Priority:   "0.5",
			})
		}
	}
	c.Header("Cache-Control", "public, max-age=300")
	writeXML(c, sitemapDocument{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: urls})
}

func (h *Controller) robots(c *gin.Context) {
	domain := strings.TrimRight(h.service.Config.External.DomainName, "/")
	body := "User-agent: *\n" +
		"Allow: /\n" +
		// 不屏蔽 /api/：搜索引擎渲染页面时要靠它加载数据；接口返回的 JSON 由 Nginx 加 X-Robots-Tag: noindex，不会被当作网页收录
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
<meta name="robots" content="noindex,follow"><title>%s - %s</title>`+siteIconLinks+`
<style>body{margin:0;background-color:#efe8da;background-image:radial-gradient(rgba(120,104,80,.18) 1px,transparent 1px);background-size:22px 22px;color:#2b2a27;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif}.box{max-width:680px;margin:12vh auto;padding:40px 28px;background:#fffdf8;border:1px solid #e6dfd1;box-shadow:0 1px 2px rgba(60,50,30,.06),0 10px 24px -16px rgba(60,50,30,.35);text-align:center}h1{font-family:"LXGW WenKai Screen","Kaiti SC","STKaiti","KaiTi","楷体",serif;font-weight:normal;font-size:30px}p{color:#7d776c;line-height:1.8}a{display:inline-block;margin-top:14px;color:#2f5d8a;text-decoration:none}</style></head>
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
	Keywords         string                 `json:"keywords,omitempty"`
	IsPartOf         any                    `json:"isPartOf,omitempty"`
}

type structuredMainEntity struct {
	Type string `json:"@type"`
	ID   string `json:"@id"`
}

type structuredPerson struct {
	Type          string   `json:"@type"`
	Name          string   `json:"name"`
	AlternateName []string `json:"alternateName,omitempty"`
	URL           string   `json:"url,omitempty"`
	Image         string   `json:"image,omitempty"`
	Description   string   `json:"description,omitempty"`
	SameAs        []string `json:"sameAs,omitempty"`
}

// ownerPerson 生成站长的 Person 结构化数据：名字、别名和外部主页让搜索引擎把
// “余南松 / YuNanSong / GitHub 账号 / 本站”识别为同一个人。
func ownerPerson(owner service.SiteOwner, domain string) structuredPerson {
	person := structuredPerson{Type: "Person", Name: owner.Name, AlternateName: owner.AlternateNames, SameAs: owner.SameAs}
	if owner.ProfilePath != "" {
		person.URL = domain + owner.ProfilePath
	}
	return person
}

// ownerDisplay 返回“余南松（YuNanSong）”形式的名字，用于标题以外的介绍文字。
func ownerDisplay(owner service.SiteOwner) string {
	if len(owner.AlternateNames) > 0 && owner.AlternateNames[0] != "" {
		return owner.Name + "（" + owner.AlternateNames[0] + "）"
	}
	return owner.Name
}

type structuredOrganization struct {
	Type string `json:"@type"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Logo string `json:"logo,omitempty"`
}

// renderHomeTopicsHTML 输出首屏的“文章系列”和“热门标签”两块，没有数据的那一块不输出
func renderHomeTopicsHTML(topics homeTopics, domain string) string {
	var output strings.Builder
	if len(topics.Series) > 0 {
		output.WriteString(`<section class="seo-home-latest" aria-labelledby="seo-series-title"><h2 id="seo-series-title">文章系列</h2><ul>`)
		for _, row := range topics.Series {
			meta := fmt.Sprintf("%d 篇", model.IntValue(row, "ARTICLE_COUNT"))
			if author := model.StringValue(row, "USERNAME"); author != "" {
				meta = authorLinkHTML(author, model.StringValue(row, "USERNUM"), domain) + " · " + meta
			}
			fmt.Fprintf(&output, `<li><a href="%s">%s</a><span class="seo-home-meta">%s</span></li>`,
				htmlstd.EscapeString(seriesURL(domain, model.StringValue(row, "ID"))), htmlstd.EscapeString(model.StringValue(row, "TITLE")), meta)
		}
		fmt.Fprintf(&output, `</ul><p class="seo-home-more"><a href="%s%s">查看全部系列 →</a></p></section>`, htmlstd.EscapeString(domain), seriesListPath)
	}
	if len(topics.Tags) > 0 {
		output.WriteString(`<section class="seo-home-latest" aria-labelledby="seo-tags-title"><h2 id="seo-tags-title">热门标签</h2><p class="seo-home-tags">`)
		for _, row := range topics.Tags {
			name := model.StringValue(row, "NAME")
			if name == "" {
				continue
			}
			fmt.Fprintf(&output, `<a href="%s" rel="tag">#%s<span>%d</span></a>`,
				htmlstd.EscapeString(tagPageURL(domain, name, 1)), htmlstd.EscapeString(name), model.IntValue(row, "TOTAL"))
		}
		output.WriteString(`</p></section>`)
	}
	return output.String()
}

func renderSEOHomeHTML(shell string, articles []map[string]any, topics homeTopics, owner service.SiteOwner, cfg *config.Config) string {
	// 首页使用服务端生成的完整站点 meta，移除 Vue 外壳中的默认副本，避免重复 description/OG。
	shell = defaultSEOElement.ReplaceAllString(shell, "\n")
	shell = javascriptNotice.ReplaceAllString(shell, "\n")
	domain := strings.TrimRight(cfg.External.DomainName, "/")
	canonical := domain + "/"
	// 首页标题和描述带上站长名字，让搜索“余南松”时也能找到本站；与前端 utils/seo.js 的 DEFAULT_TITLE 一致。
	title := homeTitle(cfg.SEO.SiteName, owner.Name)
	description := homeDescription(cfg.SEO.SiteName, ownerDisplay(owner))

	person := ownerPerson(owner, domain)
	websiteJSON, _ := json.Marshal(map[string]any{
		"@context":    "https://schema.org",
		"@type":       "WebSite",
		"name":        cfg.SEO.SiteName,
		"url":         canonical,
		"description": description,
		"inLanguage":  "zh-CN",
		"author":      person,
		"publisher":   person,
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
<style id="yns-seo-home-first-paint">.seo-home-fallback{min-height:100vh;padding:36px 16px;background-color:#efe8da;background-image:radial-gradient(rgba(120,104,80,.18) 1px,transparent 1px);background-size:22px 22px;color:#2b2a27;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif}.seo-home-intro,.seo-home-latest{max-width:1000px;margin:0 auto;background:#fffdf8;border:1px solid #e6dfd1;box-shadow:0 1px 2px rgba(60,50,30,.06),0 10px 24px -16px rgba(60,50,30,.35);padding:28px 32px}.seo-home-intro h1{margin:0 0 12px;font-family:"LXGW WenKai Screen","Kaiti SC","STKaiti","KaiTi","楷体",serif;font-weight:normal;font-size:36px}.seo-home-intro p{margin:0;color:#4d4943;line-height:1.8}.seo-home-intro .seo-home-owner{margin-top:8px;font-size:14px;color:#7d776c}.seo-home-owner a{color:#2f5d8a;text-decoration:none}.seo-home-nav{display:flex;flex-wrap:wrap;gap:18px;margin-top:18px}.seo-home-nav a,.seo-home-latest a{color:#2f5d8a;text-decoration:none}.seo-home-latest{margin-top:22px}.seo-home-latest h2{margin:0 0 12px;font-family:"LXGW WenKai Screen","Kaiti SC","STKaiti","KaiTi","楷体",serif;font-weight:normal;font-size:24px}.seo-home-latest ul{list-style:none;margin:0;padding:0}.seo-home-latest li{display:flex;justify-content:space-between;gap:20px;padding:12px 0;border-top:1px dashed #e6dfd1}.seo-home-meta{flex:none;color:#7d776c;font-size:13px}.seo-home-latest .seo-home-meta a{color:inherit}.seo-home-more{margin:16px 0 0;text-align:right}.seo-home-tags{display:flex;flex-wrap:wrap;gap:10px 18px;margin:0}.seo-home-tags span{margin-left:3px;color:#7d776c;font-size:12px}@media(max-width:640px){.seo-home-fallback{padding:14px 10px}.seo-home-intro,.seo-home-latest{padding:22px 16px}.seo-home-latest li{display:block}.seo-home-meta{display:block;margin-top:6px}}</style>
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
		author := model.PublicDisplayName(model.StringValue(article, "USERNAME"))
		dateText := htmlstd.EscapeString(formatDate(model.Lookup(article, "CREATE_TIME")))
		metaHTML := dateText
		if author != "" {
			metaHTML = authorLinkHTML(author, model.StringValue(article, "USERNUM"), domain) + " · " + dateText
		}
		fmt.Fprintf(&latest, `<li><a href="%s">%s</a><span class="seo-home-meta">%s</span></li>`,
			htmlstd.EscapeString(articleURL), htmlstd.EscapeString(articleTitle), metaHTML)
	}
	if latest.Len() == 0 {
		latest.WriteString(`<li>暂时还没有公开文章</li>`)
	}

	body := fmt.Sprintf(`<main class="seo-home-fallback">
<section class="seo-home-intro" aria-labelledby="seo-home-title">
<h1 id="seo-home-title">%s</h1>
<p>%s</p>
<p class="seo-home-owner">站长：%s</p>
<nav class="seo-home-nav" aria-label="主要栏目"><a href="%s/ynsStudy/Home">最新文章</a><a href="%s/ynsStudy/Resources">资源</a><a href="%s/ynsStudy/Notes">便签墙</a><a href="%s/ynsStudy/FriendLink">友链</a><a href="%s/ynsStudy/About">关于</a></nav>
</section>
<section class="seo-home-latest" aria-labelledby="seo-latest-title"><h2 id="seo-latest-title">最新文章</h2><ul>%s</ul><p class="seo-home-more"><a href="%s/archive">查看全部文章 →</a></p></section>
%s
</main>`, htmlstd.EscapeString(cfg.SEO.SiteName), escapedDescription, ownerLinkHTML(owner, domain),
		htmlstd.EscapeString(domain), htmlstd.EscapeString(domain), htmlstd.EscapeString(domain),
		htmlstd.EscapeString(domain), htmlstd.EscapeString(domain), latest.String(), htmlstd.EscapeString(domain),
		renderHomeTopicsHTML(topics, domain))

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

func renderSEOArticleHTML(shell string, article map[string]any, links service.ArticleLinks, owner service.SiteOwner, cfg *config.Config) string {
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
	author := model.PublicDisplayName(model.StringValue(article, "USERNAME"))
	if author == "" {
		author = cfg.SEO.SiteName
	}

	// 作者主页地址；旧数据缺少用户编号时只显示名字。
	authorURL := ""
	if userNumber := strings.TrimSpace(model.StringValue(article, "USERNUM")); userNumber != "" {
		authorURL = domain + "/user/" + url.PathEscape(userNumber)
	}
	authorHTML := `<span itemprop="author">` + htmlstd.EscapeString(author) + `</span>`
	if authorURL != "" {
		authorHTML = `<a href="` + htmlstd.EscapeString(authorURL) + `" itemprop="author">` + htmlstd.EscapeString(author) + `</a>`
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
		Author:    structuredPerson{Type: "Person", Name: author, URL: authorURL},
		Publisher: structuredOrganization{Type: "Organization", Name: cfg.SEO.SiteName, URL: domain, Logo: domain + "/icon-512.png"},
	}
	// 站长本人的文章，作者信息补充别名和外部主页
	if owner.Code != "" && model.StringValue(article, "USERCODE") == owner.Code {
		person := ownerPerson(owner, domain)
		if authorURL != "" {
			person.URL = authorURL
		}
		structured.Author = person
	}
	if firstImage != "" {
		structured.Image = []string{firstImage}
	}
	tags, _ := article["TAGS"].([]string)
	structured.Keywords = strings.Join(tags, ",")
	seriesNav, _ := article["SERIES"].(*service.SeriesNav)
	if seriesNav != nil && seriesNav.Index > 0 {
		structured.IsPartOf = map[string]any{"@type": "CreativeWorkSeries", "name": seriesNav.Title, "url": seriesURL(domain, seriesNav.ID)}
	}
	structuredJSON, _ := json.Marshal(structured)

	escapedTitle := htmlstd.EscapeString(title)
	escapedFullTitle := htmlstd.EscapeString(title + " - " + cfg.SEO.SiteName)
	escapedDescription := htmlstd.EscapeString(description)
	escapedCanonical := htmlstd.EscapeString(canonical)
	escapedImage := htmlstd.EscapeString(firstImage)
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
<meta property="article:modified_time" content="%s">%s
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="%s">
<meta name="twitter:description" content="%s">
<meta name="twitter:image" content="%s">
<script type="application/ld+json">%s</script>
<script type="application/ld+json">%s</script>
<style id="yns-seo-first-paint">.seo-article-fallback{min-height:100vh;padding:32px 16px;background-color:#efe8da;background-image:radial-gradient(rgba(120,104,80,.18) 1px,transparent 1px);background-size:22px 22px;color:#2b2a27;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif}.seo-article-fallback article{max-width:1000px;margin:0 auto;background:#fffdf8;border:1px solid #e6dfd1;box-shadow:0 1px 2px rgba(60,50,30,.06),0 10px 24px -16px rgba(60,50,30,.35);padding:36px 40px}.seo-breadcrumb{margin-bottom:22px;font-size:13px;color:#7d776c}.seo-breadcrumb a,.seo-article-meta a,.seo-article-links a{color:#2f5d8a;text-decoration:none}.seo-article-fallback h1{font-family:"LXGW WenKai Screen","Kaiti SC","STKaiti","KaiTi","楷体",serif;font-weight:normal;font-size:32px;line-height:1.35;margin:0 0 12px}.seo-article-meta{color:#7d776c;font-size:14px;margin-bottom:28px}.seo-article-content{font-size:16px;line-height:1.85;overflow-wrap:anywhere}.seo-article-content img{display:block;max-width:100%%;height:auto;margin:20px auto}.seo-article-content pre{overflow:auto;padding:16px;background:#f6f1e4;border-radius:4px}.seo-article-content table{display:block;max-width:100%%;overflow:auto;border-collapse:collapse}.seo-article-content td,.seo-article-content th{padding:8px;border:1px solid #e6dfd1}.seo-article-links{margin-top:32px;padding-top:20px;border-top:1px dashed #d6ccb8;font-size:15px;line-height:1.8}.seo-article-pager{display:flex;flex-wrap:wrap;justify-content:space-between;gap:12px}.seo-article-links h2{font-family:"LXGW WenKai Screen","Kaiti SC","STKaiti","KaiTi","楷体",serif;font-weight:normal;font-size:20px;margin:20px 0 8px}.seo-article-links ul{margin:0;padding-left:20px}.seo-article-series{margin:0 0 24px;padding:10px 14px;background:#fbf0b4;font-size:14px;line-height:1.8}.seo-article-series a{color:#2f5d8a;text-decoration:none}.seo-article-tags{margin-top:24px;color:#7d776c;font-size:14px}.seo-article-tags a{color:#2f5d8a;text-decoration:none;margin-right:6px}@media(max-width:640px){.seo-article-fallback{padding:12px 10px}.seo-article-fallback article{padding:22px 16px}.seo-article-fallback h1{font-size:25px}}</style>
`, escapedDescription, escapedCanonical, htmlstd.EscapeString(cfg.SEO.SiteName), escapedTitle,
		escapedDescription, escapedCanonical, escapedImage, escapedPublished, htmlstd.EscapeString(modified), tagMetaHTML(tags),
		escapedTitle, escapedDescription, escapedImage, structuredJSON,
		breadcrumbJSON([]breadcrumbItem{{Name: cfg.SEO.SiteName, URL: domain + "/"}, {Name: "文章归档", URL: domain + "/archive"}, {Name: title}}))

	body := fmt.Sprintf(`<main class="seo-article-fallback">
<article itemscope itemtype="https://schema.org/BlogPosting">
<nav class="seo-breadcrumb" aria-label="面包屑"><a href="%s/">YnsStudy</a> / <a href="%s/archive">文章归档</a></nav>
<header><h1 itemprop="headline">%s</h1><p class="seo-article-meta">%s · <time itemprop="datePublished" datetime="%s">%s</time></p></header>%s
<div class="seo-article-content" itemprop="articleBody">%s</div>
%s%s
</article></main>`, htmlstd.EscapeString(domain), htmlstd.EscapeString(domain), escapedTitle, authorHTML,
		escapedPublished, htmlstd.EscapeString(formatDate(model.Lookup(article, "CREATE_TIME"))), articleSeriesHTML(seriesNav, domain), safeContent,
		tagLinksForArticleHTML(tags, domain), renderArticleLinks(links, author, domain))

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

// renderArticleLinks 输出文章底部的上一篇/下一篇、同作者文章和归档入口，
// 让抓取器从任意文章都能继续发现站内其他文章。
func renderArticleLinks(links service.ArticleLinks, author, domain string) string {
	articleURL := func(row map[string]any) string {
		return domain + "/oneBlog/" + url.PathEscape(strings.TrimSpace(model.StringValue(row, "GUID")))
	}
	articleTitle := func(row map[string]any) string {
		if title := strings.TrimSpace(model.StringValue(row, "BLOG_TITLE")); title != "" {
			return title
		}
		return "未命名文章"
	}

	var output strings.Builder
	output.WriteString(`<footer class="seo-article-links">`)
	shown := make(map[string]bool)
	if links.Previous != nil || links.Next != nil {
		output.WriteString(`<nav class="seo-article-pager" aria-label="上一篇和下一篇">`)
		if links.Previous != nil {
			shown[model.StringValue(links.Previous, "GUID")] = true
			fmt.Fprintf(&output, `<a href="%s">← 上一篇：%s</a>`, htmlstd.EscapeString(articleURL(links.Previous)), htmlstd.EscapeString(articleTitle(links.Previous)))
		} else {
			output.WriteString(`<span></span>`)
		}
		if links.Next != nil {
			shown[model.StringValue(links.Next, "GUID")] = true
			fmt.Fprintf(&output, `<a href="%s">下一篇：%s →</a>`, htmlstd.EscapeString(articleURL(links.Next)), htmlstd.EscapeString(articleTitle(links.Next)))
		}
		output.WriteString(`</nav>`)
	}

	var related strings.Builder
	for _, row := range links.Related {
		guid := model.StringValue(row, "GUID")
		if guid == "" || shown[guid] {
			continue
		}
		fmt.Fprintf(&related, `<li><a href="%s">%s</a></li>`, htmlstd.EscapeString(articleURL(row)), htmlstd.EscapeString(articleTitle(row)))
	}
	if related.Len() > 0 {
		fmt.Fprintf(&output, `<h2>%s 的其他文章</h2><ul>%s</ul>`, htmlstd.EscapeString(author), related.String())
	}
	fmt.Fprintf(&output, `<p><a href="%s/archive">浏览全部文章 →</a></p></footer>`, htmlstd.EscapeString(domain))
	return output.String()
}

// siteIconLinks 与 Vue index.html 中的图标声明保持一致，供不经过 Vue 外壳的页面使用。
const siteIconLinks = `<link rel="icon" href="/favicon.ico" sizes="48x48">
<link rel="icon" type="image/png" sizes="96x96" href="/icon-96.png">
<link rel="icon" type="image/png" sizes="192x192" href="/icon-192.png">
<link rel="apple-touch-icon" href="/apple-touch-icon.png">`

// homeTitle / homeDescription 是首页的标题和描述，前端 utils/seo.js 使用同样的文字。
func homeTitle(siteName, ownerName string) string {
	return siteName + " - " + ownerName + "的技术分享与个人博客"
}

func homeDescription(siteName, ownerDisplayName string) string {
	return siteName + " 是" + ownerDisplayName + "的个人博客与技术社区，记录编程学习、技术实践与生活思考。"
}

// authorLinkHTML 输出作者名；有用户编号时链接到作者主页，帮助搜索引擎发现作者主页。
func authorLinkHTML(name, userNumber, domain string) string {
	escaped := htmlstd.EscapeString(name)
	userNumber = strings.TrimSpace(userNumber)
	if userNumber == "" {
		return escaped
	}
	return `<a href="` + htmlstd.EscapeString(domain+"/user/"+url.PathEscape(userNumber)) + `">` + escaped + `</a>`
}

// ownerLinkHTML 输出指向站长主页的链接；查不到主页时只输出名字。
func ownerLinkHTML(owner service.SiteOwner, domain string) string {
	name := htmlstd.EscapeString(ownerDisplay(owner))
	if owner.ProfilePath == "" {
		return name
	}
	return `<a href="` + htmlstd.EscapeString(domain+owner.ProfilePath) + `">` + name + `</a>`
}

func defaultFrontendShell(siteName string) string {
	return `<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>` +
		htmlstd.EscapeString(siteName) + `</title>` + siteIconLinks + `</head><body><div id="app"></div></body></html>`
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
