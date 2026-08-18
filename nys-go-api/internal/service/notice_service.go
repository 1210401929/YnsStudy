package service

import (
	"context"

	"nys-go-api/internal/model"
)

func (s *Service) AddNotice(ctx context.Context, sender, receiver, noticeType, execute, remark string) model.Result {
	if sender == "" || receiver == "" || noticeType == "" || execute == "" {
		return model.Failure("传递参数不全,请传递完整参数结构!")
	}
	result := s.SaveAll(ctx, "add", "noticeInfo", []map[string]any{{
		"SENDUSERCODE": sender, "RECEIVERUSERCODE": receiver, "TYPE": noticeType, "EXECUTE": execute, "REMARK": remark,
	}}, "GUID")
	if result.IsError {
		return model.Failure("新增消息失败,请检查!" + result.ErrMsg)
	}
	return model.Success("新增消息成功!")
}

func (s *Service) GetNotices(ctx context.Context, userCode string) model.Result {
	return s.SelectList(ctx, "SELECT * FROM noticeInfo WHERE RECEIVERUSERCODE = ? ORDER BY CREATE_TIME DESC", []any{userCode})
}

func (s *Service) ReadNotice(ctx context.Context, guid string) model.Result {
	return s.ExecuteSQL(ctx, "DELETE FROM noticeInfo WHERE GUID = ?", []any{guid})
}

func (s *Service) ReadAllNotices(ctx context.Context, userCode string) model.Result {
	return s.ExecuteSQL(ctx, "DELETE FROM noticeInfo WHERE RECEIVERUSERCODE = ?", []any{userCode})
}
