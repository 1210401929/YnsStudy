package service

import (
	"strings"
	"testing"
)

func TestResourceCategoryCondition(t *testing.T) {
	if condition, args := resourceCategoryCondition("all"); condition != "" || args != nil {
		t.Fatal("全部分类不应加筛选条件")
	}
	condition, args := resourceCategoryCondition("video")
	if !strings.Contains(condition, fileExtExpr+" IN (") || len(args) != 7 || args[0] != "mp4" {
		t.Fatalf("视频分类条件不对: %s %v", condition, args)
	}
	other, otherArgs := resourceCategoryCondition("other")
	total := 0
	for _, item := range resourceCategories {
		total += len(item.Exts)
	}
	if !strings.Contains(other, "NOT LIKE '%.%'") || !strings.Contains(other, "NOT IN (") || len(otherArgs) != total {
		t.Fatalf("其他分类应排除所有已知扩展名，并包含没有扩展名的文件: %s", other)
	}
	if condition, _ := resourceCategoryCondition("'; drop table fileInfo"); condition != "" {
		t.Fatal("未知分类不应拼进 SQL")
	}
	selectExpr, selectArgs := resourceCategorySelect()
	if strings.Count(selectExpr, "?") != len(selectArgs) || !strings.HasSuffix(selectExpr, "ELSE 'other' END") {
		t.Fatalf("分类统计表达式与参数数量不一致: %s", selectExpr)
	}
}

func TestCheckFriendLinkText(t *testing.T) {
	s := uploadTestService(t)
	if failure := s.checkFriendLinkText(map[string]any{"NAME": "  "}); failure == nil {
		t.Fatal("空名称应被拒绝")
	}
	if failure := s.checkFriendLinkText(map[string]any{"REMARK": strings.Repeat("字", maxFriendLinkRemark+1)}); failure == nil {
		t.Fatal("超长描述应被拒绝")
	}
	data := map[string]any{"NAME": "  YnsStudy ", "REMARK": " 一个博客 "}
	if failure := s.checkFriendLinkText(data); failure != nil || data["NAME"] != "YnsStudy" || data["REMARK"] != "一个博客" {
		t.Fatalf("规范化结果不对: %v %v", data, failure)
	}
	// 只改头像时不要求名称
	if failure := s.checkFriendLinkText(map[string]any{"AVATAR": "https://a.cn/a.png"}); failure != nil {
		t.Fatal("没带名称时不应校验名称")
	}
}
