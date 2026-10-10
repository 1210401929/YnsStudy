package service

import (
	"context"
	"strconv"

	"nys-go-api/internal/model"
)

func (s *Service) GetHomeData(ctx context.Context) model.Result {
	hotBlogs := s.getHotBlogs(ctx)
	if hotBlogs.IsError {
		return hotBlogs
	}
	hotFiles := s.SelectList(ctx, "SELECT * FROM fileInfo ORDER BY DOWNNUM DESC LIMIT 5", nil)
	if hotFiles.IsError {
		return hotFiles
	}
	authors := s.GetHighQualityAuthors(ctx, 4)
	if authors.IsError {
		return authors
	}
	return model.Success(map[string]any{
		"hotBlogData": hotBlogs.Result,
		"hotFileData": hotFiles.Result,
		"higAuthor":   authors.Result,
	})
}

func (s *Service) GetWebsiteStatistics(ctx context.Context) model.Result {
	query := `SELECT
(SELECT COUNT(1) FROM blogInfo) AS ARTICLENUM,
(SELECT COUNT(1) FROM communityInfo) AS COMMUNITYNUM,
(SELECT COALESCE(SUM(VIEW_PAGE), 0) FROM blogInfo) AS VIEW_PAGE,
(SELECT COUNT(1) FROM userInfo) AS USERNUM,
(SELECT COUNT(1) FROM loginHistory) AS USERLOGINNUM`
	rows, err := s.Repo.Query(ctx, query)
	if err != nil || len(rows) == 0 {
		return model.Failure("获取网站统计失败")
	}
	return model.Success(rows[0])
}

func (s *Service) GetHighQualityAuthors(ctx context.Context, limit int) model.Result {
	if limit <= 0 {
		limit = 4
	}
	if limit > 100 {
		limit = 100
	}
	query := `SELECT
u.USERCODE,
COALESCE(info.NAME, '') AS USERNAME,
COALESCE(info.REMARK, '') AS REMARK,
COALESCE(b.article_count, 0) AS ARTICLE_COUNT,
COALESCE(f.followers, 0) AS FOLLOWER_COUNT,
COALESCE(info.AVATAR, '') AS AVATAR,
COALESCE(b.article_count, 0) + COALESCE(f.followers, 0) * 2 AS TOTAL_SCORE
FROM (
  SELECT USERCODE FROM blogInfo
  UNION
  SELECT FOLLOWUSERCODE AS USERCODE FROM userFollow
) u
LEFT JOIN (SELECT USERCODE, COUNT(*) article_count FROM blogInfo GROUP BY USERCODE) b ON u.USERCODE = b.USERCODE
LEFT JOIN (SELECT FOLLOWUSERCODE, COUNT(*) followers FROM userFollow GROUP BY FOLLOWUSERCODE) f ON u.USERCODE = f.FOLLOWUSERCODE
LEFT JOIN userInfo info ON u.USERCODE = info.CODE
ORDER BY TOTAL_SCORE DESC LIMIT ?`
	return s.SelectList(ctx, query, []any{limit})
}

func (s *Service) GetAllPublicBlogIDs(ctx context.Context) model.Result {
	lastModified := "CREATE_TIME"
	if s.blogInfoHasUpdateTime(ctx) {
		lastModified = "COALESCE(UPDATE_TIME, CREATE_TIME)"
	}
	query := "SELECT GUID, CREATE_TIME, " + lastModified + " AS LAST_MODIFIED FROM blogInfo WHERE BLOG_TYPE = 'public' ORDER BY CREATE_TIME DESC"
	return s.SelectList(ctx, query, nil)
}

func (s *Service) GetLatestBlogs(ctx context.Context) model.Result {
	return s.SelectList(ctx, "SELECT * FROM blogInfo WHERE BLOG_TYPE = 'public' ORDER BY CREATE_TIME DESC LIMIT 5", nil)
}

// GetLatestBlogsForSEO 仅查询首页服务端 HTML 所需的轻量字段，避免读取大段文章正文。
func (s *Service) GetLatestBlogsForSEO(ctx context.Context, limit int) model.Result {
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}
	query := `SELECT GUID, BLOG_TITLE, USERNAME, CREATE_TIME
FROM blogInfo
WHERE BLOG_TYPE = 'public'
ORDER BY CREATE_TIME DESC
LIMIT ?`
	return s.SelectList(ctx, query, []any{limit})
}

func (s *Service) getHotBlogs(ctx context.Context) model.Result {
	query := `SELECT b.*,
COALESCE(c.COMMENT_COUNT, 0) AS COMMENT_COUNT,
COALESCE(gl.LIKE_COUNT, 0) AS LIKE_COUNT,
COALESCE(gl.COLLECT_COUNT, 0) AS COLLECT_COUNT,
(COALESCE(c.COMMENT_COUNT, 0) * 5 + COALESCE(b.VIEW_PAGE, 0)
 + COALESCE(gl.LIKE_COUNT, 0) * 3 + COALESCE(gl.COLLECT_COUNT, 0) * 4
 - (UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(b.CREATE_TIME)) / 3600 * 0.2) AS HOT_SCORE
FROM blogInfo b
LEFT JOIN (SELECT BLOGID, COUNT(DISTINCT GUID) COMMENT_COUNT FROM blogComment GROUP BY BLOGID) c ON b.GUID = c.BLOGID
LEFT JOIN (
 SELECT BLOGID,
 SUM(CASE WHEN TYPE = 'like' THEN 1 ELSE 0 END) LIKE_COUNT,
 SUM(CASE WHEN TYPE = 'collect' THEN 1 ELSE 0 END) COLLECT_COUNT
 FROM blogGiveLike GROUP BY BLOGID
) gl ON b.GUID = gl.BLOGID
WHERE b.BLOG_TYPE = 'public'
ORDER BY HOT_SCORE DESC LIMIT 10`
	result := s.SelectList(ctx, query, nil)
	// 热门列表只显示标题和数据，不需要正文
	if rows, ok := result.Result.([]map[string]any); ok {
		result.Result = stripSearchText(replaceTextWithSummary(rows, "MAINTEXT"))
	}
	return result
}

func parseLimit(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
