package service

import "testing"

func TestNormalizeSeriesID(t *testing.T) {
	for input, want := range map[string]string{
		"7": "7", " 12 ": "12", "": "", "0": "", "-3": "", "abc": "", "7;drop": "",
	} {
		if got := normalizeSeriesID(input); got != want {
			t.Errorf("%q: 期望 %q, 实际 %q", input, want, got)
		}
	}
}

func TestSeriesText(t *testing.T) {
	s := uploadTestService(t)
	if _, _, failure := s.seriesText("  ", ""); failure == nil {
		t.Fatal("空名称应被拒绝")
	}
	long := ""
	for i := 0; i < maxSeriesTitle+1; i++ {
		long += "字"
	}
	if _, _, failure := s.seriesText(long, ""); failure == nil {
		t.Fatal("超长名称应被拒绝")
	}
	title, description, failure := s.seriesText("  Go   入门  ", "  从零开始 ")
	if failure != nil || title != "Go 入门" || description != "从零开始" {
		t.Fatalf("规范化结果不对: %q %q %v", title, description, failure)
	}
}
