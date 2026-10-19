package controller

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

// seoProfile 为作者主页 /user/:u 输出包含名字、简介和文章列表的首屏 HTML，随后由 Vue 接管。
// 搜索作者名字（例如站长“余南松”）时，这个页面就是落地页。
// 带文章编号的 /user/:u/:blogId 仍由 Vue 直接处理，它的 canonical 指向文章本身。
func (h *Controller) seoProfile(c *gin.Context) {
	profile, found, err := h.service.GetPublicProfileForSEO(c.Request.Context(), c.Param("u"))
	if err != nil {
		h.renderSEOStatusPage(c, http.StatusInternalServerError, "主页暂时无法访问", "服务器读取用户信息时出现异常，请稍后重试。")
		return
	}
	if !found {
		h.renderSEOStatusPage(c, http.StatusNotFound, "用户不存在", "这个主页可能已注销或地址有误。")
		return
	}
	owner := h.service.GetSiteOwner(c.Request.Context())
	page := renderSEOProfileHTML(h.readFrontendShell(), profile, owner, h.service.Config)
	cacheSeconds := h.service.Config.SEO.ResponseCacheSeconds
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=0, s-maxage=%d, stale-while-revalidate=60", cacheSeconds))
	c.Header("X-Robots-Tag", "index, follow, max-image-preview:large, max-snippet:-1")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}

// profileTitle / profileDescription 与前端 personInformation.vue 中的计算方式保持一致，
// 避免 Vue 接管前后标题和描述不同。
func profileTitle(name, siteName string, isOwner bool, owner service.SiteOwner) string {
	if isOwner {
		name = ownerDisplay(owner)
	}
	return name + "的个人主页 - " + siteName
}

func profileDescription(name, remark, siteName string, isOwner bool, owner service.SiteOwner) string {
	remark = strings.TrimSpace(remark)
	if isOwner {
		text := ownerDisplay(owner) + "，" + siteName + " 站长。"
		if remark != "" {
			text += remark
		}
		return text
	}
	if remark != "" {
		return remark
	}
	return name + " 在 " + siteName + " 发表的文章、动态与分享的资源。"
}

func renderSEOProfileHTML(shell string, profile service.PublicProfile, owner service.SiteOwner, cfg *config.Config) string {
	shell = defaultSEOElement.ReplaceAllString(shell, "\n")
	shell = javascriptNotice.ReplaceAllString(shell, "\n")
	domain := strings.TrimRight(cfg.External.DomainName, "/")
	siteName := cfg.SEO.SiteName
	user := profile.User
	name := model.PublicDisplayName(user.Name)
	if name == "" {
		name = "用户"
	}
	canonical := fmt.Sprintf("%s/user/%s", domain, url.PathEscape(fmt.Sprint(user.UserNum)))
	title := profileTitle(name, siteName, profile.IsOwner, owner)
	description := truncateRunes(collapseWhitespace(profileDescription(name, user.Remark, siteName, profile.IsOwner, owner)), 160)

	image := cfg.SEO.DefaultImage
	avatar := ""
	if strings.TrimSpace(user.Avatar) != "" {
		avatar = absoluteArticleURL(domain, safeArticleURL(user.Avatar))
		if avatar != "" {
			image = avatar
		}
	}

	person := structuredPerson{Type: "Person", Name: name, URL: canonical, Image: avatar, Description: strings.TrimSpace(user.Remark)}
	if profile.IsOwner {
		person = ownerPerson(owner, domain)
		person.URL = canonical
		person.Image = avatar
		person.Description = description
	}
	profileJSON, _ := json.Marshal(map[string]any{
		"@context":   "https://schema.org",
		"@type":      "ProfilePage",
		"url":        canonical,
		"name":       title,
		"mainEntity": person,
	})

	escapedTitle := htmlstd.EscapeString(title)
	escapedDescription := htmlstd.EscapeString(description)
	escapedCanonical := htmlstd.EscapeString(canonical)
	escapedImage := htmlstd.EscapeString(image)
	escapedDomain := htmlstd.EscapeString(domain)

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
<meta name="twitter:card" content="summary">
<meta name="twitter:title" content="%s">
<meta name="twitter:description" content="%s">
<meta name="twitter:image" content="%s">
<script type="application/ld+json">%s</script>
<style id="yns-seo-profile-first-paint">%s</style>
`, escapedDescription, escapedCanonical, htmlstd.EscapeString(siteName), escapedTitle, escapedDescription,
		escapedCanonical, escapedImage, escapedTitle, escapedDescription, escapedImage, profileJSON, profileStyle)

	var avatarHTML string
	if avatar != "" {
		avatarHTML = fmt.Sprintf(`<img class="seo-profile-avatar" src="%s" alt="%s的头像" width="96" height="96" decoding="async">`,
			htmlstd.EscapeString(avatar), htmlstd.EscapeString(name))
	} else {
		avatarHTML = fmt.Sprintf(`<span class="seo-profile-avatar seo-profile-initial" aria-hidden="true">%s</span>`,
			htmlstd.EscapeString(string([]rune(name)[:1])))
	}

	var badge, links string
	if profile.IsOwner {
		badge = fmt.Sprintf(`<p class="seo-profile-badge">%s 站长</p>`, htmlstd.EscapeString(siteName))
		var sameAs strings.Builder
		for _, link := range owner.SameAs {
			if safe := safeArticleURL(link); safe != "" {
				fmt.Fprintf(&sameAs, `<a href="%s" rel="me noopener" target="_blank">%s</a>`, htmlstd.EscapeString(safe), htmlstd.EscapeString(sameAsLabel(safe)))
			}
		}
		if sameAs.Len() > 0 {
			links = `<p class="seo-profile-links">` + sameAs.String() + `</p>`
		}
	}
	remarkHTML := ""
	if remark := strings.TrimSpace(user.Remark); remark != "" {
		remarkHTML = `<p class="seo-profile-remark">` + htmlstd.EscapeString(remark) + `</p>`
	}

	var list strings.Builder
	for _, article := range profile.Articles {
		articleID := strings.TrimSpace(model.StringValue(article, "GUID"))
		if articleID == "" {
			continue
		}
		articleTitle := strings.TrimSpace(model.StringValue(article, "BLOG_TITLE"))
		if articleTitle == "" {
			articleTitle = "未命名文章"
		}
		created := model.Lookup(article, "CREATE_TIME")
		fmt.Fprintf(&list, `<li><a href="%s">%s</a><time datetime="%s">%s</time></li>`,
			htmlstd.EscapeString(domain+"/oneBlog/"+url.PathEscape(articleID)), htmlstd.EscapeString(articleTitle),
			htmlstd.EscapeString(formatSEOTime(created)), htmlstd.EscapeString(formatDate(created)))
	}
	articlesHTML := `<p class="seo-profile-empty">暂时还没有公开文章</p>`
	if list.Len() > 0 {
		articlesHTML = `<ul class="seo-profile-list">` + list.String() + `</ul>`
	}

	body := fmt.Sprintf(`<main class="seo-profile-fallback">
<section class="seo-profile-card" aria-labelledby="seo-profile-name">
%s
<div class="seo-profile-info"><h1 id="seo-profile-name">%s</h1>%s%s%s</div>
</section>
<section class="seo-profile-articles" aria-labelledby="seo-profile-articles-title">
<h2 id="seo-profile-articles-title">%s 的文章（%d）</h2>
%s
<p class="seo-profile-more"><a href="%s/archive">浏览全站文章 →</a></p>
</section>
</main>`, avatarHTML, htmlstd.EscapeString(name), badge, remarkHTML, links,
		htmlstd.EscapeString(name), profile.ArticleTotal, articlesHTML, escapedDomain)

	if titleElementPattern.MatchString(shell) {
		shell = titleElementPattern.ReplaceAllStringFunc(shell, func(string) string {
			return "<title>" + escapedTitle + "</title>"
		})
	} else {
		head = "<title>" + escapedTitle + "</title>\n" + head
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

// sameAsLabel 把外部主页地址显示为平台名称，例如 GitHub。
func sameAsLabel(link string) string {
	parsed, err := url.Parse(link)
	if err != nil || parsed.Host == "" {
		return link
	}
	host := strings.TrimPrefix(strings.ToLower(parsed.Host), "www.")
	switch {
	case strings.HasSuffix(host, "github.com"):
		return "GitHub"
	case strings.HasSuffix(host, "gitee.com"):
		return "Gitee"
	case strings.HasSuffix(host, "zhihu.com"):
		return "知乎"
	case strings.HasSuffix(host, "bilibili.com"):
		return "哔哩哔哩"
	case strings.HasSuffix(host, "csdn.net"):
		return "CSDN"
	case strings.HasSuffix(host, "juejin.cn"):
		return "掘金"
	default:
		return host
	}
}

// profileStyle 沿用手账风格：米色点阵桌面、纸张卡片、楷体标题。
const profileStyle = `.seo-profile-fallback{min-height:100vh;padding:32px 16px;background-color:#efe8da;background-image:radial-gradient(rgba(120,104,80,.18) 1px,transparent 1px);background-size:22px 22px;color:#2b2a27;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif;box-sizing:border-box}
.seo-profile-card,.seo-profile-articles{max-width:900px;margin:0 auto;background:#fffdf8;border:1px solid #e6dfd1;box-shadow:0 1px 2px rgba(60,50,30,.06),0 10px 24px -16px rgba(60,50,30,.35);padding:28px 32px;box-sizing:border-box}
.seo-profile-card{display:flex;gap:24px;align-items:center}
.seo-profile-avatar{flex:none;width:96px;height:96px;border-radius:50%;object-fit:cover;border:3px solid #fff;box-shadow:0 2px 8px rgba(60,50,30,.18)}
.seo-profile-initial{display:flex;align-items:center;justify-content:center;background:#3f5a57;color:#fbf6ea;font-size:40px}
.seo-profile-info{min-width:0}
.seo-profile-info h1{margin:0;font-family:"LXGW WenKai Screen","Kaiti SC","STKaiti","KaiTi","楷体",serif;font-weight:normal;font-size:32px}
.seo-profile-badge{display:inline-block;margin:8px 0 0;padding:2px 10px;border:1px solid #c2483e;color:#c2483e;font-size:13px}
.seo-profile-remark{margin:10px 0 0;color:#4d4943;line-height:1.8;overflow-wrap:anywhere}
.seo-profile-links{display:flex;flex-wrap:wrap;gap:14px;margin:10px 0 0;font-size:14px}
.seo-profile-links a,.seo-profile-articles a{color:#2f5d8a;text-decoration:none}
.seo-profile-articles{margin-top:22px}
.seo-profile-articles h2{margin:0 0 10px;font-family:"LXGW WenKai Screen","Kaiti SC","STKaiti","KaiTi","楷体",serif;font-weight:normal;font-size:22px}
.seo-profile-list{list-style:none;margin:0;padding:0}
.seo-profile-list li{display:flex;justify-content:space-between;gap:16px;padding:11px 0;border-top:1px dashed #e6dfd1}
.seo-profile-list a{min-width:0;overflow-wrap:anywhere}
.seo-profile-list time{flex:none;color:#7d776c;font-size:13px}
.seo-profile-empty{color:#7d776c}
.seo-profile-more{margin:16px 0 0;text-align:right;font-size:14px}
@media(max-width:640px){.seo-profile-fallback{padding:12px 10px}.seo-profile-card{flex-direction:column;text-align:center;padding:24px 16px}.seo-profile-links{justify-content:center}.seo-profile-articles{padding:22px 16px}.seo-profile-list li{display:block}.seo-profile-list time{display:block;margin-top:4px}}`
