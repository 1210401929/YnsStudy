package config

import (
	"path/filepath"
	"testing"
)

func TestProjectConfigLoads(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "config", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != 8889 {
		t.Fatalf("server.port = %d", cfg.Server.Port)
	}
	if !filepath.IsAbs(cfg.Upload.Directory) {
		t.Fatalf("上传目录没有解析为绝对路径: %s", cfg.Upload.Directory)
	}
	if len(cfg.Security.Whitelist) == 0 || len(cfg.Upload.AllowedExtensions) == 0 {
		t.Fatal("白名单和上传扩展名不能为空")
	}
	if cfg.QQOAuth.AuthorizeURL == "" || cfg.QQOAuth.RedirectURI == "" || cfg.QQOAuth.FrontendOrigin == "" {
		t.Fatal("QQ登录预留地址不能为空")
	}
	if cfg.SEO.SiteName == "" || cfg.SEO.DefaultDescription == "" || cfg.SEO.FrontendIndexFile == "" {
		t.Fatal("SEO 配置不能为空")
	}
}
