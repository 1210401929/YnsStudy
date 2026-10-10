package service

import (
	"strings"
	"testing"
)

func TestSummarizeHTML(t *testing.T) {
	input := `<h2>标题</h2><p>第一段&amp;内容</p><p>第二段</p>` +
		`<p><img src="data:image/png;base64,AAAA"><img src="/uploadFile/editorImage/a.webp"></p>` +
		`<script>alert(1)</script><style>p{}</style><pre><code>x := 1</code></pre>`
	excerpt, image := summarizeHTML(input, 200)
	if excerpt != "标题 第一段&内容 第二段 x := 1" {
		t.Fatalf("摘要不对: %q", excerpt)
	}
	if image != "/uploadFile/editorImage/a.webp" {
		t.Fatalf("封面图不对: %q", image)
	}

	excerpt, _ = summarizeHTML(`<p>`+strings.Repeat("字", 300)+`</p>`, 200)
	if n := len([]rune(excerpt)); n != 200 {
		t.Fatalf("摘要应截断到 200 字, 实际 %d", n)
	}
}

func TestReplaceTextWithSummary(t *testing.T) {
	rows := []map[string]any{{"GUID": "1", "MAINTEXT": `<p>正文</p><img src="/a.png">`}, {"GUID": "2"}}
	replaceTextWithSummary(rows, "MAINTEXT")
	if _, ok := rows[0]["MAINTEXT"]; ok {
		t.Fatal("列表数据不应再带正文")
	}
	if rows[0]["EXCERPT"] != "正文" || rows[0]["FIRST_IMAGE"] != "/a.png" {
		t.Fatalf("摘要字段不对: %v", rows[0])
	}
	if _, ok := rows[1]["EXCERPT"]; ok {
		t.Fatal("没有正文字段的行不应补摘要")
	}
}

func TestSanitizeRichTextDropsDataURIImages(t *testing.T) {
	output := sanitizeRichText(`<p>前<img src="data:image/png;base64,AAAA" alt="截图">后</p>`)
	if strings.Contains(output, "data:") || strings.Contains(output, "<img") {
		t.Fatalf("base64 图片应被整个去掉: %s", output)
	}
	if !strings.Contains(output, "前") || !strings.Contains(output, "后") {
		t.Fatalf("周围文字不应丢失: %s", output)
	}
}

func TestHTMLToSearchText(t *testing.T) {
	text := htmlToSearchText(`<p style="color: red;"><span>Go</span> 并发</p><p>第二段` + strings.Repeat("长", 500) + `</p>`)
	if strings.Contains(text, "span") || strings.Contains(text, "style") || strings.Contains(text, "color") {
		t.Fatalf("搜索文本不应包含标签和属性: %q", text)
	}
	if !strings.HasPrefix(text, "Go 并发 第二段") || len([]rune(text)) != 509 {
		t.Fatalf("搜索文本应保留全部文字: %q", text[:40])
	}
}

func TestLikePattern(t *testing.T) {
	if got := likePattern(" 100%_a\\ "); got != `%100\%\_a\\%` {
		t.Fatalf("关键词转义不对: %s", got)
	}
}
