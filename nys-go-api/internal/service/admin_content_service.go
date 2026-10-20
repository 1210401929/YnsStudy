package service

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

// 超级管理员后台：文章、社区帖子管理，以及 SEO 小工具。

const adminListMaxPageSize = 50

func normalizeAdminPage(page, pageSize int) (int, int) {
	page, pageSize = normalizePage(page, pageSize)
	if pageSize > adminListMaxPageSize {
		pageSize = adminListMaxPageSize
	}
	return page, pageSize
}

// AdminContentQuery 是文章、帖子列表的筛选条件。
type AdminContentQuery struct {
	Page, PageSize int
	Keyword        string
	Type           string // 文章：public / privacy；帖子：top
}

// GetAdminArticles 返回全站文章（包括私密文章），按发布时间倒序。
func (s *Service) GetAdminArticles(c *gin.Context, query AdminContentQuery) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	page, pageSize := normalizeAdminPage(query.Page, query.PageSize)
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 4)
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		conditions = append(conditions, "(BLOG_TITLE LIKE ? OR USERNAME LIKE ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}
	switch query.Type {
	case "public":
		conditions = append(conditions, "BLOG_TYPE = 'public'")
	case "privacy":
		conditions = append(conditions, "(BLOG_TYPE <> 'public' OR BLOG_TYPE IS NULL)")
	}
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM blogInfo"+where, args...)
	if err != nil {
		return dbFailure("统计文章", err)
	}
	rows, err := s.Repo.Query(ctx, `SELECT GUID, BLOG_TITLE, BLOG_TYPE, USERCODE, USERNAME, COALESCE(VIEW_PAGE, 0) AS VIEW_PAGE, CREATE_TIME
FROM blogInfo`+where+` ORDER BY CREATE_TIME DESC LIMIT ? OFFSET ?`, append(args, pageSize, (page-1)*pageSize)...)
	if err != nil {
		return dbFailure("查询文章", err)
	}
	for _, row := range rows {
		row["USERNAME"] = model.PublicDisplayName(model.StringValue(row, "USERNAME"))
	}
	return model.Success(map[string]any{"total": firstCount(counts), "data": rows})
}

// SetArticleVisibility 把文章设为公开或私密，并通知搜索引擎重新抓取。
func (s *Service) SetArticleVisibility(c *gin.Context, guid, blogType string) model.Result {
	actor, failure := s.requireSuperAdmin(c)
	if failure != nil {
		return *failure
	}
	if blogType != "public" && blogType != "privacy" {
		return model.Failure("文章类型只能是 public 或 privacy")
	}
	row, failure := s.loadRow(c, "blogInfo", guid)
	if failure != nil {
		return *failure
	}
	if _, err := s.Repo.Exec(contextOf(c), "UPDATE blogInfo SET BLOG_TYPE = ? WHERE GUID = ?", blogType, guid); err != nil {
		return dbFailure("修改文章可见性", err)
	}
	s.NotifyBlogChanged(guid, model.StringValue(row, "USERCODE"))
	label := map[string]string{"public": "设为公开", "privacy": "设为私密"}[blogType]
	s.recordAdminAction(c, actor, AdminActionArticleType, model.StringValue(row, "BLOG_TITLE"), label+"（文章 "+guid+"）")
	return model.Success("已" + label)
}

// AdminDeleteArticle 删除文章（含评论、点赞和正文图片），并记录操作日志。
func (s *Service) AdminDeleteArticle(c *gin.Context, guid string) model.Result {
	actor, failure := s.requireSuperAdmin(c)
	if failure != nil {
		return *failure
	}
	row, failure := s.loadRow(c, "blogInfo", guid)
	if failure != nil {
		return *failure
	}
	result := s.DeleteBlog(c, guid)
	if !result.IsError {
		s.recordAdminAction(c, actor, AdminActionDeleteArticle, model.StringValue(row, "BLOG_TITLE"),
			"删除文章 "+guid+"，作者 "+model.StringValue(row, "USERNAME"))
	}
	return result
}

// GetAdminCommunityPosts 返回社区帖子，置顶在前，附带评论数。
func (s *Service) GetAdminCommunityPosts(c *gin.Context, query AdminContentQuery) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	page, pageSize := normalizeAdminPage(query.Page, query.PageSize)
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 4)
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		conditions = append(conditions, "(p.TITLE LIKE ? OR p.TEXT LIKE ? OR p.USERNAME LIKE ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like, like)
	}
	if query.Type == "top" {
		conditions = append(conditions, "p.ISTOP = '1'")
	}
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM communityInfo p"+where, args...)
	if err != nil {
		return dbFailure("统计社区帖子", err)
	}
	rows, err := s.Repo.Query(ctx, `SELECT p.GUID, p.TITLE, p.TEXT, p.USERCODE, p.USERNAME, p.ISTOP, p.CREATE_TIME,
(SELECT COUNT(*) FROM communityComment cc WHERE cc.COMMUNITYID = p.GUID) AS COMMENT_COUNT
FROM communityInfo p`+where+` ORDER BY p.ISTOP DESC, p.CREATE_TIME DESC LIMIT ? OFFSET ?`, append(args, pageSize, (page-1)*pageSize)...)
	if err != nil {
		return dbFailure("查询社区帖子", err)
	}
	for _, row := range rows {
		excerpt, _ := summarizeHTML(model.StringValue(row, "TEXT"), 120)
		row["TEXT"] = excerpt
		row["USERNAME"] = model.PublicDisplayName(model.StringValue(row, "USERNAME"))
	}
	return model.Success(map[string]any{"total": firstCount(counts), "data": rows})
}

// AdminSetTopPost 置顶或取消置顶社区帖子。
func (s *Service) AdminSetTopPost(c *gin.Context, guid, isTop string) model.Result {
	// 先校验权限再读取帖子，避免未登录者借返回信息判断帖子是否存在
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	row, failure := s.loadRow(c, "communityInfo", guid)
	if failure != nil {
		return *failure
	}
	result := s.SetTopCommunity(c, guid, isTop)
	if !result.IsError {
		label := map[bool]string{true: "置顶", false: "取消置顶"}[isTop == "1"]
		s.recordAdminAction(c, nil, AdminActionTopPost, postLabel(row), label)
	}
	return result
}

// AdminDeletePost 删除社区帖子及其评论。
func (s *Service) AdminDeletePost(c *gin.Context, guid string) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	row, failure := s.loadRow(c, "communityInfo", guid)
	if failure != nil {
		return *failure
	}
	result := s.DeleteCommunity(c, guid)
	if !result.IsError {
		s.recordAdminAction(c, nil, AdminActionDeletePost, postLabel(row), "删除帖子，作者 "+model.StringValue(row, "USERNAME"))
	}
	return result
}

func postLabel(row map[string]any) string {
	if title := strings.TrimSpace(model.StringValue(row, "TITLE")); title != "" {
		return title
	}
	excerpt, _ := summarizeHTML(model.StringValue(row, "TEXT"), 40)
	return excerpt
}

// ---------------- SEO 小工具 ----------------

// GetSeoStatus 返回 IndexNow 是否启用、sitemap 中各类地址的数量和常用地址。
func (s *Service) GetSeoStatus(c *gin.Context) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	articles := s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM blogInfo WHERE BLOG_TYPE = 'public'")
	var authors any
	if rows, err := s.GetPublicAuthorNumbers(ctx); err == nil {
		authors = int64(len(rows))
	}
	domain := strings.TrimRight(s.Config.External.DomainName, "/")
	return model.Success(map[string]any{
		"indexNowEnabled": s.IndexNowEnabled(),
		"domain":          domain,
		"sitemap": map[string]any{
			"articles": articles,
			"authors":  authors,
			// 首页、归档和 5 个栏目页，与 seo_controller.go 的 sitemap 保持一致
			"pages": 7,
		},
		"links": []string{domain + "/sitemap.xml", domain + "/robots.txt", domain + "/rss.xml", domain + "/archive"},
	})
}

// indexNowMaxManualURLs 限制一次手动推送的数量，避免误把大量地址提交上去
const indexNowMaxManualURLs = 100

// SubmitIndexNowManually 立即推送一组站内地址给 IndexNow，并返回推送结果。
// 接受站内路径（/oneBlog/1）或本站完整地址（https://ynsstudy.cn/oneBlog/1），其他域名会被拒绝。
func (s *Service) SubmitIndexNowManually(c *gin.Context, inputs []string) model.Result {
	actor, failure := s.requireSuperAdmin(c)
	if failure != nil {
		return *failure
	}
	if !s.IndexNowEnabled() {
		return model.Failure("未配置 seo.indexnow_key，无法推送")
	}
	paths, invalid := normalizeSitePaths(inputs, s.Config.External.DomainName)
	if len(invalid) > 0 {
		return model.Failure("以下地址不是本站地址: " + strings.Join(invalid, "、"))
	}
	if len(paths) == 0 {
		return model.Failure("请至少填写一个地址")
	}
	if len(paths) > indexNowMaxManualURLs {
		return model.Failure("一次最多推送 100 个地址")
	}
	ctx, cancel := context.WithTimeout(context.Background(), indexNowTimeout)
	defer cancel()
	if err := s.submitIndexNow(ctx, paths); err != nil {
		return model.Failure("推送失败: " + err.Error())
	}
	s.recordAdminAction(c, actor, AdminActionIndexNowSubmit, strings.Join(paths, " "), "手动推送 "+strconv.Itoa(len(paths))+" 个地址")
	return model.Success(map[string]any{"submitted": paths, "time": time.Now().Format("2006-01-02 15:04:05")})
}

// normalizeSitePaths 把输入整理成去重后的站内路径；不属于本站的地址放进 invalid。
func normalizeSitePaths(inputs []string, domain string) (paths, invalid []string) {
	site, _ := url.Parse(strings.TrimRight(domain, "/"))
	seen := make(map[string]bool)
	for _, input := range inputs {
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		path := input
		if !strings.HasPrefix(input, "/") {
			parsed, err := url.Parse(input)
			if err != nil || site == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || !strings.EqualFold(parsed.Host, site.Host) {
				invalid = append(invalid, input)
				continue
			}
			path = parsed.RequestURI()
		}
		if strings.HasPrefix(path, "//") {
			invalid = append(invalid, input)
			continue
		}
		if !seen[path] {
			seen[path] = true
			paths = append(paths, path)
		}
	}
	return paths, invalid
}
