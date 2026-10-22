package controller

import (
	"strings"
	"testing"

	"nys-go-api/internal/config"
	"nys-go-api/internal/model"
	"nys-go-api/internal/service"
)

// 没有公开文章的作者主页除了名字之外都一样，会被谷歌当成重复网页：不收录，但链接照常可以跟着走
func TestProfileRobotsDependsOnPublicArticles(t *testing.T) {
	cfg := &config.Config{
		External: config.ExternalConfig{DomainName: "https://ynsstudy.cn"},
		SEO:      config.SEOConfig{SiteName: "YnsStudy"},
	}
	shell := `<!doctype html><html><head><title>YnsStudy</title></head><body><div id="app"></div></body></html>`
	cases := []struct {
		name    string
		profile service.PublicProfile
		want    string
	}{
		{"有公开文章的用户", service.PublicProfile{User: model.User{Code: "u1", Name: "小李", UserNum: 112}, ArticleTotal: 4}, `<meta name="robots" content="index,follow,max-image-preview:large,max-snippet:-1">`},
		{"没有公开文章的用户", service.PublicProfile{User: model.User{Code: "u2", Name: "chia", UserNum: 117}}, `<meta name="robots" content="noindex,follow">`},
		{"站长即使没有公开文章也收录", service.PublicProfile{User: model.User{Code: "yulei", Name: "余南松", UserNum: 101}, IsOwner: true}, `<meta name="robots" content="index,follow,max-image-preview:large,max-snippet:-1">`},
	}
	for _, tc := range cases {
		page := renderSEOProfileHTML(shell, tc.profile, testOwner(), cfg)
		if !strings.Contains(page, tc.want) {
			t.Errorf("%s：缺少 %s", tc.name, tc.want)
		}
		if strings.Count(page, `<meta name="robots"`) != 1 {
			t.Errorf("%s：robots 只能有一个", tc.name)
		}
		// noindex 的页面仍保留指向自己的 canonical
		if !strings.Contains(page, `<link rel="canonical" href="https://ynsstudy.cn/user/`) {
			t.Errorf("%s：缺少 canonical", tc.name)
		}
	}
}
