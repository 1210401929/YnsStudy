package service

import "testing"

func TestIsSafeNoticeLink(t *testing.T) {
	for link, want := range map[string]bool{
		"/oneBlog/12":                    true,
		"https://ynsstudy.cn/oneBlog/12": true,
		"http://127.0.0.1:8080/user/a":   true,
		"javascript:alert(1)":            false,
		" JavaScript:alert(1)":           false,
		"data:text/html,<script>":        false,
		"//evil.example/x":               false,
		"/\\evil.example":                false,
		"vbscript:msgbox":                false,
	} {
		if got := isSafeNoticeLink(link); got != want {
			t.Errorf("%q: 期望 %v, 实际 %v", link, want, got)
		}
	}
}
