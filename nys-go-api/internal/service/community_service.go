package service

// 便签墙（原“社区”）：随记的发布、列表、单条详情和评论。
// 数据表沿用 communityInfo / communityComment，接口地址沿用 /blog-api/community/*。

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

// 随记类型；旧数据 TYPE 为空，按随笔处理
const (
	NoteTypeNote     = "note"
	NoteTypeTip      = "tip"
	NoteTypeShare    = "share"
	NoteTypeQuestion = "question"

	maxNoteImages = 9
)

// 生成标题时跳过代码块，标题只取文字部分
var noteCodeBlockPattern = regexp.MustCompile(`(?is)<pre\b.*?</pre>`)

// 正文以标题开头时（例如 Markdown 的 # 标题），直接用这个标题
var noteLeadingHeadingPattern = regexp.MustCompile(`(?is)^\s*<h[1-6]\b[^>]*>(.*?)</h[1-6]>`)

var noteTypes = map[string]bool{NoteTypeNote: true, NoteTypeTip: true, NoteTypeShare: true, NoteTypeQuestion: true}

func normalizeNoteType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if noteTypes[value] {
		return value
	}
	return NoteTypeNote
}

// NotePath 是单条随记的页面地址
func NotePath(guid string) string {
	return "/note/" + guid
}

// noteImagePrefix 是随记图片允许使用的地址前缀：只接受本站上传到 editorImage 的图片
func (s *Service) noteImagePrefix() string {
	return strings.TrimSuffix(s.Config.Upload.PublicPrefix, "/") + "/editorImage/"
}

// normalizeNoteImages 接受数组或 JSON 字符串，去重并过滤掉非本站上传的地址，最多保留 9 张
func (s *Service) normalizeNoteImages(value any) []string {
	var raw []any
	switch typed := value.(type) {
	case []any:
		raw = typed
	case []string:
		for _, item := range typed {
			raw = append(raw, item)
		}
	case string:
		_ = json.Unmarshal([]byte(typed), &raw)
	}
	prefix := s.noteImagePrefix()
	seen := make(map[string]bool)
	images := make([]string, 0, len(raw))
	for _, item := range raw {
		image := strings.TrimSpace(fmt.Sprint(item))
		// 文件名由上传接口随机生成，不应包含 .. 或查询参数
		if !strings.HasPrefix(image, prefix) || strings.Contains(image, "..") || strings.ContainsAny(image, "?#\"'<> ") || seen[image] {
			continue
		}
		seen[image] = true
		images = append(images, image)
		if len(images) == maxNoteImages {
			break
		}
	}
	return images
}

// decodeNoteRows 把 IMAGES 从 JSON 字符串转成数组，并补齐 TYPE，前端可以直接使用
func decodeNoteRows(rows []map[string]any) []map[string]any {
	for _, row := range rows {
		images := []string{}
		if raw := model.StringValue(row, "IMAGES"); raw != "" {
			_ = json.Unmarshal([]byte(raw), &images)
		}
		row["IMAGES"] = images
		row["TYPE"] = normalizeNoteType(model.StringValue(row, "TYPE"))
	}
	return rows
}

// NoteTitle 给随记生成一个标题：有 TITLE 用 TITLE，否则取正文（不含代码块）前 40 个字
func NoteTitle(row map[string]any) string {
	if title := strings.TrimSpace(model.StringValue(row, "TITLE")); title != "" {
		return title
	}
	text := model.StringValue(row, "TEXT")
	if match := noteLeadingHeadingPattern.FindStringSubmatch(text); match != nil {
		if heading, _ := summarizeHTML(match[1], 40); strings.TrimSpace(heading) != "" {
			return strings.TrimSpace(heading)
		}
	}
	excerpt, _ := summarizeHTML(noteCodeBlockPattern.ReplaceAllString(text, " "), 40)
	if excerpt = strings.TrimSpace(excerpt); excerpt != "" {
		return excerpt
	}
	return "一张便签"
}

// noteContent 校验并整理随记内容：正文（Markdown 转出的 HTML）、Markdown 原文、类型和图片。
// 发布和编辑共用；只返回数据库里已有的字段（迁移 009/010 之前的表没有类型、图片、原文）。
func (s *Service) noteContent(ctx context.Context, input map[string]any) (map[string]any, *model.Result) {
	images := s.normalizeNoteImages(model.Lookup(input, "IMAGES"))
	// 正文是 Markdown 转出的 HTML，保存前过滤脚本和事件属性
	text := sanitizeRichText(model.StringValue(input, "TEXT"))
	plain, _ := summarizeHTML(text, 10)
	if strings.TrimSpace(plain) == "" && len(images) == 0 && !strings.Contains(text, "<img") {
		result := model.Failure("便签内容不能为空")
		return nil, &result
	}
	source := model.StringValue(input, "SOURCE")
	if s.containsBannedWord(text) || s.containsBannedWord(source) {
		result := model.Failure(msgBannedWord)
		return nil, &result
	}
	fields := map[string]any{"TEXT": text}
	if s.tableHasColumn(ctx, "communityInfo", "TYPE") {
		fields["TYPE"] = normalizeNoteType(model.StringValue(input, "TYPE"))
	}
	if s.tableHasColumn(ctx, "communityInfo", "IMAGES") {
		if len(images) > 0 {
			encoded, _ := json.Marshal(images)
			fields["IMAGES"] = string(encoded)
		} else {
			fields["IMAGES"] = nil
		}
	}
	if s.tableHasColumn(ctx, "communityInfo", "SOURCE") {
		if len(source) > 60000 {
			source = source[:60000]
		}
		fields["SOURCE"] = source
	}
	if s.hasSearchText(ctx, "communityInfo") {
		fields[searchTextColumn] = htmlToSearchText(text)
	}
	return fields, nil
}

func (s *Service) AddCommunity(c *gin.Context, community map[string]any) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	// 只接受内容字段；作者以登录账号为准，置顶只能由超级管理员通过 setTopCommunity 设置
	fields, failure := s.noteContent(ctx, community)
	if failure != nil {
		return *failure
	}
	fields["GUID"] = model.StringValue(community, "GUID")
	fields["USERCODE"] = actor.Code()
	fields["USERNAME"] = actor.User.Name
	if fields["IMAGES"] == nil {
		delete(fields, "IMAGES")
	}
	// 新随记使用数字编号，地址更短；编号表里没有随记这一行时继续使用前端生成的编号
	if nextID, err := s.nextSequenceID(ctx, "communityinfo"); err != nil {
		log.Printf("生成随记编号失败，使用前端编号: %v", err)
	} else if nextID != "" {
		fields["GUID"] = nextID
	}
	guid := model.StringValue(fields, "GUID")
	if guid == "" {
		return model.Failure("缺少便签编号")
	}
	result := s.SaveAll(ctx, "add", "communityInfo", []map[string]any{fields}, "GUID")
	if result.IsError {
		return stripSearchTextResult(result)
	}
	s.notifyAdminContent(ctx, actor.Code(), actor.User.Name, "便签墙新随记", "", model.StringValue(fields, "TEXT"), s.mailSiteURL(NotePath(guid)))
	return s.noteSavedResult(ctx, guid)
}

// UpdateCommunity 编辑随记。权限：自己的可以改；管理员可以改除超级管理员以外所有人的；超级管理员都能改。
func (s *Service) UpdateCommunity(c *gin.Context, community map[string]any) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	guid := model.StringValue(community, "GUID")
	row, failure := s.loadRow(c, "communityInfo", guid)
	if failure != nil {
		return *failure
	}
	if !actor.CanManage(model.StringValue(row, "USERCODE")) {
		return model.Failure(msgForbidden)
	}
	ctx := contextOf(c)
	fields, failure := s.noteContent(ctx, community)
	if failure != nil {
		return *failure
	}
	columns := make([]string, 0, len(fields)+1)
	args := make([]any, 0, len(fields)+2)
	for _, column := range []string{"TEXT", "TYPE", "IMAGES", "SOURCE", searchTextColumn} {
		if value, ok := fields[column]; ok {
			columns = append(columns, column+" = ?")
			args = append(args, value)
		}
	}
	if s.tableHasColumn(ctx, "communityInfo", "UPDATE_TIME") {
		columns = append(columns, "UPDATE_TIME = NOW()")
	}
	args = append(args, guid)
	if _, err := s.Repo.Exec(ctx, "UPDATE communityInfo SET "+strings.Join(columns, ", ")+" WHERE GUID = ?", args...); err != nil {
		return dbFailure("修改随记", err)
	}
	s.logModeration(c, actor, row, AdminActionEditPost, postLabel(row), "在便签墙修改随记")
	return s.noteSavedResult(ctx, guid)
}

// noteSavedResult 返回保存后的随记，前端直接用它刷新页面
func (s *Service) noteSavedResult(ctx context.Context, guid string) model.Result {
	note, found, err := s.FindNote(ctx, guid)
	if err != nil || !found {
		return model.Success(map[string]any{"GUID": guid})
	}
	return model.Success(note)
}

// DeleteCommunity 的规则与前端一致：超级管理员可删除任何帖子；置顶帖只有超级管理员能删；
// 非置顶帖作者本人可删，管理员可删除非超级管理员发布的帖子。
func (s *Service) DeleteCommunity(c *gin.Context, guid string) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	row, failure := s.loadRow(c, "communityInfo", guid)
	if failure != nil {
		return *failure
	}
	isTop := model.StringValue(row, "ISTOP") == "1"
	if !actor.IsSuper && (isTop || !actor.CanManage(model.StringValue(row, "USERCODE"))) {
		return model.Failure(msgForbidden)
	}
	ctx := contextOf(c)
	_, _, err := s.Repo.ExecuteBatch(ctx,
		[]string{"DELETE FROM communityInfo WHERE GUID = ?", "DELETE FROM communityComment WHERE COMMUNITYID = ?"},
		[][]any{{guid}, {guid}}, false,
	)
	if err != nil {
		return dbFailure("删除随记", err)
	}
	s.logModeration(c, actor, row, AdminActionDeletePost, postLabel(row), "在便签墙删除随记")
	return model.Success("删除成功")
}

func (s *Service) SetTopCommunity(c *gin.Context, guid, isTop string) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	var value any
	if isTop == "1" {
		value = "1"
	} else if isTop == "0" {
		value = nil
	} else {
		return model.Failure("isTop 只允许为 0 或 1")
	}
	return s.ExecuteSQL(ctx, "UPDATE communityInfo SET ISTOP = ? WHERE GUID = ?", []any{value, guid})
}

// NoteQuery 是便签墙列表的筛选条件
type NoteQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Type     string // 空表示全部
	Sort     string // latest 最新（默认）、hot 最多讨论
	Mine     bool   // 只看自己发的
}

// noteSelect 是随记列表和详情共用的查询字段：作者头像、作者主页编号、评论数
const noteSelect = `SELECT b.*, u.AVATAR, u.USERNUM,
(SELECT COUNT(1) FROM communityComment cc WHERE cc.COMMUNITYID = b.GUID) AS COMMENTNUM
FROM communityInfo b LEFT JOIN userInfo u ON b.USERCODE = u.CODE`

func (s *Service) GetAllCommunities(c *gin.Context, query NoteQuery) model.Result {
	ctx := contextOf(c)
	page, pageSize := normalizePage(query.Page, query.PageSize)
	conditions := make([]string, 0, 3)
	args := make([]any, 0, 6)
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		conditions = append(conditions, "(b.TITLE LIKE ? OR "+s.searchField(ctx, "communityInfo", "b")+" LIKE ?)")
		like := likePattern(keyword)
		args = append(args, like, like)
	}
	if noteType := strings.TrimSpace(query.Type); noteType != "" && s.tableHasColumn(ctx, "communityInfo", "TYPE") {
		noteType = normalizeNoteType(noteType)
		if noteType == NoteTypeNote {
			conditions = append(conditions, "(b.TYPE IS NULL OR b.TYPE = '' OR b.TYPE = ?)")
		} else {
			conditions = append(conditions, "b.TYPE = ?")
		}
		args = append(args, noteType)
	}
	if query.Mine {
		actor, failure := s.requireLogin(c)
		if failure != nil {
			return *failure
		}
		conditions = append(conditions, "b.USERCODE = ?")
		args = append(args, actor.Code())
	}
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}
	order := " ORDER BY b.ISTOP DESC, b.CREATE_TIME DESC"
	if query.Sort == "hot" {
		order = " ORDER BY b.ISTOP DESC, COMMENTNUM DESC, b.CREATE_TIME DESC"
	}
	rows, err := s.Repo.Query(ctx, noteSelect+where+order+" LIMIT ? OFFSET ?", append(append([]any{}, args...), pageSize, (page-1)*pageSize)...)
	if err != nil {
		return dbFailure("查询随记", err)
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM communityInfo b"+where, args...)
	if err != nil {
		return dbFailure("统计随记", err)
	}
	rows = decodeNoteRows(stripSearchText(rows))
	// 墙上只显示正文，Markdown 原文只在编辑时用（单条随记接口会返回）
	for _, row := range rows {
		delete(row, "SOURCE")
	}
	return model.Success(map[string]any{"total": firstCount(counts), "data": rows})
}

// GetNote 返回单条随记，以及按发布时间相邻的上一条（更早）和下一条（更新）。
func (s *Service) GetNote(ctx context.Context, guid string) model.Result {
	note, found, err := s.FindNote(ctx, guid)
	if err != nil {
		return dbFailure("查询随记", err)
	}
	if !found {
		return model.Failure("这张便签不存在或已删除")
	}
	previous, next := s.NoteNeighbors(ctx, note)
	return model.Success(map[string]any{"note": note, "previous": previous, "next": next})
}

// FindNote 查询单条随记，SEO 页面和接口共用。
// 旧随记改成数字编号后，用旧编号也能查到（返回的 GUID 是新编号，调用方据此跳转到新地址）。
func (s *Service) FindNote(ctx context.Context, guid string) (map[string]any, bool, error) {
	guid = strings.TrimSpace(guid)
	if guid == "" || len(guid) > 64 {
		return nil, false, nil
	}
	query, args := noteSelect+" WHERE b.GUID = ? LIMIT 1", []any{guid}
	if !isNumericID(guid) && s.tableHasColumn(ctx, "communityInfo", "LEGACY_GUID") {
		query, args = noteSelect+" WHERE b.GUID = ? OR b.LEGACY_GUID = ? LIMIT 1", []any{guid, guid}
	}
	rows, err := s.Repo.Query(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	if len(rows) == 0 {
		return nil, false, nil
	}
	return decodeNoteRows(stripSearchText(rows))[0], true, nil
}

// NoteNeighbors 返回上一条（更早发布）和下一条（更晚发布），只带编号和标题
func (s *Service) NoteNeighbors(ctx context.Context, note map[string]any) (map[string]any, map[string]any) {
	created := model.Lookup(note, "CREATE_TIME")
	guid := model.StringValue(note, "GUID")
	pick := func(query string) map[string]any {
		rows, err := s.Repo.Query(ctx, query, created, created, guid)
		if err != nil || len(rows) == 0 {
			return nil
		}
		return map[string]any{"GUID": model.StringValue(rows[0], "GUID"), "TITLE": NoteTitle(rows[0])}
	}
	previous := pick("SELECT GUID, TITLE, TEXT FROM communityInfo WHERE (CREATE_TIME < ? OR (CREATE_TIME = ? AND GUID < ?)) ORDER BY CREATE_TIME DESC, GUID DESC LIMIT 1")
	next := pick("SELECT GUID, TITLE, TEXT FROM communityInfo WHERE (CREATE_TIME > ? OR (CREATE_TIME = ? AND GUID > ?)) ORDER BY CREATE_TIME ASC, GUID ASC LIMIT 1")
	return previous, next
}

// GetLatestNotesForSEO 供便签墙首屏和站点地图使用
func (s *Service) GetLatestNotesForSEO(ctx context.Context, limit int) ([]map[string]any, error) {
	rows, err := s.Repo.Query(ctx, noteSelect+" ORDER BY b.ISTOP DESC, b.CREATE_TIME DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	return decodeNoteRows(stripSearchText(rows)), nil
}

func isNumericID(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// RenumberLegacyNotes 把编号是一长串字符的旧随记依次改成数字编号（按发布时间，早的编号小），
// 评论和站内消息里的链接跟着改，旧编号记在 LEGACY_GUID 用于旧地址跳转。
// 需要先执行迁移 010；启动时在后台执行一次，已经改过的不会重复处理。
func (s *Service) RenumberLegacyNotes(ctx context.Context) {
	if !s.tableHasColumn(ctx, "communityInfo", "LEGACY_GUID") {
		return
	}
	rows, err := s.Repo.Query(ctx, "SELECT GUID FROM communityInfo WHERE GUID NOT REGEXP '^[0-9]+$' ORDER BY CREATE_TIME ASC, GUID ASC")
	if err != nil {
		log.Printf("查询旧随记编号失败: %v", err)
		return
	}
	renumbered := 0
	for _, row := range rows {
		oldID := model.StringValue(row, "GUID")
		newID, err := s.nextSequenceID(ctx, "communityinfo")
		if err != nil || newID == "" {
			log.Printf("旧随记改数字编号失败（编号表里需要有 communityinfo 这一行，见迁移 010）: %v", err)
			return
		}
		_, _, err = s.Repo.ExecuteBatch(ctx, []string{
			"UPDATE communityInfo SET GUID = ?, LEGACY_GUID = ? WHERE GUID = ?",
			"UPDATE communityComment SET COMMUNITYID = ? WHERE COMMUNITYID = ?",
			"UPDATE noticeInfo SET EXECUTE = ? WHERE EXECUTE = ?",
		}, [][]any{{newID, oldID, oldID}, {newID, oldID}, {NotePath(newID), NotePath(oldID)}}, false)
		if err != nil {
			log.Printf("旧随记 %s 改数字编号失败: %v", oldID, err)
			return
		}
		renumbered++
	}
	if renumbered > 0 {
		log.Printf("已把 %d 条旧随记改成数字编号", renumbered)
	}
}

// GetAllNoteIDs 返回全部随记编号和最后评论时间，用于站点地图
func (s *Service) GetAllNoteIDs(ctx context.Context) ([]map[string]any, error) {
	return s.Repo.Query(ctx, `SELECT b.GUID, COALESCE((SELECT MAX(cc.CREATE_TIME) FROM communityComment cc WHERE cc.COMMUNITYID = b.GUID), b.CREATE_TIME) AS LAST_MODIFIED
FROM communityInfo b ORDER BY b.CREATE_TIME DESC`)
}

// AddCommunityComment 与文章评论一致：允许匿名评论（昵称、邮箱、网址），
// 同样受评论频率、评论限制和违禁词约束，并给随记作者或被回复的人发通知。
func (s *Service) AddCommunityComment(c *gin.Context, comment map[string]any) model.Result {
	ctx := contextOf(c)
	clientIP := ClientIP(c)
	if failure := s.checkCommentRate(ctx, clientIP); failure != nil {
		return *failure
	}
	replyTo := strings.TrimSpace(model.StringValue(comment, "REPLY_TO"))
	comment = pickFields(comment, "GUID", "COMMUNITYID", "SUPERGUID", "TEXT", "USERNAME",
		"USEREMAIL", "USERWEBSITE", "RECEIVE_USERCODE", "RECEIVE_USERNAME")
	if strings.TrimSpace(model.StringValue(comment, "TEXT")) == "" {
		return model.Failure("评论内容不能为空")
	}
	if model.StringValue(comment, "GUID") == "" {
		return model.Failure("缺少评论编号")
	}
	noteID := model.StringValue(comment, "COMMUNITYID")
	if _, found, err := s.FindNote(ctx, noteID); err != nil {
		return dbFailure("查询随记", err)
	} else if !found {
		return model.Failure("这张便签不存在或已删除")
	}
	comment["USEREMAIL"] = normalizeMailAddress(model.StringValue(comment, "USEREMAIL"))
	comment["USERWEBSITE"] = normalizeWebsite(model.StringValue(comment, "USERWEBSITE"))
	// 登录用户以当前账号身份评论；匿名评论不能冒用任何已注册账号
	if actor, ok := s.CurrentActor(c); ok {
		comment["USERCODE"] = actor.Code()
		comment["USERNAME"] = actor.User.Name
	} else if strings.TrimSpace(model.StringValue(comment, "USERNAME")) == "" {
		comment["USERNAME"] = "访客"
	}

	if failure := s.checkCommentBan(c, commentSubject{
		UserCode: model.StringValue(comment, "USERCODE"),
		IP:       clientIP,
		Website:  model.StringValue(comment, "USERWEBSITE"),
		Email:    model.StringValue(comment, "USEREMAIL"),
		Nickname: model.StringValue(comment, "USERNAME"),
	}); failure != nil {
		return *failure
	}
	for _, field := range []string{"TEXT", "USERWEBSITE", "USERNAME"} {
		if s.containsBannedWord(model.StringValue(comment, field)) {
			return model.Failure(msgBannedWord)
		}
	}
	if s.tableHasColumn(ctx, "communityComment", "IP") {
		comment["IP"] = clientIP
	}
	// 迁移 009 之前的表没有这些字段，没有就不保存
	for _, column := range []string{"USEREMAIL", "USERWEBSITE", "RECEIVE_USERCODE", "RECEIVE_USERNAME"} {
		if !s.tableHasColumn(ctx, "communityComment", column) {
			delete(comment, column)
		}
	}

	result := s.SaveAll(ctx, "add", "communityComment", []map[string]any{comment}, "GUID")
	if result.IsError {
		return result
	}
	s.countComment(ctx, clientIP)
	s.notifyNoteComment(ctx, comment, replyTo)
	return result
}

// notifyNoteComment 给随记作者或被回复的人发站内消息和邮件，并给站长发邮件。发送失败只记日志。
func (s *Service) notifyNoteComment(ctx context.Context, comment map[string]any, replyTo string) {
	noteID := model.StringValue(comment, "COMMUNITYID")
	note, found, err := s.FindNote(ctx, noteID)
	if err != nil || !found {
		return
	}
	title := NoteTitle(note)
	sender := strings.TrimSpace(model.StringValue(comment, "USERCODE"))
	senderEmail := model.StringValue(comment, "USEREMAIL")
	senderName := strings.TrimSpace(model.StringValue(comment, "USERNAME"))
	text := strings.TrimSpace(model.StringValue(comment, "TEXT"))
	link := s.mailSiteURL(NotePath(noteID))

	receiverCode, receiverEmail, remark, originalText := "", "", "", ""
	if parentID := strings.TrimSpace(model.StringValue(comment, "SUPERGUID")); parentID == "" {
		receiverCode = model.StringValue(note, "USERCODE")
		remark = senderName + "评论了你的随记「" + title + "」"
	} else {
		columns := "USERCODE, USERNAME, TEXT"
		if s.tableHasColumn(ctx, "communityComment", "USEREMAIL") {
			columns += ", USEREMAIL"
		}
		if target := s.findReplyTargetIn(ctx, "communityComment", "COMMUNITYID", columns, noteID, parentID, replyTo, model.StringValue(comment, "RECEIVE_USERCODE")); target != nil {
			originalText = mailExcerpt(strings.TrimSpace(model.StringValue(target, "TEXT")), 300)
			receiverCode = strings.TrimSpace(model.StringValue(target, "USERCODE"))
			if receiverCode == "" {
				receiverEmail = normalizeMailAddress(model.StringValue(target, "USEREMAIL"))
			}
		}
		remark = senderName + "回复了你在随记「" + title + "」下的评论"
	}
	isSelf := (receiverCode != "" && receiverCode == sender) ||
		(receiverCode == "" && receiverEmail != "" && receiverEmail == senderEmail)
	if !isSelf && receiverCode != "" {
		noticeSender := sender
		if noticeSender == "" {
			noticeSender = "guest"
		}
		if _, err := s.Repo.Exec(ctx, "INSERT INTO noticeInfo (SENDUSERCODE, RECEIVERUSERCODE, TYPE, EXECUTE, REMARK) VALUES (?, ?, 'comment', ?, ?)",
			noticeSender, receiverCode, NotePath(noteID), remark); err != nil {
			log.Printf("随记评论已保存，但发送消息失败: %v", err)
		}
		receiverEmail = s.userMailAddress(ctx, receiverCode)
	}
	adminEmail := normalizeMailAddress(s.adminMailAddress())
	if !isSelf && receiverEmail != "" && receiverEmail != adminEmail {
		s.queueMail(ctx, receiverEmail, remark,
			renderReplyMail(remark, nil, "你的评论：", originalText, mailExcerpt(text, 300), "查看随记", link))
	}
	// 站长接收全站评论用于审核；站长自己发的评论不通知自己（notifyAdminContent 内部判断）
	s.notifyAdminContent(ctx, sender, senderName, "便签墙新评论", "", text, link)
}

func (s *Service) GetCommunityComments(ctx context.Context, communityID string) model.Result {
	// 与文章评论相同的顺序：楼层按时间倒序，楼中楼按时间正序
	query := `SELECT bc.*, ui.AVATAR
FROM communityComment bc
LEFT JOIN userInfo ui ON bc.USERCODE = ui.CODE
LEFT JOIN communityComment p ON bc.SUPERGUID = p.GUID
WHERE bc.COMMUNITYID = ?
ORDER BY COALESCE(p.CREATE_TIME, bc.CREATE_TIME) DESC,
CASE WHEN bc.SUPERGUID IS NULL OR bc.SUPERGUID = '' THEN 0 ELSE 1 END ASC,
bc.CREATE_TIME ASC`
	return stripPrivateCommentFields(s.SelectList(ctx, query, []any{communityID}))
}

// DeleteCommunityComment 权限与文章评论一致：本人可删，管理员可删普通用户的，超级管理员都能删
func (s *Service) DeleteCommunityComment(c *gin.Context, guid string) model.Result {
	row, failure := s.requireOwnerOrAdmin(c, "communityComment", guid)
	if failure != nil {
		return *failure
	}
	result := s.ExecuteSQL(contextOf(c), "DELETE FROM communityComment WHERE GUID = ? OR SUPERGUID = ?", []any{guid, guid})
	if !result.IsError {
		excerpt, _ := summarizeHTML(model.StringValue(row, "TEXT"), 80)
		s.logModeration(c, nil, row, AdminActionDeleteComment, model.PublicDisplayName(model.StringValue(row, "USERNAME")), "在便签墙删除评论："+excerpt)
	}
	return result
}
