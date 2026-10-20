package service

import (
	"context"
	"net"
	"testing"
	"time"

	"nys-go-api/internal/config"
	"nys-go-api/internal/model"
)

func TestMatchCommentBanByEachType(t *testing.T) {
	_, network, _ := net.ParseCIDR("10.0.0.0/24")
	now := time.Now()
	rules := []commentBanRule{
		{Type: CommentBanUser, Value: "spammer"},
		{Type: CommentBanIP, Value: "1.2.3.4"},
		{Type: CommentBanIP, Value: "10.0.0.0/24", ipNet: network},
		{Type: CommentBanWebsite, Value: "spam-site.com"},
		{Type: CommentBanEmail, Value: "@junk.cn"},
		{Type: CommentBanNickname, Value: "广告"},
		{Type: CommentBanUser, Value: "expired-user", ExpiresAt: now.Add(-time.Hour)},
	}
	cases := []struct {
		name    string
		subject commentSubject
		want    string
	}{
		{"注册用户精确匹配", commentSubject{UserCode: "spammer"}, CommentBanUser},
		{"账号前缀相同不算", commentSubject{UserCode: "spammer2"}, ""},
		{"单个 IP", commentSubject{IP: "1.2.3.4"}, CommentBanIP},
		{"网段内的 IP", commentSubject{IP: "10.0.0.88"}, CommentBanIP},
		{"网段外的 IP", commentSubject{IP: "10.0.1.1"}, ""},
		{"网址包含即拦截，不区分大小写", commentSubject{Website: "https://WWW.Spam-Site.com/post/1"}, CommentBanWebsite},
		{"邮箱包含", commentSubject{Email: "Bob@Junk.cn"}, CommentBanEmail},
		{"昵称包含", commentSubject{Nickname: "专业广告投放"}, CommentBanNickname},
		{"已过期的规则不生效", commentSubject{UserCode: "expired-user"}, ""},
		{"正常评论", commentSubject{UserCode: "alice", IP: "8.8.8.8", Website: "https://alice.dev", Email: "a@b.c", Nickname: "Alice"}, ""},
	}
	for _, tc := range cases {
		rule := matchCommentBan(rules, tc.subject, now)
		got := ""
		if rule != nil {
			got = rule.Type
		}
		if got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestNormalizeCommentBanValue(t *testing.T) {
	s := &Service{Config: &config.Config{}}
	ctx := context.Background()
	cases := []struct {
		banType, input, want string
		ok                   bool
	}{
		{CommentBanWebsite, "HTTPS://Spam.com/", "spam.com", true},
		{CommentBanIP, " 1.2.3.4 ", "1.2.3.4", true},
		{CommentBanIP, "10.0.0.7/24", "10.0.0.0/24", true},
		{CommentBanIP, "not-an-ip", "", false},
		{CommentBanEmail, "@Junk.CN", "@junk.cn", true},
		{CommentBanNickname, "", "", false},
		{"unknown", "x", "", false},
	}
	for _, tc := range cases {
		got, problem := s.normalizeCommentBanValue(ctx, tc.banType, tc.input)
		if (problem == "") != tc.ok || (tc.ok && got != tc.want) {
			t.Errorf("normalize(%s, %q) = %q, %q", tc.banType, tc.input, got, problem)
		}
	}
	// 超级管理员不能被限制（在查询数据库之前就拒绝）
	if _, problem := s.normalizeCommentBanValue(ctx, CommentBanUser, config.DefaultSuperAdminCode); problem == "" {
		t.Fatal("不应允许限制超级管理员")
	}
}

func TestStripPrivateCommentFields(t *testing.T) {
	result := stripPrivateCommentFields(model.Success([]map[string]any{
		{"GUID": "1", "TEXT": "你好", "IP": "1.2.3.4", "USEREMAIL": "a@b.c", "ip": "x"},
	}))
	row := result.Result.([]map[string]any)[0]
	for _, field := range []string{"IP", "ip", "USEREMAIL"} {
		if _, exists := row[field]; exists {
			t.Fatalf("公开评论不应返回 %s: %#v", field, row)
		}
	}
	if row["TEXT"] != "你好" || row["GUID"] != "1" {
		t.Fatalf("其他字段不应被删除: %#v", row)
	}
}
