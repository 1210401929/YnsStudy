package service

import (
	"context"
	"log"
	"path"
	"strings"
	"time"

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

// resourceCategories 是资源页的分类，按文件扩展名划分；与前端 Resources.vue 的 fileCategories 保持一致。
// 不属于任何一类（或没有扩展名）的文件归入 other。
var resourceCategories = []struct {
	Key  string
	Exts []string
}{
	{"doc", []string{"doc", "docx", "xls", "xlsx", "ppt", "pptx", "pdf", "txt", "md", "csv"}},
	{"image", []string{"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg"}},
	{"video", []string{"mp4", "avi", "mov", "wmv", "flv", "mkv", "webm"}},
	{"audio", []string{"mp3", "wav", "flac", "aac", "ogg", "m4a"}},
	{"archive", []string{"zip", "rar", "7z", "tar", "gz"}},
}

// fileExtExpr 是取文件扩展名（小写）的 SQL 表达式；没有点的文件名取到的是整个名字，自然落入“其他”
const fileExtExpr = "LOWER(SUBSTRING_INDEX(f.ORIGINALFILENAME, '.', -1))"

// resourceCategoryCondition 返回某个分类的 SQL 条件和参数；all 或未知分类返回空字符串（不筛选）
func resourceCategoryCondition(category string) (string, []any) {
	placeholders := func(exts []string) (string, []any) {
		args := make([]any, len(exts))
		for i, ext := range exts {
			args[i] = ext
		}
		return strings.TrimSuffix(strings.Repeat("?,", len(exts)), ","), args
	}
	if category == "other" {
		all := make([]string, 0, 40)
		for _, item := range resourceCategories {
			all = append(all, item.Exts...)
		}
		marks, args := placeholders(all)
		return "(f.ORIGINALFILENAME NOT LIKE '%.%' OR " + fileExtExpr + " NOT IN (" + marks + "))", args
	}
	for _, item := range resourceCategories {
		if item.Key == category {
			marks, args := placeholders(item.Exts)
			return fileExtExpr + " IN (" + marks + ")", args
		}
	}
	return "", nil
}

// resourceCategorySelect 返回给每个文件算出分类名的 CASE 表达式，用于按分类统计数量
func resourceCategorySelect() (string, []any) {
	var expr strings.Builder
	args := make([]any, 0, 40)
	expr.WriteString("CASE WHEN f.ORIGINALFILENAME NOT LIKE '%.%' THEN 'other'")
	for _, item := range resourceCategories {
		expr.WriteString(" WHEN " + fileExtExpr + " IN (" + strings.TrimSuffix(strings.Repeat("?,", len(item.Exts)), ",") + ") THEN '" + item.Key + "'")
		for _, ext := range item.Exts {
			args = append(args, ext)
		}
	}
	expr.WriteString(" ELSE 'other' END")
	return expr.String(), args
}

// GetAllFiles 是资源页列表：按关键词和分类筛选（在后台筛选，分页和总数才准确），并返回当前关键词下每个分类的数量
func (s *Service) GetAllFiles(ctx context.Context, page, pageSize int, keyword, category string) model.Result {
	page, pageSize = normalizePage(page, pageSize)
	if pageSize > 50 {
		pageSize = 50
	}
	keywordWhere := ""
	keywordArgs := make([]any, 0, 3)
	if strings.TrimSpace(keyword) != "" {
		keywordWhere = "(f.ORIGINALFILENAME LIKE ? OR f.REMARK LIKE ? OR f.USERNAME LIKE ?)"
		like := likePattern(keyword)
		keywordArgs = append(keywordArgs, like, like, like)
	}
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 45)
	if keywordWhere != "" {
		conditions = append(conditions, keywordWhere)
		args = append(args, keywordArgs...)
	}
	if condition, categoryArgs := resourceCategoryCondition(strings.TrimSpace(category)); condition != "" {
		conditions = append(conditions, condition)
		args = append(args, categoryArgs...)
	}
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}
	rows, err := s.Repo.Query(ctx, "SELECT f.* FROM fileInfo f"+where+" ORDER BY f.CREATE_TIME DESC LIMIT ? OFFSET ?",
		append(append([]any{}, args...), pageSize, (page-1)*pageSize)...)
	if err != nil {
		return dbFailure("查询资源", err)
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM fileInfo f"+where, args...)
	if err != nil {
		return dbFailure("统计资源", err)
	}
	// 各分类数量只受关键词影响，切换分类时数字不变
	categorySelect, categoryArgs := resourceCategorySelect()
	countWhere := ""
	if keywordWhere != "" {
		countWhere = " WHERE " + keywordWhere
	}
	categoryCounts := map[string]int64{"all": 0}
	if grouped, err := s.Repo.Query(ctx, "SELECT "+categorySelect+" AS CATEGORY, COUNT(*) AS total FROM fileInfo f"+countWhere+" GROUP BY CATEGORY",
		append(append([]any{}, categoryArgs...), keywordArgs...)...); err == nil {
		for _, row := range grouped {
			count := firstCount([]map[string]any{row})
			categoryCounts[model.StringValue(row, "CATEGORY")] = count
			categoryCounts["all"] += count
		}
	}
	return model.Success(map[string]any{"total": firstCount(counts), "data": rows, "categoryCounts": categoryCounts})
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

// IncrementFileDownloads 记一次下载。游客也计数；同一 IP 同一个文件 24 小时内只算一次，防止刷下载量。
func (s *Service) IncrementFileDownloads(c *gin.Context, guid string) model.Result {
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return model.Failure(msgNotFound)
	}
	ctx := contextOf(c)
	if s.Cache != nil {
		if count, err := s.Cache.Increment(ctx, "file_down:"+guid+":"+ClientIP(c), 24*time.Hour); err == nil && count > 1 {
			return model.Success("今天已经统计过")
		}
	}
	return s.ExecuteSQL(ctx, "UPDATE fileInfo SET DOWNNUM = COALESCE(DOWNNUM, 0) + 1 WHERE GUID = ?", []any{guid})
}
