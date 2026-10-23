package service

// 便签墙“找人”：任何访客都能用的用户搜索，只返回公开信息。
// 后台的用户管理（GetAllUsers）会返回手机号、邮箱、登录 IP 等，只允许超级管理员调用，不能拿来给访客用。

import (
	"context"
	"strings"

	"nys-go-api/internal/model"
)

const maxPublicUserPageSize = 20

// SearchPublicUsers 按昵称或简介搜索用户；关键词是纯数字时只按用户编号精确查找，
// 不去匹配昵称，避免有人用数字试探出手机号形式的昵称。没有关键词时按“站长在前、公开文章多的在前”列出。
func (s *Service) SearchPublicUsers(ctx context.Context, keyword string, page, pageSize int) model.Result {
	page, pageSize = normalizePage(page, pageSize)
	if pageSize > maxPublicUserPageSize {
		pageSize = maxPublicUserPageSize
	}
	keyword = strings.TrimSpace(keyword)
	if len([]rune(keyword)) > 30 {
		keyword = string([]rune(keyword)[:30])
	}

	conditions := []string{"(u.ISBAN IS NULL OR u.ISBAN <> '1')"}
	args := make([]any, 0, 6)
	if keyword != "" {
		if isNumericID(keyword) {
			conditions = append(conditions, "u.USERNUM = ?")
			args = append(args, keyword)
		} else {
			like := likePattern(keyword)
			conditions = append(conditions, "(u.NAME LIKE ? OR u.REMARK LIKE ?)")
			args = append(args, like, like)
		}
	}
	where := " WHERE " + strings.Join(conditions, " AND ")

	rows, err := s.Repo.Query(ctx, `SELECT u.CODE, u.NAME, u.AVATAR, u.REMARK, u.USERNUM, u.ROLE, COALESCE(b.TOTAL, 0) AS ARTICLES
FROM userInfo u
LEFT JOIN (SELECT USERCODE, COUNT(*) AS TOTAL FROM blogInfo WHERE BLOG_TYPE = 'public' GROUP BY USERCODE) b ON b.USERCODE = u.CODE`+where+`
ORDER BY (u.CODE = ?) DESC, ARTICLES DESC, u.USERNUM ASC LIMIT ? OFFSET ?`,
		append(append([]any{}, args...), s.superAdminCode(), pageSize, (page-1)*pageSize)...)
	if err != nil {
		return dbFailure("搜索用户", err)
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM userInfo u"+where, args...)
	if err != nil {
		return dbFailure("统计用户", err)
	}
	superCode := s.superAdminCode()
	for _, row := range rows {
		// 手机号形式的昵称显示成“手机用户xxxx”，与全站一致
		row["NAME"] = model.PublicDisplayName(model.StringValue(row, "NAME"))
		row["ISOWNER"] = model.StringValue(row, "CODE") == superCode
		row["ISADMIN"] = model.StringValue(row, "ROLE") == "admin"
		delete(row, "ROLE")
	}
	return model.Success(map[string]any{"total": firstCount(counts), "data": rows})
}
