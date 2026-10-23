package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"nys-go-api/internal/model"
)

// blockedUploadExtensions 是浏览器打开时可能执行脚本的文件类型。上传文件和网站同一个域名，
// 这类文件一旦被打开就能读取登录信息，所以即使 config.yaml 的 allowed_extensions 里写了也不接受。
var blockedUploadExtensions = map[string]bool{
	".svg": true, ".svgz": true, ".html": true, ".htm": true, ".xhtml": true, ".xht": true, ".shtml": true,
	".xml": true, ".xsl": true, ".js": true, ".mjs": true, ".swf": true, ".php": true, ".jsp": true,
}

// SaveUploadedFile 保存上传文件；uploaderCode 是上传者账号，记录下来用于以后判断谁能删除这个文件。
func (s *Service) SaveUploadedFile(ctx context.Context, file *multipart.FileHeader, spliceURL, uploaderCode string) model.Result {
	if file == nil || file.Size == 0 {
		return model.Failure("上传文件不能为空")
	}
	if file.Size > s.Config.Server.MaxUploadMB*1024*1024 {
		return model.Failure(fmt.Sprintf("上传文件不能超过 %dMB", s.Config.Server.MaxUploadMB))
	}
	extension := strings.ToLower(filepath.Ext(file.Filename))
	if extension == "" {
		return model.Failure("非法文件名，缺少扩展名")
	}
	if blockedUploadExtensions[extension] || !containsString(s.Config.Upload.AllowedExtensions, extension) {
		return model.Failure("不支持的文件类型：" + extension)
	}

	relativeDirectory, err := cleanRelativePath(spliceURL)
	if err != nil {
		return model.Failure(err.Error())
	}
	directory := filepath.Join(s.Config.Upload.Directory, relativeDirectory)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return model.Failure("创建上传目录失败:" + err.Error())
	}
	fileName, err := randomHex(16)
	if err != nil {
		return model.Failure("生成文件名失败")
	}
	fileName += extension
	destination := filepath.Join(directory, fileName)

	source, err := file.Open()
	if err != nil {
		return model.Failure("读取上传文件失败:" + err.Error())
	}
	defer source.Close()
	target, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return model.Failure("创建目标文件失败:" + err.Error())
	}
	_, copyErr := io.Copy(target, source)
	closeErr := target.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(destination)
		if copyErr != nil {
			return model.Failure("保存上传文件失败:" + copyErr.Error())
		}
		return model.Failure("关闭上传文件失败:" + closeErr.Error())
	}

	publicParts := []string{strings.TrimSuffix(s.Config.Upload.PublicPrefix, "/")}
	if relativeDirectory != "" {
		publicParts = append(publicParts, filepath.ToSlash(relativeDirectory))
	}
	publicParts = append(publicParts, fileName)
	viewURL := strings.Join(publicParts, "/")
	s.recordUpload(ctx, viewURL, uploaderCode)
	// 不再返回服务器上的绝对路径（fileRealUrl），它会被前端存进资源表并公开
	return model.Success(map[string]string{
		"originalFileName": file.Filename,
		"savedFileName":    fileName,
		"fileViewUrl":      viewURL,
	})
}

func (s *Service) DeleteUploadedFile(urlPath string) model.Result {
	prefix := s.Config.Upload.PublicPrefix
	parsed, err := url.Parse(urlPath)
	if err == nil && parsed.Path != "" {
		urlPath = parsed.Path
	}
	index := strings.Index(urlPath, prefix)
	if index < 0 {
		return model.Failure("路径不合法!")
	}
	relative, err := cleanRelativePath(urlPath[index+len(prefix):])
	if err != nil || relative == "" {
		return model.Failure("路径不合法!")
	}
	target := filepath.Join(s.Config.Upload.Directory, relative)
	inside, err := pathInside(s.Config.Upload.Directory, target)
	if err != nil || !inside {
		return model.Failure("路径不合法!")
	}
	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return model.Failure("文件或路径不存在!")
		}
		return model.Failure("读取文件失败:" + err.Error())
	}
	if info.IsDir() {
		return model.Failure("不允许删除目录")
	}
	if err := os.Remove(target); err != nil {
		return model.Failure("删除失败:" + err.Error())
	}
	return model.Success("删除成功!")
}

// DeleteUploadedFiles 清理文章里不再使用的图片，尽力而为：
// 外链图片、base64 图片不是本站上传的文件，直接跳过；只删除 ownerCode（文章作者）自己上传的图片，
// 别人的图片、没有上传记录的旧图片都保留；文件已不存在也不算失败；其他错误只记日志，不影响文章的保存和删除。
func (s *Service) DeleteUploadedFiles(ctx context.Context, urls []string, ownerCode string) {
	prefix := s.Config.Upload.PublicPrefix
	for _, fileURL := range urls {
		if prefix == "" || !strings.Contains(fileURL, prefix) || strings.HasPrefix(strings.ToLower(strings.TrimSpace(fileURL)), "data:") {
			continue
		}
		result := s.DeleteOwnedUpload(ctx, fileURL, ownerCode)
		if result.IsError && result.ErrMsg != "文件或路径不存在!" && result.ErrMsg != msgUploadNoRecord {
			log.Printf("清理文章图片失败 %s: %s", fileURL, result.ErrMsg)
		}
	}
}

func cleanRelativePath(value string) (string, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = strings.Trim(value, "/")
	if value == "" {
		return "", nil
	}
	cleaned := filepath.Clean(filepath.FromSlash(value))
	if cleaned == "." || filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("目录路径包含非法字符")
	}
	return cleaned, nil
}

func pathInside(root, target string) (bool, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false, err
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return false, err
	}
	relative, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return false, err
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)), nil
}

func randomHex(byteCount int) (string, error) {
	raw := make([]byte, byteCount)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if strings.EqualFold(value, expected) {
			return true
		}
	}
	return false
}
