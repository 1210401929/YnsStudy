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
	if cfg.Database.Location != "Asia/Shanghai" {
		t.Fatalf("database.location = %q，Docker 部署必须显式使用 Asia/Shanghai", cfg.Database.Location)
	}
	const oneHour = 60 * 60
	if cfg.Security.JWTExpirationSeconds != oneHour || cfg.Security.SessionExpirationSecond != oneHour {
		t.Fatalf("登录有效期必须统一为 1 小时: jwt=%d session=%d", cfg.Security.JWTExpirationSeconds, cfg.Security.SessionExpirationSecond)
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
