package service

import (
	"context"
	"os"
	"path/filepath"

	"nys-go-api/internal/model"
)

func (s *Service) CheckFileConsistency(ctx context.Context, fileType string) model.Result {
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
	entries, err := os.ReadDir(directory)
	if err != nil {
		if os.IsNotExist(err) {
			return model.Failure("目录不存在: " + directory)
		}
		return model.Failure("无法读取目录: " + directory)
	}
	deleted := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if _, exists := used[entry.Name()]; exists {
			continue
		}
		if err := os.Remove(filepath.Join(directory, entry.Name())); err == nil {
			deleted = append(deleted, entry.Name())
		}
	}
	return model.Success(deleted)
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
