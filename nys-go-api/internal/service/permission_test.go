package service

import (
	"testing"

	"nys-go-api/internal/config"
	"nys-go-api/internal/model"
)

func TestActorCanManageFollowsFrontendRules(t *testing.T) {
	const superCode = "yulei"
	normal := &Actor{User: &model.User{Code: "alice"}, SuperCode: superCode}
	admin := &Actor{User: &model.User{Code: "bob", Role: "admin"}, SuperCode: superCode, IsAdmin: true}
	super := &Actor{User: &model.User{Code: superCode}, SuperCode: superCode, IsSuper: true, IsAdmin: true}

	cases := []struct {
		name  string
		actor *Actor
		owner string
		want  bool
	}{
		{"普通用户管理自己的内容", normal, "alice", true},
		{"普通用户不能管理他人内容", normal, "carol", false},
		{"普通用户不能管理匿名内容", normal, "", false},
		{"管理员可以管理普通用户内容", admin, "carol", true},
		{"管理员可以管理匿名评论", admin, "", true},
		{"管理员不能管理超级管理员内容", admin, superCode, false},
		{"超级管理员可以管理任何内容", super, "carol", true},
		{"超级管理员管理自己的内容", super, superCode, true},
	}
	for _, tc := range cases {
		if got := tc.actor.CanManage(tc.owner); got != tc.want {
			t.Errorf("%s: CanManage(%q) = %v, want %v", tc.name, tc.owner, got, tc.want)
		}
	}
}

func TestPickFieldsDropsSensitiveColumns(t *testing.T) {
	input := map[string]any{
		"NAME": "新名字", "remark": "签名", "ROLE": "admin", "ISBAN": "", "PASSWORD": "x", "GUID": "other-user",
	}
	got := pickFields(input, editableUserFields...)
	if got["NAME"] != "新名字" || got["REMARK"] != "签名" {
		t.Fatalf("白名单字段丢失: %#v", got)
	}
	for _, forbidden := range []string{"ROLE", "ISBAN", "PASSWORD", "GUID"} {
		if _, exists := got[forbidden]; exists {
			t.Errorf("不应允许客户端写入 %s: %#v", forbidden, got)
		}
	}
}

func TestPublicProfileHidesPhoneAndPassword(t *testing.T) {
	user := model.User{Name: "alice", Phone: "13800000000", Password: "hash", PasswordSalt: "salt", LoginIP: "1.2.3.4", Email: "a@b.c"}
	profile := user.PublicProfile()
	if profile.Phone != "" || profile.Password != "" || profile.PasswordSalt != "" || profile.LoginIP != "" {
		t.Fatalf("公开资料不应包含手机号、密码或 IP: %#v", profile)
	}
	if profile.Email != "a@b.c" || profile.Name != "alice" {
		t.Fatalf("公开展示的字段不应被清除: %#v", profile)
	}
}

func TestSuperAdminCodeDefaultsToFrontendValue(t *testing.T) {
	s := &Service{Config: &config.Config{}}
	if got := s.superAdminCode(); got != config.DefaultSuperAdminCode {
		t.Fatalf("未配置时应使用默认超级管理员 %q，得到 %q", config.DefaultSuperAdminCode, got)
	}
	s.Config.Security.SuperAdminCode = " owner "
	if got := s.superAdminCode(); got != "owner" {
		t.Fatalf("应使用配置的超级管理员，得到 %q", got)
	}
}

func TestLoadRowRejectsUnknownTables(t *testing.T) {
	if ownedTables["userInfo"] || ownedTables["noticeInfo"] {
		t.Fatal("归属检查只允许固定的内容表")
	}
}
