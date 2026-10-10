package service

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

// AddNotice 以当前登录用户的身份发送通知，不信任请求中的发送者，避免伪造他人发消息。
func (s *Service) AddNotice(c *gin.Context, receiver, noticeType, execute, remark string) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	if receiver == "" || noticeType == "" || execute == "" {
		return model.Failure("传递参数不全,请传递完整参数结构!")
	}
	if !isSafeNoticeLink(execute) {
		return model.Failure("消息链接不合法!")
	}
	result := s.SaveAll(contextOf(c), "add", "noticeInfo", []map[string]any{{
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
