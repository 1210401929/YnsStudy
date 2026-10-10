package service

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

// richTextPolicy 是文章正文、社区帖子等富文本保存前的白名单。
// 保留 wangEditor 和 Markdown 正常会产生的标签、样式和属性，
// 去掉 <script>、onerror 等事件属性和 javascript: 链接，防止存储型 XSS。
var richTextPolicy = newRichTextPolicy()

func newRichTextPolicy() *bluemonday.Policy {
	policy := bluemonday.UGCPolicy()

	// wangEditor 的待办、分割线、图片链接等依赖 data-* 属性
	policy.AllowDataAttributes()
	// 不放行 base64 图片：前端提交前会把它们上传成文件，避免整张图片写进数据库

	// 编辑器通过行内样式设置颜色、对齐、行高、字号和图片尺寸
	policy.AllowStyles(
		"color", "background-color", "text-align", "line-height", "text-indent",
		"font-size", "font-weight", "font-style", "text-decoration",
		"width", "height", "max-width", "vertical-align", "white-space",
		"margin-left", "padding-left", "border", "border-collapse",
	).Globally()
	// 字体名可能是中文或带引号（如 "Microsoft YaHei"），只放行字母、数字、空格、逗号、引号和连字符
	policy.AllowStyles("font-family").Matching(regexp.MustCompile(`^[\p{L}\p{N} ,'"\-]{1,100}$`)).Globally()

	// 站内链接保持可被搜索引擎跟踪，只给外部链接加 nofollow
	policy.RequireNoFollowOnLinks(false)
	policy.RequireNoFollowOnFullyQualifiedLinks(true)

	// 代码块语言（language-go 等）
	policy.AllowAttrs("class").Matching(regexp.MustCompile(`^[\w\- ]{1,100}$`)).OnElements("code", "pre", "span")
	// 待办事项的勾选框，只允许 checkbox，避免伪造输入框
	policy.AllowAttrs("type").Matching(regexp.MustCompile(`^checkbox$`)).OnElements("input")
	policy.AllowAttrs("checked", "disabled").OnElements("input")
	// 图片尺寸和新窗口打开的链接
	policy.AllowAttrs("width", "height").OnElements("img", "td", "th")
	policy.AllowAttrs("target").Matching(regexp.MustCompile(`^_blank$`)).OnElements("a")

	return policy
}

// imageTagPattern 匹配过滤后的 <img> 标签；bluemonday 输出的属性值已转义，不会包含 ">"
var imageTagPattern = regexp.MustCompile(`(?i)<img\b[^>]*>`)
var imageSrcAttrPattern = regexp.MustCompile(`(?i)\ssrc="`)

// sanitizeRichText 过滤用户提交的富文本 HTML。
func sanitizeRichText(html string) string {
	cleaned := richTextPolicy.Sanitize(html)
	// src 被过滤掉（如 base64 图片）后剩下的空 <img> 会显示成坏图，一并去掉
	return imageTagPattern.ReplaceAllStringFunc(cleaned, func(tag string) string {
		if imageSrcAttrPattern.MatchString(tag) {
			return tag
		}
		return ""
	})
}
