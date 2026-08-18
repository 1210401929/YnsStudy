package service

import (
	"context"
	"strings"

	"nys-go-api/internal/model"
)

func (s *Service) AddFileInfo(ctx context.Context, info map[string]any) model.Result {
	return s.SaveAll(ctx, "add", "FILEINFO", []map[string]any{info}, "GUID")
}

func (s *Service) DeleteFileInfo(ctx context.Context, guid, fileURL string) model.Result {
	if strings.TrimSpace(fileURL) == "" {
		return model.Failure("路径不存在或不正确!")
	}
	result := s.ExecuteSQL(ctx, "DELETE FROM fileInfo WHERE GUID = ?", []any{guid})
	if result.IsError {
		return model.Failure("删除失败!" + result.ErrMsg)
	}
	return s.DeleteUploadedFile(fileURL)
}

func (s *Service) GetAllFiles(ctx context.Context, page, pageSize int, keyword string) model.Result {
	page, pageSize = normalizePage(page, pageSize)
	where := ""
	listArgs := make([]any, 0, 5)
	countArgs := make([]any, 0, 3)
	if strings.TrimSpace(keyword) != "" {
		where = " WHERE (f.ORIGINALFILENAME LIKE ? OR f.REMARK LIKE ? OR f.USERNAME LIKE ?)"
		like := "%" + strings.TrimSpace(keyword) + "%"
		listArgs = append(listArgs, like, like, like)
		countArgs = append(countArgs, like, like, like)
	}
	listArgs = append(listArgs, pageSize, (page-1)*pageSize)
	rows, err := s.Repo.Query(ctx, "SELECT f.* FROM fileInfo f"+where+" ORDER BY f.CREATE_TIME DESC LIMIT ? OFFSET ?", listArgs...)
	if err != nil {
		return dbFailure("查询资源", err)
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM fileInfo f"+where, countArgs...)
	if err != nil {
		return dbFailure("统计资源", err)
	}
	return model.Success(map[string]any{"total": firstCount(counts), "data": rows})
}

func (s *Service) GetFilesByUser(ctx context.Context, userCode string) model.Result {
	return s.SelectList(ctx, "SELECT * FROM fileInfo WHERE USERCODE = ? ORDER BY CREATE_TIME DESC", []any{userCode})
}

func (s *Service) GetFileByID(ctx context.Context, guid string) model.Result {
	return s.SelectList(ctx, "SELECT * FROM fileInfo WHERE GUID = ?", []any{guid})
}

func (s *Service) UpdateFileInfo(ctx context.Context, guid, originalName, remark string) model.Result {
	return s.ExecuteSQL(ctx, "UPDATE fileInfo SET ORIGINALFILENAME = ?, REMARK = ? WHERE GUID = ?", []any{originalName, remark, guid})
}

func (s *Service) IncrementFileDownloads(ctx context.Context, guid string) model.Result {
	return s.ExecuteSQL(ctx, "UPDATE fileInfo SET DOWNNUM = COALESCE(DOWNNUM, 0) + 1 WHERE GUID = ?", []any{guid})
}
