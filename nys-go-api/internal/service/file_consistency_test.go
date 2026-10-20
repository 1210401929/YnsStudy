package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindOrphanFilesKeepsUsedAndRecentFiles(t *testing.T) {
	directory := t.TempDir()
	write := func(name string, age time.Duration) {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte("12345"), 0o644); err != nil {
			t.Fatal(err)
		}
		modTime := time.Now().Add(-age)
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatal(err)
		}
	}
	write("used-old.png", 30*24*time.Hour)
	write("orphan-old.png", 30*24*time.Hour)
	write("orphan-new.png", time.Hour) // 可能是正在编写的文章或草稿里的图片
	if err := os.Mkdir(filepath.Join(directory, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	used := map[string]struct{}{"used-old.png": {}}
	orphans, skippedRecent, totalSize, err := findOrphanFiles(directory, used, time.Now().Add(-orphanMinAge))
	if err != nil {
		t.Fatal(err)
	}
	if len(orphans) != 1 || orphans[0].Name != "orphan-old.png" {
		t.Fatalf("只应列出未引用且超过 7 天的文件: %#v", orphans)
	}
	if skippedRecent != 1 {
		t.Fatalf("7 天内上传的未引用文件应跳过并计数，实际 %d", skippedRecent)
	}
	if totalSize != 5 {
		t.Fatalf("待清理文件总大小不正确: %d", totalSize)
	}
	// 只是查找，不能删除任何文件
	for _, name := range []string{"used-old.png", "orphan-old.png", "orphan-new.png"} {
		if _, err := os.Stat(filepath.Join(directory, name)); err != nil {
			t.Fatalf("查找时不应删除文件 %s", name)
		}
	}
}
