package service

import (
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

// 前端直接发送的通知只有关注、点赞、收藏三种；评论、留言的通知由服务端在保存成功后自己发送。
var clientNoticeTypes = map[string]bool{"followUser": true, "giveLike": true, "collect": true}

const (
	maxNoticeRemark     = 100 // 通知文字最多字数
	noticeHourlyLimit   = 60  // 每个账号每小时最多发送的通知数
	msgNoticeTooFrequent = "操作太频繁了，请稍后再试"
)

// AddNotice 以当前登录用户的身份发送通知，不信任请求中的发送者，避免伪造他人发消息。
// 通知会出现在对方的消息中心里，为防止被用来骚扰或钓鱼：只能链接到本站页面、限制类型、字数和频率，
// 对方还没看的相同通知不重复发送（反复点赞、取消点赞不会刷屏）。
func (s *Service) AddNotice(c *gin.Context, receiver, noticeType, execute, remark string) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	receiver = strings.TrimSpace(receiver)
	if receiver == "" || noticeType == "" || execute == "" {
		return model.Failure("传递参数不全,请传递完整参数结构!")
	}
	if !clientNoticeTypes[noticeType] {
		return model.Failure("不支持的消息类型!")
	}
	execute = s.siteNoticePath(c, execute)
	if execute == "" {
		return model.Failure("消息链接不合法!")
	}
	if receiver == actor.Code() {
		return model.Success("自己的操作不发送消息")
	}
	remark = strings.TrimSpace(remark)
	if runes := []rune(remark); len(runes) > maxNoticeRemark {
		remark = string(runes[:maxNoticeRemark]) + "…"
	}
	ctx := contextOf(c)
	if receivers, err := s.Repo.Query(ctx, "SELECT GUID FROM userInfo WHERE CODE = ? LIMIT 1", receiver); err != nil {
		return dbFailure("查询接收人", err)
	} else if len(receivers) == 0 {
		return model.Failure("接收消息的用户不存在!")
	}
	if existing, err := s.Repo.Query(ctx, "SELECT GUID FROM noticeInfo WHERE SENDUSERCODE = ? AND RECEIVERUSERCODE = ? AND TYPE = ? AND EXECUTE = ? LIMIT 1",
		actor.Code(), receiver, noticeType, execute); err == nil && len(existing) > 0 {
		return model.Success("新增消息成功!")
	}
	if count, err := s.Cache.Increment(ctx, "notice_limit:"+actor.Code(), time.Hour); err == nil && count > noticeHourlyLimit {
		return model.Failure(msgNoticeTooFrequent)
	}
	result := s.SaveAll(ctx, "add", "noticeInfo", []map[string]any{{
		"SENDUSERCODE": actor.Code(), "RECEIVERUSERCODE": receiver, "TYPE": noticeType, "EXECUTE": execute, "REMARK": remark,
	}}, "GUID")
	if result.IsError {
		return model.Failure("新增消息失败,请检查!" + result.ErrMsg)
	}
	return model.Success("新增消息成功!")
}

// 通知只能由接收者本人查看和清除，忽略请求中传入的账号。
func (s *Service) GetNotices(c *gin.Context) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	return s.SelectList(contextOf(c), "SELECT * FROM noticeInfo WHERE RECEIVERUSERCODE = ? ORDER BY CREATE_TIME DESC", []any{actor.Code()})
}

func (s *Service) ReadNotice(c *gin.Context, guid string) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	return s.ExecuteSQL(contextOf(c), "DELETE FROM noticeInfo WHERE GUID = ? AND RECEIVERUSERCODE = ?", []any{guid, actor.Code()})
}

func (s *Service) ReadAllNotices(c *gin.Context) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	return s.ExecuteSQL(contextOf(c), "DELETE FROM noticeInfo WHERE RECEIVERUSERCODE = ?", []any{actor.Code()})
}

// isSafeNoticeLink 只允许站内路径或 http/https 链接。
// 前端点击消息会直接打开这个地址，javascript: 之类的链接会在接收者的浏览器里执行脚本。
func isSafeNoticeLink(link string) bool {
	link = strings.TrimSpace(link)
	if strings.HasPrefix(link, "/") {
		return !strings.HasPrefix(link, "//") && !strings.HasPrefix(link, "/\\")
	}
	parsed, err := url.Parse(link)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(parsed.Scheme)
	return (scheme == "http" || scheme == "https") && parsed.Host != ""
}

// siteNoticePath 把消息链接统一成站内路径：站内路径原样保留；完整网址只接受本站域名（或当前访问的域名，
// 便于本地开发），并去掉域名只留路径；其他网站的链接一律拒绝，返回空字符串。
func (s *Service) siteNoticePath(c *gin.Context, link string) string {
	link = strings.TrimSpace(link)
	if !isSafeNoticeLink(link) || len(link) > 500 {
		return ""
	}
	if strings.HasPrefix(link, "/") {
		return link
	}
	parsed, err := url.Parse(link)
	if err != nil {
		return ""
	}
	allowed := map[string]bool{strings.ToLower(c.Request.Host): true}
	if site, err := url.Parse(s.Config.External.DomainName); err == nil && site.Host != "" {
		allowed[strings.ToLower(site.Host)] = true
	}
	if !allowed[strings.ToLower(parsed.Host)] {
		return ""
	}
	path := parsed.EscapedPath()
	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		path = "/"
	}
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}
	if parsed.Fragment != "" {
		path += "#" + parsed.EscapedFragment()
	}
	return path
}
