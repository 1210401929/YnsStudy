package service

import (
	"context"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

// 评论限制：按用户、IP、网址、邮箱、昵称禁止某些人评论（文章评论和社区评论）。
// 规则在内存中缓存 60 秒，后台修改后立即刷新；表不存在（未执行迁移 008）时不做限制。

const (
	CommentBanUser     = "user"
	CommentBanIP       = "ip"
	CommentBanWebsite  = "website"
	CommentBanEmail    = "email"
	CommentBanNickname = "nickname"
)

var commentBanTypes = map[string]string{
	CommentBanUser:     "用户",
	CommentBanIP:       "IP",
	CommentBanWebsite:  "网址",
	CommentBanEmail:    "邮箱",
	CommentBanNickname: "昵称",
}

const (
	msgCommentBanned      = "你已被禁止在本站评论"
	commentBanCacheTTL    = 60 * time.Second
	commentBanValueMaxLen = 255
)

// commentBanRule 是一条限制规则；ipNet 仅 IP 网段规则有值。
type commentBanRule struct {
	Type      string
	Value     string
	ExpiresAt time.Time // 零值表示永久
	ipNet     *net.IPNet
}

var commentBanCache struct {
	sync.Mutex
	rules    []commentBanRule
	loadedAt time.Time
}

func invalidateCommentBanCache() {
	commentBanCache.Lock()
	commentBanCache.loadedAt = time.Time{}
	commentBanCache.Unlock()
}

func (s *Service) commentBanRules(ctx context.Context) []commentBanRule {
	commentBanCache.Lock()
	defer commentBanCache.Unlock()
	if time.Since(commentBanCache.loadedAt) < commentBanCacheTTL {
		return commentBanCache.rules
	}
	if s.Repo == nil {
		return nil
	}
	queryContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	rows, err := s.Repo.Query(queryContext, "SELECT TYPE, VALUE, EXPIRE_TIME FROM commentBan WHERE EXPIRE_TIME IS NULL OR EXPIRE_TIME > NOW()")
	commentBanCache.loadedAt = time.Now()
	if err != nil {
		commentBanCache.rules = nil
		return nil
	}
	rules := make([]commentBanRule, 0, len(rows))
	for _, row := range rows {
		rule := commentBanRule{Type: model.StringValue(row, "TYPE"), Value: strings.TrimSpace(model.StringValue(row, "VALUE"))}
		if expires := parseControllerLikeTime(model.Lookup(row, "EXPIRE_TIME")); !expires.IsZero() {
			rule.ExpiresAt = expires
		}
		if rule.Type == CommentBanIP && strings.Contains(rule.Value, "/") {
			if _, network, err := net.ParseCIDR(rule.Value); err == nil {
				rule.ipNet = network
			}
		}
		rules = append(rules, rule)
	}
	commentBanCache.rules = rules
	return rules
}

// commentSubject 是一条评论中可以用来匹配限制规则的信息。
type commentSubject struct {
	UserCode string
	IP       string
	Website  string
	Email    string
	Nickname string
}

// matchCommentBan 返回命中的第一条规则；没有命中时返回 nil。
func matchCommentBan(rules []commentBanRule, subject commentSubject, now time.Time) *commentBanRule {
	website := strings.ToLower(subject.Website)
	email := strings.ToLower(subject.Email)
	nickname := strings.ToLower(subject.Nickname)
	ip := net.ParseIP(strings.TrimSpace(subject.IP))
	for i := range rules {
		rule := &rules[i]
		if rule.Value == "" || (!rule.ExpiresAt.IsZero() && !rule.ExpiresAt.After(now)) {
			continue
		}
		value := strings.ToLower(rule.Value)
		switch rule.Type {
		case CommentBanUser:
			if subject.UserCode != "" && subject.UserCode == rule.Value {
				return rule
			}
		case CommentBanIP:
			if ip == nil {
				continue
			}
			if rule.ipNet != nil {
				if rule.ipNet.Contains(ip) {
					return rule
				}
			} else if ruleIP := net.ParseIP(rule.Value); ruleIP != nil && ruleIP.Equal(ip) {
				return rule
			}
		case CommentBanWebsite:
			if website != "" && strings.Contains(website, value) {
				return rule
			}
		case CommentBanEmail:
			if email != "" && strings.Contains(email, value) {
				return rule
			}
		case CommentBanNickname:
			if nickname != "" && strings.Contains(nickname, value) {
				return rule
			}
		}
	}
	return nil
}

// checkCommentBan 判断当前评论是否被限制。超级管理员不受限制。
func (s *Service) checkCommentBan(c *gin.Context, subject commentSubject) *model.Result {
	if subject.UserCode != "" && subject.UserCode == s.superAdminCode() {
		return nil
	}
	if rule := matchCommentBan(s.commentBanRules(contextOf(c)), subject, time.Now()); rule != nil {
		log.Printf("评论被限制规则拦截: %s=%s ip=%s", rule.Type, rule.Value, subject.IP)
		result := model.Failure(msgCommentBanned)
		return &result
	}
	return nil
}

// ---------------- 后台管理 ----------------

// normalizeCommentBanValue 校验并整理规则值。
func (s *Service) normalizeCommentBanValue(ctx context.Context, banType, value string) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "限制内容不能为空"
	}
	if len([]rune(value)) > commentBanValueMaxLen {
		return "", "限制内容过长"
	}
	switch banType {
	case CommentBanUser:
		if value == s.superAdminCode() {
			return "", "不能限制超级管理员"
		}
		rows, err := s.Repo.Query(ctx, "SELECT CODE FROM userInfo WHERE CODE = ? LIMIT 1", value)
		if err != nil || len(rows) == 0 {
			return "", "用户账号不存在: " + value
		}
	case CommentBanIP:
		if strings.Contains(value, "/") {
			if _, network, err := net.ParseCIDR(value); err == nil {
				return network.String(), ""
			}
			return "", "IP 网段格式不正确，例如 1.2.3.0/24"
		}
		ip := net.ParseIP(value)
		if ip == nil {
			return "", "IP 地址格式不正确"
		}
		return ip.String(), ""
	case CommentBanWebsite:
		value = strings.ToLower(value)
		for _, prefix := range []string{"https://", "http://"} {
			value = strings.TrimPrefix(value, prefix)
		}
		value = strings.TrimRight(value, "/")
		if value == "" {
			return "", "网址不能为空"
		}
	case CommentBanEmail, CommentBanNickname:
		value = strings.ToLower(value)
	default:
		return "", "不支持的限制类型"
	}
	return value, ""
}

// GetCommentBans 返回全部限制规则（含已过期的，前端标记出来）。
func (s *Service) GetCommentBans(c *gin.Context) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	rows, err := s.Repo.Query(contextOf(c), `SELECT ID, TYPE, VALUE, REMARK, CREATE_USER, CREATE_TIME, EXPIRE_TIME,
CASE WHEN EXPIRE_TIME IS NOT NULL AND EXPIRE_TIME <= NOW() THEN 1 ELSE 0 END AS EXPIRED
FROM commentBan ORDER BY ID DESC`)
	if err != nil {
		return model.Failure("读取评论限制失败，请确认已执行迁移 008_add_comment_ban.sql")
	}
	return model.Success(rows)
}

// AddCommentBan 新增或更新一条限制规则。days 为 0 表示永久。
func (s *Service) AddCommentBan(c *gin.Context, banType, value, remark string, days int) model.Result {
	actor, failure := s.requireSuperAdmin(c)
	if failure != nil {
		return *failure
	}
	if _, ok := commentBanTypes[banType]; !ok {
		return model.Failure("不支持的限制类型")
	}
	if days < 0 || days > 3650 {
		return model.Failure("限制天数不正确")
	}
	ctx := contextOf(c)
	value, problem := s.normalizeCommentBanValue(ctx, banType, value)
	if problem != "" {
		return model.Failure(problem)
	}
	var expire any
	if days > 0 {
		expire = time.Now().Add(time.Duration(days) * 24 * time.Hour).Format("2006-01-02 15:04:05")
	}
	remark = truncateText(remark, 255)
	// 同一类型同一值已存在时更新期限和备注
	_, err := s.Repo.Exec(ctx, `INSERT INTO commentBan (TYPE, VALUE, REMARK, CREATE_USER, EXPIRE_TIME) VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE REMARK = VALUES(REMARK), CREATE_USER = VALUES(CREATE_USER), EXPIRE_TIME = VALUES(EXPIRE_TIME), CREATE_TIME = CURRENT_TIMESTAMP`,
		banType, value, remark, actor.Code(), expire)
	if err != nil {
		return model.Failure("保存失败，请确认已执行迁移 008_add_comment_ban.sql: " + err.Error())
	}
	invalidateCommentBanCache()
	period := "永久"
	if days > 0 {
		period = strconv.Itoa(days) + " 天"
	}
	s.recordAdminAction(c, actor, AdminActionCommentBan, commentBanTypes[banType]+"："+value, "禁止评论（"+period+"）"+remarkSuffix(remark))
	return model.Success("已添加")
}

func (s *Service) DeleteCommentBan(c *gin.Context, id int) model.Result {
	actor, failure := s.requireSuperAdmin(c)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	rows, err := s.Repo.Query(ctx, "SELECT TYPE, VALUE FROM commentBan WHERE ID = ? LIMIT 1", id)
	if err != nil || len(rows) == 0 {
		return model.Failure(msgNotFound)
	}
	if _, err := s.Repo.Exec(ctx, "DELETE FROM commentBan WHERE ID = ?", id); err != nil {
		return dbFailure("删除评论限制", err)
	}
	invalidateCommentBanCache()
	s.recordAdminAction(c, actor, AdminActionCommentBan,
		commentBanTypes[model.StringValue(rows[0], "TYPE")]+"："+model.StringValue(rows[0], "VALUE"), "解除评论限制")
	return model.Success("已解除")
}

func remarkSuffix(remark string) string {
	if remark == "" {
		return ""
	}
	return "，备注：" + remark
}

// parseControllerLikeTime 解析数据库返回的时间（time.Time 或 "2006-01-02 15:04:05" 字符串）。
func parseControllerLikeTime(value any) time.Time {
	if parsed, ok := value.(time.Time); ok {
		return parsed
	}
	text := strings.TrimSpace(model.StringValue(map[string]any{"v": value}, "v"))
	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339} {
		if parsed, err := time.ParseInLocation(layout, text, time.Local); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

// privateCommentFields 是评论表中只给站长使用、不能通过公开接口返回的字段：
// 邮箱只用于回复通知，IP 只用于后台的评论限制。
var privateCommentFields = []string{"USEREMAIL", "IP"}

// stripPrivateCommentFields 删除公开评论列表中的隐私字段（大小写都处理，查询用的是 bc.*）。
func stripPrivateCommentFields(result model.Result) model.Result {
	rows, ok := result.Result.([]map[string]any)
	if !ok {
		return result
	}
	for _, row := range rows {
		for key := range row {
			for _, field := range privateCommentFields {
				if strings.EqualFold(key, field) {
					delete(row, key)
				}
			}
		}
	}
	return result
}
