package controller

// 文章标签页 /tag/标签名：服务端输出这个标签下的公开文章列表（真实内容 + Vue 入口脚本），浏览器加载后由 Vue 接管。
// 第 2 页起用 ?page=N；公开文章少于 service.BlogTagIndexableMin 篇的标签页输出 noindex，避免内容单薄的页面被收录。

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

const tagFirstPaintStyle = `<style id="yns-seo-tag-first-paint">.seo-tag-fallback{min-height:100vh;padding:32px 16px;background-color:#efe8da;background-image:radial-gradient(rgba(120,104,80,.18) 1px,transparent 1px);background-size:22px 22px;color:#2b2a27;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif}.seo-tag-fallback section{max-width:960px;margin:0 auto;background:#fffdf8;border:1px solid #e6dfd1;box-shadow:0 1px 2px rgba(60,50,30,.06),0 10px 24px -16px rgba(60,50,30,.35);padding:30px 32px}.seo-tag-fallback a{color:#2f5d8a;text-decoration:none}.seo-tag-fallback h1{font-family:"LXGW WenKai Screen","Kaiti SC","STKaiti","KaiTi","楷体",serif;font-weight:normal;font-size:30px;margin:0 0 6px}.seo-tag-meta,.seo-breadcrumb{color:#7d776c;font-size:14px;margin-bottom:16px}.seo-tag-list{list-style:none;margin:0;padding:0}.seo-tag-list li{padding:14px 0;border-top:1px dashed #e6dfd1}.seo-tag-list h2{margin:0 0 6px;font-size:18px;font-weight:600}.seo-tag-list p{margin:0 0 6px;color:#4d4943;font-size:14px;line-height:1.75;overflow-wrap:anywhere}.seo-tag-pager,.seo-tag-others{display:flex;flex-wrap:wrap;gap:10px;margin-top:20px;padding-top:16px;border-top:1px dashed #e6dfd1;font-size:14px}@media(max-width:640px){.seo-tag-fallback{padding:12px 10px}.seo-tag-fallback section{padding:22px 16px}}</style>`

// seoTag 输出文章标签页
func (h *Controller) seoTag(c *gin.Context) {
	ctx := c.Request.Context()
	requested := strings.TrimSpace(c.Param("name"))
	page := 1
	if raw, exists := c.GetQuery("page"); exists {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || strconv.Itoa(parsed) != raw {
			h.renderSEOStatusPage(c, http.StatusNotFound, "页面不存在", "标签页码有误。")
			return
		}
		page = parsed
	}
	rows, total, name, err := h.service.GetBlogTagPage(ctx, requested, page)
	if err != nil {
		h.renderSEOStatusPage(c, http.StatusInternalServerError, "标签页暂时无法访问", "服务器读取文章时出现异常，请稍后重试。")
		return
	}
	if name == "" || total == 0 {
		h.renderSEOStatusPage(c, http.StatusNotFound, "标签不存在", "这个标签下还没有公开文章。")
		return
	}
	totalPages := (total + service.BlogTagPageSize - 1) / service.BlogTagPageSize
	if page > totalPages {
		h.renderSEOStatusPage(c, http.StatusNotFound, "页面不存在", "这一页已经没有文章了。")
		return
	}
	// 大小写不同（go 和 Go）或显式写了 ?page=1 时，跳转到唯一的规范地址
	if name != requested || c.Query("page") == "1" {
		c.Redirect(http.StatusMovedPermanently, tagPageURL("", name, page))
		return
	}
	popular, _ := h.service.PopularBlogTags(ctx, 20)
	indexable := total >= service.BlogTagIndexableMin
	output := renderSEOTagHTML(h.readFrontendShell(), name, rows, total, page, totalPages, popular, indexable, h.service.Config)
	h.writeSEOPageIndexable(c, output, indexable)
}

// tagPageURL 返回标签页地址，domain 为空时返回站内路径
func tagPageURL(domain, name string, page int) string {
	address := domain + service.BlogTagPath(name)
	if page > 1 {
		address += "?page=" + strconv.Itoa(page)
	}
	return address
}

func renderSEOTagHTML(shell, name string, rows []map[string]any, total, page, totalPages int, popular []map[string]any, indexable bool, cfg *config.Config) string {
	domain := strings.TrimRight(cfg.External.DomainName, "/")
	canonical := tagPageURL(domain, name, page)
	title := name + " 相关文章"
	if page > 1 {
		title += fmt.Sprintf("（第 %d 页）", page)
	}
	titles := make([]string, 0, 3)
	for _, row := range rows {
		if len(titles) == 3 {
			break
		}
		if articleTitle := strings.TrimSpace(model.StringValue(row, "BLOG_TITLE")); articleTitle != "" {
			titles = append(titles, "《"+articleTitle+"》")
		}
	}
	description := fmt.Sprintf("%s 上标签为「%s」的 %d 篇公开文章", cfg.SEO.SiteName, name, total)
	if len(titles) > 0 {
		description += "，包括" + strings.Join(titles, "、")
	}
	description = truncateRunes(description+"。", 180)

	robots := "index,follow,max-image-preview:large,max-snippet:-1"
	if !indexable {
		robots = "noindex,follow"
	}
	collectionJSON, _ := json.Marshal(map[string]any{
		"@context":    "https://schema.org",
		"@type":       "CollectionPage",
		"name":        title,
		"url":         canonical,
		"description": description,
		"keywords":    name,
		"isPartOf":    map[string]any{"@type": "WebSite", "name": cfg.SEO.SiteName, "url": domain + "/"},
	})
	escapedDescription := htmlstd.EscapeString(description)
	escapedCanonical := htmlstd.EscapeString(canonical)
	var pagerLinks strings.Builder
	if page > 1 {
		fmt.Fprintf(&pagerLinks, "\n<link rel=\"prev\" href=\"%s\">", htmlstd.EscapeString(tagPageURL(domain, name, page-1)))
	}
	if page < totalPages {
		fmt.Fprintf(&pagerLinks, "\n<link rel=\"next\" href=\"%s\">", htmlstd.EscapeString(tagPageURL(domain, name, page+1)))
	}
	head := fmt.Sprintf(`
<meta name="description" content="%s">
<meta name="robots" content="%s">
<link rel="canonical" href="%s">%s
<meta property="og:type" content="website">
<meta property="og:site_name" content="%s">
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
<meta property="og:url" content="%s">
<meta property="og:image" content="%s">
<script type="application/ld+json">%s</script>
<script type="application/ld+json">%s</script>
%s
`, escapedDescription, robots, escapedCanonical, pagerLinks.String(), htmlstd.EscapeString(cfg.SEO.SiteName),
		htmlstd.EscapeString(title), escapedDescription, escapedCanonical, htmlstd.EscapeString(cfg.SEO.DefaultImage),
		collectionJSON,
		breadcrumbJSON([]breadcrumbItem{{Name: cfg.SEO.SiteName, URL: domain + "/"}, {Name: "文章归档", URL: domain + "/archive"}, {Name: "#" + name}}),
		tagFirstPaintStyle)

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
		created := model.Lookup(row, "CREATE_TIME")
		fmt.Fprintf(&list, `<li><h2><a href="%s/oneBlog/%s">%s</a></h2>`, htmlstd.EscapeString(domain),
			htmlstd.EscapeString(url.PathEscape(articleID)), htmlstd.EscapeString(articleTitle))
		if excerpt := truncateRunes(collapseWhitespace(plainTextFromHTML(model.StringValue(row, "EXCERPT_HTML"))), 120); excerpt != "" {
			fmt.Fprintf(&list, `<p>%s</p>`, htmlstd.EscapeString(excerpt))
		}
		meta := `<time datetime="` + htmlstd.EscapeString(formatSEOTime(created)) + `">` + htmlstd.EscapeString(formatDate(created)) + `</time>`
		if author := model.PublicDisplayName(model.StringValue(row, "USERNAME")); author != "" {
			meta = authorLinkHTML(author, model.StringValue(row, "USERNUM"), domain) + " · " + meta
		}
		list.WriteString(`<span class="seo-tag-meta">` + meta + `</span></li>`)
	}

	var pager strings.Builder
	if totalPages > 1 {
		pager.WriteString(`<nav class="seo-tag-pager" aria-label="分页">`)
		if page > 1 {
			fmt.Fprintf(&pager, `<a href="%s">← 上一页</a>`, htmlstd.EscapeString(tagPageURL(domain, name, page-1)))
		}
		fmt.Fprintf(&pager, `<span>第 %d / %d 页</span>`, page, totalPages)
		if page < totalPages {
			fmt.Fprintf(&pager, `<a href="%s">下一页 →</a>`, htmlstd.EscapeString(tagPageURL(domain, name, page+1)))
		}
		pager.WriteString(`</nav>`)
	}

	// 其他热门标签：标签页之间互相链接，抓取器能从一个标签发现其他标签
	var others strings.Builder
	for _, tag := range popular {
		other := model.StringValue(tag, "NAME")
		if other == "" || strings.EqualFold(other, name) {
			continue
		}
		if others.Len() == 0 {
			others.WriteString(`<nav class="seo-tag-others" aria-label="其他标签"><span>其他标签：</span>`)
		}
		fmt.Fprintf(&others, `<a href="%s" rel="tag">#%s</a>`, htmlstd.EscapeString(tagPageURL(domain, other, 1)), htmlstd.EscapeString(other))
	}
	if others.Len() > 0 {
		others.WriteString(`</nav>`)
	}

	body := fmt.Sprintf(`<main class="seo-tag-fallback"><section>
<nav class="seo-breadcrumb" aria-label="面包屑"><a href="%s/">YnsStudy</a> / <a href="%s/archive">文章归档</a> / #%s</nav>
<h1>#%s</h1>
<p class="seo-tag-meta">共 %d 篇公开文章</p>
<ul class="seo-tag-list">%s</ul>
%s%s
</section></main>`, htmlstd.EscapeString(domain), htmlstd.EscapeString(domain), htmlstd.EscapeString(name),
		htmlstd.EscapeString(name), total, list.String(), pager.String(), others.String())
	return injectSEOShell(shell, title+" - "+cfg.SEO.SiteName, head, body)
}

// tagLinksForArticleHTML 输出文章页的标签链接，指向各自的标签页
func tagLinksForArticleHTML(tags []string, domain string) string {
	if len(tags) == 0 {
		return ""
	}
	var output strings.Builder
	output.WriteString(`<p class="seo-article-tags">标签：`)
	for _, tag := range tags {
		fmt.Fprintf(&output, ` <a href="%s" rel="tag">#%s</a>`, htmlstd.EscapeString(tagPageURL(domain, tag, 1)), htmlstd.EscapeString(tag))
	}
	output.WriteString(`</p>`)
	return output.String()
}
