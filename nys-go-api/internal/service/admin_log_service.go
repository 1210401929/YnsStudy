package service

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

// 管理员操作日志。记录失败（例如还没执行迁移 007、表不存在）只写服务日志，不影响操作本身。

// 操作类型，前端按这些值显示中文名称和筛选
const (
	AdminActionSetAdmin       = "setAdmin"
	AdminActionRemoveAdmin    = "removeAdmin"
	AdminActionBan            = "ban"
	AdminActionRemoveBan      = "removeBan"
	AdminActionForceLogout    = "forceLogout"
	AdminActionResetPassword  = "resetPassword"
	AdminActionDeleteComment  = "deleteComment"
	AdminActionDeleteArticle  = "deleteArticle"
	AdminActionArticleType    = "articleType"
	AdminActionDeletePost     = "deletePost"
	AdminActionTopPost        = "topPost"
	AdminActionAnnouncement   = "announcement"
	AdminActionDeleteFiles    = "deleteFiles"
	AdminActionBannedWord     = "bannedWord"
	AdminActionIndexNowSubmit = "indexNow"
	AdminActionEditArticle    = "editArticle"
	AdminActionDeleteResource = "deleteResource"
	AdminActionEditResource   = "editResource"
	AdminActionFriendLink     = "friendLink"
	AdminActionCommentBan     = "commentBan"
	AdminActionEditPost       = "editPost"
)

// msgBannedWord 是内容命中违禁词时的提示
const msgBannedWord = "提交的内容包含违禁词，禁止发布！"

var adminLogWarnOnce sync.Once

// recordAdminAction 记录一条管理员操作。actor 为空时从请求中取当前用户。
func (s *Service) recordAdminAction(c *gin.Context, actor *Actor, action, target, detail string) {
	if actor == nil {
		current, ok := s.CurrentActor(c)
		if !ok {
			return
		}
		actor = current
	}
	_, err := s.Repo.Exec(contextOf(c),
		"INSERT INTO adminLog (ACTOR_CODE, ACTOR_NAME, ACTION, TARGET, DETAIL, IP) VALUES (?, ?, ?, ?, ?, ?)",
		actor.Code(), actor.User.Name, action, truncateText(target, 255), truncateText(detail, 1000), ClientIP(c))
	if err != nil {
		adminLogWarnOnce.Do(func() {
			log.Printf("记录管理员操作日志失败（是否已执行迁移 007_add_admin_tables.sql？）: %v", err)
		})
	}
}

func truncateText(value string, maximum int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= maximum {
		return string(runes)
	}
	return string(runes[:maximum])
}

// GetAdminLogs 分页返回操作日志，action 为空时返回全部。
func (s *Service) GetAdminLogs(c *gin.Context, page, pageSize int, action string) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	page, pageSize = normalizePage(page, pageSize)
	where := ""
	args := make([]any, 0, 3)
	if action = strings.TrimSpace(action); action != "" {
		where = " WHERE ACTION = ?"
		args = append(args, action)
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM adminLog"+where, args...)
	if err != nil {
		return model.Failure("读取操作日志失败，请确认已执行迁移 007_add_admin_tables.sql")
	}
	rows, err := s.Repo.Query(ctx, "SELECT * FROM adminLog"+where+" ORDER BY ID DESC LIMIT ? OFFSET ?", append(args, pageSize, (page-1)*pageSize)...)
	if err != nil {
		return dbFailure("读取操作日志", err)
	}
	return model.Success(map[string]any{"total": firstCount(counts), "data": rows})
}

// RecordFileCleanup 在文件清理（删除模式）完成后记录操作日志。
func (s *Service) RecordFileCleanup(c *gin.Context, fileType string, result model.Result) {
	data, _ := result.Result.(map[string]any)
	deleted, _ := data["deleted"].([]string)
	if len(deleted) == 0 {
		return
	}
	detail := "删除 " + strconv.Itoa(len(deleted)) + " 个未引用文件：" + strings.Join(deleted, "、")
	s.recordAdminAction(c, nil, AdminActionDeleteFiles, fileType, detail)
}

// logModeration 记录管理员在前台页面（文章页、社区、资源、友链）处理他人内容的操作。
// 自己的内容不记录；后台 /blog-api/admin/ 接口已经单独记录，这里跳过以免重复。
func (s *Service) logModeration(c *gin.Context, actor *Actor, row map[string]any, action, target, detail string) {
	if strings.HasPrefix(c.Request.URL.Path, "/blog-api/admin/") {
		return
	}
	if actor == nil {
		current, ok := s.CurrentActor(c)
		if !ok {
			return
		}
		actor = current
	}
	if !actor.IsAdmin || model.StringValue(row, "USERCODE") == actor.Code() {
		return
	}
	s.recordAdminAction(c, actor, action, target, detail)
}
