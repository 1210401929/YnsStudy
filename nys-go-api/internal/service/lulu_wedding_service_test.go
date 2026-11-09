package service

import "testing"

func TestLuluWeddingProgress(t *testing.T) {
	tests := []struct {
		level, currentChapter, next, chapterCount int
		bless, married                            bool
	}{
		{59, 0, 60, 0, false, false},
		{60, 60, 65, 1, false, false},
		{64, 60, 65, 1, false, false},
		{65, 65, 66, 2, false, false},
		{66, 66, 67, 3, false, false},
		{67, 67, 68, 4, false, false},
		{68, 68, 69, 5, true, false},
		{69, 69, 70, 6, true, false},
		{70, 70, 0, 7, true, true},
		{78, 70, 0, 7, true, true},
	}
	for _, test := range tests {
		state := luluWeddingProgress(test.level)
		chapter := 0
		if len(state.Chapters) > 0 {
			chapter = state.Chapters[len(state.Chapters)-1].Level
		}
		if chapter != test.currentChapter || len(state.Chapters) != test.chapterCount || state.NextLevel != test.next || state.CanBless != test.bless || state.Married != test.married {
			t.Errorf("unexpected wedding state at Lv.%d: %+v", test.level, state)
		}
	}
}

func TestLuluWeddingVowsAreOnlyDeliveredAfterUnlock(t *testing.T) {
	if got := luluWeddingProgress(69).Vows; len(got) != 0 {
		t.Fatalf("vows leaked before the ceremony: %+v", got)
	}
	got := luluWeddingProgress(70).Vows
	if len(got) != 2 || got[0].ID != "lulu" || got[1].ID != "lumei" || got[0].Text == "" || got[1].Text == "" || got[0].Text == got[1].Text {
		t.Fatalf("expected two distinct written vows: %+v", got)
	}
}
