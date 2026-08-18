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
	user, err := s.CurrentUser(c)
	if err != nil {
		return model.Failure("用户未登录!")
	}
	announcement["userCode"] = user.Code
	announcement["userName"] = user.Name
	return s.SaveAll(contextOf(c), "add", "announcementInfo", []map[string]any{announcement}, "GUID")
}

func (s *Service) EditAnnouncement(ctx context.Context, announcement map[string]any) model.Result {
	if fmt.Sprint(announcement["TEXT"]) == "" {
		return model.Failure("未传递有效内容:text")
	}
	return s.SaveAll(ctx, "edit", "announcementInfo", []map[string]any{announcement}, "GUID")
}

func (s *Service) DeleteAnnouncement(ctx context.Context, guid string) model.Result {
	return s.ExecuteSQL(ctx, "DELETE FROM announcementInfo WHERE GUID = ?", []any{guid})
}

func (s *Service) GetAllAnnouncements(ctx context.Context) model.Result {
	return s.SelectList(ctx, "SELECT * FROM announcementInfo ORDER BY TYPE, CREATE_TIME DESC", nil)
}

func (s *Service) GetAnnouncementsByType(ctx context.Context, announcementType string) model.Result {
	return s.SelectList(ctx, "SELECT * FROM announcementInfo WHERE ISENABLE = '1' AND TYPE = ? ORDER BY TYPE, CREATE_TIME DESC", []any{announcementType})
}

func (s *Service) AddFriendLink(ctx context.Context, value map[string]any) model.Result {
	return s.SaveAll(ctx, "add", "friendLinkInfo", []map[string]any{value}, "GUID")
}

func (s *Service) UpdateFriendLink(ctx context.Context, value map[string]any) model.Result {
	return s.SaveAll(ctx, "edit", "friendLinkInfo", []map[string]any{value}, "GUID")
}

func (s *Service) DeleteFriendLink(ctx context.Context, guid string) model.Result {
	return s.ExecuteSQL(ctx, "DELETE FROM friendLinkInfo WHERE GUID = ?", []any{guid})
}

func (s *Service) GetFriendLink(ctx context.Context, guid string) model.Result {
	return s.SelectList(ctx, "SELECT * FROM friendLinkInfo WHERE GUID = ?", []any{guid})
}

func (s *Service) GetFriendLinks(ctx context.Context) model.Result {
	return s.SelectList(ctx, "SELECT * FROM friendLinkInfo ORDER BY CREATE_TIME", nil)
}
