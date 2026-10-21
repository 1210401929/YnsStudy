package service

// 话题标签。现在用于便签墙的随记（TARGET_TYPE = note），以后文章也可以使用。需要先执行迁移 012。

import (
	"context"
	"fmt"
	"log"
	"strings"
	"unicode"

	"nys-go-api/internal/model"
)

const (
	TagTargetNote = "note"

	maxTagsPerTarget = 5
	maxTagRunes      = 20
)

func (s *Service) tagsReady(ctx context.Context) bool {
	return s.tableHasColumn(ctx, "tagRelation", "TAG_ID")
}

// normalizeTagName 去掉开头的 #，去掉首尾空白；名字里不允许空白和 , # 这类分隔符，最长 20 个字
func normalizeTagName(name string) string {
	name = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(name), "#＃"))
	if name == "" || len([]rune(name)) > maxTagRunes {
		return ""
	}
	for _, r := range name {
		if unicode.IsSpace(r) || strings.ContainsRune(",，#＃;；、<>\"'", r) || unicode.IsControl(r) {
			return ""
		}
	}
	return name
}

// normalizeTags 接受数组或用逗号分隔的字符串，去重（不区分大小写），最多保留 5 个
func normalizeTags(value any) []string {
	var raw []string
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			raw = append(raw, fmt.Sprint(item))
		}
	case []string:
		raw = typed
	case string:
		raw = strings.FieldsFunc(typed, func(r rune) bool { return r == ',' || r == '，' })
	}
	seen := make(map[string]bool)
	tags := make([]string, 0, len(raw))
	for _, item := range raw {
		name := normalizeTagName(item)
		key := strings.ToLower(name)
		if name == "" || seen[key] {
			continue
		}
		seen[key] = true
		tags = append(tags, name)
		if len(tags) == maxTagsPerTarget {
			break
		}
	}
	return tags
}

// saveTargetTags 用新的话题列表替换一条内容原有的话题。失败只记日志，不影响内容本身的保存。
func (s *Service) saveTargetTags(ctx context.Context, targetType, targetID string, tags []string) {
	if !s.tagsReady(ctx) || targetID == "" {
		return
	}
	if _, err := s.Repo.Exec(ctx, "DELETE FROM tagRelation WHERE TARGET_TYPE = ? AND TARGET_ID = ?", targetType, targetID); err != nil {
		log.Printf("清除话题失败: %v", err)
		return
	}
	for _, name := range tags {
		if _, err := s.Repo.Exec(ctx, "INSERT IGNORE INTO tag (NAME) VALUES (?)", name); err != nil {
			log.Printf("保存话题 %s 失败: %v", name, err)
			continue
		}
		if _, err := s.Repo.Exec(ctx, "INSERT IGNORE INTO tagRelation (TAG_ID, TARGET_TYPE, TARGET_ID) SELECT ID, ?, ? FROM tag WHERE NAME = ?", targetType, targetID, name); err != nil {
			log.Printf("关联话题 %s 失败: %v", name, err)
		}
	}
}

func (s *Service) deleteTargetTags(ctx context.Context, targetType, targetID string) {
	if !s.tagsReady(ctx) {
		return
	}
	if _, err := s.Repo.Exec(ctx, "DELETE FROM tagRelation WHERE TARGET_TYPE = ? AND TARGET_ID = ?", targetType, targetID); err != nil {
		log.Printf("删除话题关联失败: %v", err)
	}
}

// attachTags 给每一行加上 TAGS（话题名数组），没有话题时为空数组
func (s *Service) attachTags(ctx context.Context, targetType string, rows []map[string]any) {
	for _, row := range rows {
		row["TAGS"] = []string{}
	}
	if len(rows) == 0 || !s.tagsReady(ctx) {
		return
	}
	placeholders := make([]string, len(rows))
	args := []any{targetType}
	byID := make(map[string][]map[string]any, len(rows))
	for index, row := range rows {
		placeholders[index] = "?"
		id := model.StringValue(row, "GUID")
		args = append(args, id)
		byID[id] = append(byID[id], row)
	}
	relations, err := s.Repo.Query(ctx, `SELECT r.TARGET_ID, t.NAME FROM tagRelation r JOIN tag t ON t.ID = r.TAG_ID
WHERE r.TARGET_TYPE = ? AND r.TARGET_ID IN (`+strings.Join(placeholders, ",")+`) ORDER BY r.CREATE_TIME ASC, t.ID ASC`, args...)
	if err != nil {
		log.Printf("读取话题失败: %v", err)
		return
	}
	for _, relation := range relations {
		for _, row := range byID[model.StringValue(relation, "TARGET_ID")] {
			row["TAGS"] = append(row["TAGS"].([]string), model.StringValue(relation, "NAME"))
		}
	}
}

// GetPopularTags 返回使用最多的话题和使用次数
func (s *Service) GetPopularTags(ctx context.Context, targetType string, limit int) model.Result {
	if !s.tagsReady(ctx) {
		return model.Success([]map[string]any{})
	}
	if limit < 1 || limit > 100 {
		limit = 30
	}
	return s.SelectList(ctx, `SELECT t.NAME, COUNT(*) AS TOTAL FROM tagRelation r JOIN tag t ON t.ID = r.TAG_ID
WHERE r.TARGET_TYPE = ? GROUP BY t.ID, t.NAME ORDER BY TOTAL DESC, t.NAME ASC LIMIT ?`, []any{targetType, limit})
}
