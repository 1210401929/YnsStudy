package service

// 上传文件的归属：上传时把“访问地址 → 上传者账号”记在 uploadRecord 表（迁移 014），
// 删除文件前核对归属，防止有人把别人的文件地址填进头像、文章或资源记录，再借删除操作把别人的文件删掉。

import (
	"context"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"nys-go-api/internal/model"
)

const (
	msgUploadNoRecord   = "这个文件没有上传记录，不能删除"
	msgUploadNotReady   = "文件归属记录还没有启用（需要先执行数据库迁移 014）"
	msgUploadNotOwner   = "只能删除自己上传的文件"
	msgUploadInvalidURL = "文件地址不正确"
)

// uploadRecordReady 迁移 014 执行后才有 uploadRecord 表；没有数据库连接（测试）时视为未启用
func (s *Service) uploadRecordReady(ctx context.Context) bool {
	return s.Repo != nil && s.tableHasColumn(ctx, "uploadRecord", "PATH")
}

// uploadKey 把文件地址统一成 /uploadFile/目录/文件名 的形式（去掉域名、参数），不是本站上传的地址返回空字符串
func (s *Service) uploadKey(fileURL string) string {
	fileURL = strings.TrimSpace(fileURL)
	if parsed, err := url.Parse(fileURL); err == nil && parsed.Path != "" {
		fileURL = parsed.Path
	}
	prefix := s.Config.Upload.PublicPrefix
	index := strings.Index(fileURL, prefix)
	if prefix == "" || index < 0 {
		return ""
	}
	relative, err := cleanRelativePath(fileURL[index+len(prefix):])
	if err != nil || relative == "" {
		return ""
	}
	return strings.TrimSuffix(prefix, "/") + "/" + filepath.ToSlash(relative)
}

// recordUpload 记录文件的上传者。失败只记日志：文件已经保存，只是以后不能通过业务操作删除。
func (s *Service) recordUpload(ctx context.Context, fileURL, uploaderCode string) {
	key := s.uploadKey(fileURL)
	if key == "" || strings.TrimSpace(uploaderCode) == "" || !s.uploadRecordReady(ctx) {
		return
	}
	if _, err := s.Repo.Exec(ctx, "INSERT IGNORE INTO uploadRecord (PATH, USERCODE) VALUES (?, ?)", key, uploaderCode); err != nil {
		log.Printf("记录上传者失败 %s: %v", key, err)
	}
}

// uploadOwner 返回文件的上传者；found 为 false 表示没有上传记录
func (s *Service) uploadOwner(ctx context.Context, key string) (owner string, found bool, err error) {
	rows, err := s.Repo.Query(ctx, "SELECT USERCODE FROM uploadRecord WHERE PATH = ? LIMIT 1", key)
	if err != nil || len(rows) == 0 {
		return "", false, err
	}
	return model.StringValue(rows[0], "USERCODE"), true, nil
}

// checkUploadOwner 判断 fileURL 是否是 ownerCode 上传的本站文件，是则返回统一后的地址
func (s *Service) checkUploadOwner(ctx context.Context, fileURL, ownerCode string) (string, *model.Result) {
	fail := func(message string) (string, *model.Result) {
		result := model.Failure(message)
		return "", &result
	}
	key := s.uploadKey(fileURL)
	if key == "" {
		return fail(msgUploadInvalidURL)
	}
	if !s.uploadRecordReady(ctx) {
		return fail(msgUploadNotReady)
	}
	owner, found, err := s.uploadOwner(ctx, key)
	if err != nil {
		result := dbFailure("查询文件归属", err)
		return "", &result
	}
	if !found {
		return fail(msgUploadNoRecord)
	}
	if strings.TrimSpace(ownerCode) == "" || owner != ownerCode {
		return fail(msgUploadNotOwner)
	}
	return key, nil
}

// DeleteOwnedUpload 只有文件确实由 ownerCode 上传时才删除，并清除上传记录。
// ownerCode 是内容的所有者（头像的主人、文章或资源的作者），管理员删除别人的内容时传内容作者的账号。
func (s *Service) DeleteOwnedUpload(ctx context.Context, fileURL, ownerCode string) model.Result {
	key, failure := s.checkUploadOwner(ctx, fileURL, ownerCode)
	if failure != nil {
		return *failure
	}
	result := s.DeleteUploadedFile(key)
	if !result.IsError || result.ErrMsg == "文件或路径不存在!" {
		if _, err := s.Repo.Exec(ctx, "DELETE FROM uploadRecord WHERE PATH = ?", key); err != nil {
			log.Printf("清除上传记录失败 %s: %v", key, err)
		}
	}
	return result
}
