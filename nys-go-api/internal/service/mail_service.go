package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"log"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	"nys-go-api/internal/model"
)

// mailSendInterval 是两封邮件之间的间隔，避免触发 SMTP 频率限制
var mailSendInterval = time.Second

// mailJob 是一封待发送的通知邮件
type mailJob struct {
	To      string
	Subject string
	HTML    string
}

const (
	mailQueueSize          = 200
	defaultMailDailyLimit  = 10
	mailUnsubscribeTable   = "mailUnsubscribe"
	mailConnectTimeout     = 15 * time.Second
	mailUnsubscribeSubject = "unsubscribe:"
)

// adminMailAddress 是站长接收通知的邮箱
func (s *Service) adminMailAddress() string {
	if address := strings.TrimSpace(s.Config.Mail.AdminEmail); address != "" {
		return address
	}
	return strings.TrimSpace(s.Config.Mail.Username)
}

// normalizeMailAddress 校验并规范化邮箱地址，不合法时返回空字符串
func normalizeMailAddress(address string) string {
	address = strings.TrimSpace(address)
	if address == "" || len(address) > 191 || strings.ContainsAny(address, "\r\n<>,; ") {
		return ""
	}
	parsed, err := mail.ParseAddress(address)
	if err != nil || parsed.Address != address || !strings.Contains(address, "@") {
		return ""
	}
	return strings.ToLower(address)
}

// queueMail 把邮件放进后台队列，不阻塞当前请求。
// 退订过的邮箱不发；非站长邮箱每天有发送上限，防止有人填别人的邮箱刷邮件。
func (s *Service) queueMail(ctx context.Context, to, subject, body string) {
	if !s.Config.Mail.Enabled {
		return
	}
	to = normalizeMailAddress(to)
	if to == "" {
		return
	}
	isAdmin := strings.EqualFold(to, s.adminMailAddress())
	if !isAdmin {
		if s.isMailUnsubscribed(ctx, to) {
			return
		}
		limit := s.Config.Mail.DailyLimitPerAddress
		if limit <= 0 {
			limit = defaultMailDailyLimit
		}
		count, err := s.Cache.Increment(ctx, "mail:daily:"+to, 24*time.Hour)
		if err != nil || count > int64(limit) {
			return
		}
		body += s.unsubscribeFooter(to)
	}
	s.mailOnce.Do(func() {
		s.mailQueue = make(chan mailJob, mailQueueSize)
		go s.mailWorker()
	})
	select {
	case s.mailQueue <- mailJob{To: to, Subject: subject, HTML: body}:
	default:
		log.Printf("邮件队列已满，丢弃发给 %s 的通知", to)
	}
}

func (s *Service) mailWorker() {
	for job := range s.mailQueue {
		send := s.sendMailFunc
		if send == nil {
			send = s.sendSMTP
		}
		if err := send(job); err != nil {
			log.Printf("发送邮件给 %s 失败: %v", job.To, err)
		}
		time.Sleep(mailSendInterval)
	}
}

// sendSMTP 通过 SMTP 发送一封 HTML 邮件。465 端口使用 SSL 直连，其他端口使用 STARTTLS。
func (s *Service) sendSMTP(job mailJob) error {
	cfg := s.Config.Mail
	address := net.JoinHostPort(cfg.SMTPHost, strconv.Itoa(cfg.SMTPPort))
	dialer := &net.Dialer{Timeout: mailConnectTimeout}
	var client *smtp.Client
	if cfg.SMTPPort == 465 {
		conn, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{ServerName: cfg.SMTPHost})
		if err != nil {
			return err
		}
		client, err = smtp.NewClient(conn, cfg.SMTPHost)
		if err != nil {
			conn.Close()
			return err
		}
	} else {
		conn, err := dialer.Dial("tcp", address)
		if err != nil {
			return err
		}
		client, err = smtp.NewClient(conn, cfg.SMTPHost)
		if err != nil {
			conn.Close()
			return err
		}
		if err := client.StartTLS(&tls.Config{ServerName: cfg.SMTPHost}); err != nil {
			client.Close()
			return err
		}
	}
	defer client.Close()
	if err := client.Auth(smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)); err != nil {
		return err
	}
	if err := client.Mail(cfg.Username); err != nil {
		return err
	}
	if err := client.Rcpt(job.To); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(buildMailMessage(cfg.FromName, cfg.Username, job)); err != nil {
		writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

// buildMailMessage 生成邮件原文。标题和发件人名称用 UTF-8 编码，正文 base64，避免中文乱码和头部注入。
func buildMailMessage(fromName, from string, job mailJob) []byte {
	if strings.TrimSpace(fromName) == "" {
		fromName = "YnsStudy"
	}
	var builder strings.Builder
	builder.WriteString("From: " + mime.BEncoding.Encode("UTF-8", fromName) + " <" + from + ">\r\n")
	builder.WriteString("To: " + job.To + "\r\n")
	builder.WriteString("Subject: " + mime.BEncoding.Encode("UTF-8", job.Subject) + "\r\n")
	builder.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	builder.WriteString("MIME-Version: 1.0\r\n")
	builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	builder.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	encoded := base64.StdEncoding.EncodeToString([]byte(job.HTML))
	for len(encoded) > 76 {
		builder.WriteString(encoded[:76] + "\r\n")
		encoded = encoded[76:]
	}
	builder.WriteString(encoded + "\r\n")
	return []byte(builder.String())
}

/* ---------- 退订 ---------- */

func (s *Service) unsubscribeToken(address string) string {
	mac := hmac.New(sha256.New, []byte(s.Config.Security.JWTSecret))
	mac.Write([]byte(mailUnsubscribeSubject + strings.ToLower(address)))
	return hex.EncodeToString(mac.Sum(nil))[:32]
}

func (s *Service) unsubscribeURL(address string) string {
	base := strings.TrimRight(strings.TrimSpace(s.Config.Mail.APIBaseURL), "/")
	if base == "" {
		base = strings.TrimRight(s.Config.External.DomainName, "/") + "/api"
	}
	return base + "/pub-api/mail/getUnsubscribe?email=" + url.QueryEscape(address) + "&token=" + s.unsubscribeToken(address)
}

func (s *Service) unsubscribeFooter(address string) string {
	return `<p style="margin-top:28px;color:#a39c90;font-size:12px">不想再收到这类邮件？<a style="color:#a39c90" href="` +
		html.EscapeString(s.unsubscribeURL(address)) + `">点此退订</a></p>`
}

func (s *Service) isMailUnsubscribed(ctx context.Context, address string) bool {
	if s.Repo == nil {
		return false
	}
	rows, err := s.Repo.Query(ctx, "SELECT 1 FROM "+mailUnsubscribeTable+" WHERE EMAIL = ? LIMIT 1", address)
	return err == nil && len(rows) > 0
}

// Unsubscribe 处理邮件里的退订链接，返回给用户看的提示文字
func (s *Service) Unsubscribe(ctx context.Context, address, token string) (string, bool) {
	address = normalizeMailAddress(address)
	if address == "" || !hmac.Equal([]byte(token), []byte(s.unsubscribeToken(address))) {
		return "退订链接无效。", false
	}
	if _, err := s.Repo.Exec(ctx, "INSERT IGNORE INTO "+mailUnsubscribeTable+" (EMAIL) VALUES (?)", address); err != nil {
		log.Printf("记录邮件退订失败: %v", err)
		return "退订失败，请稍后再试。", false
	}
	return fmt.Sprintf("已退订，%s 不会再收到 YnsStudy 的通知邮件。", address), true
}

/* ---------- 邮件内容 ---------- */

// mailSiteURL 返回站内页面的完整地址
func (s *Service) mailSiteURL(path string) string {
	return strings.TrimRight(s.Config.External.DomainName, "/") + path
}

// renderNoticeMail 生成通知邮件正文，所有内容都做 HTML 转义
func renderNoticeMail(heading string, lines []string, quote, linkText, link string) string {
	var builder strings.Builder
	builder.WriteString(`<div style="max-width:560px;margin:0 auto;padding:24px;font-family:-apple-system,'PingFang SC','Microsoft YaHei',sans-serif;color:#2b2a27;background:#fffdf8;border:1px solid #e6dfd1">`)
	builder.WriteString(`<h2 style="margin:0 0 16px;font-size:18px;font-weight:normal">` + html.EscapeString(heading) + `</h2>`)
	for _, line := range lines {
		builder.WriteString(`<p style="margin:6px 0;color:#4d4943;font-size:14px">` + html.EscapeString(line) + `</p>`)
	}
	if quote != "" {
		builder.WriteString(`<blockquote style="margin:16px 0;padding:10px 14px;border-left:4px solid #e8c95b;background:#fdf6d8;color:#4d4943;font-size:14px;white-space:pre-wrap">` + html.EscapeString(quote) + `</blockquote>`)
	}
	if link != "" {
		builder.WriteString(`<p style="margin:18px 0 0"><a style="color:#2f5d8a" href="` + html.EscapeString(link) + `">` + html.EscapeString(linkText) + `</a></p>`)
	}
	builder.WriteString(`</div>`)
	return builder.String()
}

// mailExcerpt 截取纯文本，用于邮件里引用评论或文章开头
func mailExcerpt(content string, max int) string {
	text, _ := summarizeHTML(content, max)
	if len([]rune(text)) >= max {
		text += "…"
	}
	return text
}

// userMailAddress 读取登录用户在个人中心填写的邮箱
func (s *Service) userMailAddress(ctx context.Context, userCode string) string {
	userCode = strings.TrimSpace(userCode)
	if userCode == "" || s.Repo == nil {
		return ""
	}
	rows, err := s.Repo.Query(ctx, "SELECT EMAIL FROM userInfo WHERE CODE = ? LIMIT 1", userCode)
	if err != nil || len(rows) == 0 {
		return ""
	}
	return normalizeMailAddress(model.StringValue(rows[0], "EMAIL"))
}

// notifyAdminContent 有人发表公开文章、社区帖子或社区评论时通知站长，方便审核内容。
// 站长自己发的不通知。
func (s *Service) notifyAdminContent(ctx context.Context, authorCode, authorName, kind, title, content, link string) {
	if !s.Config.Mail.Enabled || strings.TrimSpace(authorCode) == s.superAdminCode() {
		return
	}
	adminEmail := s.adminMailAddress()
	if strings.TrimSpace(authorName) == "" {
		authorName = authorCode
	}
	lines := []string{"作者：" + authorName + "（" + authorCode + "）"}
	if title != "" {
		lines = append([]string{"标题：" + title}, lines...)
	}
	subject := "【" + kind + "】" + authorName
	if title != "" {
		subject += "：" + title
	} else {
		subject += "：" + mailExcerpt(content, 30)
	}
	s.queueMail(ctx, adminEmail, subject, renderNoticeMail(kind, lines, mailExcerpt(content, 500), "去看看", link))
}
