package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

func (s *Service) AddAnnouncement(c *gin.Context, announcement map[string]any) model.Result {
	if fmt.Sprint(announcement["TEXT"]) == "" {
		return model.Failure("未传递有效内容:text")
	}
	actor, failure := s.requireSuperAdmin(c)
	if failure != nil {
		return *failure
	}
	announcement["userCode"] = actor.Code()
	announcement["userName"] = actor.User.Name
	result := s.SaveAll(contextOf(c), "add", "announcementInfo", []map[string]any{announcement}, "GUID")
	if !result.IsError {
		s.recordAdminAction(c, actor, AdminActionAnnouncement, fmt.Sprint(announcement["TEXT"]), "新增公告")
	}
	return result
}

func (s *Service) EditAnnouncement(c *gin.Context, announcement map[string]any) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	if fmt.Sprint(announcement["TEXT"]) == "" {
		return model.Failure("未传递有效内容:text")
	}
	result := s.SaveAll(ctx, "edit", "announcementInfo", []map[string]any{announcement}, "GUID")
	if !result.IsError {
		s.recordAdminAction(c, nil, AdminActionAnnouncement, fmt.Sprint(announcement["TEXT"]), "修改公告")
	}
	return result
}

func (s *Service) DeleteAnnouncement(c *gin.Context, guid string) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	var text string
	if rows, err := s.Repo.Query(contextOf(c), "SELECT TEXT FROM announcementInfo WHERE GUID = ? LIMIT 1", guid); err == nil && len(rows) > 0 {
		text = model.StringValue(rows[0], "TEXT")
	}
	result := s.ExecuteSQL(contextOf(c), "DELETE FROM announcementInfo WHERE GUID = ?", []any{guid})
	if !result.IsError {
		s.recordAdminAction(c, nil, AdminActionAnnouncement, text, "删除公告")
	}
	return result
}

func (s *Service) GetAllAnnouncements(ctx context.Context) model.Result {
	return s.SelectList(ctx, "SELECT * FROM announcementInfo ORDER BY TYPE, CREATE_TIME DESC", nil)
}

func (s *Service) GetAnnouncementsByType(ctx context.Context, announcementType string) model.Result {
	return s.SelectList(ctx, "SELECT * FROM announcementInfo WHERE ISENABLE = '1' AND TYPE = ? ORDER BY TYPE, CREATE_TIME DESC", []any{announcementType})
}

// friendLinkFields 是友链允许由客户端填写的字段。
var friendLinkFields = []string{"NAME", "LINK", "AVATAR", "REMARK", "LINK_TYPE"}

func (s *Service) AddFriendLink(c *gin.Context, value map[string]any) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	data := pickFields(value, append([]string{"GUID"}, friendLinkFields...)...)
	if model.Lookup(data, "LINK") == nil {
		return model.Failure(msgFriendLinkURL)
	}
	if failure := s.normalizeFriendLinkURLs(data); failure != nil {
		return *failure
	}
	if failure := s.checkFriendLinkText(data); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	// “推荐好站”（LINK_TYPE = 2）只有管理员能添加，普通用户只能申请友链，每人最多 3 个
	if !actor.IsAdmin {
		data["LINK_TYPE"] = 1
		counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM friendLinkInfo WHERE USERCODE = ? AND LINK_TYPE = 1", actor.Code())
		if err != nil {
			return dbFailure("统计友链", err)
		}
		if firstCount(counts) >= maxFriendLinksPerUser {
			return model.Failure(fmt.Sprintf("为保证友链质量，每人最多申请 %d 个友链", maxFriendLinksPerUser))
		}
	}
	if model.StringValue(data, "GUID") == "" {
		guid, err := randomHex(16)
		if err != nil {
			return model.Failure("生成编号失败")
		}
		data["GUID"] = strings.ToUpper(guid)
	}
	// 普通用户提交的友链先待审核，管理员添加的直接通过
	pending := s.friendLinkReviewReady(ctx) && !actor.IsAdmin
	if s.friendLinkReviewReady(ctx) {
		data["STATUS"] = friendLinkApproved
		if pending {
			data["STATUS"] = friendLinkPending
		}
	}
	data["USERCODE"] = actor.Code()
	data["USERNAME"] = actor.User.Name
	result := s.SaveAll(ctx, "add", "friendLinkInfo", []map[string]any{data}, "GUID")
	if !result.IsError && pending {
		s.notifyFriendLinkApplied(ctx, actor, data)
	}
	return result
}

const (
	friendLinkPending     = "pending"
	friendLinkApproved    = "approved"
	maxFriendLinksPerUser = 3
	maxFriendLinkName     = 40
	maxFriendLinkRemark   = 50
	friendLinkPagePath    = "/ynsStudy/FriendLink"
)

// friendLinkReviewReady 执行迁移 016 后友链才有审核状态；之前提交的友链照旧直接显示
func (s *Service) friendLinkReviewReady(ctx context.Context) bool {
	return s.Repo != nil && s.tableHasColumn(ctx, "friendLinkInfo", "STATUS")
}

// checkFriendLinkText 校验请求里出现的网站名称和介绍（没出现的不处理）：长度和违禁词
func (s *Service) checkFriendLinkText(data map[string]any) *model.Result {
	fail := func(message string) *model.Result {
		result := model.Failure(message)
		return &result
	}
	if value := model.Lookup(data, "NAME"); value != nil {
		name := strings.TrimSpace(fmt.Sprint(value))
		if name == "" {
			return fail("网站名称不能为空")
		}
		if len([]rune(name)) > maxFriendLinkName {
			return fail(fmt.Sprintf("网站名称最多 %d 个字", maxFriendLinkName))
		}
		data["NAME"] = name
	}
	if value := model.Lookup(data, "REMARK"); value != nil {
		remark := strings.TrimSpace(fmt.Sprint(value))
		if len([]rune(remark)) > maxFriendLinkRemark {
			return fail(fmt.Sprintf("一句话描述最多 %d 个字", maxFriendLinkRemark))
		}
		data["REMARK"] = remark
	}
	if s.containsBannedWord(model.StringValue(data, "NAME")) || s.containsBannedWord(model.StringValue(data, "REMARK")) {
		return fail(msgBannedWord)
	}
	return nil
}

// notifyFriendLinkApplied 有人申请或修改了友链，等待审核：给站长发站内消息和邮件。失败只记日志。
func (s *Service) notifyFriendLinkApplied(ctx context.Context, actor *Actor, data map[string]any) {
	name := model.StringValue(data, "NAME")
	remark := actor.User.Name + "申请了友链「" + name + "」，等待审核"
	if _, err := s.Repo.Exec(ctx, "INSERT INTO noticeInfo (SENDUSERCODE, RECEIVERUSERCODE, TYPE, EXECUTE, REMARK) VALUES (?, ?, 'friendLink', ?, ?)",
		actor.Code(), s.superAdminCode(), friendLinkPagePath, remark); err != nil {
		log.Printf("友链待审核消息发送失败: %v", err)
	}
	s.notifyAdminContent(ctx, actor.Code(), actor.User.Name, "友链待审核", name,
		model.StringValue(data, "LINK")+"\n"+model.StringValue(data, "REMARK"), s.mailSiteURL(friendLinkPagePath))
}

// ReviewFriendLink 审核友链（管理员）：approve 通过并公开显示，reject 拒绝并删除。都会通知申请人。
func (s *Service) ReviewFriendLink(c *gin.Context, guid, action string) model.Result {
	actor, failure := s.requireAdmin(c)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	if !s.friendLinkReviewReady(ctx) {
		return model.Failure("友链审核还没有启用（需要先执行数据库迁移 016）")
	}
	row, failure := s.loadRow(c, "friendLinkInfo", guid)
	if failure != nil {
		return *failure
	}
	name := model.StringValue(row, "NAME")
	var remark string
	switch action {
	case "approve":
		if _, err := s.Repo.Exec(ctx, "UPDATE friendLinkInfo SET STATUS = ? WHERE GUID = ?", friendLinkApproved, guid); err != nil {
			return dbFailure("审核友链", err)
		}
		remark = "你申请的友链「" + name + "」已通过审核，现在会显示在友链页"
	case "reject":
		if _, err := s.Repo.Exec(ctx, "DELETE FROM friendLinkInfo WHERE GUID = ?", guid); err != nil {
			return dbFailure("审核友链", err)
		}
		remark = "你申请的友链「" + name + "」没有通过审核，可以修改后重新申请"
	default:
		return model.Failure("不支持的审核操作")
	}
	if applicant := model.StringValue(row, "USERCODE"); applicant != "" && applicant != actor.Code() {
		if _, err := s.Repo.Exec(ctx, "INSERT INTO noticeInfo (SENDUSERCODE, RECEIVERUSERCODE, TYPE, EXECUTE, REMARK) VALUES (?, ?, 'friendLink', ?, ?)",
			actor.Code(), applicant, friendLinkPagePath, remark); err != nil {
			log.Printf("友链审核结果通知失败: %v", err)
		}
	}
	label := map[string]string{"approve": "通过友链", "reject": "拒绝友链"}[action]
	s.recordAdminAction(c, actor, AdminActionFriendLink, name, label+" "+model.StringValue(row, "LINK"))
	if action == "approve" {
		return model.Success("已通过")
	}
	return model.Success("已拒绝")
}

const (
	msgFriendLinkURL    = "网站地址必须是 http:// 或 https:// 开头的网址"
	msgFriendLinkAvatar = "网站图标必须是 http:// 或 https:// 开头的图片地址，或本站上传的图片"
)

// normalizeFriendLinkURLs 校验并规范化请求里出现的 LINK、AVATAR（没出现的不处理，编辑时可以只改部分字段）。
// 友链地址会直接放进页面的 <a href>，必须拦住 javascript: 之类的地址，否则访客一点击就会执行脚本。
func (s *Service) normalizeFriendLinkURLs(data map[string]any) *model.Result {
	fail := func(message string) *model.Result {
		result := model.Failure(message)
		return &result
	}
	for key := range data {
		switch strings.ToUpper(key) {
		case "LINK":
			link := normalizeWebsite(model.StringValue(data, key))
			if link == "" {
				return fail(msgFriendLinkURL)
			}
			data[key] = link
		case "AVATAR":
			avatar := strings.TrimSpace(model.StringValue(data, key))
			switch {
			case avatar == "":
			case strings.HasPrefix(avatar, "/") && !strings.HasPrefix(avatar, "//"):
				if avatar = s.uploadKey(avatar); avatar == "" {
					return fail(msgFriendLinkAvatar)
				}
			default:
				if avatar = normalizeWebsite(avatar); avatar == "" {
					return fail(msgFriendLinkAvatar)
				}
			}
			data[key] = avatar
		}
	}
	return nil
}

// requireFriendLinkManager：友链发布者本人或任意管理员可以修改、删除（与前端 isAdmin || 本人 的显示规则一致）。
func (s *Service) requireFriendLinkManager(c *gin.Context, guid string) (map[string]any, *model.Result) {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return nil, failure
	}
	row, failure := s.loadRow(c, "friendLinkInfo", guid)
	if failure != nil {
		return nil, failure
	}
	if !actor.IsAdmin && model.StringValue(row, "USERCODE") != actor.Code() {
		result := model.Failure(msgForbidden)
		return nil, &result
	}
	return row, nil
}

func (s *Service) UpdateFriendLink(c *gin.Context, value map[string]any) model.Result {
	guid := model.StringValue(value, "GUID")
	row, failure := s.requireFriendLinkManager(c, guid)
	if failure != nil {
		return *failure
	}
	data := pickFields(value, friendLinkFields...)
	if failure := s.normalizeFriendLinkURLs(data); failure != nil {
		return *failure
	}
	if failure := s.checkFriendLinkText(data); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	actor, _ := s.CurrentActor(c)
	isAdmin := actor != nil && actor.IsAdmin
	if !isAdmin {
		delete(data, "LINK_TYPE")
	}
	// 普通用户修改后要重新审核，防止通过审核后再把网址换成别的网站
	pending := !isAdmin && s.friendLinkReviewReady(ctx)
	if pending {
		data["STATUS"] = friendLinkPending
	}
	data["GUID"] = guid
	result := s.SaveAll(ctx, "edit", "friendLinkInfo", []map[string]any{data}, "GUID")
	if !result.IsError {
		s.logModeration(c, nil, row, AdminActionFriendLink, model.StringValue(row, "NAME"), "修改了他人的友链")
		if pending && actor != nil {
			s.notifyFriendLinkApplied(ctx, actor, data)
		}
	}
	return result
}

func (s *Service) DeleteFriendLink(c *gin.Context, guid string) model.Result {
	row, failure := s.requireFriendLinkManager(c, guid)
	if failure != nil {
		return *failure
	}
	result := s.ExecuteSQL(contextOf(c), "DELETE FROM friendLinkInfo WHERE GUID = ?", []any{guid})
	if !result.IsError {
		s.logModeration(c, nil, row, AdminActionFriendLink, model.StringValue(row, "NAME"), "删除了他人的友链 "+model.StringValue(row, "LINK"))
	}
	return result
}

func (s *Service) GetFriendLink(ctx context.Context, guid string) model.Result {
	return s.SelectList(ctx, "SELECT * FROM friendLinkInfo WHERE GUID = ?", []any{guid})
}

// GetFriendLinks 返回友链列表：访客只看到已通过的；登录用户还能看到自己待审核的；管理员能看到全部待审核的
func (s *Service) GetFriendLinks(c *gin.Context) model.Result {
	ctx := contextOf(c)
	if !s.friendLinkReviewReady(ctx) {
		return s.SelectList(ctx, "SELECT * FROM friendLinkInfo ORDER BY CREATE_TIME", nil)
	}
	if actor, ok := s.CurrentActor(c); ok {
		if actor.IsAdmin {
			return s.SelectList(ctx, "SELECT * FROM friendLinkInfo ORDER BY CREATE_TIME", nil)
		}
		return s.SelectList(ctx, "SELECT * FROM friendLinkInfo WHERE STATUS = ? OR USERCODE = ? ORDER BY CREATE_TIME", []any{friendLinkApproved, actor.Code()})
	}
	return s.SelectList(ctx, "SELECT * FROM friendLinkInfo WHERE STATUS = ? ORDER BY CREATE_TIME", []any{friendLinkApproved})
}
