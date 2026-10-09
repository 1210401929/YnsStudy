package service

import (
	"context"
	"strings"

	"nys-go-api/internal/model"
)

// ArchivePageSize 是文章归档每页的文章数。
const ArchivePageSize = 20

// archiveExcerptChars 只截取正文开头用于生成摘要，避免归档页一次读取所有文章全文。
const archiveExcerptChars = 1500

// GetBlogArchivePage 按发布时间倒序返回一页公开文章，以及公开文章总数。
func (s *Service) GetBlogArchivePage(ctx context.Context, page int) ([]map[string]any, int, error) {
	if page < 1 {
		page = 1
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM blogInfo WHERE BLOG_TYPE = 'public'")
	if err != nil {
		return nil, 0, err
	}
	total := firstCount(counts)
	rows, err := s.Repo.Query(ctx, `SELECT GUID, BLOG_TITLE, USERNAME, CREATE_TIME, SUBSTRING(MAINTEXT, 1, ?) AS EXCERPT_HTML
FROM blogInfo
WHERE BLOG_TYPE = 'public'
ORDER BY CREATE_TIME DESC, GUID DESC
LIMIT ? OFFSET ?`, archiveExcerptChars, ArchivePageSize, (page-1)*ArchivePageSize)
	if err != nil {
		return nil, 0, err
	}
	return rows, int(total), nil
}

// ArticleLinks 是文章页底部的站内链接：相邻文章和同作者的其他文章。
type ArticleLinks struct {
	Previous map[string]any   // 更早发布的一篇
	Next     map[string]any   // 更晚发布的一篇
	Related  []map[string]any // 同作者的其他公开文章
}

// GetArticleLinksForSEO 读取文章页的内链数据。查询失败只影响内链，不影响正文输出，
// 因此这里吞掉错误并返回已取得的部分。
func (s *Service) GetArticleLinksForSEO(ctx context.Context, article map[string]any) ArticleLinks {
	var links ArticleLinks
	guid := strings.TrimSpace(model.StringValue(article, "GUID"))
	created := model.Lookup(article, "CREATE_TIME")
	if guid == "" || created == nil {
		return links
	}
	const fields = "SELECT GUID, BLOG_TITLE, CREATE_TIME FROM blogInfo WHERE BLOG_TYPE = 'public' AND GUID <> ?"
	// 发布时间相同时按 GUID 排序，保证上一篇/下一篇的顺序稳定且互为对应。
	if rows, err := s.Repo.Query(ctx, fields+` AND (CREATE_TIME < ? OR (CREATE_TIME = ? AND GUID < ?))
ORDER BY CREATE_TIME DESC, GUID DESC LIMIT 1`, guid, created, created, guid); err == nil && len(rows) > 0 {
		links.Previous = rows[0]
	}
	if rows, err := s.Repo.Query(ctx, fields+` AND (CREATE_TIME > ? OR (CREATE_TIME = ? AND GUID > ?))
ORDER BY CREATE_TIME ASC, GUID ASC LIMIT 1`, guid, created, created, guid); err == nil && len(rows) > 0 {
		links.Next = rows[0]
	}
	if userCode := strings.TrimSpace(model.StringValue(article, "USERCODE")); userCode != "" {
		if rows, err := s.Repo.Query(ctx, fields+` AND USERCODE = ?
ORDER BY CREATE_TIME DESC LIMIT 5`, guid, userCode); err == nil {
			links.Related = rows
		}
	}
	return links
}

// GetArticleLinks 给前端文章页使用，返回与服务端 HTML 相同的上一篇、下一篇和同作者文章。
func (s *Service) GetArticleLinks(ctx context.Context, blogID string) model.Result {
	rows, err := s.Repo.Query(ctx, "SELECT GUID, USERCODE, CREATE_TIME FROM blogInfo WHERE GUID = ? AND BLOG_TYPE = 'public' LIMIT 1", strings.TrimSpace(blogID))
	if err != nil {
		return dbFailure("查询文章", err)
	}
	if len(rows) == 0 {
		return model.Success(map[string]any{"previous": nil, "next": nil, "related": []map[string]any{}})
	}
	links := s.GetArticleLinksForSEO(ctx, rows[0])
	related := links.Related
	if related == nil {
		related = []map[string]any{}
	}
	return model.Success(map[string]any{"previous": links.Previous, "next": links.Next, "related": related})
}

// GetPublicAuthorNumbers 返回至少有一篇公开文章的作者编号，用于 sitemap。
// 没有公开内容的空主页不提交，避免大量低质量页面拉低整站评价。
func (s *Service) GetPublicAuthorNumbers(ctx context.Context) ([]map[string]any, error) {
	return s.Repo.Query(ctx, `SELECT u.USERNUM, MAX(b.CREATE_TIME) AS LAST_MODIFIED
FROM userInfo u
JOIN blogInfo b ON b.USERCODE = u.CODE AND b.BLOG_TYPE = 'public'
WHERE u.USERNUM IS NOT NULL AND u.USERNUM <> ''
GROUP BY u.USERNUM`)
}
