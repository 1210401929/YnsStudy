package service

// 文章系列：作者把自己的几篇文章编成一个有顺序的合集，需要先执行迁移 015。
// 规则：每个作者都能建自己的系列；一篇文章最多属于一个系列，且只能放进文章作者自己的系列；
// 管理员可以管理除超级管理员以外所有人的系列（与文章权限一致）；删除系列时文章只是退出系列。
// 公开页面只显示公开文章，作者本人（以及能管理它的管理员）还能看到其中的私密文章。

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

const (
	maxSeriesTitle       = 60
	maxSeriesDescription = 300
	maxSeriesPerUser     = 50
	// SeriesIndexableMin：公开文章少于这个数的系列页不让搜索引擎收录，也不放进站点地图
	SeriesIndexableMin = 2

	msgSeriesNotReady = "文章系列还没有启用（需要先执行数据库迁移 015）"
	msgSeriesNotFound = "系列不存在或已删除"
	msgSeriesNotOwner = "只能把文章放进作者自己的系列"

	// seriesOrderExpr 是系列内文章的排序：先按手动调整的顺序，没有顺序的按发布时间排在后面
	seriesOrderExpr = "COALESCE(b.SERIES_ORDER, 2147483647), b.CREATE_TIME, b.GUID"
)

func (s *Service) seriesReady(ctx context.Context) bool {
	return s.Repo != nil && s.tableHasColumn(ctx, "blogSeries", "ID") && s.tableHasColumn(ctx, "blogInfo", "SERIES_ID")
}

// normalizeSeriesID 把请求里的系列编号整理成数字字符串；空、0、非数字都视为“不属于任何系列”
func normalizeSeriesID(value string) string {
	value = strings.TrimSpace(value)
	if id, err := strconv.ParseInt(value, 10, 64); err == nil && id > 0 {
		return strconv.FormatInt(id, 10)
	}
	return ""
}

func (s *Service) loadSeries(ctx context.Context, id string) (map[string]any, bool, error) {
	id = normalizeSeriesID(id)
	if id == "" {
		return nil, false, nil
	}
	rows, err := s.Repo.Query(ctx, "SELECT * FROM blogSeries WHERE ID = ? LIMIT 1", id)
	if err != nil || len(rows) == 0 {
		return nil, false, err
	}
	return rows[0], true, nil
}

// seriesText 校验系列名称和简介
func (s *Service) seriesText(title, description string) (string, string, *model.Result) {
	fail := func(message string) (string, string, *model.Result) {
		result := model.Failure(message)
		return "", "", &result
	}
	title = strings.Join(strings.Fields(title), " ")
	description = strings.TrimSpace(description)
	if title == "" {
		return fail("系列名称不能为空")
	}
	if len([]rune(title)) > maxSeriesTitle {
		return fail(fmt.Sprintf("系列名称最多 %d 个字", maxSeriesTitle))
	}
	if len([]rune(description)) > maxSeriesDescription {
		return fail(fmt.Sprintf("系列简介最多 %d 个字", maxSeriesDescription))
	}
	if s.containsBannedWord(title) || s.containsBannedWord(description) {
		return fail(msgBannedWord)
	}
	return title, description, nil
}

// requireSeriesManager 读取系列并检查当前用户能否管理它
func (s *Service) requireSeriesManager(c *gin.Context, id string) (*Actor, map[string]any, *model.Result) {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return nil, nil, failure
	}
	if !s.seriesReady(contextOf(c)) {
		result := model.Failure(msgSeriesNotReady)
		return nil, nil, &result
	}
	series, found, err := s.loadSeries(contextOf(c), id)
	if err != nil {
		result := dbFailure("查询系列", err)
		return nil, nil, &result
	}
	if !found {
		result := model.Failure(msgSeriesNotFound)
		return nil, nil, &result
	}
	if !actor.CanManage(model.StringValue(series, "USERCODE")) {
		result := model.Failure(msgForbidden)
		return nil, nil, &result
	}
	return actor, series, nil
}

// AddSeries 新建一个属于当前用户的系列
func (s *Service) AddSeries(c *gin.Context, title, description string) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	if !s.seriesReady(ctx) {
		return model.Failure(msgSeriesNotReady)
	}
	title, description, failure = s.seriesText(title, description)
	if failure != nil {
		return *failure
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM blogSeries WHERE USERCODE = ?", actor.Code())
	if err != nil {
		return dbFailure("统计系列", err)
	}
	if firstCount(counts) >= maxSeriesPerUser {
		return model.Failure(fmt.Sprintf("每人最多建 %d 个系列", maxSeriesPerUser))
	}
	result := s.SaveAll(ctx, "add", "blogSeries", []map[string]any{{
		"USERCODE": actor.Code(), "TITLE": title, "DESCRIPTION": nullIfEmpty(description),
	}}, "ID")
	if rows, ok := result.Result.([]map[string]any); ok && len(rows) > 0 {
		rows[0]["ARTICLE_COUNT"] = 0
		return model.Success(rows[0])
	}
	return result
}

// UpdateSeries 修改系列名称和简介
func (s *Service) UpdateSeries(c *gin.Context, id, title, description string) model.Result {
	actor, series, failure := s.requireSeriesManager(c, id)
	if failure != nil {
		return *failure
	}
	title, description, failure = s.seriesText(title, description)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	if _, err := s.Repo.Exec(ctx, "UPDATE blogSeries SET TITLE = ?, DESCRIPTION = ?, UPDATE_TIME = NOW() WHERE ID = ?",
		title, nullIfEmpty(description), model.StringValue(series, "ID")); err != nil {
		return dbFailure("修改系列", err)
	}
	s.logModeration(c, actor, series, AdminActionEditSeries, model.StringValue(series, "TITLE"), "修改了他人的系列，新名称："+title)
	s.notifySeriesChanged(ctx, model.StringValue(series, "ID"))
	return model.Success("已保存")
}

// DeleteSeries 删除系列，系列里的文章退出系列（文章本身不删）
func (s *Service) DeleteSeries(c *gin.Context, id string) model.Result {
	actor, series, failure := s.requireSeriesManager(c, id)
	if failure != nil {
		return *failure
	}
	seriesID := model.StringValue(series, "ID")
	_, _, err := s.Repo.ExecuteBatch(contextOf(c), []string{
		"UPDATE blogInfo SET SERIES_ID = NULL, SERIES_ORDER = NULL WHERE SERIES_ID = ?",
		"DELETE FROM blogSeries WHERE ID = ?",
	}, [][]any{{seriesID}, {seriesID}}, false)
	if err != nil {
		return dbFailure("删除系列", err)
	}
	s.logModeration(c, actor, series, AdminActionEditSeries, model.StringValue(series, "TITLE"), "删除了他人的系列")
	s.SubmitIndexNow(SeriesPath(seriesID))
	return model.Success("已删除，系列里的文章都还在")
}

// SortSeries 按 guids 的顺序重新排列系列里的文章；不在列表里的文章排在后面
func (s *Service) SortSeries(c *gin.Context, id string, guids []string) model.Result {
	_, series, failure := s.requireSeriesManager(c, id)
	if failure != nil {
		return *failure
	}
	seriesID := model.StringValue(series, "ID")
	queries := []string{"UPDATE blogInfo SET SERIES_ORDER = NULL WHERE SERIES_ID = ?"}
	args := [][]any{{seriesID}}
	seen := make(map[string]bool, len(guids))
	for _, guid := range guids {
		guid = strings.TrimSpace(guid)
		if guid == "" || seen[guid] {
			continue
		}
		seen[guid] = true
		queries = append(queries, "UPDATE blogInfo SET SERIES_ORDER = ? WHERE GUID = ? AND SERIES_ID = ?")
		args = append(args, []any{len(seen), guid, seriesID})
	}
	queries = append(queries, "UPDATE blogSeries SET UPDATE_TIME = NOW() WHERE ID = ?")
	args = append(args, []any{seriesID})
	if _, _, err := s.Repo.ExecuteBatch(contextOf(c), queries, args, false); err != nil {
		return dbFailure("调整系列顺序", err)
	}
	s.notifySeriesChanged(contextOf(c), seriesID)
	return model.Success("顺序已保存")
}

// GetMySeries 返回当前用户的全部系列（含私密文章的数量），写作页选择系列、管理系列时使用
func (s *Service) GetMySeries(c *gin.Context) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	if !s.seriesReady(ctx) {
		return model.Success([]map[string]any{})
	}
	return s.SelectList(ctx, `SELECT s.ID, s.TITLE, s.DESCRIPTION, s.CREATE_TIME, s.UPDATE_TIME,
(SELECT COUNT(*) FROM blogInfo b WHERE b.SERIES_ID = s.ID) AS ARTICLE_COUNT
FROM blogSeries s WHERE s.USERCODE = ?
ORDER BY COALESCE(s.UPDATE_TIME, s.CREATE_TIME) DESC, s.ID DESC`, []any{actor.Code()})
}

// PublicSeriesList 返回至少有一篇公开文章的系列。sort = hot 按文章数，否则按最近更新；userCode 不为空时只看这个作者的。
func (s *Service) PublicSeriesList(ctx context.Context, userCode, sort string, limit int) ([]map[string]any, error) {
	if !s.seriesReady(ctx) {
		return []map[string]any{}, nil
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	order := "LAST_MODIFIED DESC, s.ID DESC"
	if sort == "hot" {
		order = "ARTICLE_COUNT DESC, LAST_MODIFIED DESC, s.ID DESC"
	}
	rows, err := s.Repo.Query(ctx, `SELECT s.ID, s.TITLE, s.DESCRIPTION, s.USERCODE, u.NAME AS USERNAME, u.USERNUM, u.AVATAR,
COUNT(b.GUID) AS ARTICLE_COUNT,
GREATEST(MAX(b.CREATE_TIME), COALESCE(s.UPDATE_TIME, s.CREATE_TIME)) AS LAST_MODIFIED
FROM blogSeries s
JOIN blogInfo b ON b.SERIES_ID = s.ID AND b.BLOG_TYPE = 'public'
LEFT JOIN userInfo u ON u.CODE = s.USERCODE
WHERE (? = '' OR s.USERCODE = ?)
GROUP BY s.ID, s.TITLE, s.DESCRIPTION, s.USERCODE, s.UPDATE_TIME, s.CREATE_TIME, u.NAME, u.USERNUM, u.AVATAR
ORDER BY `+order+`
LIMIT ?`, userCode, userCode, limit)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		row["USERNAME"] = model.PublicDisplayName(model.StringValue(row, "USERNAME"))
	}
	return rows, nil
}

// GetSeriesList 是“全部系列”接口
func (s *Service) GetSeriesList(ctx context.Context, userCode, sort string, limit int) model.Result {
	rows, err := s.PublicSeriesList(ctx, strings.TrimSpace(userCode), sort, limit)
	if err != nil {
		return dbFailure("查询系列", err)
	}
	return model.Success(rows)
}

// SeriesDetail 返回系列信息（附作者名字、主页编号、头像和公开文章数）和按顺序排列的文章。
// includePrivate 为 true 时包含私密文章（作者本人或能管理它的管理员）。系列不存在时 found 为 false。
func (s *Service) SeriesDetail(ctx context.Context, id string, includePrivate bool) (series map[string]any, articles []map[string]any, found bool, err error) {
	if !s.seriesReady(ctx) {
		return nil, nil, false, nil
	}
	series, found, err = s.loadSeries(ctx, id)
	if err != nil || !found {
		return nil, nil, false, err
	}
	if users, err := s.Repo.Query(ctx, "SELECT NAME, USERNUM, AVATAR FROM userInfo WHERE CODE = ? LIMIT 1", model.StringValue(series, "USERCODE")); err == nil && len(users) > 0 {
		series["USERNAME"] = model.PublicDisplayName(model.StringValue(users[0], "NAME"))
		series["USERNUM"] = model.Lookup(users[0], "USERNUM")
		series["AVATAR"] = model.Lookup(users[0], "AVATAR")
	}
	visibility := "b.BLOG_TYPE = 'public'"
	if includePrivate {
		visibility = "1 = 1"
	}
	articles, err = s.Repo.Query(ctx, `SELECT b.GUID, b.BLOG_TITLE, b.BLOG_TYPE, b.CREATE_TIME, b.VIEW_PAGE, b.SERIES_ORDER,
SUBSTRING(b.MAINTEXT, 1, ?) AS EXCERPT_HTML
FROM blogInfo b
WHERE b.SERIES_ID = ? AND `+visibility+`
ORDER BY `+seriesOrderExpr, archiveExcerptChars, model.StringValue(series, "ID"))
	if err != nil {
		return nil, nil, false, err
	}
	publicCount := 0
	for _, article := range articles {
		if isPublicBlogType(model.StringValue(article, "BLOG_TYPE")) {
			publicCount++
		}
	}
	series["PUBLIC_COUNT"] = publicCount
	return series, articles, true, nil
}

// GetSeries 是系列页接口：系列信息、文章目录（附摘要和首图）、当前用户能否管理。
// 没有公开文章的系列对其他人来说等于不存在。
func (s *Service) GetSeries(c *gin.Context, id string) model.Result {
	ctx := contextOf(c)
	if !s.seriesReady(ctx) {
		return model.Failure(msgSeriesNotFound)
	}
	series, found, err := s.loadSeries(ctx, id)
	if err != nil {
		return dbFailure("查询系列", err)
	}
	if !found {
		return model.Failure(msgSeriesNotFound)
	}
	canManage := false
	if actor, ok := s.CurrentActor(c); ok {
		canManage = actor.CanManage(model.StringValue(series, "USERCODE"))
	}
	series, articles, found, err := s.SeriesDetail(ctx, id, canManage)
	if err != nil {
		return dbFailure("查询系列", err)
	}
	if !found || (!canManage && model.IntValue(series, "PUBLIC_COUNT") == 0) {
		return model.Failure(msgSeriesNotFound)
	}
	s.attachTags(ctx, TagTargetBlog, articles)
	return model.Success(map[string]any{
		"series":    series,
		"articles":  replaceTextWithSummary(articles, "EXCERPT_HTML"),
		"canManage": canManage,
	})
}

// checkArticleSeries 检查文章能否放进这个系列（系列必须属于文章作者），返回整理后的系列编号；空字符串表示不属于任何系列
func (s *Service) checkArticleSeries(ctx context.Context, seriesID, ownerCode string) (string, *model.Result) {
	seriesID = normalizeSeriesID(seriesID)
	if seriesID == "" {
		return "", nil
	}
	if !s.seriesReady(ctx) {
		result := model.Failure(msgSeriesNotReady)
		return "", &result
	}
	series, found, err := s.loadSeries(ctx, seriesID)
	if err != nil {
		result := dbFailure("查询系列", err)
		return "", &result
	}
	if !found {
		result := model.Failure(msgSeriesNotFound)
		return "", &result
	}
	if model.StringValue(series, "USERCODE") != ownerCode {
		result := model.Failure(msgSeriesNotOwner)
		return "", &result
	}
	return seriesID, nil
}

// assignArticleSeries 把文章放进系列（排在最后）或移出系列；已经在这个系列里时不改变顺序
func (s *Service) assignArticleSeries(ctx context.Context, guid, seriesID string) error {
	if !s.seriesReady(ctx) || guid == "" {
		return nil
	}
	current, err := s.Repo.Query(ctx, "SELECT SERIES_ID FROM blogInfo WHERE GUID = ? LIMIT 1", guid)
	if err != nil {
		return err
	}
	previous := ""
	if len(current) > 0 {
		previous = normalizeSeriesID(model.StringValue(current[0], "SERIES_ID"))
	}
	if previous == seriesID {
		return nil
	}
	if seriesID == "" {
		_, err = s.Repo.Exec(ctx, "UPDATE blogInfo SET SERIES_ID = NULL, SERIES_ORDER = NULL WHERE GUID = ?", guid)
	} else {
		_, _, err = s.Repo.ExecuteBatch(ctx, []string{
			"UPDATE blogInfo b JOIN (SELECT COALESCE(MAX(SERIES_ORDER), 0) + 1 AS NEXT_ORDER FROM blogInfo WHERE SERIES_ID = ?) n SET b.SERIES_ID = ?, b.SERIES_ORDER = n.NEXT_ORDER WHERE b.GUID = ?",
			"UPDATE blogSeries SET UPDATE_TIME = NOW() WHERE ID = ?",
		}, [][]any{{seriesID, seriesID, guid}, {seriesID}}, false)
	}
	if err == nil {
		for _, id := range []string{previous, seriesID} {
			if id != "" {
				s.notifySeriesChanged(ctx, id)
			}
		}
	}
	return err
}

// SeriesNav 是文章页的系列信息：系列名、本文是第几篇、上一篇和下一篇、完整目录
type SeriesNav struct {
	ID       string           `json:"ID"`
	Title    string           `json:"TITLE"`
	Total    int              `json:"TOTAL"`
	Index    int              `json:"INDEX"` // 从 1 开始；本文不在可见目录里时为 0
	Previous map[string]any   `json:"PREV"`
	Next     map[string]any   `json:"NEXT"`
	Items    []map[string]any `json:"ITEMS"`
}

// ArticleSeriesNav 返回文章所属系列的导航；文章不属于系列时返回 nil。includePrivate 见 SeriesDetail。
func (s *Service) ArticleSeriesNav(ctx context.Context, article map[string]any, includePrivate bool) *SeriesNav {
	seriesID := normalizeSeriesID(model.StringValue(article, "SERIES_ID"))
	if seriesID == "" || !s.seriesReady(ctx) {
		return nil
	}
	series, articles, found, err := s.SeriesDetail(ctx, seriesID, includePrivate)
	if err != nil || !found || len(articles) == 0 {
		return nil
	}
	nav := &SeriesNav{ID: seriesID, Title: model.StringValue(series, "TITLE"), Total: len(articles), Items: make([]map[string]any, 0, len(articles))}
	guid := model.StringValue(article, "GUID")
	for index, item := range articles {
		entry := map[string]any{"GUID": model.StringValue(item, "GUID"), "BLOG_TITLE": model.StringValue(item, "BLOG_TITLE"), "BLOG_TYPE": model.StringValue(item, "BLOG_TYPE")}
		nav.Items = append(nav.Items, entry)
		if entry["GUID"] == guid {
			nav.Index = index + 1
		}
	}
	if nav.Index > 1 {
		nav.Previous = nav.Items[nav.Index-2]
	}
	if nav.Index > 0 && nav.Index < len(nav.Items) {
		nav.Next = nav.Items[nav.Index]
	}
	return nav
}

// attachSeriesBadges 给文章列表的每一行加上 SERIES（系列编号、名称、第几篇、共几篇，只按公开文章计算），不属于系列时为 nil
func (s *Service) attachSeriesBadges(ctx context.Context, rows []map[string]any) {
	ids := make([]any, 0)
	seen := make(map[string]bool)
	for _, row := range rows {
		row["SERIES"] = nil
		if id := normalizeSeriesID(model.StringValue(row, "SERIES_ID")); id != "" && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 || !s.seriesReady(ctx) {
		return
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	titles, err := s.Repo.Query(ctx, "SELECT ID, TITLE FROM blogSeries WHERE ID IN ("+placeholders+")", ids...)
	if err != nil {
		return
	}
	members, err := s.Repo.Query(ctx, "SELECT b.GUID, b.SERIES_ID FROM blogInfo b WHERE b.BLOG_TYPE = 'public' AND b.SERIES_ID IN ("+placeholders+") ORDER BY b.SERIES_ID, "+seriesOrderExpr, ids...)
	if err != nil {
		return
	}
	titleByID := make(map[string]string, len(titles))
	for _, row := range titles {
		titleByID[model.StringValue(row, "ID")] = model.StringValue(row, "TITLE")
	}
	indexByGUID := make(map[string]int)
	totalByID := make(map[string]int)
	for _, member := range members {
		id := model.StringValue(member, "SERIES_ID")
		totalByID[id]++
		indexByGUID[model.StringValue(member, "GUID")] = totalByID[id]
	}
	for _, row := range rows {
		id := normalizeSeriesID(model.StringValue(row, "SERIES_ID"))
		if title, ok := titleByID[id]; ok && id != "" {
			row["SERIES"] = map[string]any{"ID": id, "TITLE": title, "INDEX": indexByGUID[model.StringValue(row, "GUID")], "TOTAL": totalByID[id]}
		}
	}
}

// GetIndexableSeries 返回公开文章数达到收录标准的系列，用于站点地图
func (s *Service) GetIndexableSeries(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.PublicSeriesList(ctx, "", "latest", 200)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if model.IntValue(row, "ARTICLE_COUNT") >= SeriesIndexableMin {
			result = append(result, row)
		}
	}
	return result, nil
}

// SeriesPath 是系列页的地址
func SeriesPath(id string) string {
	return "/series/" + id
}

// notifySeriesChanged 系列内容变化（改名、调整顺序、文章加入或移出）时通知搜索引擎重新抓取系列页
func (s *Service) notifySeriesChanged(ctx context.Context, seriesID string) {
	if seriesID = normalizeSeriesID(seriesID); seriesID != "" {
		s.SubmitIndexNow(SeriesPath(seriesID))
	}
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
