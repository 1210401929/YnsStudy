package database

import (
	"testing"
	"time"
)

func TestResolveLocationUsesShanghaiOffset(t *testing.T) {
	location, err := ResolveLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	_, offset := time.Date(2026, 8, 18, 17, 30, 0, 0, location).Zone()
	if offset != 8*60*60 {
		t.Fatalf("Asia/Shanghai offset = %d, want %d", offset, 8*60*60)
	}
}
