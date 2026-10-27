package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

func (s *Service) FollowUser(c *gin.Context, followCode, followName string) model.Result {
	user, err := s.CurrentUser(c)
	if err != nil {
		return model.Failure("用户未登录!")
	}
	return s.SaveAll(contextOf(c), "add", "userFollow", []map[string]any{{
		"USERCODE": user.Code, "USERNAME": user.Name, "FOLLOWUSERCODE": followCode, "FOLLOWUSERNAME": followName,
	}}, "GUID")
}

func (s *Service) UnfollowUser(c *gin.Context, followCode string) model.Result {
	user, err := s.CurrentUser(c)
	if err != nil {
		return model.Failure("用户未登录!")
	}
	return s.ExecuteSQL(contextOf(c), "DELETE FROM userFollow WHERE FOLLOWUSERCODE = ? AND USERCODE = ?", []any{followCode, user.Code})
}

func (s *Service) GetFollowUsers(c *gin.Context, userCode string, countOnly bool) model.Result {
	currentCode := ""
	if current, err := s.CurrentUser(c); err == nil {
		currentCode = current.Code
	}
	ctx := contextOf(c)
	result := make(map[string]any)
	if countOnly {
		following, err := s.Repo.Query(ctx, "SELECT COUNT(1) AS TOTAL FROM userFollow WHERE USERCODE = ?", userCode)
		if err != nil {
			return dbFailure("统计关注数", err)
		}
		followers, err := s.Repo.Query(ctx, "SELECT COUNT(1) AS TOTAL FROM userFollow WHERE FOLLOWUSERCODE = ?", userCode)
		if err != nil {
			return dbFailure("统计粉丝数", err)
		}
		result["followingNum"] = firstCount(following)
		result["followersNum"] = firstCount(followers)
		if currentCode != "" {
			rows, err := s.Repo.Query(ctx, "SELECT COUNT(1) AS IS_FOLLOW FROM userFollow WHERE USERCODE = ? AND FOLLOWUSERCODE = ?", currentCode, userCode)
			if err != nil {
				return dbFailure("查询关注状态", err)
			}
			if len(rows) > 0 {
				result["isFollowingCount"] = model.Lookup(rows[0], "IS_FOLLOW")
			}
		}
		return model.Success(result)
	}

	following, err := s.Repo.Query(ctx, `SELECT u.CODE, u.NAME, u.REMARK, u.ROLE, u.AVATAR
FROM userInfo u LEFT JOIN userFollow f ON u.CODE = f.FOLLOWUSERCODE
WHERE f.USERCODE = ? ORDER BY f.CREATE_TIME`, userCode)
	if err != nil {
		return dbFailure("查询关注列表", err)
	}
	followers, err := s.Repo.Query(ctx, `SELECT u.CODE, u.NAME, u.REMARK, u.ROLE, u.AVATAR
FROM userInfo u LEFT JOIN userFollow f ON u.CODE = f.USERCODE
WHERE f.FOLLOWUSERCODE = ? ORDER BY f.CREATE_TIME`, userCode)
	if err != nil {
		return dbFailure("查询粉丝列表", err)
	}
	result["followingUser"] = following
	result["followersUser"] = followers
	return model.Success(result)
}

func (s *Service) GetBlogResourceCommunityByUser(ctx context.Context, userCode string) model.Result {
	blogs := s.GetBlogsByUser(ctx, userCode, true)
	communities := stripSearchTextResult(s.SelectList(ctx, "SELECT * FROM communityInfo WHERE USERCODE = ? ORDER BY CREATE_TIME DESC", []any{userCode}))
	resources := s.GetFilesByUser(ctx, userCode)
	return model.Success(map[string]any{
		"blog":      blogs.Result,
		"community": communities.Result,
		"resource":  resources.Result,
	})
}

// GetBlogAndCommunityByUser 是作者主页的内容列表：公开文章和随记按时间倒序混排。
// contentType 为 blog 时只看文章，community 时只看随记，其他值看全部（在后台筛选，分页才准确）。
func (s *Service) GetBlogAndCommunityByUser(ctx context.Context, userCode string, page, pageSize int, keyword, contentType string) model.Result {
	page, pageSize = normalizePage(page, pageSize)
	if pageSize > 50 {
		pageSize = 50
	}
	like := likePattern(keyword)
	blogText := s.searchField(ctx, "blogInfo", "")
	communityText := s.searchField(ctx, "communityInfo", "")
	blogPart := `SELECT GUID, BLOG_TITLE, MAINTEXT, USERCODE, USERNAME, CREATE_TIME, 'blog' AS TYPE
FROM blogInfo WHERE USERCODE = ? AND BLOG_TYPE = 'public' AND (BLOG_TITLE LIKE ? OR ` + blogText + ` LIKE ?)`
	communityPart := `SELECT GUID, TITLE AS BLOG_TITLE, TEXT AS MAINTEXT, USERCODE, USERNAME, CREATE_TIME, 'community' AS TYPE
FROM communityInfo WHERE USERCODE = ? AND (TITLE LIKE ? OR ` + communityText + ` LIKE ?)`
	parts := make([]string, 0, 2)
	args := make([]any, 0, 8)
	if contentType != "community" {
		parts = append(parts, blogPart)
		args = append(args, userCode, like, like)
	}
	if contentType != "blog" {
		parts = append(parts, communityPart)
		args = append(args, userCode, like, like)
	}
	union := strings.Join(parts, "\nUNION ALL\n")
	rows, err := s.Repo.Query(ctx, "SELECT * FROM ("+union+") combined ORDER BY CREATE_TIME DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), pageSize, (page-1)*pageSize)...)
	if err != nil {
		return dbFailure("查询用户内容", err)
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM ("+union+") combined", args...)
	if err != nil {
		return dbFailure("统计用户内容", err)
	}
	return model.Success(map[string]any{"total": firstCount(counts), "data": replaceTextWithSummary(rows, "MAINTEXT")})
}

// authorHotBlogLimit 是作者主页“热门文章”的条数
const authorHotBlogLimit = 5

// GetAuthorOverview 是作者主页的概览：创作数据（公开文章数、随记数、公开文章总阅读）、
// 按阅读量排的热门文章，以及作者有公开文章的系列。三块互不依赖，某一块查询失败时返回空值，不影响其他内容。
func (s *Service) GetAuthorOverview(ctx context.Context, userCode string) model.Result {
	userCode = strings.TrimSpace(userCode)
	if userCode == "" {
		return model.Failure("用户编码出错!")
	}
	stats := map[string]any{"articles": int64(0), "notes": int64(0), "views": int64(0)}
	if rows, err := s.Repo.Query(ctx, `SELECT
(SELECT COUNT(*) FROM blogInfo WHERE USERCODE = ? AND BLOG_TYPE = 'public') AS ARTICLES,
(SELECT COUNT(*) FROM communityInfo WHERE USERCODE = ?) AS NOTES,
(SELECT COALESCE(SUM(VIEW_PAGE), 0) FROM blogInfo WHERE USERCODE = ? AND BLOG_TYPE = 'public') AS VIEWS`, userCode, userCode, userCode); err == nil && len(rows) > 0 {
		stats["articles"] = model.Int64Value(rows[0], "ARTICLES")
		stats["notes"] = model.Int64Value(rows[0], "NOTES")
		stats["views"] = model.Int64Value(rows[0], "VIEWS")
	}
	hotBlogs, err := s.Repo.Query(ctx, `SELECT GUID, BLOG_TITLE, VIEW_PAGE, CREATE_TIME FROM blogInfo
WHERE USERCODE = ? AND BLOG_TYPE = 'public'
ORDER BY COALESCE(VIEW_PAGE, 0) DESC, CREATE_TIME DESC LIMIT ?`, userCode, authorHotBlogLimit)
	if err != nil {
		hotBlogs = []map[string]any{}
	}
	series, err := s.PublicSeriesList(ctx, userCode, "hot", 10)
	if err != nil {
		series = []map[string]any{}
	}
	for _, row := range series {
		delete(row, "USERCODE")
	}
	return model.Success(map[string]any{"stats": stats, "hotBlogs": hotBlogs, "series": series})
}

// personInfoFields 是个人主页允许设置的字段（背景图、背景音乐）。
var personInfoFields = map[string]bool{"BGIMAGEURL": true, "BGMUSICURL": true}

// SetPersonInfo 只能修改当前登录用户自己的主页设置。
func (s *Service) SetPersonInfo(c *gin.Context, userCode, fieldName, fieldValue string) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	if userCode != "" && userCode != actor.Code() {
		return model.Failure(msgForbidden)
	}
	userCode = actor.Code()
	fieldName = strings.ToUpper(strings.TrimSpace(fieldName))
	if !personInfoFields[fieldName] {
		return model.Failure("不支持修改该字段")
	}
	ctx := contextOf(c)
	rows, err := s.Repo.Query(ctx, "SELECT GUID FROM personInfo WHERE USERCODE = ? LIMIT 1", userCode)
	if err != nil {
		return dbFailure("查询个人主页配置", err)
	}
	data := map[string]any{"USERCODE": userCode, fieldName: fieldValue}
	operation := "add"
	if len(rows) > 0 {
		operation = "edit"
		data["GUID"] = model.StringValue(rows[0], "GUID")
	}
	return s.SaveAll(ctx, operation, "personInfo", []map[string]any{data}, "GUID")
}

func (s *Service) GetPersonInfo(ctx context.Context, userCode string) model.Result {
	return s.SelectList(ctx, "SELECT * FROM personInfo WHERE USERCODE = ?", []any{userCode})
}

func asBool(value any) bool {
	return strings.EqualFold(fmt.Sprint(value), "true")
}
