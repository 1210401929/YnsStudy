package controller

// 便签墙的服务端首屏：单条随记 /note/:id 和便签墙 /ynsStudy/Notes。
// 和文章页一样输出“真实内容 + Vue 入口脚本”，搜索引擎无需执行 JavaScript 就能收录，浏览器加载后由 Vue 接管。

import (
	"encoding/json"
	"fmt"
	htmlstd "html"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/config"
	"nys-go-api/internal/model"
	"nys-go-api/internal/service"
)

// 随记类型的中文名，与前端 noteTypes.js 一致
var noteTypeNames = map[string]string{
	service.NoteTypeNote:     "随笔",
	service.NoteTypeTip:      "技巧",
	service.NoteTypeShare:    "分享",
	service.NoteTypeQuestion: "提问",
}

const noteWallPath = "/ynsStudy/Notes"

const noteFirstPaintStyle = `<style id="yns-seo-note-first-paint">.seo-note-fallback{min-height:100vh;padding:32px 16px;background-color:#efe8da;background-image:radial-gradient(rgba(120,104,80,.18) 1px,transparent 1px);background-size:22px 22px;color:#2b2a27;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif}.seo-note-fallback article,.seo-note-fallback section{max-width:760px;margin:0 auto;background:#fdf6d8;box-shadow:0 1px 2px rgba(60,50,30,.1),0 10px 18px -12px rgba(60,50,30,.4);padding:30px 34px}.seo-note-fallback h1{font-family:"LXGW WenKai Screen","Kaiti SC","STKaiti","KaiTi","楷体",serif;font-weight:normal;font-size:28px;line-height:1.4;margin:0 0 10px}.seo-note-meta,.seo-breadcrumb{color:#7d776c;font-size:14px;margin-bottom:18px}.seo-note-fallback a{color:#2f5d8a;text-decoration:none}.seo-note-content{font-size:16px;line-height:1.85;overflow-wrap:anywhere}.seo-note-content img,.seo-note-images img{max-width:100%;height:auto}.seo-note-images{display:flex;flex-wrap:wrap;gap:10px;margin-top:16px}.seo-note-images img{width:220px;object-fit:cover}.seo-note-links{margin-top:28px;padding-top:16px;border-top:1px dashed #d6ccb8;line-height:1.9}.seo-note-list{list-style:none;margin:0;padding:0}.seo-note-list li{padding:12px 0;border-top:1px dashed #e6dfd1}.seo-note-list .seo-note-meta{margin:4px 0 0}@media(max-width:640px){.seo-note-fallback{padding:12px 10px}.seo-note-fallback article,.seo-note-fallback section{padding:22px 16px}}</style>`

func noteTypeName(row map[string]any) string {
	if name := noteTypeNames[model.StringValue(row, "TYPE")]; name != "" {
		return name
	}
	return noteTypeNames[service.NoteTypeNote]
}

func noteImages(row map[string]any) []string {
	images, _ := row["IMAGES"].([]string)
	return images
}

// injectSEOShell 把标题、head 内容和首屏正文放进 Vue 的 index.html 外壳
func injectSEOShell(shell, fullTitle, head, body string) string {
	shell = defaultSEOElement.ReplaceAllString(shell, "\n")
	shell = javascriptNotice.ReplaceAllString(shell, "\n")
	escapedFullTitle := htmlstd.EscapeString(fullTitle)
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

func (h *Controller) writeSEOPage(c *gin.Context, page string) {
	cacheSeconds := h.service.Config.SEO.ResponseCacheSeconds
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=0, s-maxage=%d, stale-while-revalidate=60", cacheSeconds))
	c.Header("X-Robots-Tag", "index, follow, max-image-preview:large, max-snippet:-1")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}

// seoNote 输出单条随记
func (h *Controller) seoNote(c *gin.Context) {
	ctx := c.Request.Context()
	note, found, err := h.service.FindNote(ctx, c.Param("id"))
	if err != nil {
		h.renderSEOStatusPage(c, http.StatusInternalServerError, "便签暂时无法访问", "服务器读取便签时出现异常，请稍后重试。")
		return
	}
	if !found {
		h.renderSEOStatusPage(c, http.StatusNotFound, "便签不存在", "这张便签可能已被删除，或地址有误。")
		return
	}
	previous, next := h.service.NoteNeighbors(ctx, note)
	owner := h.service.GetSiteOwner(ctx)
	h.writeSEOPage(c, renderSEONoteHTML(h.readFrontendShell(), note, previous, next, owner, h.service.Config))
}

func renderSEONoteHTML(shell string, note, previous, next map[string]any, owner service.SiteOwner, cfg *config.Config) string {
	domain := strings.TrimRight(cfg.External.DomainName, "/")
	noteID := model.StringValue(note, "GUID")
	canonical := domain + "/note/" + url.PathEscape(noteID)
	title := service.NoteTitle(note)
	author := model.PublicDisplayName(model.StringValue(note, "USERNAME"))
	if author == "" {
		author = cfg.SEO.SiteName
	}
	authorURL := ""
	if userNumber := strings.TrimSpace(model.StringValue(note, "USERNUM")); userNumber != "" {
		authorURL = domain + "/user/" + url.PathEscape(userNumber)
	}

	safeContent, firstImage := sanitizeArticleHTML(model.StringValue(note, "TEXT"), title, domain)
	images := noteImages(note)
	if firstImage == "" && len(images) > 0 {
		firstImage = domain + images[0]
	}
	if firstImage == "" {
		firstImage = cfg.SEO.DefaultImage
	}
	description := truncateRunes(collapseWhitespace(plainTextFromHTML(safeContent)), 160)
	if description == "" {
		description = author + " 在 " + cfg.SEO.SiteName + " 便签墙贴的一张便签"
	}
	published := formatSEOTime(model.Lookup(note, "CREATE_TIME"))

	structured := articleStructuredData{
		Context:          "https://schema.org",
		Type:             "SocialMediaPosting",
		Headline:         title,
		Description:      description,
		URL:              canonical,
		DatePublished:    published,
		DateModified:     published,
		MainEntityOfPage: structuredMainEntity{Type: "WebPage", ID: canonical},
		Author:           structuredPerson{Type: "Person", Name: author, URL: authorURL},
		Publisher:        structuredOrganization{Type: "Organization", Name: cfg.SEO.SiteName, URL: domain, Logo: domain + "/icon-512.png"},
		Image:            []string{firstImage},
	}
	if owner.Code != "" && model.StringValue(note, "USERCODE") == owner.Code {
		person := ownerPerson(owner, domain)
		if authorURL != "" {
			person.URL = authorURL
		}
		structured.Author = person
	}
	structuredJSON, _ := json.Marshal(structured)

	escapedTitle := htmlstd.EscapeString(title)
	escapedDescription := htmlstd.EscapeString(description)
	escapedCanonical := htmlstd.EscapeString(canonical)
	escapedImage := htmlstd.EscapeString(firstImage)
	head := fmt.Sprintf(`
<meta name="description" content="%s">
<meta name="robots" content="index,follow,max-image-preview:large,max-snippet:-1">
<link rel="canonical" href="%s">
<meta property="og:type" content="article">
<meta property="og:site_name" content="%s">
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
<meta property="og:url" content="%s">
<meta property="og:image" content="%s">
<meta property="article:published_time" content="%s">
<meta name="twitter:card" content="summary">
<meta name="twitter:title" content="%s">
<meta name="twitter:description" content="%s">
<script type="application/ld+json">%s</script>
<script type="application/ld+json">%s</script>
%s
`, escapedDescription, escapedCanonical, htmlstd.EscapeString(cfg.SEO.SiteName), escapedTitle, escapedDescription,
		escapedCanonical, escapedImage, htmlstd.EscapeString(published), escapedTitle, escapedDescription, structuredJSON,
		breadcrumbJSON([]breadcrumbItem{{Name: cfg.SEO.SiteName, URL: domain + "/"}, {Name: "便签墙", URL: domain + noteWallPath}, {Name: title}}),
		noteFirstPaintStyle)

	authorHTML := authorLinkHTML(author, model.StringValue(note, "USERNUM"), domain)
	var imageHTML strings.Builder
	if len(images) > 0 {
		imageHTML.WriteString(`<div class="seo-note-images">`)
		for index, image := range images {
			fmt.Fprintf(&imageHTML, `<img src="%s" alt="%s 配图 %d" loading="lazy">`, htmlstd.EscapeString(image), escapedTitle, index+1)
		}
		imageHTML.WriteString(`</div>`)
	}
	var links strings.Builder
	links.WriteString(`<footer class="seo-note-links">`)
	if previous != nil {
		fmt.Fprintf(&links, `<p><a href="%s/note/%s">← 上一条：%s</a></p>`, htmlstd.EscapeString(domain), url.PathEscape(model.StringValue(previous, "GUID")), htmlstd.EscapeString(model.StringValue(previous, "TITLE")))
	}
	if next != nil {
		fmt.Fprintf(&links, `<p><a href="%s/note/%s">下一条：%s →</a></p>`, htmlstd.EscapeString(domain), url.PathEscape(model.StringValue(next, "GUID")), htmlstd.EscapeString(model.StringValue(next, "TITLE")))
	}
	fmt.Fprintf(&links, `<p><a href="%s%s">回到便签墙 →</a></p></footer>`, htmlstd.EscapeString(domain), noteWallPath)

	body := fmt.Sprintf(`<main class="seo-note-fallback">
<article itemscope itemtype="https://schema.org/SocialMediaPosting">
<nav class="seo-breadcrumb" aria-label="面包屑"><a href="%s/">YnsStudy</a> / <a href="%s%s">便签墙</a></nav>
<h1 itemprop="headline">%s</h1>
<p class="seo-note-meta">%s · %s · <time itemprop="datePublished" datetime="%s">%s</time></p>
<div class="seo-note-content" itemprop="articleBody">%s</div>
%s
%s
</article></main>`, htmlstd.EscapeString(domain), htmlstd.EscapeString(domain), noteWallPath, escapedTitle,
		htmlstd.EscapeString(noteTypeName(note)), authorHTML, htmlstd.EscapeString(published),
		htmlstd.EscapeString(formatDate(model.Lookup(note, "CREATE_TIME"))), safeContent, imageHTML.String(), links.String())

	return injectSEOShell(shell, title+" - 便签墙 - "+cfg.SEO.SiteName, head, body)
}

// seoNoteWall 输出便签墙首屏：栏目介绍和最新随记链接
func (h *Controller) seoNoteWall(c *gin.Context) {
	notes, err := h.service.GetLatestNotesForSEO(c.Request.Context(), 30)
	if err != nil {
		h.renderSEOStatusPage(c, http.StatusInternalServerError, "便签墙暂时无法访问", "服务器读取便签时出现异常，请稍后重试。")
		return
	}
	h.writeSEOPage(c, renderSEONoteWallHTML(h.readFrontendShell(), notes, h.service.Config))
}

func renderSEONoteWallHTML(shell string, notes []map[string]any, cfg *config.Config) string {
	domain := strings.TrimRight(cfg.External.DomainName, "/")
	canonical := domain + noteWallPath
	title := "便签墙 - " + cfg.SEO.SiteName
	description := cfg.SEO.SiteName + " 便签墙：随手记下的技巧、踩坑、分享和想法，一张张贴在墙上。"
	escapedDescription := htmlstd.EscapeString(description)
	escapedCanonical := htmlstd.EscapeString(canonical)
	head := fmt.Sprintf(`
<meta name="description" content="%s">
<meta name="robots" content="index,follow,max-image-preview:large,max-snippet:-1">
<link rel="canonical" href="%s">
<meta property="og:type" content="website">
<meta property="og:site_name" content="%s">
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
<meta property="og:url" content="%s">
<meta property="og:image" content="%s">
<script type="application/ld+json">%s</script>
%s
`, escapedDescription, escapedCanonical, htmlstd.EscapeString(cfg.SEO.SiteName), htmlstd.EscapeString(title), escapedDescription,
		escapedCanonical, htmlstd.EscapeString(cfg.SEO.DefaultImage),
		breadcrumbJSON([]breadcrumbItem{{Name: cfg.SEO.SiteName, URL: domain + "/"}, {Name: "便签墙"}}), noteFirstPaintStyle)

	var list strings.Builder
	for _, note := range notes {
		noteID := strings.TrimSpace(model.StringValue(note, "GUID"))
		if noteID == "" {
			continue
		}
		author := model.PublicDisplayName(model.StringValue(note, "USERNAME"))
		fmt.Fprintf(&list, `<li><a href="%s/note/%s">%s</a><p class="seo-note-meta">%s · %s · %s</p></li>`,
			htmlstd.EscapeString(domain), url.PathEscape(noteID), htmlstd.EscapeString(service.NoteTitle(note)),
			htmlstd.EscapeString(noteTypeName(note)), htmlstd.EscapeString(author),
			htmlstd.EscapeString(formatDate(model.Lookup(note, "CREATE_TIME"))))
	}
	if list.Len() == 0 {
		list.WriteString(`<li>墙上还没有便签</li>`)
	}
	body := fmt.Sprintf(`<main class="seo-note-fallback"><section>
<nav class="seo-breadcrumb" aria-label="面包屑"><a href="%s/">YnsStudy</a> / 便签墙</nav>
<h1>便签墙</h1>
<p class="seo-note-meta">%s</p>
<ul class="seo-note-list">%s</ul>
</section></main>`, htmlstd.EscapeString(domain), escapedDescription, list.String())
	return injectSEOShell(shell, title, head, body)
}
