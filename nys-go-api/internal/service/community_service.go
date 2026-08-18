package service

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

func (s *Service) AddCommunity(c *gin.Context, community map[string]any) model.Result {
	if _, err := s.CurrentUser(c); err != nil {
		return model.Failure("用户未登录!")
	}
	return s.SaveAll(contextOf(c), "add", "communityInfo", []map[string]any{community}, "GUID")
}

func (s *Service) DeleteCommunity(ctx context.Context, guid string) model.Result {
	_, _, err := s.Repo.ExecuteBatch(ctx,
		[]string{"DELETE FROM communityInfo WHERE GUID = ?", "DELETE FROM communityComment WHERE COMMUNITYID = ?"},
		[][]any{{guid}, {guid}}, false,
	)
	if err != nil {
		return dbFailure("删除社区内容", err)
	}
	return model.Success("删除成功")
}

func (s *Service) SetTopCommunity(ctx context.Context, guid, isTop string) model.Result {
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
		where = " WHERE (b.TITLE LIKE ? OR b.TEXT LIKE ?)"
		like := "%" + strings.TrimSpace(keyword) + "%"
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
	return model.Success(map[string]any{"total": firstCount(counts), "data": rows})
}

func (s *Service) AddCommunityComment(c *gin.Context, comment map[string]any) model.Result {
	if _, err := s.CurrentUser(c); err != nil {
		return model.Failure("用户未登录!")
	}
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
