package service

import "testing"

func TestNormalizeNewBlogFiltersUnknownFieldsAndMapsCategoryAlias(t *testing.T) {
	input := map[string]any{
		"BLOG_TITLE":  "标题",
		"MAINTEXT":    "正文",
		"CATEGORY_ID": "category-1",
		"UNKNOWN":     "不能进入 SQL",
	}

	blog := normalizeNewBlog(input)
	if blog["CAT_ID"] != "category-1" {
		t.Fatalf("CATEGORY_ID 没有映射为 CAT_ID: %#v", blog)
	}
	if _, exists := blog["CATEGORY_ID"]; exists {
		t.Fatalf("CATEGORY_ID 不应直接进入数据库字段: %#v", blog)
	}
	if _, exists := blog["UNKNOWN"]; exists {
		t.Fatalf("未知字段不应进入数据库 INSERT: %#v", blog)
	}
}
