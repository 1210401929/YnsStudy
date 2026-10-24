package controller

import (
	"strings"
	"testing"

	"nys-go-api/internal/service"
)

func TestRenderSEOAboutHTML(t *testing.T) {
	owner := testOwner()
	owner.Avatar = "/uploadFile/userAvatar/a.jpg"
	page := renderSEOAboutHTML(noteTestShell, service.AboutContentData(), owner, noteTestConfig())
	for _, expected := range []string{
		"<title>关于本站 - YnsStudy</title>",
		`<link rel="canonical" href="https://ynsstudy.cn/ynsStudy/About">`,
		`"@type":"AboutPage"`,
		`"mainEntity":{"@type":"Person","name":"余南松"`,
		`"image":"https://ynsstudy.cn/uploadFile/userAvatar/a.jpg"`,
		`"@type":"BreadcrumbList"`,
		`<h1>欢迎来到 YnsStudy</h1>`,
		`<h2>站长想说</h2>`,
		`<h2>建站历程</h2>`,
		`<li><time>2025-08-23</time> 网站 1.0 上线：`,
		`<a href="https://ynsstudy.cn/oneBlog/46140E57DA81405EAB4C48871B7FE790">查看完整更新日志 →</a>`,
		`<h2>常见问题</h2>`,
		`<a href="https://ynsstudy.cn/ynsStudy/Notes">便签墙</a>`,
		`<a href="mailto:yvnansong@qq.com">`,
		`src="/assets/app.js"`,
	} {
		if !strings.Contains(page, expected) {
			t.Errorf("关于页缺少 %q", expected)
		}
	}
	if strings.Contains(page, "%!") {
		t.Fatal("关于页格式化参数数量不匹配")
	}
}
