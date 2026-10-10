package service

import (
	"context"
	"encoding/base64"
	"strings"
	"sync"
	"testing"
	"time"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
)

func newMailTestService(t *testing.T) (*Service, func() []mailJob) {
	t.Helper()
	mailSendInterval = 0
	cfg := &config.Config{
		Mail:     config.MailConfig{Enabled: true, Username: "site@qq.com", AdminEmail: "Admin@QQ.com"},
		External: config.ExternalConfig{DomainName: "https://ynsstudy.cn"},
		Security: config.SecurityConfig{JWTSecret: "test-secret"},
	}
	var mu sync.Mutex
	var sent []mailJob
	service := &Service{Config: cfg, Cache: cache.NewMemoryStore(), sendMailFunc: func(job mailJob) error {
		mu.Lock()
		defer mu.Unlock()
		sent = append(sent, job)
		return nil
	}}
	return service, func() []mailJob {
		time.Sleep(50 * time.Millisecond)
		mu.Lock()
		defer mu.Unlock()
		return append([]mailJob(nil), sent...)
	}
}

func TestQueueMailDailyLimitAndUnsubscribeFooter(t *testing.T) {
	service, sent := newMailTestService(t)
	ctx := context.Background()
	for i := 0; i < defaultMailDailyLimit+3; i++ {
		service.queueMail(ctx, "guest@example.com", "回复通知", "<p>内容</p>")
	}
	for i := 0; i < defaultMailDailyLimit+3; i++ {
		service.queueMail(ctx, "admin@qq.com", "新评论", "<p>内容</p>")
	}
	service.queueMail(ctx, "bad address\r\nBcc: x@y.com", "注入", "x")

	jobs := sent()
	guest, admin := 0, 0
	for _, job := range jobs {
		switch job.To {
		case "guest@example.com":
			guest++
			if !strings.Contains(job.HTML, "getUnsubscribe?email=guest%40example.com&amp;token=") {
				t.Fatalf("访客邮件应带退订链接: %s", job.HTML)
			}
		case "admin@qq.com":
			admin++
			if strings.Contains(job.HTML, "getUnsubscribe") {
				t.Fatal("站长邮件不需要退订链接")
			}
		default:
			t.Fatalf("不应发给 %q", job.To)
		}
	}
	if guest != defaultMailDailyLimit {
		t.Fatalf("访客邮箱每天最多 %d 封, 实际 %d", defaultMailDailyLimit, guest)
	}
	if admin != defaultMailDailyLimit+3 {
		t.Fatalf("站长邮箱不受限制, 实际 %d", admin)
	}
}

func TestQueueMailDisabled(t *testing.T) {
	service, sent := newMailTestService(t)
	service.Config.Mail.Enabled = false
	service.queueMail(context.Background(), "guest@example.com", "x", "x")
	if len(sent()) != 0 {
		t.Fatal("未启用邮件时不应发送")
	}
}

func TestBuildMailMessageEncodesHeaders(t *testing.T) {
	raw := string(buildMailMessage("YnsStudy", "site@qq.com", mailJob{To: "a@b.com", Subject: "标题\r\nBcc: evil@x.com", HTML: "<p>你好</p>"}))
	headers, body, _ := strings.Cut(raw, "\r\n\r\n")
	if strings.Contains(headers, "Bcc:") {
		t.Fatalf("标题里的换行不应变成新的邮件头: %s", headers)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(body, "\r\n", ""))
	if err != nil || string(decoded) != "<p>你好</p>" {
		t.Fatalf("正文编码不对: %v %q", err, decoded)
	}
}

func TestNormalizeMailAddressAndWebsite(t *testing.T) {
	for input, want := range map[string]string{
		"User@Example.com": "user@example.com",
		" a@b.cn ":         "a@b.cn",
		"not-an-email":     "",
		"a@b.com\r\nBcc:x": "",
		"Name <a@b.com>":   "",
		"a@b.com, c@d.com": "",
	} {
		if got := normalizeMailAddress(input); got != want {
			t.Errorf("邮箱 %q: 期望 %q, 实际 %q", input, want, got)
		}
	}
	for input, want := range map[string]string{
		"https://ynsstudy.cn":       "https://ynsstudy.cn",
		"www.example.com":           "https://www.example.com",
		"javascript:alert(1)":       "",
		"javascript://x%0aalert(1)": "",
		"data:text/html,<script>":   "",
		"":                          "",
	} {
		if got := normalizeWebsite(input); got != want {
			t.Errorf("网址 %q: 期望 %q, 实际 %q", input, want, got)
		}
	}
}

func TestUnsubscribeRejectsBadToken(t *testing.T) {
	service, _ := newMailTestService(t)
	if _, ok := service.Unsubscribe(context.Background(), "guest@example.com", "wrong"); ok {
		t.Fatal("错误的令牌不应退订成功")
	}
	if service.unsubscribeToken("Guest@Example.com") != service.unsubscribeToken("guest@example.com") {
		t.Fatal("退订令牌应与邮箱大小写无关")
	}
}

func TestRenderReplyMailQuotesOriginal(t *testing.T) {
	body := renderReplyMail("小宋回复了你", nil, "你的评论：", "原来的<评论>", "新的回复", "查看文章", "https://ynsstudy.cn/oneBlog/1")
	original := strings.Index(body, "原来的&lt;评论&gt;")
	reply := strings.Index(body, "新的回复")
	if original < 0 || reply < 0 || original > reply {
		t.Fatalf("应先引用原评论再显示回复，且内容需转义: %s", body)
	}
	if strings.Contains(renderNoticeMail("新评论", nil, "内容", "", ""), "回复内容") {
		t.Fatal("普通评论邮件不应出现原评论区块")
	}
}
