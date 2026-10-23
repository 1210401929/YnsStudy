package controller

// 文章系列的服务端首屏：/series（全部系列）和 /series/编号（一个系列的目录）。
// 输出真实内容 + Vue 入口脚本，浏览器加载后由 Vue 接管。只显示公开文章；
// 公开文章少于 service.SeriesIndexableMin 篇的系列页输出 noindex。

import (
	"encoding/json"
	"fmt"
	htmlstd "html"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/config"
	"nys-go-api/internal/model"
	"nys-go-api/internal/service"
)

const seriesListPath = "/series"

const seriesFirstPaintStyle = `<style id="yns-seo-series-first-paint">.seo-series-fallback{min-height:100vh;padding:32px 16px;background-color:#efe8da;background-image:radial-gradient(rgba(120,104,80,.18) 1px,transparent 1px);background-size:22px 22px;color:#2b2a27;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif}.seo-series-fallback section{max-width:960px;margin:0 auto;background:#fffdf8;border:1px solid #e6dfd1;box-shadow:0 1px 2px rgba(60,50,30,.06),0 10px 24px -16px rgba(60,50,30,.35);padding:30px 32px}.seo-series-fallback a{color:#2f5d8a;text-decoration:none}.seo-series-fallback h1{font-family:"LXGW WenKai Screen","Kaiti SC","STKaiti","KaiTi","楷体",serif;font-weight:normal;font-size:30px;margin:0 0 6px}.seo-series-meta,.seo-breadcrumb{color:#7d776c;font-size:14px;margin-bottom:16px}.seo-series-desc{line-height:1.8;color:#4d4943}.seo-series-list{margin:0;padding-left:28px}.seo-series-list li{padding:12px 0;border-top:1px dashed #e6dfd1}.seo-series-list h2{margin:0 0 6px;font-size:18px;font-weight:600}.seo-series-list p{margin:0;color:#4d4943;font-size:14px;line-height:1.75;overflow-wrap:anywhere}.seo-series-start{display:inline-block;margin:6px 0 18px}@media(max-width:640px){.seo-series-fallback{padding:12px 10px}.seo-series-fallback section{padding:22px 16px}}</style>`

// canonicalHeadHTML 返回规范网址和 og:url 两个标签。页面不收录（noindex）时都不输出，
// 避免“不收录”和“规范网址”两个信号互相矛盾，与前端 utils/seo.js 的 useSeo 规则一致。
func canonicalHeadHTML(canonical string, indexable bool) (link, ogURL string) {
	if !indexable {
		return "", ""
	}
	escaped := htmlstd.EscapeString(canonical)
	return `<link rel="canonical" href="` + escaped + `">`, `<meta property="og:url" content="` + escaped + `">`
}

func seriesURL(domain, id string) string {
	return domain + service.SeriesPath(url.PathEscape(id))
}

// seoSeries 输出一个系列的目录页
func (h *Controller) seoSeries(c *gin.Context) {
	raw := strings.TrimSpace(c.Param("id"))
	if id, err := strconv.ParseInt(raw, 10, 64); err != nil || id <= 0 || strconv.FormatInt(id, 10) != raw {
		h.renderSEOStatusPage(c, http.StatusNotFound, "系列不存在", "系列地址有误。")
		return
	}
	series, articles, found, err := h.service.SeriesDetail(c.Request.Context(), raw, false)
	if err != nil {
		h.renderSEOStatusPage(c, http.StatusInternalServerError, "系列暂时无法访问", "服务器读取系列时出现异常，请稍后重试。")
		return
	}
	if !found || len(articles) == 0 {
		h.renderSEOStatusPage(c, http.StatusNotFound, "系列不存在", "这个系列可能已删除，或者还没有公开文章。")
		return
	}
	indexable := len(articles) >= service.SeriesIndexableMin
	page := renderSEOSeriesHTML(h.readFrontendShell(), series, articles, indexable, h.service.Config)
	h.writeSEOPageIndexable(c, page, indexable)
}

// writeSEOPageIndexable 和 writeSEOPage 相同，indexable 为 false 时响应头声明 noindex
func (h *Controller) writeSEOPageIndexable(c *gin.Context, page string, indexable bool) {
	if indexable {
		h.writeSEOPage(c, page)
		return
	}
	cacheSeconds := h.service.Config.SEO.ResponseCacheSeconds
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=0, s-maxage=%d, stale-while-revalidate=60", cacheSeconds))
	c.Header("X-Robots-Tag", "noindex, follow")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}

func renderSEOSeriesHTML(shell string, series map[string]any, articles []map[string]any, indexable bool, cfg *config.Config) string {
	domain := strings.TrimRight(cfg.External.DomainName, "/")
	seriesID := model.StringValue(series, "ID")
	canonical := seriesURL(domain, seriesID)
	name := strings.TrimSpace(model.StringValue(series, "TITLE"))
	author := model.StringValue(series, "USERNAME")
	authorURL := ""
	if number := strings.TrimSpace(model.StringValue(series, "USERNUM")); number != "" {
		authorURL = domain + "/user/" + url.PathEscape(number)
	}
	description := strings.TrimSpace(model.StringValue(series, "DESCRIPTION"))
	if description == "" {
		titles := make([]string, 0, 3)
		for _, article := range articles {
			if len(titles) == 3 {
				break
			}
			titles = append(titles, "《"+strings.TrimSpace(model.StringValue(article, "BLOG_TITLE"))+"》")
		}
		description = fmt.Sprintf("%s的系列文章《%s》，共 %d 篇：%s", author, name, len(articles), strings.Join(titles, "、"))
	}
	description = truncateRunes(collapseWhitespace(description), 180)

	parts := make([]map[string]any, 0, len(articles))
	for index, article := range articles {
		parts = append(parts, map[string]any{
			"@type":    "BlogPosting",
			"headline": model.StringValue(article, "BLOG_TITLE"),
			"url":      domain + "/oneBlog/" + url.PathEscape(model.StringValue(article, "GUID")),
			"position": index + 1,
		})
	}
	structured := map[string]any{
		"@context":    "https://schema.org",
		"@type":       "CreativeWorkSeries",
		"name":        name,
		"description": description,
		"url":         canonical,
		"inLanguage":  "zh-CN",
		"hasPart":     parts,
	}
	if author != "" {
		structured["author"] = structuredPerson{Type: "Person", Name: author, URL: authorURL}
	}
	structuredJSON, _ := json.Marshal(structured)

	robots := "index,follow,max-image-preview:large,max-snippet:-1"
	if !indexable {
		robots = "noindex,follow"
	}
	title := name + "（系列）"
	escapedDescription := htmlstd.EscapeString(description)
	canonicalLink, ogURL := canonicalHeadHTML(canonical, indexable)
	head := fmt.Sprintf(`
<meta name="description" content="%s">
<meta name="robots" content="%s">
%s
<meta property="og:type" content="website">
<meta property="og:site_name" content="%s">
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
%s
<meta property="og:image" content="%s">
<script type="application/ld+json">%s</script>
<script type="application/ld+json">%s</script>
%s
`, escapedDescription, robots, canonicalLink, htmlstd.EscapeString(cfg.SEO.SiteName), htmlstd.EscapeString(title),
		escapedDescription, ogURL, htmlstd.EscapeString(cfg.SEO.DefaultImage), structuredJSON,
		breadcrumbJSON([]breadcrumbItem{{Name: cfg.SEO.SiteName, URL: domain + "/"}, {Name: "文章系列", URL: domain + seriesListPath}, {Name: name}}),
		seriesFirstPaintStyle)

	var list strings.Builder
	for _, article := range articles {
		articleTitle := strings.TrimSpace(model.StringValue(article, "BLOG_TITLE"))
		if articleTitle == "" {
			articleTitle = "未命名文章"
		}
		fmt.Fprintf(&list, `<li><h2><a href="%s/oneBlog/%s">%s</a></h2>`, htmlstd.EscapeString(domain),
			htmlstd.EscapeString(url.PathEscape(model.StringValue(article, "GUID"))), htmlstd.EscapeString(articleTitle))
		if excerpt := truncateRunes(collapseWhitespace(plainTextFromHTML(model.StringValue(article, "EXCERPT_HTML"))), 120); excerpt != "" {
			fmt.Fprintf(&list, `<p>%s</p>`, htmlstd.EscapeString(excerpt))
		}
		list.WriteString(`</li>`)
	}

	authorHTML := ""
	if author != "" {
		authorHTML = authorLinkHTML(author, model.StringValue(series, "USERNUM"), domain) + " · "
	}
	descriptionHTML := ""
	if text := strings.TrimSpace(model.StringValue(series, "DESCRIPTION")); text != "" {
		descriptionHTML = `<p class="seo-series-desc">` + htmlstd.EscapeString(text) + `</p>`
	}
	firstURL := domain + "/oneBlog/" + url.PathEscape(model.StringValue(articles[0], "GUID"))
	body := fmt.Sprintf(`<main class="seo-series-fallback"><section>
<nav class="seo-breadcrumb" aria-label="面包屑"><a href="%s/">YnsStudy</a> / <a href="%s%s">文章系列</a></nav>
<h1>%s</h1>
<p class="seo-series-meta">%s共 %d 篇</p>
%s
<a class="seo-series-start" href="%s">从第一篇开始读 →</a>
<ol class="seo-series-list">%s</ol>
</section></main>`, htmlstd.EscapeString(domain), htmlstd.EscapeString(domain), seriesListPath, htmlstd.EscapeString(name),
		authorHTML, len(articles), descriptionHTML, htmlstd.EscapeString(firstURL), list.String())
	return injectSEOShell(shell, title+" - "+cfg.SEO.SiteName, head, body)
}

// seoSeriesList 输出全部系列
func (h *Controller) seoSeriesList(c *gin.Context) {
	rows, err := h.service.PublicSeriesList(c.Request.Context(), "", "latest", 200)
	if err != nil {
		h.renderSEOStatusPage(c, http.StatusInternalServerError, "系列暂时无法访问", "服务器读取系列时出现异常，请稍后重试。")
		return
	}
	h.writeSEOPageIndexable(c, renderSEOSeriesListHTML(h.readFrontendShell(), rows, h.service.Config), len(rows) > 0)
}

func renderSEOSeriesListHTML(shell string, rows []map[string]any, cfg *config.Config) string {
	domain := strings.TrimRight(cfg.External.DomainName, "/")
	canonical := domain + seriesListPath
	title := "文章系列"
	description := fmt.Sprintf("%s 的文章系列：按顺序编排的专题文章合集，共 %d 个系列。", cfg.SEO.SiteName, len(rows))
	robots := "index,follow,max-image-preview:large,max-snippet:-1"
	if len(rows) == 0 {
		robots = "noindex,follow"
	}
	collectionJSON, _ := json.Marshal(map[string]any{
		"@context": "https://schema.org", "@type": "CollectionPage", "name": title, "url": canonical, "description": description,
		"isPartOf": map[string]any{"@type": "WebSite", "name": cfg.SEO.SiteName, "url": domain + "/"},
	})
	escapedDescription := htmlstd.EscapeString(description)
	canonicalLink, ogURL := canonicalHeadHTML(canonical, len(rows) > 0)
	head := fmt.Sprintf(`
<meta name="description" content="%s">
<meta name="robots" content="%s">
%s
<meta property="og:type" content="website">
<meta property="og:site_name" content="%s">
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
%s
<script type="application/ld+json">%s</script>
<script type="application/ld+json">%s</script>
%s
`, escapedDescription, robots, canonicalLink, htmlstd.EscapeString(cfg.SEO.SiteName), htmlstd.EscapeString(title),
		escapedDescription, ogURL, collectionJSON,
		breadcrumbJSON([]breadcrumbItem{{Name: cfg.SEO.SiteName, URL: domain + "/"}, {Name: "文章系列"}}), seriesFirstPaintStyle)

	var list strings.Builder
	for _, row := range rows {
		fmt.Fprintf(&list, `<li><h2><a href="%s">%s</a></h2><p>`, htmlstd.EscapeString(seriesURL(domain, model.StringValue(row, "ID"))),
			htmlstd.EscapeString(model.StringValue(row, "TITLE")))
		if author := model.StringValue(row, "USERNAME"); author != "" {
			list.WriteString(authorLinkHTML(author, model.StringValue(row, "USERNUM"), domain) + " · ")
		}
		fmt.Fprintf(&list, `共 %d 篇`, model.IntValue(row, "ARTICLE_COUNT"))
		if text := strings.TrimSpace(model.StringValue(row, "DESCRIPTION")); text != "" {
			list.WriteString(` · ` + htmlstd.EscapeString(truncateRunes(text, 80)))
		}
		list.WriteString(`</p></li>`)
	}
	if list.Len() == 0 {
		list.WriteString(`<li>还没有系列</li>`)
	}
	body := fmt.Sprintf(`<main class="seo-series-fallback"><section>
<nav class="seo-breadcrumb" aria-label="面包屑"><a href="%s/">YnsStudy</a> / 文章系列</nav>
<h1>文章系列</h1>
<p class="seo-series-meta">%s</p>
<ul class="seo-series-list">%s</ul>
</section></main>`, htmlstd.EscapeString(domain), escapedDescription, list.String())
	return injectSEOShell(shell, title+" - "+cfg.SEO.SiteName, head, body)
}

// articleSeriesHTML 输出文章页的系列信息：属于哪个系列、第几篇，以及系列里的上一篇和下一篇
func articleSeriesHTML(nav *service.SeriesNav, domain string) string {
	if nav == nil || nav.Index == 0 {
		return ""
	}
	link := func(item map[string]any, label string) string {
		if item == nil {
			return ""
		}
		return fmt.Sprintf(` · <a href="%s/oneBlog/%s">%s：%s</a>`, htmlstd.EscapeString(domain),
			htmlstd.EscapeString(url.PathEscape(model.StringValue(item, "GUID"))), label, htmlstd.EscapeString(model.StringValue(item, "BLOG_TITLE")))
	}
	return fmt.Sprintf(`<aside class="seo-article-series">本文属于系列《<a href="%s">%s</a>》第 %d / %d 篇%s%s</aside>`,
		htmlstd.EscapeString(seriesURL(domain, nav.ID)), htmlstd.EscapeString(nav.Title), nav.Index, nav.Total,
		link(nav.Previous, "上一篇"), link(nav.Next, "下一篇"))
}
