package service

// 文章标签：和便签墙的话题共用 tag / tagRelation 两张表（TARGET_TYPE = blog），不需要新的迁移。
// 统计和标签页只算公开文章，私密文章的标签只有作者自己能看到。

import (
	"context"
	"net/url"
	"regexp"
	"sync"

	"nys-go-api/internal/model"
)

const (
	// BlogTagIndexableMin：公开文章少于这个数的标签页不让搜索引擎收录，也不放进站点地图，避免大量内容单薄的页面
	BlogTagIndexableMin = 2
	// BlogTagPageSize 是标签页每页的文章数，与归档页一致
	BlogTagPageSize = ArchivePageSize
)

// 线上 blogInfo 是旧表（utf8mb4_unicode_ci），tagRelation 是迁移 012 新建的表（数据库默认的 utf8mb4_0900_ai_ci），
// 两列直接比较会报 “Illegal mix of collations”。比较时把 tagRelation.TARGET_ID 转成 blogInfo.GUID 的排序规则，
// blogInfo 一侧不加转换，仍能用上主键索引。排序规则从 information_schema 读取一次后缓存。
var (
	blogIDCollation struct {
		sync.Mutex
		name     string
		resolved bool // 查询失败时不记为已查，下次再试
	}
	collationPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
)

// tagTargetAsBlogID 返回可以和 blogInfo.GUID 直接比较的 r.TARGET_ID 表达式
func (s *Service) tagTargetAsBlogID(ctx context.Context) string {
	blogIDCollation.Lock()
	defer blogIDCollation.Unlock()
	if !blogIDCollation.resolved {
		rows, err := s.Repo.Query(ctx, `SELECT COLLATION_NAME FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND LOWER(TABLE_NAME) = 'bloginfo' AND UPPER(COLUMN_NAME) = 'GUID' LIMIT 1`)
		if err == nil {
			blogIDCollation.resolved = true
			if len(rows) > 0 {
				if name := model.StringValue(rows[0], "COLLATION_NAME"); collationPattern.MatchString(name) {
					blogIDCollation.name = name
				}
			}
		}
	}
	if blogIDCollation.name == "" {
		return "r.TARGET_ID"
	}
	return "r.TARGET_ID COLLATE " + blogIDCollation.name
}

// blogTagFilter 是“文章带有某个标签”的 SQL 条件，alias 是 blogInfo 的别名，参数是标签名
func (s *Service) blogTagFilter(ctx context.Context, alias string) string {
	return alias + ".GUID IN (SELECT " + s.tagTargetAsBlogID(ctx) + " FROM tagRelation r JOIN tag t ON t.ID = r.TAG_ID WHERE r.TARGET_TYPE = '" + TagTargetBlog + "' AND t.NAME = ?)"
}

// NormalizeTags 供控制器使用：整理请求里的标签（去重、最多 5 个），结果不会是 nil
func NormalizeTags(value any) []string {
	return normalizeTags(value)
}

// containsBannedTag 检查标签里有没有违禁词
func (s *Service) containsBannedTag(tags []string) bool {
	for _, tag := range tags {
		if s.containsBannedWord(tag) {
			return true
		}
	}
	return false
}

// PopularBlogTags 返回公开文章里使用最多的标签和对应的公开文章数
func (s *Service) PopularBlogTags(ctx context.Context, limit int) ([]map[string]any, error) {
	if s.Repo == nil || !s.tagsReady(ctx) {
		return []map[string]any{}, nil
	}
	if limit < 1 || limit > 100 {
		limit = 30
	}
	return s.Repo.Query(ctx, `SELECT t.NAME, COUNT(*) AS TOTAL
FROM tagRelation r
JOIN tag t ON t.ID = r.TAG_ID
JOIN blogInfo b ON b.GUID = `+s.tagTargetAsBlogID(ctx)+` AND b.BLOG_TYPE = 'public'
WHERE r.TARGET_TYPE = ?
GROUP BY t.ID, t.NAME
ORDER BY TOTAL DESC, t.NAME ASC
LIMIT ?`, TagTargetBlog, limit)
}

// GetBlogTags 是热门文章标签接口
func (s *Service) GetBlogTags(ctx context.Context, limit int) model.Result {
	rows, err := s.PopularBlogTags(ctx, limit)
	if err != nil {
		return dbFailure("查询文章标签", err)
	}
	return model.Success(rows)
}

// GetBlogTagPage 返回带某个标签的一页公开文章、公开文章总数，以及标签在库里的写法（大小写以第一次创建时为准）。
// 标签不存在时 name 返回空字符串。
func (s *Service) GetBlogTagPage(ctx context.Context, rawName string, page int) (rows []map[string]any, total int, name string, err error) {
	name = normalizeTagName(rawName)
	if name == "" || !s.tagsReady(ctx) {
		return nil, 0, "", nil
	}
	tags, err := s.Repo.Query(ctx, "SELECT NAME FROM tag WHERE NAME = ? LIMIT 1", name)
	if err != nil || len(tags) == 0 {
		return nil, 0, "", err
	}
	name = model.StringValue(tags[0], "NAME")
	if page < 1 {
		page = 1
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM blogInfo b WHERE b.BLOG_TYPE = 'public' AND "+s.blogTagFilter(ctx, "b"), name)
	if err != nil {
		return nil, 0, name, err
	}
	total = int(firstCount(counts))
	rows, err = s.Repo.Query(ctx, `SELECT b.GUID, b.BLOG_TITLE, b.USERNAME, b.USERCODE, b.CREATE_TIME, b.VIEW_PAGE, u.USERNUM, u.AVATAR,
SUBSTRING(b.MAINTEXT, 1, ?) AS EXCERPT_HTML
FROM blogInfo b
LEFT JOIN userInfo u ON b.USERCODE = u.CODE
WHERE b.BLOG_TYPE = 'public' AND `+s.blogTagFilter(ctx, "b")+`
ORDER BY b.CREATE_TIME DESC, b.GUID DESC
LIMIT ? OFFSET ?`, archiveExcerptChars, name, BlogTagPageSize, (page-1)*BlogTagPageSize)
	if err != nil {
		return nil, total, name, err
	}
	s.attachTags(ctx, TagTargetBlog, rows)
	return rows, total, name, nil
}

// GetBlogTagArticles 给前端标签页使用：一页文章（附摘要和首图）、总数和标签的规范写法
func (s *Service) GetBlogTagArticles(ctx context.Context, rawName string, page int) model.Result {
	rows, total, name, err := s.GetBlogTagPage(ctx, rawName, page)
	if err != nil {
		return dbFailure("查询标签文章", err)
	}
	if name == "" {
		return model.Success(map[string]any{"tag": "", "total": 0, "pageSize": BlogTagPageSize, "data": []map[string]any{}})
	}
	rows = replaceTextWithSummary(rows, "EXCERPT_HTML")
	return model.Success(map[string]any{"tag": name, "total": total, "pageSize": BlogTagPageSize, "data": rows})
}

// GetIndexableBlogTags 返回公开文章数达到收录标准的标签和最近一篇文章的时间，用于站点地图
func (s *Service) GetIndexableBlogTags(ctx context.Context) ([]map[string]any, error) {
	if !s.tagsReady(ctx) {
		return []map[string]any{}, nil
	}
	return s.Repo.Query(ctx, `SELECT t.NAME, COUNT(*) AS TOTAL, MAX(b.CREATE_TIME) AS LAST_MODIFIED
FROM tagRelation r
JOIN tag t ON t.ID = r.TAG_ID
JOIN blogInfo b ON b.GUID = `+s.tagTargetAsBlogID(ctx)+` AND b.BLOG_TYPE = 'public'
WHERE r.TARGET_TYPE = ?
GROUP BY t.ID, t.NAME
HAVING COUNT(*) >= ?
ORDER BY TOTAL DESC`, TagTargetBlog, BlogTagIndexableMin)
}

// BlogTagPath 是标签页的地址（名字已做路径转义）
func BlogTagPath(name string) string {
	return "/tag/" + url.PathEscape(name)
}
