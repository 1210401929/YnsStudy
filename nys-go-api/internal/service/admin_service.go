package service

import (
	"context"
	"crypto/rand"
	"log"
	"math/big"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
	"nys-go-api/internal/security"
)

// 超级管理员后台：数据概览、评论与留言、用户管理。所有接口都只允许超级管理员调用。

// columnCache 缓存“某表是否有某列”。旧库的部分表（如 loginHistory、userInfo）没有统一的时间字段，
// 统计前先查 information_schema，缺列时对应数据显示为空，而不是让整个页面报错。
var columnCache sync.Map

func (s *Service) tableHasColumn(ctx context.Context, table, column string) bool {
	key := strings.ToLower(table + "." + column)
	if cached, ok := columnCache.Load(key); ok {
		return cached.(bool)
	}
	rows, err := s.Repo.Query(ctx, `SELECT COUNT(*) AS total FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND LOWER(TABLE_NAME) = LOWER(?) AND UPPER(COLUMN_NAME) = UPPER(?)`, table, column)
	if err != nil {
		return false
	}
	exists := firstCount(rows) > 0
	columnCache.Store(key, exists)
	return exists
}

// loginTimeColumn 返回 loginHistory 的登录时间列名，没有时返回空字符串。
func (s *Service) loginTimeColumn(ctx context.Context) string {
	for _, column := range []string{"CREATE_TIME", "LOGIN_TIME", "LOGINTIME", "CREATETIME"} {
		if s.tableHasColumn(ctx, "loginHistory", column) {
			return column
		}
	}
	return ""
}

// countOrNil 执行 COUNT/SUM 查询；失败时返回 nil，前端显示“无数据”。
func (s *Service) countOrNil(ctx context.Context, query string, args ...any) any {
	rows, err := s.Repo.Query(ctx, query, args...)
	if err != nil {
		log.Printf("后台统计查询失败: %v", err)
		return nil
	}
	return firstCount(rows)
}

// GetAdminDashboard 返回站点概览：总量、近 7/30 天新增、阅读量最高的文章、最近登录。
func (s *Service) GetAdminDashboard(c *gin.Context) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	ctx := contextOf(c)

	totals := map[string]any{
		"publicArticles":    s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM blogInfo WHERE BLOG_TYPE = 'public'"),
		"privateArticles":   s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM blogInfo WHERE BLOG_TYPE <> 'public' OR BLOG_TYPE IS NULL"),
		"communityPosts":    s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM communityInfo"),
		"blogComments":      s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM blogComment"),
		"communityComments": s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM communityComment"),
		"users":             s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM userInfo"),
		"totalViews":        s.countOrNil(ctx, "SELECT COALESCE(SUM(VIEW_PAGE), 0) AS total FROM blogInfo"),
		"logins":            s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM loginHistory"),
		"guestbook":         nil,
	}
	// 执行迁移 011 之后才有留言板
	guestbookReady := s.guestbookReady(ctx)
	if guestbookReady {
		totals["guestbook"] = s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM guestbook WHERE SUPERGUID IS NULL OR SUPERGUID = ''")
	}

	recent := make(map[string]any, 2)
	userHasTime := s.tableHasColumn(ctx, "userInfo", "CREATE_TIME")
	for label, days := range map[string]int{"days7": 7, "days30": 30} {
		since := "CREATE_TIME >= DATE_SUB(NOW(), INTERVAL ? DAY)"
		item := map[string]any{
			"articles":       s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM blogInfo WHERE "+since, days),
			"communityPosts": s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM communityInfo WHERE "+since, days),
			"comments": sumCounts(
				s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM blogComment WHERE "+since, days),
				s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM communityComment WHERE "+since, days),
			),
			"users":     nil,
			"guestbook": nil,
		}
		if guestbookReady {
			item["guestbook"] = s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM guestbook WHERE (SUPERGUID IS NULL OR SUPERGUID = '') AND "+since, days)
		}
		if userHasTime {
			item["users"] = s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM userInfo WHERE "+since, days)
		}
		recent[label] = item
	}

	topArticles, err := s.Repo.Query(ctx, `SELECT GUID, BLOG_TITLE, USERNAME, COALESCE(VIEW_PAGE, 0) AS VIEW_PAGE, CREATE_TIME
FROM blogInfo WHERE BLOG_TYPE = 'public' ORDER BY VIEW_PAGE DESC LIMIT 10`)
	if err != nil {
		topArticles = []map[string]any{}
	}
	for _, row := range topArticles {
		row["USERNAME"] = model.PublicDisplayName(model.StringValue(row, "USERNAME"))
	}

	return model.Success(map[string]any{
		"totals":       totals,
		"recent":       recent,
		"topArticles":  topArticles,
		"recentLogins": s.recentLogins(ctx, "", 20),
	})
}

func sumCounts(values ...any) any {
	var total int64
	for _, value := range values {
		number, ok := value.(int64)
		if !ok {
			return nil
		}
		total += number
	}
	return total
}

// recentLogins 返回最近的登录记录；userID 不为空时只看该用户。表里没有时间列时返回空列表。
func (s *Service) recentLogins(ctx context.Context, userID string, limit int) []map[string]any {
	timeColumn := s.loginTimeColumn(ctx)
	if timeColumn == "" {
		return []map[string]any{}
	}
	query := "SELECT USERID, USERNAME, LOGINADDRESS, LOGINIP, " + timeColumn + " AS LOGIN_TIME FROM loginHistory"
	args := make([]any, 0, 2)
	if userID != "" {
		query += " WHERE USERID = ?"
		args = append(args, userID)
	}
	query += " ORDER BY " + timeColumn + " DESC LIMIT ?"
	args = append(args, limit)
	rows, err := s.Repo.Query(ctx, query, args...)
	if err != nil {
		log.Printf("查询登录记录失败: %v", err)
		return []map[string]any{}
	}
	for _, row := range rows {
		row["USERNAME"] = model.PublicDisplayName(model.StringValue(row, "USERNAME"))
	}
	return rows
}

// ---------------- 评论与留言 ----------------

// AdminCommentQuery 是评论列表的筛选条件。
type AdminCommentQuery struct {
	Page, PageSize int
	Keyword        string
	Source         string // all / blog / community
	AnonymousOnly  bool
}

const adminCommentMaxPageSize = 50

// GetAdminComments 合并文章评论和社区评论，按时间倒序分页。
// 两张表分别查询再在内存中合并，避免 UNION 两张表时的字符集排序规则冲突。
func (s *Service) GetAdminComments(c *gin.Context, query AdminCommentQuery) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > adminCommentMaxPageSize {
		query.PageSize = 20
	}
	// 每张表最多取前 page*pageSize 条，合并排序后再截取当前页
	fetch := query.Page * query.PageSize

	where, args := adminCommentFilter("c", query)
	var all []map[string]any
	var total int64
	// 执行迁移 008 之后评论才有 IP 列
	ipColumn := func(table string) string {
		if s.tableHasColumn(ctx, table, "IP") {
			return ", c.IP"
		}
		return ""
	}
	// source 为空或 all 时三种评论都查
	include := func(source string) bool {
		return query.Source == "" || query.Source == "all" || query.Source == source
	}
	// 执行迁移 009 之后随记评论才有邮箱、网址
	communityContact := ""
	if s.tableHasColumn(ctx, "communityComment", "USEREMAIL") {
		communityContact = " c.USEREMAIL, c.USERWEBSITE,"
	}
	if include("blog") {
		rows, err := s.Repo.Query(ctx, `SELECT 'blog' AS SOURCE, c.GUID, c.SUPERGUID, c.TEXT, c.USERCODE, c.USERNAME, c.USEREMAIL, c.USERWEBSITE,
c.CREATE_TIME, c.BLOGID AS TARGETID, b.BLOG_TITLE AS TARGETTITLE`+ipColumn("blogComment")+`
FROM blogComment c LEFT JOIN blogInfo b ON b.GUID = c.BLOGID`+where+` ORDER BY c.CREATE_TIME DESC LIMIT ?`, append(args, fetch)...)
		if err != nil {
			return dbFailure("查询文章评论", err)
		}
		all = append(all, rows...)
		total += firstCount(mustRows(s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM blogComment c"+where, args...)))
	}
	if include("community") {
		rows, err := s.Repo.Query(ctx, `SELECT 'community' AS SOURCE, c.GUID, c.SUPERGUID, c.TEXT, c.USERCODE, c.USERNAME,`+communityContact+`
c.CREATE_TIME, c.COMMUNITYID AS TARGETID, p.TITLE AS TARGETTITLE, p.TEXT AS TARGETTEXT`+ipColumn("communityComment")+`
FROM communityComment c LEFT JOIN communityInfo p ON p.GUID = c.COMMUNITYID`+where+` ORDER BY c.CREATE_TIME DESC LIMIT ?`, append(args, fetch)...)
		if err != nil {
			return dbFailure("查询社区评论", err)
		}
		// 随记大多没有标题，用正文开头当标题
		for _, row := range rows {
			if model.StringValue(row, "TARGETTITLE") == "" && model.StringValue(row, "TARGETTEXT") != "" {
				row["TARGETTITLE"] = NoteTitle(map[string]any{"TEXT": row["TARGETTEXT"]})
			}
			delete(row, "TARGETTEXT")
		}
		all = append(all, rows...)
		total += firstCount(mustRows(s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM communityComment c"+where, args...)))
	}
	if include("guestbook") && s.guestbookReady(ctx) {
		rows, err := s.Repo.Query(ctx, `SELECT 'guestbook' AS SOURCE, c.GUID, c.SUPERGUID, c.TEXT, c.USERCODE, c.USERNAME, c.USEREMAIL, c.USERWEBSITE,
c.CREATE_TIME, '' AS TARGETID, '便签墙留言板' AS TARGETTITLE, c.IP, c.ISTOP
FROM guestbook c`+where+` ORDER BY c.CREATE_TIME DESC LIMIT ?`, append(args, fetch)...)
		if err != nil {
			return dbFailure("查询留言", err)
		}
		all = append(all, rows...)
		total += firstCount(mustRows(s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM guestbook c"+where, args...)))
	}

	sort.SliceStable(all, func(i, j int) bool {
		return model.StringValue(all[i], "CREATE_TIME") > model.StringValue(all[j], "CREATE_TIME")
	})
	start := (query.Page - 1) * query.PageSize
	if start > len(all) {
		start = len(all)
	}
	end := start + query.PageSize
	if end > len(all) {
		end = len(all)
	}
	page := all[start:end]
	for _, row := range page {
		excerpt, _ := summarizeHTML(model.StringValue(row, "TEXT"), 200)
		row["TEXT"] = excerpt
		row["USERNAME"] = model.PublicDisplayName(model.StringValue(row, "USERNAME"))
		row["ISANONYMOUS"] = strings.TrimSpace(model.StringValue(row, "USERCODE")) == ""
		row["ISREPLY"] = strings.TrimSpace(model.StringValue(row, "SUPERGUID")) != ""
	}
	return model.Success(map[string]any{"total": total, "data": page})
}

func adminCommentFilter(alias string, query AdminCommentQuery) (string, []any) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 2)
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		conditions = append(conditions, "("+alias+".TEXT LIKE ? OR "+alias+".USERNAME LIKE ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}
	if query.AnonymousOnly {
		conditions = append(conditions, "("+alias+".USERCODE IS NULL OR "+alias+".USERCODE = '')")
	}
	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func mustRows(rows []map[string]any, err error) []map[string]any {
	if err != nil {
		return nil
	}
	return rows
}

// DeleteAdminComment 删除一条评论及其回复。source 为 blog、community 或 guestbook。
func (s *Service) DeleteAdminComment(c *gin.Context, source, guid string) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	guid = strings.TrimSpace(guid)
	if guid == "" {
		return model.Failure(msgNotFound)
	}
	table := map[string]string{"blog": "blogComment", "community": "communityComment", "guestbook": "guestbook"}[source]
	if table == "" {
		return model.Failure("未知的评论类型")
	}
	// 先读出评论内容写进操作日志，删除后就查不到了
	var excerpt, author string
	if rows, err := s.Repo.Query(contextOf(c), "SELECT TEXT, USERNAME FROM "+table+" WHERE GUID = ? LIMIT 1", guid); err == nil && len(rows) > 0 {
		excerpt, _ = summarizeHTML(model.StringValue(rows[0], "TEXT"), 80)
		author = model.PublicDisplayName(model.StringValue(rows[0], "USERNAME"))
	}
	var result model.Result
	if source == "blog" {
		result = s.DeleteBlogComment(c, guid)
	} else {
		result = s.ExecuteSQL(contextOf(c), "DELETE FROM "+table+" WHERE GUID = ? OR SUPERGUID = ?", []any{guid, guid})
	}
	if !result.IsError {
		s.recordAdminAction(c, nil, AdminActionDeleteComment, author, map[string]string{"blog": "文章评论：", "community": "随记评论：", "guestbook": "留言："}[source]+excerpt)
	}
	return result
}

// ---------------- 用户管理 ----------------

// adminTargetUser 读取被操作的用户；超级管理员本人不能被强制下线或重置密码。
func (s *Service) adminTargetUser(c *gin.Context, userID string) (map[string]any, *model.Result) {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return nil, failure
	}
	rows, err := s.Repo.Query(contextOf(c), "SELECT * FROM userInfo WHERE GUID = ? LIMIT 1", strings.TrimSpace(userID))
	if err != nil {
		result := dbFailure("查询用户", err)
		return nil, &result
	}
	if len(rows) == 0 {
		result := model.Failure("用户不存在")
		return nil, &result
	}
	return rows[0], nil
}

// GetAdminUserDetail 返回用户的发文、评论数量和最近登录记录。
func (s *Service) GetAdminUserDetail(c *gin.Context, userID string) model.Result {
	row, failure := s.adminTargetUser(c, userID)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	user := model.UserFromRow(row)
	user.Password, user.PasswordSalt = "", ""
	return model.Success(map[string]any{
		"user": user,
		"stats": map[string]any{
			"publicArticles":  s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM blogInfo WHERE USERCODE = ? AND BLOG_TYPE = 'public'", user.Code),
			"privateArticles": s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM blogInfo WHERE USERCODE = ? AND (BLOG_TYPE <> 'public' OR BLOG_TYPE IS NULL)", user.Code),
			"communityPosts":  s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM communityInfo WHERE USERCODE = ?", user.Code),
			"comments": sumCounts(
				s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM blogComment WHERE USERCODE = ?", user.Code),
				s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM communityComment WHERE USERCODE = ?", user.Code),
			),
			"logins": s.countOrNil(ctx, "SELECT COUNT(*) AS total FROM loginHistory WHERE USERID = ?", user.GUID),
		},
		"recentLogins": s.recentLogins(ctx, user.GUID, 20),
	})
}

// ForceLogoutUser 让该用户在所有设备上的登录立即失效。
func (s *Service) ForceLogoutUser(c *gin.Context, userID string) model.Result {
	row, failure := s.adminTargetUser(c, userID)
	if failure != nil {
		return *failure
	}
	if model.StringValue(row, "CODE") == s.superAdminCode() {
		return model.Failure("不能对超级管理员执行此操作")
	}
	if err := s.Sessions.RevokeUser(contextOf(c), model.StringValue(row, "GUID")); err != nil {
		return model.Failure("强制下线失败: " + err.Error())
	}
	s.recordAdminAction(c, nil, AdminActionForceLogout, userLabel(row), "强制下线")
	return model.Success("已强制下线")
}

// ResetUserPassword 为用户生成一个随机临时密码（只返回这一次），并让其所有登录失效。
func (s *Service) ResetUserPassword(c *gin.Context, userID string) model.Result {
	row, failure := s.adminTargetUser(c, userID)
	if failure != nil {
		return *failure
	}
	if model.StringValue(row, "CODE") == s.superAdminCode() {
		return model.Failure("超级管理员请在个人中心自行修改密码")
	}
	temporary, err := randomReadablePassword(12)
	if err != nil {
		return model.Failure("生成临时密码失败")
	}
	salt, err := security.NewSalt()
	if err != nil {
		return model.Failure(err.Error())
	}
	hashed, err := security.HashPassword(temporary, salt)
	if err != nil {
		return model.Failure(err.Error())
	}
	ctx := contextOf(c)
	guid := model.StringValue(row, "GUID")
	if _, err := s.Repo.Exec(ctx, "UPDATE userInfo SET PASSWORD = ?, PASSWORDSALT = ? WHERE GUID = ?", hashed, salt, guid); err != nil {
		return dbFailure("重置密码", err)
	}
	if err := s.Sessions.RevokeUser(ctx, guid); err != nil {
		log.Printf("重置密码后作废旧登录失败: %v", err)
	}
	// 日志里不记录临时密码本身
	s.recordAdminAction(c, nil, AdminActionResetPassword, userLabel(row), "重置为临时密码")
	return model.Success(map[string]any{
		"password": temporary,
		"account":  model.StringValue(row, "CODE"),
	})
}

// readablePasswordAlphabet 去掉了 0/O、1/l/I 等容易看错的字符，方便管理员口头或截图告诉用户。
const readablePasswordAlphabet = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789"

func randomReadablePassword(length int) (string, error) {
	var builder strings.Builder
	limit := big.NewInt(int64(len(readablePasswordAlphabet)))
	for i := 0; i < length; i++ {
		index, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		builder.WriteByte(readablePasswordAlphabet[index.Int64()])
	}
	return builder.String(), nil
}
