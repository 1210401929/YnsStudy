package service

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
	"nys-go-api/internal/security"
	"nys-go-api/internal/session"
)

// 后台接口里 getDashboard 等以 get 开头的地址不经过 JWT 中间件，权限完全依赖 service 层的校验。
func TestAdminEndpointsRejectAnonymousRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := cache.NewMemoryStore()
	cfg := &config.Config{Security: config.SecurityConfig{SessionCookieName: "NYS_SESSION"}}
	s := &Service{
		Config:   cfg,
		Cache:    store,
		Sessions: session.NewManager(store, cfg.Security),
		JWT:      security.NewJWT("test-secret", time.Hour, time.Hour),
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/blog-api/admin/getDashboard", nil)

	for name, call := range map[string]func() bool{
		"getDashboard":     func() bool { return s.GetAdminDashboard(c).IsError },
		"getComments":      func() bool { return s.GetAdminComments(c, AdminCommentQuery{}).IsError },
		"deleteComment":    func() bool { return s.DeleteAdminComment(c, "blog", "x").IsError },
		"getUserDetail":    func() bool { return s.GetAdminUserDetail(c, "x").IsError },
		"forceLogout":      func() bool { return s.ForceLogoutUser(c, "x").IsError },
		"resetPassword":    func() bool { return s.ResetUserPassword(c, "x").IsError },
		"getArticles":      func() bool { return s.GetAdminArticles(c, AdminContentQuery{}).IsError },
		"setArticleType":   func() bool { return s.SetArticleVisibility(c, "x", "public").IsError },
		"deleteArticle":    func() bool { return s.AdminDeleteArticle(c, "x").IsError },
		"getPosts":         func() bool { return s.GetAdminCommunityPosts(c, AdminContentQuery{}).IsError },
		"setPostTop":       func() bool { return s.AdminSetTopPost(c, "x", "1").IsError },
		"deletePost":       func() bool { return s.AdminDeletePost(c, "x").IsError },
		"getBannedWords":   func() bool { return s.GetBannedWords(c).IsError },
		"addBannedWord":    func() bool { return s.AddBannedWord(c, "x").IsError },
		"deleteBannedWord": func() bool { return s.DeleteBannedWord(c, "x").IsError },
		"getSeoStatus":     func() bool { return s.GetSeoStatus(c).IsError },
		"submitIndexNow":   func() bool { return s.SubmitIndexNowManually(c, []string{"/"}).IsError },
		"getLogs":          func() bool { return s.GetAdminLogs(c, 1, 20, "").IsError },
		"getCommentBans":   func() bool { return s.GetCommentBans(c).IsError },
		"addCommentBan":    func() bool { return s.AddCommentBan(c, CommentBanIP, "1.2.3.4", "", 0).IsError },
		"deleteCommentBan": func() bool { return s.DeleteCommentBan(c, 1).IsError },
	} {
		if !call() {
			t.Errorf("%s 未登录时应被拒绝", name)
		}
	}
}

func TestRandomReadablePassword(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		password, err := randomReadablePassword(12)
		if err != nil {
			t.Fatal(err)
		}
		if len(password) != 12 {
			t.Fatalf("临时密码长度应为 12: %q", password)
		}
		if strings.ContainsAny(password, "0O1lI") {
			t.Fatalf("临时密码不应包含容易看错的字符: %q", password)
		}
		seen[password] = true
	}
	if len(seen) < 50 {
		t.Fatal("临时密码出现重复，随机性不足")
	}
}

func TestAdminCommentFilterAndSumCounts(t *testing.T) {
	where, args := adminCommentFilter("c", AdminCommentQuery{Keyword: " 广告 ", AnonymousOnly: true})
	if where != " WHERE (c.TEXT LIKE ? OR c.USERNAME LIKE ?) AND (c.USERCODE IS NULL OR c.USERCODE = '')" {
		t.Fatalf("筛选条件不正确: %s", where)
	}
	if len(args) != 2 || args[0] != "%广告%" {
		t.Fatalf("关键词参数不正确: %#v", args)
	}
	if where, args := adminCommentFilter("c", AdminCommentQuery{}); where != "" || len(args) != 0 {
		t.Fatal("没有筛选条件时不应生成 WHERE")
	}
	if got := sumCounts(int64(3), int64(4)); got != int64(7) {
		t.Fatalf("sumCounts = %v", got)
	}
	if got := sumCounts(int64(3), nil); got != nil {
		t.Fatal("任一统计失败时合计应为空，而不是只算一部分")
	}
}

func TestNormalizeSitePathsOnlyAcceptsOwnSite(t *testing.T) {
	paths, invalid := normalizeSitePaths([]string{
		"/oneBlog/441",
		"https://ynsstudy.cn/user/101",
		"https://YNSSTUDY.cn/archive?x=1",
		" /oneBlog/441 ", // 重复，去重
		"",
		"https://evil.example.com/a",
		"//evil.example.com/a",
		"ftp://ynsstudy.cn/x",
	}, "https://ynsstudy.cn/")
	want := []string{"/oneBlog/441", "/user/101", "/archive?x=1"}
	if len(paths) != len(want) {
		t.Fatalf("paths = %#v", paths)
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Fatalf("paths[%d] = %q, want %q", i, paths[i], want[i])
		}
	}
	if len(invalid) != 3 || invalid[0] != "https://evil.example.com/a" || invalid[1] != "//evil.example.com/a" || invalid[2] != "ftp://ynsstudy.cn/x" {
		t.Fatalf("外站地址应被拒绝: %#v", invalid)
	}
}

func TestContainsBannedWordUsesConfigWithoutDatabase(t *testing.T) {
	invalidateBannedWordCache()
	s := &Service{Config: &config.Config{Content: config.ContentConfig{BannedWords: []string{"赌博", "SPAM"}}}}
	if !s.containsBannedWord("这里有赌博广告") || !s.containsBannedWord("buy spam now") {
		t.Fatal("配置文件中的违禁词应生效，且不区分大小写")
	}
	if s.containsBannedWord("正常的评论") || s.containsBannedWord("") {
		t.Fatal("正常内容不应被拦截")
	}
}
