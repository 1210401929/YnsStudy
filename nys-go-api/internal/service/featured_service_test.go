package service

import (
	"reflect"
	"testing"
	"time"
)

func TestSiteDays(t *testing.T) {
	loc := time.Local
	for _, c := range []struct {
		now  time.Time
		want int
	}{
		{time.Date(2025, 8, 23, 0, 1, 0, 0, loc), 1},   // 上线当天是第 1 天
		{time.Date(2025, 8, 24, 23, 59, 0, 0, loc), 2}, // 一天中的时间不影响
		{time.Date(2026, 8, 23, 12, 0, 0, 0, loc), 366},
		{time.Date(2025, 8, 1, 12, 0, 0, 0, loc), 0}, // 上线之前
	} {
		if got := SiteDays(c.now); got != c.want {
			t.Errorf("%s: 期望 %d, 实际 %d", c.now.Format("2006-01-02 15:04"), c.want, got)
		}
	}
}

func TestUniqueIDs(t *testing.T) {
	got := uniqueIDs([]string{" 446 ", "", "446", "21", "7", "9"}, 3)
	if !reflect.DeepEqual(got, []string{"446", "21", "7"}) {
		t.Fatalf("去重、去空、限制数量后不对: %v", got)
	}
}
