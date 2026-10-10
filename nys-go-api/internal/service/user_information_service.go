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
	communities := s.SelectList(ctx, "SELECT * FROM communityInfo WHERE USERCODE = ? ORDER BY CREATE_TIME DESC", []any{userCode})
	resources := s.GetFilesByUser(ctx, userCode)
	return model.Success(map[string]any{
		"blog":      blogs.Result,
		"community": communities.Result,
		"resource":  resources.Result,
	})
}

func (s *Service) GetBlogAndCommunityByUser(ctx context.Context, userCode string, page, pageSize int, keyword string) model.Result {
	page, pageSize = normalizePage(page, pageSize)
	like := "%" + strings.TrimSpace(keyword) + "%"
	query := `SELECT GUID, BLOG_TITLE, MAINTEXT, USERCODE, USERNAME, CREATE_TIME, 'blog' AS TYPE
FROM blogInfo WHERE USERCODE = ? AND BLOG_TYPE = 'public' AND (BLOG_TITLE LIKE ? OR MAINTEXT LIKE ?)
UNION ALL
SELECT GUID, TITLE AS BLOG_TITLE, TEXT AS MAINTEXT, USERCODE, USERNAME, CREATE_TIME, 'community' AS TYPE
FROM communityInfo WHERE USERCODE = ? AND (TITLE LIKE ? OR TEXT LIKE ?)
ORDER BY CREATE_TIME DESC LIMIT ? OFFSET ?`
	args := []any{userCode, like, like, userCode, like, like, pageSize, (page - 1) * pageSize}
	rows, err := s.Repo.Query(ctx, query, args...)
	if err != nil {
		return dbFailure("查询用户内容", err)
	}
	countQuery := `SELECT COUNT(*) AS total FROM (
SELECT 1 FROM blogInfo WHERE USERCODE = ? AND BLOG_TYPE = 'public' AND (BLOG_TITLE LIKE ? OR MAINTEXT LIKE ?)
UNION ALL
SELECT 1 FROM communityInfo WHERE USERCODE = ? AND (TITLE LIKE ? OR TEXT LIKE ?)
) combined`
	counts, err := s.Repo.Query(ctx, countQuery, userCode, like, like, userCode, like, like)
	if err != nil {
		return dbFailure("统计用户内容", err)
	}
	return model.Success(map[string]any{"total": firstCount(counts), "data": replaceTextWithSummary(rows, "MAINTEXT")})
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
