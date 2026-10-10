package service

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

func (s *Service) AddCommunity(c *gin.Context, community map[string]any) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	// 只接受帖子编号和内容；作者以登录账号为准，置顶只能由超级管理员通过 setTopCommunity 设置
	community = pickFields(community, "GUID", "TEXT")
	if model.Lookup(community, "TEXT") != nil {
		// 帖子是 Markdown 转出的 HTML，保存前过滤脚本和事件属性
		community["TEXT"] = sanitizeRichText(model.StringValue(community, "TEXT"))
	}
	community["USERCODE"] = actor.Code()
	community["USERNAME"] = actor.User.Name
	ctx := contextOf(c)
	if s.hasSearchText(ctx, "communityInfo") {
		community[searchTextColumn] = htmlToSearchText(model.StringValue(community, "TEXT"))
	}
	return stripSearchTextResult(s.SaveAll(ctx, "add", "communityInfo", []map[string]any{community}, "GUID"))
}

// DeleteCommunity 的规则与前端一致：超级管理员可删除任何帖子；置顶帖只有超级管理员能删；
// 非置顶帖作者本人可删，管理员可删除非超级管理员发布的帖子。
func (s *Service) DeleteCommunity(c *gin.Context, guid string) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	row, failure := s.loadRow(c, "communityInfo", guid)
	if failure != nil {
		return *failure
	}
	isTop := model.StringValue(row, "ISTOP") == "1"
	if !actor.IsSuper && (isTop || !actor.CanManage(model.StringValue(row, "USERCODE"))) {
		return model.Failure(msgForbidden)
	}
	ctx := contextOf(c)
	_, _, err := s.Repo.ExecuteBatch(ctx,
		[]string{"DELETE FROM communityInfo WHERE GUID = ?", "DELETE FROM communityComment WHERE COMMUNITYID = ?"},
		[][]any{{guid}, {guid}}, false,
	)
	if err != nil {
		return dbFailure("删除社区内容", err)
	}
	return model.Success("删除成功")
}

func (s *Service) SetTopCommunity(c *gin.Context, guid, isTop string) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	var value any
	if isTop == "1" {
		value = "1"
	} else if isTop == "0" {
		value = nil
	} else {
		return model.Failure("isTop 只允许为 0 或 1")
	}
	return s.ExecuteSQL(ctx, "UPDATE communityInfo SET ISTOP = ? WHERE GUID = ?", []any{value, guid})
}

func (s *Service) GetAllCommunities(ctx context.Context, page, pageSize int, keyword string) model.Result {
	page, pageSize = normalizePage(page, pageSize)
	where := ""
	args := make([]any, 0, 4)
	countArgs := make([]any, 0, 2)
	if strings.TrimSpace(keyword) != "" {
		where = " WHERE (b.TITLE LIKE ? OR " + s.searchField(ctx, "communityInfo", "b") + " LIKE ?)"
		like := likePattern(keyword)
		args = append(args, like, like)
		countArgs = append(countArgs, like, like)
	}
	args = append(args, pageSize, (page-1)*pageSize)
	listQuery := "SELECT b.*, u.AVATAR FROM communityInfo b LEFT JOIN userInfo u ON b.USERCODE = u.CODE" + where + " ORDER BY b.ISTOP DESC, b.CREATE_TIME DESC LIMIT ? OFFSET ?"
	rows, err := s.Repo.Query(ctx, listQuery, args...)
	if err != nil {
		return dbFailure("查询社区内容", err)
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM communityInfo b"+where, countArgs...)
	if err != nil {
		return dbFailure("统计社区内容", err)
	}
	return model.Success(map[string]any{"total": firstCount(counts), "data": stripSearchText(rows)})
}

func (s *Service) AddCommunityComment(c *gin.Context, comment map[string]any) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	comment = pickFields(comment, "GUID", "COMMUNITYID", "SUPERGUID", "TEXT")
	comment["USERCODE"] = actor.Code()
	comment["USERNAME"] = actor.User.Name
	return s.SaveAll(contextOf(c), "add", "communityComment", []map[string]any{comment}, "GUID")
}

func (s *Service) GetCommunityComments(ctx context.Context, communityID string) model.Result {
	query := `SELECT bc.*, ui.AVATAR
FROM communityComment bc
LEFT JOIN userInfo ui ON bc.USERCODE = ui.CODE
WHERE bc.COMMUNITYID = ?
ORDER BY bc.CREATE_TIME DESC`
	return s.SelectList(ctx, query, []any{communityID})
}
