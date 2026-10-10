package controller

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

// seoArchive 输出完整的服务端文章归档页（不依赖 Vue），让搜索引擎能顺着分页发现所有旧文章。
// /archive 是第 1 页，/archive/page/N 是后续页；/archive/page/1 永久跳转到 /archive，避免重复地址。
func (h *Controller) seoArchive(c *gin.Context) {
	page := 1
	if raw := c.Param("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || strconv.Itoa(parsed) != raw {
			h.renderSEOStatusPage(c, http.StatusNotFound, "页面不存在", "归档页码有误。")
			return
		}
		if parsed == 1 {
			c.Redirect(http.StatusMovedPermanently, "/archive")
			return
		}
		page = parsed
	}

	rows, total, err := h.service.GetBlogArchivePage(c.Request.Context(), page)
	if err != nil {
		h.renderSEOStatusPage(c, http.StatusInternalServerError, "归档暂时无法访问", "服务器读取文章列表时出现异常，请稍后重试。")
		return
	}
	totalPages := archiveTotalPages(total)
	if page > totalPages {
		h.renderSEOStatusPage(c, http.StatusNotFound, "页面不存在", "这一页已经没有文章了。")
		return
	}

	cacheSeconds := h.service.Config.SEO.ResponseCacheSeconds
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=0, s-maxage=%d, stale-while-revalidate=60", cacheSeconds))
	c.Header("X-Robots-Tag", "index, follow, max-image-preview:large, max-snippet:-1")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(renderSEOArchiveHTML(rows, page, totalPages, total, h.service.Config)))
}

func archiveTotalPages(total int) int {
	if total <= 0 {
		return 1
	}
	return (total + service.ArchivePageSize - 1) / service.ArchivePageSize
}

func archivePageURL(domain string, page int) string {
	if page <= 1 {
		return domain + "/archive"
	}
	return fmt.Sprintf("%s/archive/page/%d", domain, page)
}

type breadcrumbItem struct {
	Name string
	URL  string
}

// breadcrumbJSON 生成 schema.org BreadcrumbList，最后一项可以不带 URL（表示当前页）。
func breadcrumbJSON(items []breadcrumbItem) []byte {
	elements := make([]map[string]any, 0, len(items))
	for index, item := range items {
		element := map[string]any{"@type": "ListItem", "position": index + 1, "name": item.Name}
		if item.URL != "" {
			element["item"] = item.URL
		}
		elements = append(elements, element)
	}
	data, _ := json.Marshal(map[string]any{
		"@context":        "https://schema.org",
		"@type":           "BreadcrumbList",
		"itemListElement": elements,
	})
	return data
}

func renderSEOArchiveHTML(rows []map[string]any, page, totalPages, total int, cfg *config.Config) string {
	domain := strings.TrimRight(cfg.External.DomainName, "/")
	siteName := cfg.SEO.SiteName
	canonical := archivePageURL(domain, page)

	title := "文章归档 - " + siteName
	heading := "文章归档"
	description := fmt.Sprintf("%s 全部公开文章，按发布时间倒序排列，共 %d 篇。", siteName, total)
	if page > 1 {
		title = fmt.Sprintf("文章归档（第 %d 页） - %s", page, siteName)
		heading = fmt.Sprintf("文章归档 · 第 %d 页", page)
		description = fmt.Sprintf("%s 全部公开文章第 %d 页，共 %d 篇。", siteName, page, total)
	}

	var pagerLinks strings.Builder
	if page > 1 {
		fmt.Fprintf(&pagerLinks, "<link rel=\"prev\" href=\"%s\">\n", htmlstd.EscapeString(archivePageURL(domain, page-1)))
	}
	if page < totalPages {
		fmt.Fprintf(&pagerLinks, "<link rel=\"next\" href=\"%s\">\n", htmlstd.EscapeString(archivePageURL(domain, page+1)))
	}

	collectionJSON, _ := json.Marshal(map[string]any{
		"@context":    "https://schema.org",
		"@type":       "CollectionPage",
		"name":        title,
		"url":         canonical,
		"description": description,
		"isPartOf":    map[string]any{"@type": "WebSite", "name": siteName, "url": domain + "/"},
	})
	crumbs := []breadcrumbItem{{Name: siteName, URL: domain + "/"}, {Name: "文章归档", URL: domain + "/archive"}}
	if page > 1 {
		crumbs = append(crumbs, breadcrumbItem{Name: fmt.Sprintf("第 %d 页", page)})
	}

	var list strings.Builder
	for _, row := range rows {
		articleID := strings.TrimSpace(model.StringValue(row, "GUID"))
		if articleID == "" {
			continue
		}
		articleTitle := strings.TrimSpace(model.StringValue(row, "BLOG_TITLE"))
		if articleTitle == "" {
			articleTitle = "未命名文章"
		}
		excerpt := truncateRunes(collapseWhitespace(plainTextFromHTML(model.StringValue(row, "EXCERPT_HTML"))), 120)
		author := strings.TrimSpace(model.StringValue(row, "USERNAME"))
		created := model.Lookup(row, "CREATE_TIME")
		fmt.Fprintf(&list, `<li class="archive-item"><time datetime="%s">%s</time><div class="archive-body"><h2><a href="%s">%s</a></h2>`,
			htmlstd.EscapeString(formatSEOTime(created)), htmlstd.EscapeString(formatDate(created)),
			htmlstd.EscapeString(domain+"/oneBlog/"+url.PathEscape(articleID)), htmlstd.EscapeString(articleTitle))
		if excerpt != "" {
			fmt.Fprintf(&list, `<p>%s</p>`, htmlstd.EscapeString(excerpt))
		}
		if author != "" {
			fmt.Fprintf(&list, `<span class="archive-author">%s</span>`, htmlstd.EscapeString(author))
		}
		list.WriteString(`</div></li>`)
	}
	if list.Len() == 0 {
		list.WriteString(`<li class="archive-empty">暂时还没有公开文章</li>`)
	}

	escapedTitle := htmlstd.EscapeString(title)
	escapedDescription := htmlstd.EscapeString(description)
	escapedCanonical := htmlstd.EscapeString(canonical)
	escapedSiteName := htmlstd.EscapeString(siteName)
	escapedDomain := htmlstd.EscapeString(domain)

	return fmt.Sprintf(`<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s</title>
<meta name="description" content="%s">
<meta name="robots" content="index,follow,max-image-preview:large,max-snippet:-1">
<link rel="canonical" href="%s">
%s%s
<link rel="alternate" type="application/rss+xml" title="%s RSS" href="%s/rss.xml">
<meta property="og:type" content="website">
<meta property="og:site_name" content="%s">
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
<meta property="og:url" content="%s">
<meta property="og:image" content="%s">
<script type="application/ld+json">%s</script>
<script type="application/ld+json">%s</script>
<style>%s</style>
</head>
<body>
<header class="archive-top"><div class="archive-top-inner"><a class="archive-brand" href="%s/">%s</a>
<nav aria-label="主要栏目"><a href="%s/ynsStudy/Home">最新文章</a><a href="%s/ynsStudy/Resources">资源</a><a href="%s/ynsStudy/Community">社区</a><a href="%s/ynsStudy/About">关于</a></nav></div></header>
<main class="archive-page">
<nav class="archive-crumb" aria-label="面包屑"><a href="%s/">%s</a> / <a href="%s/archive">文章归档</a></nav>
<section class="archive-paper">
<h1>%s</h1>
<p class="archive-summary">共 %d 篇公开文章 · 第 %d / %d 页</p>
<ol class="archive-list">%s</ol>
%s
</section>
</main>
<footer class="archive-foot">©2025 %s · <a href="%s/rss.xml">RSS 订阅</a></footer>
</body>
</html>`,
		escapedTitle, escapedDescription, escapedCanonical, pagerLinks.String(), siteIconLinks,
		escapedSiteName, escapedDomain,
		escapedSiteName, escapedTitle, escapedDescription, escapedCanonical, htmlstd.EscapeString(cfg.SEO.DefaultImage),
		collectionJSON, breadcrumbJSON(crumbs), archiveStyle,
		escapedDomain, escapedSiteName,
		escapedDomain, escapedDomain, escapedDomain, escapedDomain,
		escapedDomain, escapedSiteName, escapedDomain,
		htmlstd.EscapeString(heading), total, page, totalPages, list.String(),
		renderArchivePager(domain, page, totalPages),
		escapedSiteName, escapedDomain)
}

// renderArchivePager 输出上一页、下一页和附近页码，所有页面都能通过普通链接互相到达。
func renderArchivePager(domain string, page, totalPages int) string {
	if totalPages <= 1 {
		return ""
	}
	var pager strings.Builder
	pager.WriteString(`<nav class="archive-pager" aria-label="分页">`)
	if page > 1 {
		fmt.Fprintf(&pager, `<a href="%s">← 上一页</a>`, htmlstd.EscapeString(archivePageURL(domain, page-1)))
	}
	start, end := page-3, page+3
	if start < 1 {
		start = 1
	}
	if end > totalPages {
		end = totalPages
	}
	if start > 1 {
		fmt.Fprintf(&pager, `<a href="%s">1</a>`, htmlstd.EscapeString(archivePageURL(domain, 1)))
		if start > 2 {
			pager.WriteString(`<span class="gap">…</span>`)
		}
	}
	for number := start; number <= end; number++ {
		if number == page {
			fmt.Fprintf(&pager, `<span class="current" aria-current="page">%d</span>`, number)
			continue
		}
		fmt.Fprintf(&pager, `<a href="%s">%d</a>`, htmlstd.EscapeString(archivePageURL(domain, number)), number)
	}
	if end < totalPages {
		if end < totalPages-1 {
			pager.WriteString(`<span class="gap">…</span>`)
		}
		fmt.Fprintf(&pager, `<a href="%s">%d</a>`, htmlstd.EscapeString(archivePageURL(domain, totalPages)), totalPages)
	}
	if page < totalPages {
		fmt.Fprintf(&pager, `<a href="%s">下一页 →</a>`, htmlstd.EscapeString(archivePageURL(domain, page+1)))
	}
	pager.WriteString(`</nav>`)
	return pager.String()
}

// archiveStyle 沿用站点手账风格的配色：桌面点阵底、纸张卡片、楷体标题和钢笔蓝链接。
const archiveStyle = `:root{--ink:#2b2a27;--muted:#7d776c;--paper:#fffdf8;--desk:#efe8da;--rule:#e6dfd1;--pen:#2f5d8a;--stamp:#c2483e;--hand:"LXGW WenKai Screen","Kaiti SC","STKaiti","KaiTi","楷体",serif}
*{box-sizing:border-box}
body{margin:0;min-height:100vh;color:var(--ink);background-color:var(--desk);background-image:radial-gradient(rgba(120,104,80,.18) 1px,transparent 1px);background-size:22px 22px;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif}
a{color:var(--pen);text-decoration:none}
a:hover{text-decoration:underline}
.archive-top{background:rgba(255,253,248,.92);border-bottom:1px dashed #d6ccb8}
.archive-top-inner{max-width:960px;margin:0 auto;padding:14px 16px;display:flex;flex-wrap:wrap;align-items:center;justify-content:space-between;gap:10px}
.archive-brand{font-family:var(--hand);font-size:22px;color:var(--ink)}
.archive-top nav{display:flex;flex-wrap:wrap;gap:18px;font-size:14px}
.archive-page{max-width:960px;margin:0 auto;padding:22px 16px 40px}
.archive-crumb{font-size:13px;color:var(--muted);margin-bottom:14px}
.archive-paper{background:var(--paper);border:1px solid var(--rule);box-shadow:0 1px 2px rgba(60,50,30,.06),0 10px 24px -16px rgba(60,50,30,.35);padding:30px 32px}
.archive-paper h1{margin:0 0 6px;font-family:var(--hand);font-weight:normal;font-size:30px}
.archive-summary{margin:0 0 18px;color:var(--muted);font-size:14px}
.archive-list{list-style:none;margin:0;padding:0}
.archive-item{display:flex;gap:20px;padding:16px 0;border-top:1px dashed var(--rule)}
.archive-item time{flex:none;width:92px;color:var(--muted);font-size:13px;padding-top:4px}
.archive-body{min-width:0}
.archive-body h2{margin:0 0 6px;font-size:18px;line-height:1.5;font-weight:600}
.archive-body p{margin:0 0 6px;color:#4d4943;font-size:14px;line-height:1.75;overflow-wrap:anywhere}
.archive-author{color:var(--muted);font-size:12px}
.archive-empty{padding:24px 0;color:var(--muted)}
.archive-pager{display:flex;flex-wrap:wrap;gap:8px;margin-top:24px;padding-top:18px;border-top:1px dashed var(--rule);font-size:14px}
.archive-pager a,.archive-pager span{padding:4px 10px;border:1px solid var(--rule);background:#fff}
.archive-pager .current{border-color:var(--stamp);color:var(--stamp)}
.archive-pager .gap{border:0;background:none}
.archive-foot{text-align:center;color:var(--muted);font-size:12px;padding:0 16px 28px}
@media(max-width:640px){.archive-paper{padding:22px 16px}.archive-item{display:block}.archive-item time{display:block;width:auto;padding:0 0 4px}}`
