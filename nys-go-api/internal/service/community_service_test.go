package service

import (
	"reflect"
	"testing"

	"nys-go-api/internal/config"
)

func TestNormalizeNoteImages(t *testing.T) {
	s := &Service{Config: &config.Config{Upload: config.UploadConfig{PublicPrefix: "/uploadFile/"}}}
	ok := "/uploadFile/editorImage/a1.webp"
	got := s.normalizeNoteImages([]any{
		ok,
		ok,                                  // 重复
		"https://evil.com/x.png",            // 外站图片
		"/uploadFile/userAvatar/b.png",      // 其他目录
		"/uploadFile/editorImage/../x.png",  // 路径穿越
		"/uploadFile/editorImage/c.png?x=1", // 带参数
		"/uploadFile/editorImage/d.png",
	})
	want := []string{ok, "/uploadFile/editorImage/d.png"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	// JSON 字符串也能识别，最多保留 9 张
	many := `["/uploadFile/editorImage/1.png","/uploadFile/editorImage/2.png","/uploadFile/editorImage/3.png","/uploadFile/editorImage/4.png","/uploadFile/editorImage/5.png","/uploadFile/editorImage/6.png","/uploadFile/editorImage/7.png","/uploadFile/editorImage/8.png","/uploadFile/editorImage/9.png","/uploadFile/editorImage/10.png"]`
	if images := s.normalizeNoteImages(many); len(images) != maxNoteImages {
		t.Fatalf("应最多保留 %d 张，实际 %d", maxNoteImages, len(images))
	}
	if images := s.normalizeNoteImages(nil); len(images) != 0 {
		t.Fatalf("空值应返回空数组: %v", images)
	}
}

func TestNormalizeNoteType(t *testing.T) {
	cases := map[string]string{"": NoteTypeNote, "TIP": NoteTypeTip, "share": NoteTypeShare, "question": NoteTypeQuestion, "hack": NoteTypeNote}
	for input, want := range cases {
		if got := normalizeNoteType(input); got != want {
			t.Errorf("normalizeNoteType(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNoteTitleAndDecode(t *testing.T) {
	if got := NoteTitle(map[string]any{"TITLE": " 标题 ", "TEXT": "<p>正文</p>"}); got != "标题" {
		t.Fatalf("有标题时应使用标题: %q", got)
	}
	if got := NoteTitle(map[string]any{"TEXT": "<p>Go 的 <b>小技巧</b></p>"}); got != "Go 的 小技巧" {
		t.Fatalf("没有标题时应取正文: %q", got)
	}
	if got := NoteTitle(map[string]any{"TEXT": "<p>用 errors.Is 判断</p><pre><code>if errors.Is(err, x) {}</code></pre>"}); got != "用 errors.Is 判断" {
		t.Fatalf("标题不应包含代码块: %q", got)
	}
	if got := NoteTitle(map[string]any{"TEXT": "<h1>Markdown 语法简明教程</h1><p>本文介绍基础语法</p>"}); got != "Markdown 语法简明教程" {
		t.Fatalf("正文以标题开头时应使用该标题: %q", got)
	}
	if got := NoteTitle(map[string]any{"TEXT": `<p><img src="/a.png"></p>`}); got != "一张便签" {
		t.Fatalf("只有图片时应使用默认标题: %q", got)
	}
	rows := decodeNoteRows([]map[string]any{{"IMAGES": `["/uploadFile/editorImage/a.png"]`}, {"TYPE": "tip"}})
	if images, _ := rows[0]["IMAGES"].([]string); len(images) != 1 || rows[0]["TYPE"] != NoteTypeNote {
		t.Fatalf("解码结果不对: %#v", rows[0])
	}
	if images, _ := rows[1]["IMAGES"].([]string); images == nil || len(images) != 0 || rows[1]["TYPE"] != NoteTypeTip {
		t.Fatalf("没有图片时应为空数组: %#v", rows[1])
	}
}

func TestNormalizeTags(t *testing.T) {
	got := normalizeTags([]any{" #Go ", "go", "踩坑", "带 空格", "a,b", "", "这个话题的名字实在是太长了已经超过了二十个字", "Vue", "Nginx", "MySQL", "Redis"})
	want := []string{"Go", "踩坑", "Vue", "Nginx", "MySQL"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got := normalizeTags("Go，前端,  "); !reflect.DeepEqual(got, []string{"Go", "前端"}) {
		t.Fatalf("逗号分隔的字符串解析不对: %v", got)
	}
	// 文章标签有自己的地址 /tag/名字，含 / \ ? 的名字会把地址拆开
	if got := normalizeTags([]any{"CI/CD", `a\b`, "why?", "C++", "Node.js"}); !reflect.DeepEqual(got, []string{"C++", "Node.js"}) {
		t.Fatalf("含路径分隔符的标签应被丢弃: %v", got)
	}
	if NormalizeTags(nil) == nil {
		t.Fatal("NormalizeTags 不应返回 nil，否则修改文章时会被当成“没带标签”")
	}
}

func TestIsNumericID(t *testing.T) {
	cases := map[string]bool{"12": true, "0": true, "": false, "12a": false, "BBA624F7B5BD4546B014AA4335A692C2": false, "-1": false}
	for input, want := range cases {
		if got := isNumericID(input); got != want {
			t.Errorf("isNumericID(%q) = %v, want %v", input, got, want)
		}
	}
}
