package service

import (
	"context"
	"log"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

func (s *Service) AddFileInfo(c *gin.Context, info map[string]any) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	// 只接受这几个字段；文件地址必须是本人上传的文件，否则删除资源时会顺带删掉别人的文件
	data := pickFields(info, "GUID", "ORIGINALFILENAME", "REMARK")
	viewURL := s.uploadKey(model.StringValue(info, "FILEVIEWURL"))
	if viewURL == "" {
		return model.Failure(msgUploadInvalidURL)
	}
	if s.uploadRecordReady(ctx) {
		if _, failure := s.checkUploadOwner(ctx, viewURL, actor.Code()); failure != nil {
			return *failure
		}
	}
	data["FILEVIEWURL"] = viewURL
	data["SAVEDFILENAME"] = path.Base(viewURL)
	// 以前这里保存服务器上的绝对路径并公开返回，现在不再保存
	data["FILEREALURL"] = ""
	data["USERCODE"] = actor.Code()
	data["USERNAME"] = actor.User.Name
	return s.SaveAll(ctx, "add", "FILEINFO", []map[string]any{data}, "GUID")
}

// DeleteFileInfo 删除资源记录和对应文件。文件地址从数据库读取，不使用请求传入的地址，
// 否则可以借此删除服务器上任意上传文件。
func (s *Service) DeleteFileInfo(c *gin.Context, guid string) model.Result {
	row, failure := s.requireOwnerOrAdmin(c, "fileInfo", guid)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	fileURL := strings.TrimSpace(model.StringValue(row, "FILEVIEWURL"))
	if fileURL == "" {
		return model.Failure("路径不存在或不正确!")
	}
	result := s.ExecuteSQL(ctx, "DELETE FROM fileInfo WHERE GUID = ?", []any{guid})
	if result.IsError {
		return model.Failure("删除失败!" + result.ErrMsg)
	}
	s.logModeration(c, nil, row, AdminActionDeleteResource, model.StringValue(row, "ORIGINALFILENAME"), "在资源页删除了他人的资源")
	// 资源记录已删除；文件只在确实由资源作者上传时才删除，删不掉只记日志，不影响结果
	if fileResult := s.DeleteOwnedUpload(ctx, fileURL, model.StringValue(row, "USERCODE")); fileResult.IsError && fileResult.ErrMsg != "文件或路径不存在!" {
		log.Printf("删除资源 %s 的文件未执行 %s: %s", guid, fileURL, fileResult.ErrMsg)
	}
	return model.Success("删除成功!")
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

func (s *Service) UpdateFileInfo(c *gin.Context, guid, originalName, remark string) model.Result {
	row, failure := s.requireOwnerOrAdmin(c, "fileInfo", guid)
	if failure != nil {
		return *failure
	}
	result := s.ExecuteSQL(contextOf(c), "UPDATE fileInfo SET ORIGINALFILENAME = ?, REMARK = ? WHERE GUID = ?", []any{originalName, remark, guid})
	if !result.IsError {
		s.logModeration(c, nil, row, AdminActionEditResource, model.StringValue(row, "ORIGINALFILENAME"), "在资源页修改为："+originalName)
	}
	return result
}

func (s *Service) IncrementFileDownloads(ctx context.Context, guid string) model.Result {
	return s.ExecuteSQL(ctx, "UPDATE fileInfo SET DOWNNUM = COALESCE(DOWNNUM, 0) + 1 WHERE GUID = ?", []any{guid})
}
