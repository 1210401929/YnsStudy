package controller

import (
	"strings"
	"testing"

	"nys-go-api/internal/service"
)

func TestRenderSEOSeriesHTML(t *testing.T) {
	series := map[string]any{"ID": int64(7), "TITLE": "Go 入门<一>", "DESCRIPTION": "", "USERNAME": "余南松", "USERNUM": "101"}
	articles := []map[string]any{
		{"GUID": "21", "BLOG_TITLE": "安装 Go", "EXCERPT_HTML": "<p>下载安装包</p>"},
		{"GUID": "22", "BLOG_TITLE": "第一个程序", "EXCERPT_HTML": "<p>Hello</p>"},
	}
	page := renderSEOSeriesHTML(noteTestShell, series, articles, true, noteTestConfig())
	for _, expected := range []string{
		"<title>Go 入门&lt;一&gt;（系列） - YnsStudy</title>",
		`<link rel="canonical" href="https://ynsstudy.cn/series/7">`,
		`<meta name="robots" content="index,follow,max-image-preview:large,max-snippet:-1">`,
		`"@type":"CreativeWorkSeries"`,
		`"position":2`,
		`"@type":"BreadcrumbList"`,
		`<h1>Go 入门&lt;一&gt;</h1>`,
		`共 2 篇`,
		`<a class="seo-series-start" href="https://ynsstudy.cn/oneBlog/21">从第一篇开始读 →</a>`,
		`<a href="https://ynsstudy.cn/oneBlog/22">第一个程序</a>`,
		`<a href="https://ynsstudy.cn/series">文章系列</a>`,
		`src="/assets/app.js"`,
	} {
		if !strings.Contains(page, expected) {
			t.Errorf("系列页缺少 %q", expected)
		}
	}
	thin := renderSEOSeriesHTML(noteTestShell, series, articles[:1], false, noteTestConfig())
	if !strings.Contains(thin, `<meta name="robots" content="noindex,follow">`) {
		t.Fatal("只有一篇公开文章的系列页应为 noindex")
	}
	if strings.Contains(thin, `rel="canonical"`) || strings.Contains(thin, `og:url`) {
		t.Fatal("noindex 的系列页不应输出规范网址")
	}
	if !strings.Contains(page, `<meta property="og:url" content="https://ynsstudy.cn/series/7">`) {
		t.Fatal("可收录的系列页应输出 og:url")
	}
}

func TestRenderSEOSeriesListHTML(t *testing.T) {
	page := renderSEOSeriesListHTML(noteTestShell, []map[string]any{
		{"ID": int64(7), "TITLE": "Go 入门", "USERNAME": "余南松", "USERNUM": "101", "ARTICLE_COUNT": int64(3), "DESCRIPTION": "从零开始"},
	}, noteTestConfig())
	for _, expected := range []string{
		`<link rel="canonical" href="https://ynsstudy.cn/series">`,
		`<a href="https://ynsstudy.cn/series/7">Go 入门</a>`,
		`共 3 篇 · 从零开始`,
	} {
		if !strings.Contains(page, expected) {
			t.Errorf("全部系列页缺少 %q", expected)
		}
	}
	empty := renderSEOSeriesListHTML(noteTestShell, nil, noteTestConfig())
	if !strings.Contains(empty, `noindex,follow`) || strings.Contains(empty, `rel="canonical"`) {
		t.Fatal("没有系列时应为 noindex，且不输出规范网址")
	}
}

func TestArticleSeriesHTML(t *testing.T) {
	if articleSeriesHTML(nil, "https://ynsstudy.cn") != "" {
		t.Fatal("不属于系列时不输出")
	}
	nav := &service.SeriesNav{ID: "7", Title: "Go 入门", Total: 3, Index: 2,
		Previous: map[string]any{"GUID": "21", "BLOG_TITLE": "安装 Go"},
		Next:     map[string]any{"GUID": "23", "BLOG_TITLE": "变量"}}
	got := articleSeriesHTML(nav, "https://ynsstudy.cn")
	for _, expected := range []string{
		`本文属于系列《<a href="https://ynsstudy.cn/series/7">Go 入门</a>》第 2 / 3 篇`,
		`<a href="https://ynsstudy.cn/oneBlog/21">上一篇：安装 Go</a>`,
		`<a href="https://ynsstudy.cn/oneBlog/23">下一篇：变量</a>`,
	} {
		if !strings.Contains(got, expected) {
			t.Errorf("文章系列信息缺少 %q: %s", expected, got)
		}
	}
	// 私密文章在公开目录里找不到自己的位置时不输出
	if articleSeriesHTML(&service.SeriesNav{ID: "7", Title: "Go 入门", Total: 3}, "https://ynsstudy.cn") != "" {
		t.Fatal("不在公开目录里的文章不应输出系列信息")
	}
}
