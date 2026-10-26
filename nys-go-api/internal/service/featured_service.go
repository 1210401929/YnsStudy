package service

// 关于页的“站长推荐”：站长在后台挑选的几篇文章和几个系列，保存在 siteSetting 表（迁移 017）。
// 只展示仍然公开的文章和仍有公开文章的系列；文章被删除或改为私密后自动不再显示，不需要重新配置。

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

const (
	settingAboutFeatured = "about_featured"
	maxFeaturedArticles  = 6
	maxFeaturedSeries    = 3
	msgSettingNotReady   = "网站设置还没有启用（需要先执行数据库迁移 017）"
)

// FeaturedConfig 是保存在 siteSetting 里的推荐配置，按站长设置的顺序排列
type FeaturedConfig struct {
	Articles []string `json:"articles"`
	Series   []string `json:"series"`
}

// Featured 是解析后的推荐内容，供关于页展示
type Featured struct {
	Articles []map[string]any `json:"articles"`
	Series   []map[string]any `json:"series"`
}

func (s *Service) siteSettingReady(ctx context.Context) bool {
	return s.Repo != nil && s.tableHasColumn(ctx, "siteSetting", "SETTING_KEY")
}

// loadSiteSetting 读取一项设置并解析到 target；没有这项设置时 found 为 false
func (s *Service) loadSiteSetting(ctx context.Context, key string, target any) (found bool, err error) {
	if !s.siteSettingReady(ctx) {
		return false, nil
	}
	rows, err := s.Repo.Query(ctx, "SELECT SETTING_VALUE FROM siteSetting WHERE SETTING_KEY = ? LIMIT 1", key)
	if err != nil || len(rows) == 0 {
		return false, err
	}
	return true, json.Unmarshal([]byte(model.StringValue(rows[0], "SETTING_VALUE")), target)
}

// saveSiteSetting 保存一项设置（没有则新增）
func (s *Service) saveSiteSetting(ctx context.Context, key string, value any, userCode string) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.Repo.Exec(ctx, `INSERT INTO siteSetting (SETTING_KEY, SETTING_VALUE, UPDATE_USER) VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE SETTING_VALUE = VALUES(SETTING_VALUE), UPDATE_USER = VALUES(UPDATE_USER)`, key, string(encoded), userCode)
	return err
}

// uniqueIDs 去掉空值和重复值，最多保留 limit 个
func uniqueIDs(values []string, limit int) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
		if len(result) == limit {
			break
		}
	}
	return result
}

// resolveFeaturedArticles 按给定顺序读取公开文章（附摘要、首图、标签），不公开或已删除的跳过
func (s *Service) resolveFeaturedArticles(ctx context.Context, guids []string) ([]map[string]any, error) {
	result := make([]map[string]any, 0, len(guids))
	if len(guids) == 0 {
		return result, nil
	}
	args := make([]any, 0, len(guids)+1)
	args = append(args, archiveExcerptChars)
	for _, guid := range guids {
		args = append(args, guid)
	}
	rows, err := s.Repo.Query(ctx, `SELECT b.GUID, b.BLOG_TITLE, b.CREATE_TIME, b.USERNAME, b.VIEW_PAGE, u.USERNUM,
SUBSTRING(b.MAINTEXT, 1, ?) AS EXCERPT_HTML
FROM blogInfo b LEFT JOIN userInfo u ON b.USERCODE = u.CODE
WHERE b.BLOG_TYPE = 'public' AND b.GUID IN (`+strings.TrimSuffix(strings.Repeat("?,", len(guids)), ",")+`)`, args...)
	if err != nil {
		return nil, err
	}
	byGUID := make(map[string]map[string]any, len(rows))
	for _, row := range rows {
		row["USERNAME"] = model.PublicDisplayName(model.StringValue(row, "USERNAME"))
		byGUID[model.StringValue(row, "GUID")] = row
	}
	for _, guid := range guids {
		if row, ok := byGUID[guid]; ok {
			result = append(result, row)
		}
	}
	s.attachTags(ctx, TagTargetBlog, result)
	return replaceTextWithSummary(result, "EXCERPT_HTML"), nil
}

// resolveFeaturedSeries 按给定顺序读取仍有公开文章的系列，附上第一篇公开文章的编号（“从第一篇开始读”）
func (s *Service) resolveFeaturedSeries(ctx context.Context, ids []string) ([]map[string]any, error) {
	result := make([]map[string]any, 0, len(ids))
	if len(ids) == 0 || !s.seriesReady(ctx) {
		return result, nil
	}
	all, err := s.PublicSeriesList(ctx, "", "latest", 200)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]map[string]any, len(all))
	for _, row := range all {
		byID[model.StringValue(row, "ID")] = row
	}
	for _, id := range ids {
		row, ok := byID[normalizeSeriesID(id)]
		if !ok {
			continue
		}
		if _, articles, found, err := s.SeriesDetail(ctx, model.StringValue(row, "ID"), false); err == nil && found && len(articles) > 0 {
			row["FIRST_GUID"] = model.StringValue(articles[0], "GUID")
		}
		delete(row, "USERCODE")
		result = append(result, row)
	}
	return result, nil
}

// AboutFeatured 返回关于页的站长推荐；没有配置或还没执行迁移时返回空列表
func (s *Service) AboutFeatured(ctx context.Context) Featured {
	featured := Featured{Articles: []map[string]any{}, Series: []map[string]any{}}
	var config FeaturedConfig
	if found, err := s.loadSiteSetting(ctx, settingAboutFeatured, &config); err != nil || !found {
		return featured
	}
	if articles, err := s.resolveFeaturedArticles(ctx, uniqueIDs(config.Articles, maxFeaturedArticles)); err == nil {
		featured.Articles = articles
	}
	if series, err := s.resolveFeaturedSeries(ctx, uniqueIDs(config.Series, maxFeaturedSeries)); err == nil {
		featured.Series = series
	}
	return featured
}

// GetAdminAboutFeatured 后台读取当前的站长推荐（超级管理员）
func (s *Service) GetAdminAboutFeatured(c *gin.Context) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	ready := s.siteSettingReady(ctx)
	var config FeaturedConfig
	if ready {
		_, _ = s.loadSiteSetting(ctx, settingAboutFeatured, &config)
	}
	featured := s.AboutFeatured(ctx)
	return model.Success(map[string]any{
		"ready":       ready,
		"articles":    featured.Articles,
		"series":      featured.Series,
		"maxArticles": maxFeaturedArticles,
		"maxSeries":   maxFeaturedSeries,
		// 配置里有、但文章已删除或改为私密的数量，后台提示站长
		"hiddenArticles": len(uniqueIDs(config.Articles, maxFeaturedArticles)) - len(featured.Articles),
		"hiddenSeries":   len(uniqueIDs(config.Series, maxFeaturedSeries)) - len(featured.Series),
	})
}

// SetAboutFeatured 保存站长推荐（超级管理员）。只接受公开文章和有公开文章的系列，顺序即展示顺序。
func (s *Service) SetAboutFeatured(c *gin.Context, articles, series []string) model.Result {
	actor, failure := s.requireSuperAdmin(c)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	if !s.siteSettingReady(ctx) {
		return model.Failure(msgSettingNotReady)
	}
	articles = uniqueIDs(articles, maxFeaturedArticles+1)
	series = uniqueIDs(series, maxFeaturedSeries+1)
	if len(articles) > maxFeaturedArticles {
		return model.Failure(fmt.Sprintf("最多推荐 %d 篇文章", maxFeaturedArticles))
	}
	if len(series) > maxFeaturedSeries {
		return model.Failure(fmt.Sprintf("最多推荐 %d 个系列", maxFeaturedSeries))
	}
	resolvedArticles, err := s.resolveFeaturedArticles(ctx, articles)
	if err != nil {
		return dbFailure("查询文章", err)
	}
	if len(resolvedArticles) != len(articles) {
		return model.Failure("只能推荐公开文章，有文章不存在或不是公开的，请刷新后重新选择")
	}
	resolvedSeries, err := s.resolveFeaturedSeries(ctx, series)
	if err != nil {
		return dbFailure("查询系列", err)
	}
	if len(resolvedSeries) != len(series) {
		return model.Failure("只能推荐有公开文章的系列，请刷新后重新选择")
	}
	if err := s.saveSiteSetting(ctx, settingAboutFeatured, FeaturedConfig{Articles: articles, Series: series}, actor.Code()); err != nil {
		return dbFailure("保存站长推荐", err)
	}
	s.recordAdminAction(c, actor, AdminActionAboutFeatured, "关于页站长推荐", fmt.Sprintf("推荐 %d 篇文章、%d 个系列", len(articles), len(series)))
	s.SubmitIndexNow("/ynsStudy/About")
	return model.Success(map[string]any{"articles": resolvedArticles, "series": resolvedSeries})
}
