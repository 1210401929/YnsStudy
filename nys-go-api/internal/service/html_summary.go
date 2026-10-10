package service

import (
	"strings"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"nys-go-api/internal/model"
)

// listExcerptRunes 是列表摘要保留的字数，首页和个人主页只展示两三行。
const listExcerptRunes = 200

// summarizeHTML 从正文 HTML 中提取纯文本摘要和第一张图片地址。
// 块级标签之间补空格，避免相邻段落的文字连在一起；base64 图片不作为封面返回。
func summarizeHTML(content string, maxRunes int) (excerpt, firstImage string) {
	tokenizer := html.NewTokenizer(strings.NewReader(content))
	var builder strings.Builder
	runes := 0
	skipDepth := 0 // 位于 script/style 内部时跳过文字
	needSpace := false

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return strings.TrimSpace(builder.String()), firstImage
		case html.StartTagToken, html.SelfClosingTagToken, html.EndTagToken:
			name, hasAttr := tokenizer.TagName()
			tag := atom.Lookup(name)
			if tag == atom.Script || tag == atom.Style {
				if tokenType == html.StartTagToken {
					skipDepth++
				} else if tokenType == html.EndTagToken && skipDepth > 0 {
					skipDepth--
				}
				continue
			}
			if tag == atom.Img && firstImage == "" && hasAttr {
				for {
					key, value, more := tokenizer.TagAttr()
					if string(key) == "src" {
						src := strings.TrimSpace(string(value))
						if src != "" && !strings.HasPrefix(strings.ToLower(src), "data:") {
							firstImage = src
						}
						break
					}
					if !more {
						break
					}
				}
			}
			if isBlockTag(tag) {
				needSpace = true
			}
		case html.TextToken:
			if skipDepth > 0 || runes >= maxRunes {
				continue
			}
			raw := string(tokenizer.Text())
			text := strings.Join(strings.Fields(raw), " ")
			if text == "" {
				// 只有空白的文字（如 </span> <span> 之间）也算一个分隔
				if raw != "" {
					needSpace = true
				}
				continue
			}
			if strings.TrimLeftFunc(raw, unicode.IsSpace) != raw {
				needSpace = true
			}
			if needSpace && builder.Len() > 0 {
				builder.WriteByte(' ')
				runes++
			}
			needSpace = false
			for _, r := range text {
				if runes >= maxRunes {
					break
				}
				builder.WriteRune(r)
				runes++
			}
			if strings.TrimRightFunc(raw, unicode.IsSpace) != raw {
				needSpace = true
			}
		}
		// 摘要和封面都拿到后不必再解析剩余正文
		if runes >= maxRunes && firstImage != "" {
			return strings.TrimSpace(builder.String()), firstImage
		}
	}
}

func isBlockTag(tag atom.Atom) bool {
	switch tag {
	case atom.P, atom.Div, atom.Br, atom.Li, atom.Ul, atom.Ol, atom.Blockquote, atom.Pre,
		atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6, atom.Tr, atom.Td, atom.Th, atom.Table, atom.Hr:
		return true
	}
	return false
}

// replaceTextWithSummary 把列表数据里的正文字段换成 EXCERPT（纯文本摘要）和 FIRST_IMAGE（封面图），
// 列表接口不再把整篇正文传给前端。
func replaceTextWithSummary(rows []map[string]any, field string) []map[string]any {
	for _, row := range rows {
		if _, ok := row[field]; !ok {
			continue
		}
		excerpt, image := summarizeHTML(model.StringValue(row, field), listExcerptRunes)
		row["EXCERPT"] = excerpt
		row["FIRST_IMAGE"] = image
		delete(row, field)
	}
	return rows
}

// replaceTextWithKeywordSummary 和 replaceTextWithSummary 相同，但搜索时摘要从关键词附近开始截取，
// 关键词在正文后半部分时也能在列表里看到；正文里没有关键词（只匹配了标题）时仍取开头。
func replaceTextWithKeywordSummary(rows []map[string]any, field, keyword string) []map[string]any {
	keywordRunes := []rune(strings.ToLower(strings.TrimSpace(keyword)))
	if len(keywordRunes) == 0 {
		return replaceTextWithSummary(rows, field)
	}
	for _, row := range rows {
		if _, ok := row[field]; !ok {
			continue
		}
		content := model.StringValue(row, field)
		excerpt, image := summarizeHTML(content, listExcerptRunes)
		full := []rune(htmlToSearchText(content))
		if index := runeIndexFold(full, keywordRunes); index > 0 {
			start := index - 40
			prefix := "…"
			if start <= 0 {
				start, prefix = 0, ""
			}
			end := start + listExcerptRunes
			if end > len(full) {
				end = len(full)
			}
			excerpt = prefix + string(full[start:end])
		}
		row["EXCERPT"] = excerpt
		row["FIRST_IMAGE"] = image
		delete(row, field)
	}
	return rows
}

// runeIndexFold 不区分大小写查找子串，返回按字符计算的位置，找不到返回 -1
func runeIndexFold(text, keyword []rune) int {
	lower := []rune(strings.ToLower(string(text)))
	if len(lower) != len(text) {
		// 个别字符转小写后长度变化时，退回逐字比较原文
		lower = text
	}
	for i := 0; i+len(keyword) <= len(lower); i++ {
		match := true
		for j, r := range keyword {
			if lower[i+j] != r {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
