package service

import (
	"strings"
	"testing"
)

func TestSanitizeRichTextRemovesScripts(t *testing.T) {
	cases := map[string]string{
		"事件属性":          `<img src="x" onerror="alert(1)">`,
		"script 标签":     `<p>正文</p><script>alert(1)</script>`,
		"javascript 链接": `<a href="javascript:alert(1)">点我</a>`,
		"svg onload":    `<svg onload="alert(1)"></svg>`,
		"iframe":        `<iframe src="https://evil.example"></iframe>`,
		"style 里的 url":  `<p style="background-image: url(javascript:alert(1))">x</p>`,
		"文本输入框":         `<input type="text" name="password">`,
	}
	for name, input := range cases {
		output := strings.ToLower(sanitizeRichText(input))
		for _, bad := range []string{"onerror", "onload", "<script", "javascript:", "<iframe", "url(", `type="text"`} {
			if strings.Contains(output, bad) {
				t.Fatalf("%s: 过滤后仍包含 %q: %s", name, bad, output)
			}
		}
	}
}

func TestSanitizeRichTextKeepsEditorFormatting(t *testing.T) {
	input := `<h2>标题</h2>` +
		`<p style="text-align: center; line-height: 1.8;"><span style="color: rgb(225, 60, 57);">红字</span></p>` +
		`<pre><code class="language-go">fmt.Println("hi")</code></pre>` +
		`<blockquote>引用</blockquote>` +
		`<div data-w-e-type="todo"><input type="checkbox" disabled checked>待办</div>` +
		`<img src="/uploadFile/editorImage/a.png" alt="图" style="width: 50%;" data-href="https://example.com">` +
		`<table><tr><th width="100">表头</th></tr><tr><td colspan="2">单元格</td></tr></table>` +
		`<a href="https://example.com" target="_blank">链接</a>` +
		`<span style="font-family: 黑体;">黑体</span><span style="font-family: &quot;Microsoft YaHei&quot;;">雅黑</span>` +
		`<a href="/oneBlog/12">站内链接</a>`
	output := sanitizeRichText(input)
	for _, want := range []string{
		"<h2>标题</h2>", "text-align: center", "line-height: 1.8", "color: rgb(225, 60, 57)",
		`class="language-go"`, "<blockquote>", `data-w-e-type="todo"`, `type="checkbox"`, "checked",
		`src="/uploadFile/editorImage/a.png"`, "width: 50%", `data-href="https://example.com"`,
		`colspan="2"`, `width="100"`, `href="https://example.com"`, `target="_blank"`,
		"font-family: 黑体", "Microsoft YaHei", `<a href="/oneBlog/12">站内链接</a>`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("过滤后丢失了 %q\n输出: %s", want, output)
		}
	}
}
