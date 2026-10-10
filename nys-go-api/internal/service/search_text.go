package service

import (
	"context"
	"log"
	"math"
	"strings"
	"sync"

	"nys-go-api/internal/model"
)

// searchTextColumn 保存正文去掉 HTML 标签后的纯文本，搜索只匹配这一列，
// 不会因为关键词出现在 span、style 等标签里而误命中。字段由 deploy/migrations/004 添加。
const searchTextColumn = "SEARCH_TEXT"

// searchTables 是需要纯文本搜索的表及其 HTML 正文字段
var searchTables = map[string]string{"blogInfo": "MAINTEXT", "communityInfo": "TEXT"}

// 字段一旦确认存在就不再查询；不存在时每次都重新判断，执行迁移后无需重启即可生效
var searchColumnReady sync.Map

func (s *Service) hasSearchText(ctx context.Context, table string) bool {
	if _, ok := searchColumnReady.Load(table); ok {
		return true
	}
	rows, err := s.Repo.Query(ctx, `SELECT COUNT(*) AS total
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND LOWER(TABLE_NAME) = LOWER(?)
  AND UPPER(COLUMN_NAME) = ?`, table, searchTextColumn)
	if err != nil || firstCount(rows) == 0 {
		return false
	}
	searchColumnReady.Store(table, true)
	return true
}

// htmlToSearchText 提取 HTML 中的全部文字
func htmlToSearchText(content string) string {
	text, _ := summarizeHTML(content, math.MaxInt32)
	return text
}

// searchField 返回搜索正文时使用的 SQL 表达式。
// 已补齐纯文本时搜 SEARCH_TEXT；尚未补齐的行（或还没执行迁移）退回搜原 HTML 字段。
func (s *Service) searchField(ctx context.Context, table, alias string) string {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	if s.hasSearchText(ctx, table) {
		return "COALESCE(" + prefix + searchTextColumn + ", " + prefix + searchTables[table] + ")"
	}
	return prefix + searchTables[table]
}

// likePattern 转义关键词里的 % 和 _，让它们按普通字符匹配
func likePattern(keyword string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + replacer.Replace(strings.TrimSpace(keyword)) + "%"
}

// stripSearchText 去掉返回给前端的数据里的 SEARCH_TEXT，它只在数据库里用于搜索
func stripSearchText(rows []map[string]any) []map[string]any {
	for _, row := range rows {
		delete(row, searchTextColumn)
	}
	return rows
}

// stripSearchTextResult 用于直接返回查询结果的接口
func stripSearchTextResult(result model.Result) model.Result {
	if rows, ok := result.Result.([]map[string]any); ok {
		stripSearchText(rows)
	}
	return result
}

// BackfillSearchText 为还没有纯文本的旧文章、旧帖子补齐 SEARCH_TEXT，服务启动时在后台执行一次。
func (s *Service) BackfillSearchText(ctx context.Context) {
	for table, field := range searchTables {
		if !s.hasSearchText(ctx, table) {
			continue
		}
		filled := 0
		lastFirst := ""
		for {
			rows, err := s.Repo.Query(ctx, "SELECT GUID, "+field+" AS CONTENT FROM "+table+" WHERE "+searchTextColumn+" IS NULL LIMIT 100")
			if err != nil {
				log.Printf("补齐 %s 搜索文本失败: %v", table, err)
				break
			}
			// 没有取到新数据（例如更新没有生效）时停止，避免死循环
			if len(rows) == 0 || model.StringValue(rows[0], "GUID") == lastFirst {
				break
			}
			lastFirst = model.StringValue(rows[0], "GUID")
			for _, row := range rows {
				text := htmlToSearchText(model.StringValue(row, "CONTENT"))
				if _, err := s.Repo.Exec(ctx, "UPDATE "+table+" SET "+searchTextColumn+" = ? WHERE GUID = ?", text, model.Lookup(row, "GUID")); err != nil {
					log.Printf("补齐 %s 搜索文本失败: %v", table, err)
					return
				}
				filled++
			}
		}
		if filled > 0 {
			log.Printf("已为 %d 条 %s 补齐搜索文本", filled, table)
		}
	}
}
