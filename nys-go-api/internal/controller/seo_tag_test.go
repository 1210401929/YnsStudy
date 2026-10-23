package controller

import (
	"strings"
	"testing"
	"time"
)

func TestRenderSEOTagHTML(t *testing.T) {
	rows := []map[string]any{{
		"GUID":         "21",
		"BLOG_TITLE":   "Go 错误处理<实践>",
		"EXCERPT_HTML": `<p>用 errors.Is 判断错误</p><script>alert(1)</script>`,
		"USERNAME":     "余南松",
		"USERNUM":      "101",
		"CREATE_TIME":  time.Date(2026, 10, 1, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60)),
	}}
	popular := []map[string]any{{"NAME": "Go"}, {"NAME": "C++"}, {"NAME": "前端"}}
	page := renderSEOTagHTML(noteTestShell, "Go", rows, 25, 2, 2, popular, true, noteTestConfig())
	for _, expected := range []string{
		"<title>Go 相关文章（第 2 页） - YnsStudy</title>",
		`<link rel="canonical" href="https://ynsstudy.cn/tag/Go?page=2">`,
		`<link rel="prev" href="https://ynsstudy.cn/tag/Go">`,
		`<meta name="robots" content="index,follow,max-image-preview:large,max-snippet:-1">`,
		`"@type":"CollectionPage"`,
		`"@type":"BreadcrumbList"`,
		`<h1>#Go</h1>`,
		`共 25 篇公开文章`,
		`<a href="https://ynsstudy.cn/oneBlog/21">Go 错误处理&lt;实践&gt;</a>`,
		`<a href="https://ynsstudy.cn/tag/C++" rel="tag">#C++</a>`,
		`<a href="https://ynsstudy.cn/tag/%E5%89%8D%E7%AB%AF" rel="tag">#前端</a>`,
		`src="/assets/app.js"`,
	} {
		if !strings.Contains(page, expected) {
			t.Errorf("标签页缺少 %q", expected)
		}
	}
	if strings.Contains(page, "<script>alert(1)") {
		t.Fatal("标签页不应输出脚本")
	}
	if strings.Contains(page, `rel="tag">#Go</a>`) {
		t.Fatal("其他标签里不应出现当前标签")
	}

	thin := renderSEOTagHTML(noteTestShell, "Go", rows, 1, 1, 1, nil, false, noteTestConfig())
	if !strings.Contains(thin, `<meta name="robots" content="noindex,follow">`) {
		t.Fatal("文章太少的标签页应为 noindex")
	}
	if strings.Contains(thin, `rel="canonical"`) || strings.Contains(thin, `og:url`) {
		t.Fatal("noindex 的标签页不应输出规范网址，否则两个信号互相矛盾")
	}
	if !strings.Contains(page, `<meta property="og:url" content="https://ynsstudy.cn/tag/Go?page=2">`) {
		t.Fatal("可收录的标签页应输出 og:url")
	}
	if strings.Contains(thin, `rel="next"`) || strings.Contains(thin, "第 1 / 1 页") {
		t.Fatal("只有一页时不应输出分页")
	}
}

func TestTagLinksForArticleHTML(t *testing.T) {
	if tagLinksForArticleHTML(nil, "https://ynsstudy.cn") != "" {
		t.Fatal("没有标签时不输出")
	}
	got := tagLinksForArticleHTML([]string{"Go", "踩坑"}, "https://ynsstudy.cn")
	if !strings.Contains(got, `<a href="https://ynsstudy.cn/tag/Go" rel="tag">#Go</a>`) ||
		!strings.Contains(got, `<a href="https://ynsstudy.cn/tag/%E8%B8%A9%E5%9D%91" rel="tag">#踩坑</a>`) {
		t.Fatalf("文章标签链接不对: %s", got)
	}
}
