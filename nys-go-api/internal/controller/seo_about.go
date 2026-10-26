package controller

// 关于页 /ynsStudy/About 的服务端首屏：站点介绍和建站天数、站长、站长想说、站长推荐、功能、常见问题和使用技术。
// 内容来自 service.AboutContentData（与前端关于页同一份），结构化数据声明这是关于页，主体是站长本人，
// 让搜索站长名字时更容易找到这一页。浏览器加载后由 Vue 接管。

import (
	"encoding/json"
	"fmt"
	htmlstd "html"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/config"
	"nys-go-api/internal/model"
	"nys-go-api/internal/service"
)

const aboutPath = "/ynsStudy/About"

const aboutFirstPaintStyle = `<style id="yns-seo-about-first-paint">.seo-about-fallback{min-height:100vh;padding:32px 16px;background-color:#efe8da;background-image:radial-gradient(rgba(120,104,80,.18) 1px,transparent 1px);background-size:22px 22px;color:#2b2a27;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif}.seo-about-fallback article{max-width:880px;margin:0 auto;background:#fffdf8;border:1px solid #e6dfd1;box-shadow:0 1px 2px rgba(60,50,30,.06),0 10px 24px -16px rgba(60,50,30,.35);padding:32px 36px}.seo-about-fallback a{color:#2f5d8a;text-decoration:none}.seo-about-fallback h1,.seo-about-fallback h2{font-family:"LXGW WenKai Screen","Kaiti SC","STKaiti","KaiTi","楷体",serif;font-weight:normal}.seo-about-fallback h1{font-size:30px;margin:0 0 8px}.seo-about-fallback h2{font-size:22px;margin:28px 0 10px}.seo-about-fallback p,.seo-about-fallback li,.seo-about-fallback dd{line-height:1.8;color:#4d4943}.seo-about-fallback dt{font-weight:600;margin-top:10px}.seo-about-fallback dd{margin:2px 0 0}.seo-breadcrumb{color:#7d776c;font-size:14px;margin-bottom:16px}@media(max-width:640px){.seo-about-fallback{padding:12px 10px}.seo-about-fallback article{padding:22px 16px}}</style>`

// seoAbout 输出关于页首屏
func (h *Controller) seoAbout(c *gin.Context) {
	owner := h.service.GetSiteOwner(c.Request.Context())
	featured := h.service.AboutFeatured(c.Request.Context())
	h.writeSEOPage(c, renderSEOAboutHTML(h.readFrontendShell(), service.AboutContentData(), owner, featured, service.SiteDays(time.Now()), h.service.Config))
}

func renderSEOAboutHTML(shell string, content service.AboutContent, owner service.SiteOwner, featured service.Featured, siteDays int, cfg *config.Config) string {
	domain := strings.TrimRight(cfg.External.DomainName, "/")
	canonical := domain + aboutPath
	title := "关于本站"
	description := truncateRunes(fmt.Sprintf("%s：%s。站长 %s，介绍网站的由来、站长推荐的文章、功能和常见问题。",
		cfg.SEO.SiteName, content.Intro, ownerDisplay(owner)), 180)

	person := ownerPerson(owner, domain)
	if owner.Avatar != "" {
		person.Image = absoluteArticleURL(domain, owner.Avatar)
	}
	aboutJSON, _ := json.Marshal(map[string]any{
		"@context":    "https://schema.org",
		"@type":       "AboutPage",
		"name":        title + " - " + cfg.SEO.SiteName,
		"url":         canonical,
		"description": description,
		"inLanguage":  "zh-CN",
		"mainEntity":  person,
		"isPartOf":    map[string]any{"@type": "WebSite", "name": cfg.SEO.SiteName, "url": domain + "/"},
	})
	escapedDescription := htmlstd.EscapeString(description)
	escapedCanonical := htmlstd.EscapeString(canonical)
	head := fmt.Sprintf(`
<meta name="description" content="%s">
<meta name="robots" content="index,follow,max-image-preview:large,max-snippet:-1">
<link rel="canonical" href="%s">
<meta property="og:type" content="profile">
<meta property="og:site_name" content="%s">
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
<meta property="og:url" content="%s">
<meta property="og:image" content="%s">
<script type="application/ld+json">%s</script>
<script type="application/ld+json">%s</script>
%s
`, escapedDescription, escapedCanonical, htmlstd.EscapeString(cfg.SEO.SiteName), htmlstd.EscapeString(title+" - "+cfg.SEO.SiteName),
		escapedDescription, escapedCanonical, htmlstd.EscapeString(cfg.SEO.DefaultImage), aboutJSON,
		breadcrumbJSON([]breadcrumbItem{{Name: cfg.SEO.SiteName, URL: domain + "/"}, {Name: title}}), aboutFirstPaintStyle)

	link := func(path, text string) string {
		if path == "" {
			return htmlstd.EscapeString(text)
		}
		href := path
		if strings.HasPrefix(path, "/") {
			href = domain + path
		}
		return `<a href="` + htmlstd.EscapeString(href) + `">` + htmlstd.EscapeString(text) + `</a>`
	}

	var body strings.Builder
	body.WriteString(`<main class="seo-about-fallback"><article>`)
	fmt.Fprintf(&body, `<nav class="seo-breadcrumb" aria-label="面包屑"><a href="%s/">YnsStudy</a> / 关于本站</nav>`, htmlstd.EscapeString(domain))
	fmt.Fprintf(&body, `<h1>%s</h1><p>%s</p>`, htmlstd.EscapeString(content.Title), htmlstd.EscapeString(content.Intro))
	if siteDays > 0 {
		fmt.Fprintf(&body, `<p>自 %s 上线，%s 已经陪伴了 %d 天。</p>`, htmlstd.EscapeString(content.SiteStartDate), htmlstd.EscapeString(cfg.SEO.SiteName), siteDays)
	}

	body.WriteString(`<h2>站长</h2><p>`)
	body.WriteString(ownerLinkHTML(owner, domain))
	for _, sameAs := range owner.SameAs {
		body.WriteString(` · ` + link(sameAs, strings.TrimPrefix(strings.TrimPrefix(sameAs, "https://"), "http://")))
	}
	if content.ContactEmail != "" {
		fmt.Fprintf(&body, ` · 联系方式：<a href="mailto:%s">%s</a>`, htmlstd.EscapeString(content.ContactEmail), htmlstd.EscapeString(content.ContactEmail))
	}
	body.WriteString(`</p>`)

	body.WriteString(`<h2>站长想说</h2>`)
	for _, paragraph := range content.OwnerMessage {
		body.WriteString(`<p>` + htmlstd.EscapeString(paragraph) + `</p>`)
	}

	// 站长推荐：站长在后台挑选的文章和系列，没有配置时不输出
	if len(featured.Articles) > 0 || len(featured.Series) > 0 {
		body.WriteString(`<h2>站长推荐</h2>`)
		if len(featured.Series) > 0 {
			body.WriteString(`<ul>`)
			for _, row := range featured.Series {
				fmt.Fprintf(&body, `<li>系列《<a href="%s">%s</a>》，共 %d 篇`, htmlstd.EscapeString(seriesURL(domain, model.StringValue(row, "ID"))),
					htmlstd.EscapeString(model.StringValue(row, "TITLE")), model.IntValue(row, "ARTICLE_COUNT"))
				if text := strings.TrimSpace(model.StringValue(row, "DESCRIPTION")); text != "" {
					body.WriteString(`：` + htmlstd.EscapeString(truncateRunes(text, 80)))
				}
				body.WriteString(`</li>`)
			}
			body.WriteString(`</ul>`)
		}
		if len(featured.Articles) > 0 {
			body.WriteString(`<ul>`)
			for _, row := range featured.Articles {
				fmt.Fprintf(&body, `<li><a href="%s/oneBlog/%s">%s</a>`, htmlstd.EscapeString(domain),
					htmlstd.EscapeString(url.PathEscape(model.StringValue(row, "GUID"))), htmlstd.EscapeString(model.StringValue(row, "BLOG_TITLE")))
				if excerpt := strings.TrimSpace(model.StringValue(row, "EXCERPT")); excerpt != "" {
					body.WriteString(`：` + htmlstd.EscapeString(truncateRunes(excerpt, 80)))
				}
				body.WriteString(`</li>`)
			}
			body.WriteString(`</ul>`)
		}
	}

	body.WriteString(`<h2>能做什么</h2><ul>`)
	for _, item := range content.Features {
		body.WriteString(`<li>` + link(item.Link, item.Title) + `：` + htmlstd.EscapeString(item.Text) + `</li>`)
	}
	body.WriteString(`</ul>`)

	body.WriteString(`<h2>常见问题</h2><dl>`)
	for _, item := range content.FAQ {
		answer := htmlstd.EscapeString(item.Text)
		if item.Link != "" {
			answer += ` ` + link(item.Link, "去看看 →")
		}
		body.WriteString(`<dt>` + htmlstd.EscapeString(item.Title) + `</dt><dd>` + answer + `</dd>`)
	}
	body.WriteString(`</dl>`)

	body.WriteString(`<h2>使用技术</h2><ul>`)
	for _, item := range content.TechStack {
		body.WriteString(`<li>` + htmlstd.EscapeString(item.Title) + `：` + htmlstd.EscapeString(item.Text) + `</li>`)
	}
	body.WriteString(`</ul>`)
	if content.SourceURL != "" {
		body.WriteString(`<p>网站源码：` + link(content.SourceURL, content.SourceURL) + `</p>`)
	}
	if content.ChangelogPath != "" {
		body.WriteString(`<p>` + link(content.ChangelogPath, "网站更新日志 →") + `</p>`)
	}
	body.WriteString(`</article></main>`)
	return injectSEOShell(shell, title+" - "+cfg.SEO.SiteName, head, body.String())
}
