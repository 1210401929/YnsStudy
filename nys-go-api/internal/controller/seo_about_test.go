package controller

import (
	"strings"
	"testing"

	"nys-go-api/internal/service"
)

func TestRenderSEOAboutHTML(t *testing.T) {
	owner := testOwner()
	owner.Avatar = "/uploadFile/userAvatar/a.jpg"
	featured := service.Featured{
		Articles: []map[string]any{{"GUID": "446", "BLOG_TITLE": "博客圈<就是>自娱自乐", "EXCERPT": "我一直觉得个人博客挺矛盾的"}},
		Series:   []map[string]any{{"ID": int64(2), "TITLE": "AI", "ARTICLE_COUNT": int64(3), "DESCRIPTION": "AI 工具的使用记录"}},
	}
	page := renderSEOAboutHTML(noteTestShell, service.AboutContentData(), owner, featured, 430, noteTestConfig())
	for _, expected := range []string{
		"<title>关于本站 - YnsStudy</title>",
		`<link rel="canonical" href="https://ynsstudy.cn/ynsStudy/About">`,
		`"@type":"AboutPage"`,
		`"mainEntity":{"@type":"Person","name":"余南松"`,
		`"image":"https://ynsstudy.cn/uploadFile/userAvatar/a.jpg"`,
		`"@type":"BreadcrumbList"`,
		`<h1>欢迎来到 YnsStudy</h1>`,
		`<p>自 2025-08-23 上线，YnsStudy 已经陪伴了 430 天。</p>`,
		`<h2>站长想说</h2>`,
		`<h2>站长推荐</h2>`,
		`<li>系列《<a href="https://ynsstudy.cn/series/2">AI</a>》，共 3 篇：AI 工具的使用记录</li>`,
		`<li><a href="https://ynsstudy.cn/oneBlog/446">博客圈&lt;就是&gt;自娱自乐</a>：我一直觉得个人博客挺矛盾的</li>`,
		`<h2>常见问题</h2>`,
		`<a href="https://ynsstudy.cn/ynsStudy/Notes">便签墙</a>`,
		`<a href="mailto:yvnansong@qq.com">`,
		`<a href="https://ynsstudy.cn/oneBlog/46140E57DA81405EAB4C48871B7FE790">网站更新日志 →</a>`,
		`src="/assets/app.js"`,
	} {
		if !strings.Contains(page, expected) {
			t.Errorf("关于页缺少 %q", expected)
		}
	}
	if strings.Contains(page, "<h2>建站历程</h2>") {
		t.Fatal("建站历程已去掉，不应再输出")
	}
	if strings.Contains(page, "%!") {
		t.Fatal("关于页格式化参数数量不匹配")
	}
	// 没有配置推荐时不输出这一节
	empty := renderSEOAboutHTML(noteTestShell, service.AboutContentData(), owner, service.Featured{}, 430, noteTestConfig())
	if strings.Contains(empty, "<h2>站长推荐</h2>") {
		t.Fatal("没有推荐内容时不应输出站长推荐")
	}
}
