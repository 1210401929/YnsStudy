package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/model"
)

var imageSourcePattern = regexp.MustCompile(`(?i)<img\b[^>]*?\bsrc\s*=\s*["']([^"']+)["']`)

func (s *Service) AddBlog(c *gin.Context, blog map[string]any) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	blog = normalizeNewBlog(blog)
	// 作者以登录用户为准，不信任请求体中的 USERCODE/USERNAME
	blog["USERCODE"] = actor.Code()
	blog["USERNAME"] = actor.User.Name
	// 正文是编辑器生成的 HTML，保存前过滤脚本和事件属性
	if mainText := model.Lookup(blog, "MAINTEXT"); mainText != nil {
		blog["MAINTEXT"] = sanitizeRichText(fmt.Sprint(mainText))
	}
	if s.hasSearchText(ctx, "blogInfo") {
		blog[searchTextColumn] = htmlToSearchText(model.StringValue(blog, "MAINTEXT"))
	}
	if fmt.Sprint(blog["GUID"]) == "" || blog["GUID"] == nil {
		tx, err := s.Repo.DB().BeginTx(ctx, nil)
		if err != nil {
			return dbFailure("生成文章编号", err)
		}
		defer tx.Rollback()
		result, err := tx.ExecContext(ctx, "UPDATE blogInfo_guid_sequence SET current_value = LAST_INSERT_ID(current_value + 1) WHERE table_name = ?", "bloginfo")
		if err != nil {
			return dbFailure("生成文章编号", err)
		}
		updated, _ := result.RowsAffected()
		if updated > 0 {
			var nextID int64
			if err := tx.QueryRowContext(ctx, "SELECT LAST_INSERT_ID()").Scan(&nextID); err != nil {
				return dbFailure("读取文章编号", err)
			}
			blog["GUID"] = fmt.Sprint(nextID)
		}
		if err := tx.Commit(); err != nil {
			return dbFailure("提交文章编号", err)
		}
	}
	result := stripSearchTextResult(s.SaveAll(ctx, "add", "BLOGINFO", []map[string]any{blog}, "GUID"))
	if !result.IsError && isPublicBlogType(model.StringValue(blog, "BLOG_TYPE")) {
		guid := model.StringValue(blog, "GUID")
		s.NotifyBlogChanged(guid)
		s.notifyAdminContent(ctx, actor.Code(), actor.User.Name, "新文章", model.StringValue(blog, "BLOG_TITLE"),
			model.StringValue(blog, "MAINTEXT"), s.mailSiteURL("/oneBlog/"+guid))
	}
	return result
}

func isPublicBlogType(blogType string) bool {
	return strings.EqualFold(strings.TrimSpace(blogType), "public")
}

// normalizeNewBlog reproduces the field filtering previously provided by the
// Java BlogBean and accepts the CATEGORY_ID alias used by older frontends.
func normalizeNewBlog(input map[string]any) map[string]any {
	blog := make(map[string]any)
	for _, field := range []string{"GUID", "BLOG_TITLE", "BLOG_TYPE", "MAINTEXT", "USERCODE", "USERNAME", "CAT_ID"} {
		if value := model.Lookup(input, field); value != nil {
			blog[field] = value
		}
	}
	if _, exists := blog["CAT_ID"]; !exists {
		if value := model.Lookup(input, "CATEGORY_ID"); value != nil {
			blog["CAT_ID"] = value
		}
	}
	return blog
}

func (s *Service) GetCurrentUserBlogs(c *gin.Context) model.Result {
	user, err := s.CurrentUser(c)
	if err != nil {
		return model.Failure("用户未登录!")
	}
	return s.GetBlogsByUser(contextOf(c), user.Code, false)
}

func (s *Service) GetBlogsByUser(ctx context.Context, userCode string, publicOnly bool) model.Result {
	if strings.TrimSpace(userCode) == "" {
		return model.Failure("用户编码出错!")
	}
	query := "SELECT GUID, BLOG_TITLE, BLOG_TYPE, USERCODE, USERNAME, VIEW_PAGE, CREATE_TIME, CAT_ID FROM blogInfo WHERE USERCODE = ?"
	if publicOnly {
		query += " AND BLOG_TYPE = 'public'"
	}
	query += " ORDER BY CASE BLOG_TYPE WHEN 'public' THEN 0 WHEN 'privacy' THEN 1 ELSE 2 END, CREATE_TIME DESC"
	return s.SelectList(ctx, query, []any{userCode})
}

func (s *Service) UpdateBlogCategory(c *gin.Context, blogID, categoryID string) model.Result {
	blog, failure := s.requireOwnerOrAdmin(c, "blogInfo", blogID)
	if failure != nil {
		return *failure
	}
	// 只能归入文章作者自己的分类
	if categoryID != "" {
		category, failure := s.loadRow(c, "blogCatInfo", categoryID)
		if failure != nil {
			return *failure
		}
		if model.StringValue(category, "USERCODE") != model.StringValue(blog, "USERCODE") {
			return model.Failure(msgForbidden)
		}
	}
	ctx := contextOf(c)
	var value any = categoryID
	if categoryID == "" {
		value = nil
	}
	return s.ExecuteSQL(ctx, "UPDATE blogInfo SET CAT_ID = ? WHERE GUID = ?", []any{value, blogID})
}

func (s *Service) GetBlog(c *gin.Context, blogID string) model.Result {
	userCode := ""
	if user, err := s.CurrentUser(c); err == nil {
		userCode = user.Code
	}
	query := `SELECT b.*, u.AVATAR, u.USERNUM FROM blogInfo b
LEFT JOIN userInfo u ON b.USERCODE = u.CODE
WHERE b.GUID = ? AND (b.USERCODE = ? OR b.BLOG_TYPE = 'public')`
	rows, err := s.Repo.Query(contextOf(c), query, blogID, userCode)
	if err != nil {
		return dbFailure("查询文章", err)
	}
	if len(rows) > 0 {
		_, _ = s.Repo.Exec(contextOf(c), "UPDATE blogInfo SET VIEW_PAGE = COALESCE(VIEW_PAGE, 0) + 1 WHERE GUID = ?", blogID)
	}
	return model.Success(stripSearchText(rows))
}

// GetPublicBlogForSEO 只读取公开文章，不增加阅读量。
// 搜索引擎抓取会比较频繁，如果复用普通详情接口会把爬虫访问错误地计入 VIEW_PAGE。
func (s *Service) GetPublicBlogForSEO(ctx context.Context, blogID string) (map[string]any, bool, error) {
	rows, err := s.Repo.Query(ctx, `SELECT b.*, u.AVATAR, u.USERNUM
FROM blogInfo b
LEFT JOIN userInfo u ON b.USERCODE = u.CODE
WHERE b.GUID = ? AND b.BLOG_TYPE = 'public'
LIMIT 1`, strings.TrimSpace(blogID))
	if err != nil {
		return nil, false, err
	}
	if len(rows) == 0 {
		return nil, false, nil
	}
	return rows[0], true, nil
}

func (s *Service) GetAllBlogs(ctx context.Context, page, pageSize int, keyword string) model.Result {
	page, pageSize = normalizePage(page, pageSize)
	where := " WHERE b.BLOG_TYPE = 'public'"
	listArgs := make([]any, 0, 4)
	countArgs := make([]any, 0, 2)
	if strings.TrimSpace(keyword) != "" {
		where += " AND (b.BLOG_TITLE LIKE ? OR " + s.searchField(ctx, "blogInfo", "b") + " LIKE ?)"
		like := likePattern(keyword)
		listArgs = append(listArgs, like, like)
		countArgs = append(countArgs, like, like)
	}
	listArgs = append(listArgs, pageSize, (page-1)*pageSize)
	rows, err := s.Repo.Query(ctx, "SELECT b.*, u.AVATAR, u.USERNUM FROM blogInfo b LEFT JOIN userInfo u ON b.USERCODE = u.CODE"+where+" ORDER BY b.CREATE_TIME DESC LIMIT ? OFFSET ?", listArgs...)
	if err != nil {
		return dbFailure("查询文章", err)
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM blogInfo b"+where, countArgs...)
	if err != nil {
		return dbFailure("统计文章", err)
	}
	return model.Success(map[string]any{"total": firstCount(counts), "data": stripSearchText(replaceTextWithSummary(rows, "MAINTEXT"))})
}

func (s *Service) UpdateBlog(c *gin.Context, guid, title, content, blogType string) model.Result {
	row, failure := s.requireOwnerOrAdmin(c, "blogInfo", guid)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	content = sanitizeRichText(content)
	rows := []map[string]any{row}
	wasPublic := isPublicBlogType(model.StringValue(rows[0], "BLOG_TYPE"))
	oldURLs := extractImageURLs(model.StringValue(rows[0], "MAINTEXT"))
	newURLs := extractImageURLs(content)
	sets := "BLOG_TITLE = ?, MAINTEXT = ?, BLOG_TYPE = ?"
	args := []any{title, content, blogType}
	if s.hasSearchText(ctx, "blogInfo") {
		sets += ", " + searchTextColumn + " = ?"
		args = append(args, htmlToSearchText(content))
	}
	if s.blogInfoHasUpdateTime(ctx) {
		// UPDATE_TIME 只在文章内容真正编辑时变化，阅读量增加不会污染 sitemap 的 lastmod。
		sets += ", UPDATE_TIME = NOW()"
	}
	if _, err := s.Repo.Exec(ctx, "UPDATE blogInfo SET "+sets+" WHERE GUID = ?", append(args, guid)...); err != nil {
		return dbFailure("修改文章", err)
	}
	// 文章保存成功后再删除不再使用的图片，保存失败时图片不会丢
	s.DeleteUploadedFiles(difference(oldURLs, newURLs))
	// 公开文章被修改或改为私密时都要通知，后者让搜索引擎发现页面已不可访问。
	if wasPublic || isPublicBlogType(blogType) {
		s.NotifyBlogChanged(guid)
	}
	// 私密文章改为公开，等同于新发表，同样通知站长审核
	if !wasPublic && isPublicBlogType(blogType) {
		s.notifyAdminContent(ctx, model.StringValue(row, "USERCODE"), model.StringValue(row, "USERNAME"), "文章改为公开", title,
			content, s.mailSiteURL("/oneBlog/"+guid))
	}
	return model.Success("执行成功，影响行数：1")
}

// blogInfoHasUpdateTime 让新代码在执行数据库迁移前后都能运行。
// 迁移前回退到 CREATE_TIME；执行迁移并重启服务后自动使用真实更新时间。
func (s *Service) blogInfoHasUpdateTime(ctx context.Context) bool {
	rows, err := s.Repo.Query(ctx, `SELECT COUNT(*) AS total
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND LOWER(TABLE_NAME) = 'bloginfo'
  AND UPPER(COLUMN_NAME) = 'UPDATE_TIME'`)
	return err == nil && firstCount(rows) > 0
}

func (s *Service) DeleteBlog(c *gin.Context, guid string) model.Result {
	if strings.TrimSpace(guid) == "" {
		return model.Failure("未传入删除主键!")
	}
	row, failure := s.requireOwnerOrAdmin(c, "blogInfo", guid)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	rows := []map[string]any{row}
	wasPublic := isPublicBlogType(model.StringValue(rows[0], "BLOG_TYPE"))
	_, _, err := s.Repo.ExecuteBatch(ctx,
		[]string{"DELETE FROM blogInfo WHERE GUID = ?", "DELETE FROM blogComment WHERE BLOGID = ?", "DELETE FROM blogGiveLike WHERE BLOGID = ?"},
		[][]any{{guid}, {guid}, {guid}}, false,
	)
	if err != nil {
		return dbFailure("删除文章", err)
	}
	s.DeleteUploadedFiles(extractImageURLs(model.StringValue(rows[0], "MAINTEXT")))
	if wasPublic {
		s.NotifyBlogChanged(guid)
	}
	return model.Success("删除成功")
}

func (s *Service) AddBlogCategory(c *gin.Context, category map[string]any) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	data := pickFields(category, "GUID", "CATNAME", "SUPERGUID", "REMARK", "ORDERNO")
	data["USERCODE"] = actor.Code()
	// 父分类必须属于自己
	if parent := strings.TrimSpace(model.StringValue(data, "SUPERGUID")); parent != "" {
		parentRow, failure := s.loadRow(c, "blogCatInfo", parent)
		if failure != nil {
			return *failure
		}
		if model.StringValue(parentRow, "USERCODE") != actor.Code() {
			return model.Failure(msgForbidden)
		}
	}
	return s.SaveAll(contextOf(c), "add", "BLOGCATINFO", []map[string]any{data}, "GUID")
}

func (s *Service) UpdateBlogCategoryInfo(c *gin.Context, category map[string]any) model.Result {
	guid := model.StringValue(category, "GUID")
	if _, failure := s.requireOwnerOrAdmin(c, "blogCatInfo", guid); failure != nil {
		return *failure
	}
	data := pickFields(category, "CATNAME", "REMARK", "ORDERNO")
	data["GUID"] = guid
	return s.SaveAll(contextOf(c), "edit", "BLOGCATINFO", []map[string]any{data}, "GUID")
}

func (s *Service) DeleteBlogCategory(c *gin.Context, guid string) model.Result {
	if _, failure := s.requireOwnerOrAdmin(c, "blogCatInfo", guid); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	_, _, err := s.Repo.ExecuteBatch(ctx,
		[]string{
			"UPDATE blogInfo SET CAT_ID = NULL WHERE CAT_ID IN (SELECT GUID FROM BLOGCATINFO WHERE GUID = ? OR SUPERGUID = ?)",
			"DELETE FROM BLOGCATINFO WHERE GUID = ? OR SUPERGUID = ?",
		}, [][]any{{guid, guid}, {guid, guid}}, false)
	if err != nil {
		return dbFailure("删除文章分类", err)
	}
	return model.Success("批量执行成功")
}

func (s *Service) GetUserBlogCategories(ctx context.Context, userCode string) model.Result {
	return s.SelectList(ctx, "SELECT * FROM blogCatInfo WHERE USERCODE = ? ORDER BY ORDERNO", []any{userCode})
}

func (s *Service) AddBlogComment(c *gin.Context, comment map[string]any) model.Result {
	ctx := contextOf(c)
	ip := ClientIP(c)
	limitKey := "comment_limit:" + ip

	currentCount := int64(0)
	currentValue, err := s.Cache.Get(ctx, limitKey)
	if err == nil {
		currentCount, err = strconv.ParseInt(currentValue, 10, 64)
		if err != nil {
			return model.Failure("评论限流计数格式异常: " + err.Error())
		}
	} else if !errors.Is(err, cache.ErrNotFound) {
		return model.Failure("评论限流服务异常: " + err.Error())
	}
	if currentCount >= int64(s.Config.Content.CommentLimit) {
		return model.Failure("评论过于频繁，请稍后再试！")
	}

	// 被回复的那条评论（不入库，只用于确定通知谁）
	replyTo := strings.TrimSpace(model.StringValue(comment, "REPLY_TO"))
	// 只保留评论表单会提交的字段，其余（如 CREATE_TIME）由数据库生成
	comment = pickFields(comment, "GUID", "BLOGID", "SUPERGUID", "TEXT", "USERCODE", "USERNAME",
		"USEREMAIL", "USERWEBSITE", "RECEIVE_USERCODE", "RECEIVE_USERNAME")
	// 邮箱用于回复通知，格式不对就不保存；网址只接受 http/https，防止 javascript: 链接
	comment["USEREMAIL"] = normalizeMailAddress(model.StringValue(comment, "USEREMAIL"))
	comment["USERWEBSITE"] = normalizeWebsite(model.StringValue(comment, "USERWEBSITE"))
	// 登录用户以当前账号身份评论；匿名评论不能冒用任何已注册账号
	if actor, ok := s.CurrentActor(c); ok {
		comment["USERCODE"] = actor.Code()
		comment["USERNAME"] = actor.User.Name
	} else {
		delete(comment, "USERCODE")
		delete(comment, "usercode")
	}

	for _, field := range []string{"TEXT", "USERWEBSITE", "USERNAME"} {
		if s.containsBannedWord(fmt.Sprint(comment[field])) {
			return model.Failure("提交的内容包含违禁词，禁止发布！")
		}
	}

	result := s.SaveAll(ctx, "add", "BLOGCOMMENT", []map[string]any{comment}, "GUID")
	if result.IsError {
		return result
	}

	window := time.Duration(s.Config.Content.CommentWindowSecond) * time.Second
	if _, err := s.Cache.Increment(ctx, limitKey, window); err != nil {
		// 评论已经保存，避免向客户端返回失败而导致用户重复提交。
		log.Printf("评论已保存，但更新限流计数失败: %v", err)
	}
	s.notifyBlogComment(ctx, comment, replyTo)
	return result
}

// notifyBlogComment 给文章作者或被回复的人发站内消息和邮件，并给站长发邮件。
// 匿名访客没有登录，无法调用发消息接口，所以评论的通知统一由这里发送；
// 接收人从数据库确认，不信任请求里的 RECEIVE_USERCODE。发送失败只记日志，不影响评论。
func (s *Service) notifyBlogComment(ctx context.Context, comment map[string]any, replyTo string) {
	blogID := model.StringValue(comment, "BLOGID")
	blogs, err := s.Repo.Query(ctx, "SELECT USERCODE, BLOG_TITLE FROM blogInfo WHERE GUID = ? LIMIT 1", blogID)
	if err != nil || len(blogs) == 0 {
		return
	}
	title := model.StringValue(blogs[0], "BLOG_TITLE")
	authorCode := model.StringValue(blogs[0], "USERCODE")
	sender := strings.TrimSpace(model.StringValue(comment, "USERCODE"))
	senderEmail := model.StringValue(comment, "USEREMAIL")
	senderName := strings.TrimSpace(model.StringValue(comment, "USERNAME"))
	if senderName == "" {
		senderName = "访客"
	}
	text := strings.TrimSpace(model.StringValue(comment, "TEXT"))
	link := s.mailSiteURL("/oneBlog/" + blogID)

	// receiverCode 收站内消息；receiverEmail 收邮件（匿名访客只有邮箱）
	receiverCode, receiverEmail, remark := "", "", ""
	// 回复时被回复的那条原评论，邮件里一并引用
	originalName, originalText := "", ""
	if parentID := strings.TrimSpace(model.StringValue(comment, "SUPERGUID")); parentID == "" {
		receiverCode = authorCode
		remark = senderName + "评论了你的文章《" + title + "》"
	} else {
		// 找不到被回复的评论时只是不通知被回复人，站长邮件照常发送
		if target := s.findReplyTarget(ctx, blogID, parentID, replyTo, model.StringValue(comment, "RECEIVE_USERCODE")); target != nil {
			originalName = strings.TrimSpace(model.StringValue(target, "USERNAME"))
			if originalName == "" {
				originalName = "访客"
			}
			originalText = mailExcerpt(strings.TrimSpace(model.StringValue(target, "TEXT")), 300)
			receiverCode = strings.TrimSpace(model.StringValue(target, "USERCODE"))
			if receiverCode == "" {
				receiverEmail = normalizeMailAddress(model.StringValue(target, "USEREMAIL"))
			}
		} else {
			log.Printf("回复评论时未找到被回复的评论: blog=%s parent=%s replyTo=%s", blogID, parentID, replyTo)
		}
		remark = senderName + "回复了你在文章《" + title + "》下的评论"
	}
	// 自己评论自己的文章、回复自己不通知
	isSelf := (receiverCode != "" && receiverCode == sender) ||
		(receiverCode == "" && receiverEmail != "" && receiverEmail == senderEmail)
	if !isSelf && receiverCode != "" {
		noticeSender := sender
		if noticeSender == "" {
			noticeSender = "guest"
		}
		if _, err := s.Repo.Exec(ctx, "INSERT INTO noticeInfo (SENDUSERCODE, RECEIVERUSERCODE, TYPE, EXECUTE, REMARK) VALUES (?, ?, 'comment', ?, ?)",
			noticeSender, receiverCode, "/oneBlog/"+blogID, remark); err != nil {
			log.Printf("评论已保存，但发送消息失败: %v", err)
		}
		receiverEmail = s.userMailAddress(ctx, receiverCode)
	}

	adminEmail := normalizeMailAddress(s.adminMailAddress())
	senderIsAdmin := sender != "" && sender == s.superAdminCode()
	quote := mailExcerpt(text, 300)
	if !isSelf && receiverEmail != "" && receiverEmail != adminEmail {
		s.queueMail(ctx, receiverEmail, remark,
			renderReplyMail(remark, nil, "你的评论：", originalText, quote, "查看文章", link))
	}
	// 站长接收全站评论，用于审核内容；站长自己发的评论不再通知自己
	if senderIsAdmin {
		log.Printf("评论由站长本人发表，不发站长通知邮件: blog=%s", blogID)
	}
	if adminEmail != "" && !senderIsAdmin {
		kind := "评论"
		if model.StringValue(comment, "SUPERGUID") != "" {
			kind = "回复"
		}
		lines := []string{"文章：《" + title + "》", "评论人：" + senderName}
		if sender == "" {
			lines = append(lines, "身份：匿名访客")
			if senderEmail != "" {
				lines = append(lines, "邮箱："+senderEmail)
			}
		}
		if website := model.StringValue(comment, "USERWEBSITE"); website != "" {
			lines = append(lines, "网址："+website)
		}
		s.queueMail(ctx, adminEmail, "【新"+kind+"】"+senderName+"："+mailExcerpt(text, 30),
			renderReplyMail("《"+title+"》有新"+kind, lines, originalName+" 的原评论：", originalText, quote, "查看文章", link))
	}
}

// findReplyTarget 找到被回复的那条评论，必须在同一篇文章的同一楼层里。
// 优先用前端传来的被回复评论编号；旧版前端没有传时，退回按 RECEIVE_USERCODE 查找。
func (s *Service) findReplyTarget(ctx context.Context, blogID, parentID, replyTo, receiverCode string) map[string]any {
	if replyTo != "" {
		rows, err := s.Repo.Query(ctx, "SELECT USERCODE, USERNAME, USEREMAIL, TEXT FROM BLOGCOMMENT WHERE BLOGID = ? AND GUID = ? AND (GUID = ? OR SUPERGUID = ?) LIMIT 1", blogID, replyTo, parentID, parentID)
		if err == nil && len(rows) > 0 {
			return rows[0]
		}
	}
	receiverCode = strings.TrimSpace(receiverCode)
	if receiverCode == "" {
		return nil
	}
	rows, err := s.Repo.Query(ctx, "SELECT USERCODE, USERNAME, USEREMAIL, TEXT FROM BLOGCOMMENT WHERE BLOGID = ? AND (GUID = ? OR SUPERGUID = ?) AND USERCODE = ? LIMIT 1", blogID, parentID, parentID, receiverCode)
	if err != nil || len(rows) == 0 {
		return nil
	}
	return rows[0]
}

// normalizeWebsite 规范化评论者填写的网址：只接受 http/https，没写协议时补上 https://
func normalizeWebsite(website string) string {
	website = strings.TrimSpace(website)
	if website == "" || len(website) > 255 {
		return ""
	}
	if !strings.Contains(website, "://") {
		website = "https://" + website
	}
	parsed, err := url.Parse(website)
	if err != nil || parsed.Host == "" {
		return ""
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return ""
	}
	return website
}

func (s *Service) GetBlogComments(ctx context.Context, blogID string) model.Result {
	query := `SELECT bc.*, ui.AVATAR
FROM BLOGCOMMENT bc
LEFT JOIN userInfo ui ON bc.USERCODE = ui.CODE
LEFT JOIN BLOGCOMMENT p ON bc.SUPERGUID = p.GUID
WHERE bc.BLOGID = ?
ORDER BY COALESCE(p.CREATE_TIME, bc.CREATE_TIME) DESC,
CASE WHEN bc.SUPERGUID IS NULL OR bc.SUPERGUID = '' THEN 0 ELSE 1 END ASC,
bc.CREATE_TIME ASC`
	result := s.SelectList(ctx, query, []any{blogID})
	// 评论者留的邮箱只用于回复通知，不对外返回
	if rows, ok := result.Result.([]map[string]any); ok {
		for _, row := range rows {
			delete(row, "USEREMAIL")
		}
	}
	return result
}

func (s *Service) DeleteBlogComment(c *gin.Context, guid string) model.Result {
	if _, failure := s.requireOwnerOrAdmin(c, "blogComment", guid); failure != nil {
		return *failure
	}
	return s.ExecuteSQL(contextOf(c), "DELETE FROM BLOGCOMMENT WHERE GUID = ? OR SUPERGUID = ?", []any{guid, guid})
}

func (s *Service) AddBlogReaction(c *gin.Context, blogID, reactionType string) model.Result {
	user, err := s.CurrentUser(c)
	if err != nil {
		return model.Failure("用户未登录!")
	}
	return s.SaveAll(contextOf(c), "add", "blogGiveLike", []map[string]any{{
		"BLOGID": blogID, "USERCODE": user.Code, "USERNAME": user.Name, "TYPE": reactionType,
	}}, "GUID")
}

func (s *Service) RemoveBlogReaction(c *gin.Context, blogID, reactionType string) model.Result {
	user, err := s.CurrentUser(c)
	if err != nil {
		return model.Failure("用户未登录!")
	}
	return s.ExecuteSQL(contextOf(c), "DELETE FROM blogGiveLike WHERE BLOGID = ? AND USERCODE = ? AND TYPE = ?", []any{blogID, user.Code, reactionType})
}

func (s *Service) GetBlogReactions(ctx context.Context, blogID, reactionType string) model.Result {
	query := "SELECT * FROM blogGiveLike WHERE BLOGID = ?"
	args := []any{blogID}
	if reactionType != "" {
		query += " AND TYPE = ?"
		args = append(args, reactionType)
	}
	return s.SelectList(ctx, query, args)
}

func (s *Service) GetUserReactions(ctx context.Context, userCode string, countOnly bool, reactionType string) model.Result {
	if countOnly {
		return s.SelectList(ctx, "SELECT TYPE, COUNT(1) AS TOTAL FROM blogGiveLike WHERE USERCODE = ? GROUP BY TYPE", []any{userCode})
	}
	query := `SELECT b.GUID, b.BLOG_TITLE, b.BLOG_TYPE, b.USERCODE, b.USERNAME, b.VIEW_PAGE, b.CREATE_TIME, g.TYPE
FROM blogInfo b JOIN blogGiveLike g ON g.BLOGID = b.GUID WHERE g.USERCODE = ?`
	args := []any{userCode}
	if reactionType != "" {
		query += " AND g.TYPE = ?"
		args = append(args, reactionType)
	}
	return s.SelectList(ctx, query, args)
}

func (s *Service) containsBannedWord(content string) bool {
	content = strings.ToLower(strings.TrimSpace(content))
	for _, word := range s.Config.Content.BannedWords {
		if word != "" && strings.Contains(content, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

func extractImageURLs(html string) []string {
	matches := imageSourcePattern.FindAllStringSubmatch(html, -1)
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			result = append(result, match[1])
		}
	}
	return result
}

func difference(left, right []string) []string {
	rightSet := make(map[string]struct{}, len(right))
	for _, value := range right {
		rightSet[value] = struct{}{}
	}
	result := make([]string, 0)
	for _, value := range left {
		if _, found := rightSet[value]; !found {
			result = append(result, value)
		}
	}
	return result
}
