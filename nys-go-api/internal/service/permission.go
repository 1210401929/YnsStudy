package service

import (
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/config"
	"nys-go-api/internal/model"
)

// 权限规则与前端 getCurrentUserAdminObject 保持一致：
//   - 超级管理员：账号等于 security.super_admin_code，只有一个；
//   - 管理员：userInfo.ROLE = 'admin'；
//   - 普通用户只能管理自己的内容，管理员可以管理除超级管理员以外所有人的内容。
// 前端只是隐藏按钮，真正的限制必须在这里完成。

const (
	msgNotLoggedIn = "用户未登录!"
	msgForbidden   = "无权限执行此操作!"
	msgNotFound    = "数据不存在或已删除!"
)

// Actor 是当前请求的登录用户及其权限。
type Actor struct {
	User      *model.User
	SuperCode string
	IsSuper   bool
	IsAdmin   bool // 包含超级管理员
}

// Code 返回当前用户账号。
func (a *Actor) Code() string { return a.User.Code }

// CanManage 判断能否修改或删除属于 ownerCode 的内容。
func (a *Actor) CanManage(ownerCode string) bool {
	ownerCode = strings.TrimSpace(ownerCode)
	if a.IsSuper || (ownerCode != "" && ownerCode == a.User.Code) {
		return true
	}
	return a.IsAdmin && ownerCode != a.SuperCode
}

func (s *Service) superAdminCode() string {
	if code := strings.TrimSpace(s.Config.Security.SuperAdminCode); code != "" {
		return code
	}
	return config.DefaultSuperAdminCode
}

// CurrentActor 返回当前登录用户及权限。会话里的用户信息可能是登录时的旧数据，
// 这里按账号重新读取，保证刚被设置/撤销管理员、封禁的状态立即生效。
func (s *Service) CurrentActor(c *gin.Context) (*Actor, bool) {
	user, err := s.CurrentUser(c)
	if err != nil || user == nil || user.Code == "" {
		return nil, false
	}
	if rows, err := s.Repo.Query(contextOf(c), "SELECT * FROM userInfo WHERE CODE = ? LIMIT 1", user.Code); err == nil && len(rows) > 0 {
		fresh := model.UserFromRow(rows[0])
		user = &fresh
	}
	if user.IsBan == "1" {
		return nil, false
	}
	superCode := s.superAdminCode()
	isSuper := user.Code == superCode
	return &Actor{
		User:      user,
		SuperCode: superCode,
		IsSuper:   isSuper,
		IsAdmin:   isSuper || user.Role == "admin",
	}, true
}

// requireLogin 返回当前登录用户；未登录时返回失败结果。
func (s *Service) requireLogin(c *gin.Context) (*Actor, *model.Result) {
	actor, ok := s.CurrentActor(c)
	if !ok {
		result := model.Failure(msgNotLoggedIn)
		return nil, &result
	}
	return actor, nil
}

// requireSuperAdmin 只允许超级管理员继续执行。
func (s *Service) requireSuperAdmin(c *gin.Context) (*Actor, *model.Result) {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return nil, failure
	}
	if !actor.IsSuper {
		result := model.Failure(msgForbidden)
		return nil, &result
	}
	return actor, nil
}

// AuthorizeSuperAdmin 供控制器直接使用：非超级管理员时返回失败结果。
func (s *Service) AuthorizeSuperAdmin(c *gin.Context) *model.Result {
	_, failure := s.requireSuperAdmin(c)
	return failure
}

// requireAdmin 允许管理员和超级管理员继续执行。
func (s *Service) requireAdmin(c *gin.Context) (*Actor, *model.Result) {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return nil, failure
	}
	if !actor.IsAdmin {
		result := model.Failure(msgForbidden)
		return nil, &result
	}
	return actor, nil
}

// requireOwnerOrAdmin 读取 table 中主键为 guid 的数据的 USERCODE，并检查当前用户能否管理它。
// 返回该行数据，供调用方继续使用（例如删除前读取文件地址）。
func (s *Service) requireOwnerOrAdmin(c *gin.Context, table, guid string) (map[string]any, *model.Result) {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return nil, failure
	}
	row, failure := s.loadRow(c, table, guid)
	if failure != nil {
		return nil, failure
	}
	if !actor.CanManage(model.StringValue(row, "USERCODE")) {
		result := model.Failure(msgForbidden)
		return nil, &result
	}
	return row, nil
}

// ownedTables 是允许按 GUID 读取归属信息的表，表名固定，避免拼接任意输入。
var ownedTables = map[string]bool{
	"blogInfo": true, "blogComment": true, "blogCatInfo": true, "communityInfo": true,
	"fileInfo": true, "friendLinkInfo": true,
}

func (s *Service) loadRow(c *gin.Context, table, guid string) (map[string]any, *model.Result) {
	if !ownedTables[table] {
		result := model.Failure("不支持的数据类型")
		return nil, &result
	}
	guid = strings.TrimSpace(guid)
	if guid == "" {
		result := model.Failure(msgNotFound)
		return nil, &result
	}
	rows, err := s.Repo.Query(contextOf(c), "SELECT * FROM "+table+" WHERE GUID = ? LIMIT 1", guid)
	if err != nil {
		result := dbFailure("查询数据", err)
		return nil, &result
	}
	if len(rows) == 0 {
		result := model.Failure(msgNotFound)
		return nil, &result
	}
	return rows[0], nil
}

// pickFields 只保留允许由客户端写入的字段，防止通过请求体修改 USERCODE、ROLE 等敏感列。
func pickFields(input map[string]any, allowed ...string) map[string]any {
	output := make(map[string]any, len(allowed))
	for _, field := range allowed {
		if value := model.Lookup(input, field); value != nil {
			output[field] = value
		}
	}
	return output
}
