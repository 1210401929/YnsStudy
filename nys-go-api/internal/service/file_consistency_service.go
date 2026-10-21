package service

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"nys-go-api/internal/model"
)

// orphanMinAge：只清理上传超过 7 天的未引用文件。
// 正在编写的文章、只保存在浏览器本地草稿里的文章，已经上传的图片还没有出现在正文中，留出时间避免误删。
const orphanMinAge = 7 * 24 * time.Hour

// OrphanFile 是一个未被引用的上传文件。
type OrphanFile struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

// CheckFileConsistency 查找未被引用的上传文件。mode 为 "delete" 时删除这些文件，
// 其他值（默认）只返回预览列表，不做任何修改。删除时会重新计算，不使用客户端传来的列表。
func (s *Service) CheckFileConsistency(ctx context.Context, fileType, mode string) model.Result {
	if fileType == "" {
		fileType = "editorImage"
	}
	if fileType != "editorImage" && fileType != "resourcesFile" && fileType != "userAvatar" {
		return model.Failure("不支持的文件类型目录")
	}
	used, err := s.usedFileNames(ctx, fileType)
	if err != nil {
		return dbFailure("查询已使用文件", err)
	}
	directory := filepath.Join(s.Config.Upload.Directory, fileType)
	orphans, skippedRecent, totalSize, err := findOrphanFiles(directory, used, time.Now().Add(-orphanMinAge))
	if err != nil {
		if os.IsNotExist(err) {
			return model.Failure("目录不存在: " + directory)
		}
		return model.Failure("无法读取目录: " + directory)
	}

	deleted := make([]string, 0)
	if mode == "delete" {
		for _, file := range orphans {
			if err := os.Remove(filepath.Join(directory, file.Name)); err == nil {
				deleted = append(deleted, file.Name)
			}
		}
	}
	return model.Success(map[string]any{
		"mode":          map[bool]string{true: "delete", false: "preview"}[mode == "delete"],
		"files":         orphans,
		"totalSize":     totalSize,
		"skippedRecent": skippedRecent,
		"deleted":       deleted,
		"minAgeDays":    int(orphanMinAge.Hours() / 24),
	})
}

// findOrphanFiles 找出目录中未被引用、且修改时间早于 cutoff 的文件；较新的未引用文件只计数不返回。
func findOrphanFiles(directory string, used map[string]struct{}, cutoff time.Time) ([]OrphanFile, int, int64, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, 0, 0, err
	}
	orphans := make([]OrphanFile, 0)
	var totalSize int64
	skippedRecent := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if _, exists := used[entry.Name()]; exists {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(cutoff) {
			skippedRecent++
			continue
		}
		orphans = append(orphans, OrphanFile{Name: entry.Name(), Size: info.Size(), ModTime: info.ModTime().Format("2006-01-02 15:04:05")})
		totalSize += info.Size()
	}
	return orphans, skippedRecent, totalSize, nil
}

func (s *Service) usedFileNames(ctx context.Context, fileType string) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	var rows []map[string]any
	var err error
	switch fileType {
	case "editorImage":
		rows, err = s.Repo.Query(ctx, "SELECT MAINTEXT AS TEXT FROM blogInfo")
		if err == nil {
			for _, row := range rows {
				for _, fileURL := range extractImageURLs(model.StringValue(row, "TEXT")) {
					result[filepath.Base(fileURL)] = struct{}{}
				}
			}
			// 便签墙随记的图片也上传在 editorImage 目录：正文里的图片和 IMAGES 列都算“正在使用”
			noteColumns := "TEXT"
			if s.tableHasColumn(ctx, "communityInfo", "IMAGES") {
				noteColumns += ", IMAGES"
			}
			var notes []map[string]any
			notes, err = s.Repo.Query(ctx, "SELECT "+noteColumns+" FROM communityInfo")
			if err == nil {
				for _, note := range decodeNoteRows(notes) {
					for _, fileURL := range extractImageURLs(model.StringValue(note, "TEXT")) {
						result[filepath.Base(fileURL)] = struct{}{}
					}
					if images, ok := note["IMAGES"].([]string); ok {
						for _, fileURL := range images {
							result[filepath.Base(fileURL)] = struct{}{}
						}
					}
				}
			}
		}
	case "resourcesFile":
		rows, err = s.Repo.Query(ctx, "SELECT FILEVIEWURL AS URL FROM fileInfo")
	case "userAvatar":
		rows, err = s.Repo.Query(ctx, "SELECT AVATAR AS URL FROM userInfo")
	}
	if err != nil {
		return nil, err
	}
	if fileType != "editorImage" {
		for _, row := range rows {
			if fileURL := model.StringValue(row, "URL"); fileURL != "" {
				result[filepath.Base(fileURL)] = struct{}{}
			}
		}
	}
	return result, nil
}
