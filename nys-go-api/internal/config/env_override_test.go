package config

import "testing"

func TestEnvOverridesDatabaseAndRedisAddress(t *testing.T) {
	// 没有环境变量时使用配置文件里的值（本地开发连远程）
	cfg, err := Load("../../config/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	fileDatabaseHost, fileRedisHost := cfg.Database.Host, cfg.Redis.Host

	t.Setenv("NYS_DATABASE_HOST", "127.0.0.1")
	t.Setenv("NYS_REDIS_HOST", " 127.0.0.1 ")
	t.Setenv("NYS_DATABASE_PORT", "3307")
	cfg, err = Load("../../config/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.Host != "127.0.0.1" || cfg.Redis.Host != "127.0.0.1" || cfg.Database.Port != 3307 {
		t.Fatalf("环境变量没有生效: db=%s:%d redis=%s", cfg.Database.Host, cfg.Database.Port, cfg.Redis.Host)
	}
	if fileDatabaseHost == "127.0.0.1" || fileRedisHost == "127.0.0.1" {
		t.Fatal("测试前提不成立：配置文件里本来就是 127.0.0.1")
	}

	t.Setenv("NYS_REDIS_PORT", "abc")
	if _, err := Load("../../config/config.yaml"); err == nil {
		t.Fatal("无效端口应该报错，而不是悄悄用默认值")
	}
}
