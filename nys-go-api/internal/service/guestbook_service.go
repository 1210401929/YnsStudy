package service

// 便签墙的留言板：访客不登录也能留言，可以回复（一层，和文章评论一样）。
// 和文章评论共用评论频率限制、评论限制规则和违禁词检查。需要先执行迁移 011。

import (
	"context"
	"log"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

const (
	guestbookPath       = "/ynsStudy/Notes?tab=guestbook"
	msgGuestbookMissing = "留言板还没有启用（需要先执行数据库迁移 011）"
	maxGuestbookText    = 1000
)

func (s *Service) guestbookReady(ctx context.Context) bool {
	return s.tableHasColumn(ctx, "guestbook", "GUID")
}

// GetGuestbook 按页返回留言（置顶在前，新的在前），每条留言带上它下面的全部回复（按时间正序）。
// USERROLE 是留言人的角色，页面据此给管理员的留言盖“管理员”章。
func (s *Service) GetGuestbook(ctx context.Context, page, pageSize int) model.Result {
	if !s.guestbookReady(ctx) {
		return model.Success(map[string]any{"total": 0, "data": []map[string]any{}})
	}
	page, pageSize = normalizePage(page, pageSize)
	rows, err := s.Repo.Query(ctx, `SELECT g.*, u.AVATAR, u.ROLE AS USERROLE FROM guestbook g LEFT JOIN userInfo u ON g.USERCODE = u.CODE
WHERE g.SUPERGUID IS NULL OR g.SUPERGUID = ''
ORDER BY g.ISTOP DESC, g.CREATE_TIME DESC LIMIT ? OFFSET ?`, pageSize, (page-1)*pageSize)
	if err != nil {
		return dbFailure("查询留言", err)
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM guestbook WHERE SUPERGUID IS NULL OR SUPERGUID = ''")
	if err != nil {
		return dbFailure("统计留言", err)
	}
	stripPrivateCommentFields(model.Success(rows))

	if len(rows) > 0 {
		placeholders := make([]string, len(rows))
		args := make([]any, len(rows))
		byID := make(map[string]map[string]any, len(rows))
		for index, row := range rows {
			placeholders[index] = "?"
			args[index] = model.StringValue(row, "GUID")
			row["children"] = []map[string]any{}
			byID[model.StringValue(row, "GUID")] = row
		}
		replies, err := s.Repo.Query(ctx, `SELECT g.*, u.AVATAR, u.ROLE AS USERROLE FROM guestbook g LEFT JOIN userInfo u ON g.USERCODE = u.CODE
WHERE g.SUPERGUID IN (`+strings.Join(placeholders, ",")+`) ORDER BY g.CREATE_TIME ASC`, args...)
		if err != nil {
			return dbFailure("查询留言回复", err)
		}
		stripPrivateCommentFields(model.Success(replies))
		for _, reply := range replies {
			if parent := byID[model.StringValue(reply, "SUPERGUID")]; parent != nil {
				parent["children"] = append(parent["children"].([]map[string]any), reply)
			}
		}
	}
	return model.Success(map[string]any{"total": firstCount(counts), "data": rows})
}

// AddGuestbookMessage 发表留言或回复留言。
func (s *Service) AddGuestbookMessage(c *gin.Context, message map[string]any) model.Result {
	ctx := contextOf(c)
	if !s.guestbookReady(ctx) {
		return model.Failure(msgGuestbookMissing)
	}
	clientIP := ClientIP(c)
	if failure := s.checkCommentRate(ctx, clientIP); failure != nil {
		return *failure
	}
	replyTo := strings.TrimSpace(model.StringValue(message, "REPLY_TO"))
	message = pickFields(message, "GUID", "SUPERGUID", "TEXT", "USERNAME",
		"USEREMAIL", "USERWEBSITE", "RECEIVE_USERCODE", "RECEIVE_USERNAME")
	text := strings.TrimSpace(model.StringValue(message, "TEXT"))
	if text == "" {
		return model.Failure("留言内容不能为空")
	}
	if len([]rune(text)) > maxGuestbookText {
		return model.Failure("留言最多 1000 个字")
	}
	message["TEXT"] = text
	if model.StringValue(message, "GUID") == "" {
		return model.Failure("缺少留言编号")
	}
	// 回复只能挂在一条留言下面（不能回复到回复下面形成多层）
	if parentID := strings.TrimSpace(model.StringValue(message, "SUPERGUID")); parentID != "" {
		parents, err := s.Repo.Query(ctx, "SELECT GUID FROM guestbook WHERE GUID = ? AND (SUPERGUID IS NULL OR SUPERGUID = '') LIMIT 1", parentID)
		if err != nil {
			return dbFailure("查询留言", err)
		}
		if len(parents) == 0 {
			return model.Failure("要回复的留言不存在或已删除")
		}
	} else {
		delete(message, "SUPERGUID")
	}
	message["USEREMAIL"] = normalizeMailAddress(model.StringValue(message, "USEREMAIL"))
	message["USERWEBSITE"] = normalizeWebsite(model.StringValue(message, "USERWEBSITE"))
	// 登录用户以当前账号身份留言；匿名访客不能冒用任何已注册账号
	if actor, ok := s.CurrentActor(c); ok {
		message["USERCODE"] = actor.Code()
		message["USERNAME"] = actor.User.Name
	} else if strings.TrimSpace(model.StringValue(message, "USERNAME")) == "" {
		message["USERNAME"] = "访客"
	}
	if failure := s.checkCommentBan(c, commentSubject{
		UserCode: model.StringValue(message, "USERCODE"),
		IP:       clientIP,
		Website:  model.StringValue(message, "USERWEBSITE"),
		Email:    model.StringValue(message, "USEREMAIL"),
		Nickname: model.StringValue(message, "USERNAME"),
	}); failure != nil {
		return *failure
	}
	for _, field := range []string{"TEXT", "USERWEBSITE", "USERNAME"} {
		if s.containsBannedWord(model.StringValue(message, field)) {
			return model.Failure(msgBannedWord)
		}
	}
	message["IP"] = clientIP

	result := s.SaveAll(ctx, "add", "guestbook", []map[string]any{message}, "GUID")
	if result.IsError {
		return result
	}
	s.countComment(ctx, clientIP)
	s.notifyGuestbook(ctx, message, replyTo)
	return result
}

// notifyGuestbook 新留言通知超级管理员（站内消息 + 邮件）；回复通知被回复的人。发送失败只记日志。
func (s *Service) notifyGuestbook(ctx context.Context, message map[string]any, replyTo string) {
	sender := strings.TrimSpace(model.StringValue(message, "USERCODE"))
	senderEmail := model.StringValue(message, "USEREMAIL")
	senderName := strings.TrimSpace(model.StringValue(message, "USERNAME"))
	text := model.StringValue(message, "TEXT")
	link := s.mailSiteURL(guestbookPath)
	noticeSender := sender
	if noticeSender == "" {
		noticeSender = "guest"
	}
	notice := func(receiver, remark string) {
		if _, err := s.Repo.Exec(ctx, "INSERT INTO noticeInfo (SENDUSERCODE, RECEIVERUSERCODE, TYPE, EXECUTE, REMARK) VALUES (?, ?, 'comment', ?, ?)",
			noticeSender, receiver, guestbookPath, remark); err != nil {
			log.Printf("留言已保存，但发送消息失败: %v", err)
		}
	}

	parentID := strings.TrimSpace(model.StringValue(message, "SUPERGUID"))
	if parentID == "" {
		// 新留言：超级管理员收站内消息；邮件由 notifyAdminContent 发送（超级管理员自己留言不通知）
		if superCode := s.superAdminCode(); sender != superCode {
			notice(superCode, senderName+"在便签墙给你留言了")
		}
		s.notifyAdminContent(ctx, sender, senderName, "便签墙新留言", "", text, link)
		return
	}

	target := s.guestbookReplyTarget(ctx, parentID, replyTo, model.StringValue(message, "RECEIVE_USERCODE"))
	if target == nil {
		s.notifyAdminContent(ctx, sender, senderName, "便签墙留言回复", "", text, link)
		return
	}
	remark := senderName + "回复了你在便签墙的留言"
	receiverCode := strings.TrimSpace(model.StringValue(target, "USERCODE"))
	receiverEmail := ""
	if receiverCode == "" {
		receiverEmail = normalizeMailAddress(model.StringValue(target, "USEREMAIL"))
	}
	isSelf := (receiverCode != "" && receiverCode == sender) ||
		(receiverCode == "" && receiverEmail != "" && receiverEmail == senderEmail)
	if !isSelf && receiverCode != "" {
		notice(receiverCode, remark)
		receiverEmail = s.userMailAddress(ctx, receiverCode)
	}
	adminEmail := normalizeMailAddress(s.adminMailAddress())
	if !isSelf && receiverEmail != "" && receiverEmail != adminEmail {
		s.queueMail(ctx, receiverEmail, remark,
			renderReplyMail(remark, nil, "你的留言：", mailExcerpt(model.StringValue(target, "TEXT"), 300), mailExcerpt(text, 300), "查看留言板", link))
	}
	s.notifyAdminContent(ctx, sender, senderName, "便签墙留言回复", "", text, link)
}

// guestbookReplyTarget 找到被回复的那条留言或回复，必须在同一条留言下面。
// 优先用前端传来的编号；没有时按被回复人账号查找。
func (s *Service) guestbookReplyTarget(ctx context.Context, parentID, replyTo, receiverCode string) map[string]any {
	const columns = "SELECT USERCODE, USERNAME, USEREMAIL, TEXT FROM guestbook WHERE (GUID = ? OR SUPERGUID = ?)"
	if replyTo != "" {
		if rows, err := s.Repo.Query(ctx, columns+" AND GUID = ? LIMIT 1", parentID, parentID, replyTo); err == nil && len(rows) > 0 {
			return rows[0]
		}
	}
	if receiverCode = strings.TrimSpace(receiverCode); receiverCode != "" {
		if rows, err := s.Repo.Query(ctx, columns+" AND USERCODE = ? LIMIT 1", parentID, parentID, receiverCode); err == nil && len(rows) > 0 {
			return rows[0]
		}
	}
	return nil
}

// DeleteGuestbookMessage 删除留言（连同下面的回复）或一条回复。
// 权限与评论一致：本人可删，管理员可删普通用户的，超级管理员都能删；匿名留言只有管理员能删。
func (s *Service) DeleteGuestbookMessage(c *gin.Context, guid string) model.Result {
	if !s.guestbookReady(contextOf(c)) {
		return model.Failure(msgGuestbookMissing)
	}
	row, failure := s.requireOwnerOrAdmin(c, "guestbook", guid)
	if failure != nil {
		return *failure
	}
	result := s.ExecuteSQL(contextOf(c), "DELETE FROM guestbook WHERE GUID = ? OR SUPERGUID = ?", []any{guid, guid})
	if !result.IsError {
		excerpt, _ := summarizeHTML(model.StringValue(row, "TEXT"), 80)
		s.logModeration(c, nil, row, AdminActionDeleteComment, model.PublicDisplayName(model.StringValue(row, "USERNAME")), "在便签墙删除留言："+excerpt)
	}
	return result
}

// SetTopGuestbook 置顶或取消置顶一条留言，只有超级管理员可以操作
func (s *Service) SetTopGuestbook(c *gin.Context, guid, isTop string) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	var value any
	switch isTop {
	case "1":
		value = "1"
	case "0":
		value = nil
	default:
		return model.Failure("isTop 只允许为 0 或 1")
	}
	return s.ExecuteSQL(contextOf(c), "UPDATE guestbook SET ISTOP = ? WHERE GUID = ? AND (SUPERGUID IS NULL OR SUPERGUID = '')", []any{value, guid})
}
