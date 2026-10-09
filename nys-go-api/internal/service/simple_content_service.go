package service

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

func (s *Service) AddAnnouncement(c *gin.Context, announcement map[string]any) model.Result {
	if fmt.Sprint(announcement["TEXT"]) == "" {
		return model.Failure("未传递有效内容:text")
	}
	actor, failure := s.requireSuperAdmin(c)
	if failure != nil {
		return *failure
	}
	announcement["userCode"] = actor.Code()
	announcement["userName"] = actor.User.Name
	return s.SaveAll(contextOf(c), "add", "announcementInfo", []map[string]any{announcement}, "GUID")
}

func (s *Service) EditAnnouncement(c *gin.Context, announcement map[string]any) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	if fmt.Sprint(announcement["TEXT"]) == "" {
		return model.Failure("未传递有效内容:text")
	}
	return s.SaveAll(ctx, "edit", "announcementInfo", []map[string]any{announcement}, "GUID")
}

func (s *Service) DeleteAnnouncement(c *gin.Context, guid string) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	return s.ExecuteSQL(contextOf(c), "DELETE FROM announcementInfo WHERE GUID = ?", []any{guid})
}

func (s *Service) GetAllAnnouncements(ctx context.Context) model.Result {
	return s.SelectList(ctx, "SELECT * FROM announcementInfo ORDER BY TYPE, CREATE_TIME DESC", nil)
}

func (s *Service) GetAnnouncementsByType(ctx context.Context, announcementType string) model.Result {
	return s.SelectList(ctx, "SELECT * FROM announcementInfo WHERE ISENABLE = '1' AND TYPE = ? ORDER BY TYPE, CREATE_TIME DESC", []any{announcementType})
}

// friendLinkFields 是友链允许由客户端填写的字段。
var friendLinkFields = []string{"NAME", "LINK", "AVATAR", "REMARK", "LINK_TYPE"}

func (s *Service) AddFriendLink(c *gin.Context, value map[string]any) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	data := pickFields(value, append([]string{"GUID"}, friendLinkFields...)...)
	data["USERCODE"] = actor.Code()
	data["USERNAME"] = actor.User.Name
	return s.SaveAll(contextOf(c), "add", "friendLinkInfo", []map[string]any{data}, "GUID")
}

// requireFriendLinkManager：友链发布者本人或任意管理员可以修改、删除（与前端 isAdmin || 本人 的显示规则一致）。
func (s *Service) requireFriendLinkManager(c *gin.Context, guid string) *model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return failure
	}
	row, failure := s.loadRow(c, "friendLinkInfo", guid)
	if failure != nil {
		return failure
	}
	if !actor.IsAdmin && model.StringValue(row, "USERCODE") != actor.Code() {
		result := model.Failure(msgForbidden)
		return &result
	}
	return nil
}

func (s *Service) UpdateFriendLink(c *gin.Context, value map[string]any) model.Result {
	guid := model.StringValue(value, "GUID")
	if failure := s.requireFriendLinkManager(c, guid); failure != nil {
		return *failure
	}
	data := pickFields(value, friendLinkFields...)
	data["GUID"] = guid
	return s.SaveAll(contextOf(c), "edit", "friendLinkInfo", []map[string]any{data}, "GUID")
}

func (s *Service) DeleteFriendLink(c *gin.Context, guid string) model.Result {
	if failure := s.requireFriendLinkManager(c, guid); failure != nil {
		return *failure
	}
	return s.ExecuteSQL(contextOf(c), "DELETE FROM friendLinkInfo WHERE GUID = ?", []any{guid})
}

func (s *Service) GetFriendLink(ctx context.Context, guid string) model.Result {
	return s.SelectList(ctx, "SELECT * FROM friendLinkInfo WHERE GUID = ?", []any{guid})
}

func (s *Service) GetFriendLinks(ctx context.Context) model.Result {
	return s.SelectList(ctx, "SELECT * FROM friendLinkInfo ORDER BY CREATE_TIME", nil)
}
