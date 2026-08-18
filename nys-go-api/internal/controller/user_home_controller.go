package controller

import (
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

func (h *Controller) registerUserInformationRoutes(group *gin.RouterGroup) {
	group.Any("/followUser", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.FollowUser(c, stringParam(body, "followUserCode"), stringParam(body, "followUserName")))
		}
	})
	group.Any("/noFollowUser", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.UnfollowUser(c, stringParam(body, "followUserCode")))
		}
	})
	group.Any("/getFollowUser", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetFollowUsers(c, stringParam(body, "userCode"), asBoolValue(body["isCountOnly"])))
		}
	})
	group.Any("/getBlogAndResourceByUserCode", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetBlogResourceCommunityByUser(c.Request.Context(), stringParam(body, "userCode")))
		}
	})
	group.Any("/getBlogAndCommunityByUserCode", h.getUserBlogAndCommunity)
	group.Any("/getResourceByUserCode", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetFilesByUser(c.Request.Context(), stringParam(body, "userCode")))
		}
	})
	group.Any("/getBlogByUserCode", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetBlogsByUser(c.Request.Context(), stringParam(body, "userCode"), true))
		}
	})
	group.Any("/setPersonInfo", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.SetPersonInfo(c.Request.Context(), stringParam(body, "userCode"), stringParam(body, "fieldName"), stringParam(body, "fieldValue")))
		}
	})
	group.Any("/getPersonInfo", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetPersonInfo(c.Request.Context(), stringParam(body, "userCode")))
		}
	})
}

func (h *Controller) getUserBlogAndCommunity(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.GetBlogAndCommunityByUser(c.Request.Context(), stringParam(body, "userCode"), intParam(body, "page", 1), intParam(body, "pageSize", 10), stringParam(body, "keyword")))
}

func (h *Controller) registerHomeRoutes(group *gin.RouterGroup) {
	group.Any("/getHomeData", func(c *gin.Context) { writeResult(c, h.service.GetHomeData(c.Request.Context())) })
	group.Any("/getWebsiteStatistics", func(c *gin.Context) { writeResult(c, h.service.GetWebsiteStatistics(c.Request.Context())) })
	group.Any("/getHigAuthor", func(c *gin.Context) {
		body, _ := readBody(c)
		writeResult(c, h.service.GetHighQualityAuthors(c.Request.Context(), parseIntOr(stringParam(body, "num"), 4)))
	})
	group.GET("/sitemap_blog.xml", h.sitemapBlogs)
	group.GET("/sitemap_user.xml", h.sitemapUsers)
	group.GET("/rss.xml", h.rss)
}

type sitemapDocument struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Location   string `xml:"loc"`
	LastModify string `xml:"lastmod,omitempty"`
	Frequency  string `xml:"changefreq"`
	Priority   string `xml:"priority"`
}

func (h *Controller) sitemapBlogs(c *gin.Context) {
	result := h.service.GetAllPublicBlogIDs(c.Request.Context())
	if result.IsError {
		writeResult(c, result)
		return
	}
	rows, _ := result.Result.([]map[string]any)
	document := sitemapDocument{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: make([]sitemapURL, 0, len(rows))}
	for _, row := range rows {
		document.URLs = append(document.URLs, sitemapURL{
			Location:   h.service.Config.External.DomainName + "/oneBlog/" + model.StringValue(row, "GUID"),
			LastModify: formatDate(model.Lookup(row, "LAST_MODIFIED")), Frequency: "weekly", Priority: "0.8",
		})
	}
	writeXML(c, document)
}

func (h *Controller) sitemapUsers(c *gin.Context) {
	rows, err := h.service.Repo.Query(c.Request.Context(), "SELECT USERNUM FROM userInfo")
	if err != nil {
		writeResult(c, model.Failure("查询用户失败:"+err.Error()))
		return
	}
	document := sitemapDocument{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: make([]sitemapURL, 0, len(rows))}
	for _, row := range rows {
		document.URLs = append(document.URLs, sitemapURL{
			Location: h.service.Config.External.DomainName + "/user/" + model.StringValue(row, "USERNUM"), Frequency: "weekly", Priority: "0.5",
		})
	}
	writeXML(c, document)
}

func writeXML(c *gin.Context, value any) {
	data, err := xml.MarshalIndent(value, "", "  ")
	if err != nil {
		writeResult(c, model.Failure("生成 XML 失败:"+err.Error()))
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", append([]byte(xml.Header), data...))
}

func (h *Controller) rss(c *gin.Context) {
	result := h.service.GetLatestBlogs(c.Request.Context())
	if result.IsError {
		writeResult(c, result)
		return
	}
	rows, _ := result.Result.([]map[string]any)
	lastBuild := time.Now().Format(time.RFC1123Z)
	if len(rows) > 0 {
		lastBuild = formatRSSDate(model.Lookup(rows[0], "CREATE_TIME"))
	}
	domain := html.EscapeString(h.service.Config.External.DomainName)
	var output strings.Builder
	output.WriteString(xml.Header)
	output.WriteString(`<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom"><channel>`)
	fmt.Fprintf(&output, `<atom:link href="%s/rss.xml" rel="self" type="application/rss+xml"></atom:link>`, domain)
	output.WriteString(`<title>YnsStudy</title><link>` + domain + `</link><description>YnsStudy - 技术分享、学习记录与生活思考</description><language>zh-CN</language>`)
	output.WriteString(`<lastBuildDate>` + html.EscapeString(lastBuild) + `</lastBuildDate>`)
	for _, row := range rows {
		title := model.StringValue(row, "BLOG_TITLE")
		if title == "" {
			title = "文章 ID: " + model.StringValue(row, "GUID")
		}
		link := h.service.Config.External.DomainName + "/oneBlog/" + model.StringValue(row, "GUID")
		// RSS 只保留摘要并明确指向原文，避免完整正文形成一个比文章页更容易收录的重复页面。
		excerpt := truncateRunes(collapseWhitespace(plainTextFromHTML(model.StringValue(row, "MAINTEXT"))), 240)
		fmt.Fprintf(&output, `<item><title><![CDATA[%s]]></title><link>%s</link><guid isPermaLink="true">%s</guid><pubDate>%s</pubDate><description><![CDATA[%s]]></description></item>`,
			safeCDATA(title), html.EscapeString(link), html.EscapeString(link), html.EscapeString(formatRSSDate(model.Lookup(row, "CREATE_TIME"))), safeCDATA(excerpt))
	}
	output.WriteString(`</channel></rss>`)
	c.Header("X-Robots-Tag", "noindex, follow")
	c.Header("Cache-Control", "public, max-age=300")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(output.String()))
}

func formatDate(value any) string {
	if parsed := parseControllerTime(value); !parsed.IsZero() {
		return parsed.Format("2006-01-02")
	}
	text := fmt.Sprint(value)
	if len(text) >= 10 {
		return text[:10]
	}
	return text
}

func formatRSSDate(value any) string {
	if parsed := parseControllerTime(value); !parsed.IsZero() {
		return parsed.Format(time.RFC1123Z)
	}
	return fmt.Sprint(value)
}

func parseControllerTime(value any) time.Time {
	if parsed, ok := value.(time.Time); ok {
		return parsed
	}
	text := fmt.Sprint(value)
	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339, "2006-01-02T15:04:05"} {
		if parsed, err := time.ParseInLocation(layout, strings.Split(text, ".")[0], time.Local); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func safeCDATA(value string) string { return strings.ReplaceAll(value, "]]>", "]]]]><![CDATA[>") }

func parseIntOr(value string, fallback int) int {
	var result int
	if _, err := fmt.Sscan(value, &result); err != nil {
		return fallback
	}
	return result
}
