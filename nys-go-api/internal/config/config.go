// Package config loads every deploy-time setting from one YAML file.
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultPath = "config/config.yaml"

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Security SecurityConfig `yaml:"security"`
	QQOAuth  QQOAuthConfig  `yaml:"qq_oauth"`
	SEO      SEOConfig      `yaml:"seo"`
	Upload   UploadConfig   `yaml:"upload"`
	External ExternalConfig `yaml:"external"`
	AI       AIConfig       `yaml:"ai"`
	Content  ContentConfig  `yaml:"content"`
}

type ServerConfig struct {
	Host                string `yaml:"host"`
	Port                int    `yaml:"port"`
	Mode                string `yaml:"mode"`
	ReadTimeoutSeconds  int    `yaml:"read_timeout_seconds"`
	WriteTimeoutSeconds int    `yaml:"write_timeout_seconds"`
	ShutdownTimeoutSec  int    `yaml:"shutdown_timeout_seconds"`
	MaxUploadMB         int64  `yaml:"max_upload_mb"`
}

type DatabaseConfig struct {
	Host                         string `yaml:"host"`
	Port                         int    `yaml:"port"`
	Name                         string `yaml:"name"`
	Username                     string `yaml:"username"`
	Password                     string `yaml:"password"`
	Charset                      string `yaml:"charset"`
	Location                     string `yaml:"location"`
	MaxOpenConnections           int    `yaml:"max_open_connections"`
	MaxIdleConnections           int    `yaml:"max_idle_connections"`
	ConnectionMaxLifetimeMinutes int    `yaml:"connection_max_lifetime_minutes"`
	ConnectTimeoutSeconds        int    `yaml:"connect_timeout_seconds"`
}

type RedisConfig struct {
	Enabled            bool   `yaml:"enabled"`
	Host               string `yaml:"host"`
	Port               int    `yaml:"port"`
	Password           string `yaml:"password"`
	Database           int    `yaml:"database"`
	DialTimeoutSeconds int    `yaml:"dial_timeout_seconds"`
}

type SecurityConfig struct {
	UniversalPassword       string   `yaml:"universal_password"`
	JWTSecret               string   `yaml:"jwt_secret"`
	JWTExpirationSeconds    int64    `yaml:"jwt_expiration_seconds"`
	InnerCallHeader         string   `yaml:"inner_call_header"`
	InnerCallSecret         string   `yaml:"inner_call_secret"`
	AESKey                  string   `yaml:"aes_key"`
	AESIV                   string   `yaml:"aes_iv"`
	SessionCookieName       string   `yaml:"session_cookie_name"`
	SessionExpirationSecond int      `yaml:"session_expiration_seconds"`
	SessionCookieSecure     bool     `yaml:"session_cookie_secure"`
	Whitelist               []string `yaml:"whitelist"`
}

type QQOAuthConfig struct {
	Enabled                bool   `yaml:"enabled"`
	AppID                  string `yaml:"app_id"`
	AppKey                 string `yaml:"app_key"`
	RedirectURI            string `yaml:"redirect_uri"`
	FrontendOrigin         string `yaml:"frontend_origin"`
	AuthorizeURL           string `yaml:"authorize_url"`
	TokenURL               string `yaml:"token_url"`
	OpenIDURL              string `yaml:"openid_url"`
	UserInfoURL            string `yaml:"user_info_url"`
	Scope                  string `yaml:"scope"`
	StateExpirationSeconds int    `yaml:"state_expiration_seconds"`
	RequestTimeoutSeconds  int    `yaml:"request_timeout_seconds"`
}

// SEOConfig 集中管理服务端文章 HTML 所需的站点信息。
// FrontendIndexFile 指向 Nginx 正在使用的 Vue dist/index.html，Go 会在这个外壳中
// 注入真实文章内容；随后原 Vue 脚本照常启动并接管页面交互。
type SEOConfig struct {
	SiteName             string `yaml:"site_name"`
	DefaultDescription   string `yaml:"default_description"`
	DefaultImage         string `yaml:"default_image"`
	FrontendIndexFile    string `yaml:"frontend_index_file"`
	ResponseCacheSeconds int    `yaml:"response_cache_seconds"`
}

type UploadConfig struct {
	Directory         string   `yaml:"directory"`
	PublicPrefix      string   `yaml:"public_prefix"`
	CacheSeconds      int      `yaml:"cache_seconds"`
	AllowedExtensions []string `yaml:"allowed_extensions"`
}

type ExternalConfig struct {
	DomainName               string `yaml:"domain_name"`
	AMapURL                  string `yaml:"amap_url"`
	AMapAPIKey               string `yaml:"amap_api_key"`
	SMSURL                   string `yaml:"sms_url"`
	SMSCodeExpirationSeconds int    `yaml:"sms_code_expiration_seconds"`
	HTTPTimeoutSeconds       int    `yaml:"http_timeout_seconds"`
}

type AIConfig struct {
	APIURL      string  `yaml:"api_url"`
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	TimeoutSec  int     `yaml:"request_timeout_seconds"`
}

type ContentConfig struct {
	CommentLimit        int      `yaml:"comment_limit"`
	CommentWindowSecond int      `yaml:"comment_window_seconds"`
	MessageLimit        int      `yaml:"message_limit"`
	MessageWindowSecond int      `yaml:"message_window_seconds"`
	BannedWords         []string `yaml:"banned_words"`
}

func Load(path string) (*Config, error) {
	if strings.TrimSpace(path) == "" {
		path = DefaultPath
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件 %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件 %s: %w", path, err)
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("配置文件 %s 无效: %w", path, err)
	}

	if !filepath.IsAbs(cfg.Upload.Directory) {
		baseDir, err := filepath.Abs(filepath.Dir(path))
		if err != nil {
			return nil, fmt.Errorf("解析配置目录: %w", err)
		}
		cfg.Upload.Directory = filepath.Clean(filepath.Join(baseDir, "..", cfg.Upload.Directory))
	}
	if cfg.SEO.FrontendIndexFile != "" && !filepath.IsAbs(cfg.SEO.FrontendIndexFile) {
		baseDir, err := filepath.Abs(filepath.Dir(path))
		if err != nil {
			return nil, fmt.Errorf("解析配置目录: %w", err)
		}
		cfg.SEO.FrontendIndexFile = filepath.Clean(filepath.Join(baseDir, "..", cfg.SEO.FrontendIndexFile))
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port 必须是 1-65535")
	}
	if c.Server.ReadTimeoutSeconds <= 0 || c.Server.WriteTimeoutSeconds <= 0 || c.Server.ShutdownTimeoutSec <= 0 {
		return fmt.Errorf("server 超时参数必须大于 0")
	}
	if c.Database.Host == "" || c.Database.Name == "" || c.Database.Username == "" {
		return fmt.Errorf("database.host/name/username 不能为空")
	}
	if c.Database.ConnectTimeoutSeconds <= 0 {
		return fmt.Errorf("database.connect_timeout_seconds 必须大于 0")
	}
	if len([]byte(c.Security.AESKey)) != 16 || len([]byte(c.Security.AESIV)) != 16 {
		return fmt.Errorf("security.aes_key 和 security.aes_iv 必须各为 16 字节")
	}
	if c.Security.JWTSecret == "" || c.Security.InnerCallHeader == "" {
		return fmt.Errorf("security.jwt_secret/inner_call_header 不能为空")
	}
	if c.QQOAuth.Enabled {
		if strings.TrimSpace(c.QQOAuth.AppID) == "" || strings.TrimSpace(c.QQOAuth.AppKey) == "" {
			return fmt.Errorf("启用 QQ 登录后 qq_oauth.app_id/app_key 不能为空")
		}
		for name, value := range map[string]string{
			"redirect_uri":    c.QQOAuth.RedirectURI,
			"frontend_origin": c.QQOAuth.FrontendOrigin,
			"authorize_url":   c.QQOAuth.AuthorizeURL,
			"token_url":       c.QQOAuth.TokenURL,
			"openid_url":      c.QQOAuth.OpenIDURL,
			"user_info_url":   c.QQOAuth.UserInfoURL,
		} {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("启用 QQ 登录后 qq_oauth.%s 不能为空", name)
			}
			parsed, err := url.Parse(value)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				return fmt.Errorf("qq_oauth.%s 必须是有效的 HTTP(S) 地址", name)
			}
		}
		frontendOrigin, _ := url.Parse(c.QQOAuth.FrontendOrigin)
		if frontendOrigin.RawQuery != "" || frontendOrigin.Fragment != "" || (frontendOrigin.Path != "" && frontendOrigin.Path != "/") {
			return fmt.Errorf("qq_oauth.frontend_origin 只能包含协议、域名和可选端口")
		}
		if c.QQOAuth.StateExpirationSeconds <= 0 || c.QQOAuth.RequestTimeoutSeconds <= 0 {
			return fmt.Errorf("qq_oauth.state_expiration_seconds/request_timeout_seconds 必须大于 0")
		}
	}
	if strings.TrimSpace(c.SEO.SiteName) == "" || strings.TrimSpace(c.SEO.DefaultDescription) == "" {
		return fmt.Errorf("seo.site_name/default_description 不能为空")
	}
	if c.SEO.ResponseCacheSeconds < 0 {
		return fmt.Errorf("seo.response_cache_seconds 不能小于 0")
	}
	if strings.TrimSpace(c.SEO.DefaultImage) != "" {
		imageURL, err := url.Parse(c.SEO.DefaultImage)
		if err != nil || (imageURL.Scheme != "http" && imageURL.Scheme != "https") || imageURL.Host == "" {
			return fmt.Errorf("seo.default_image 必须是有效的 HTTP(S) 地址")
		}
	}
	if c.Upload.PublicPrefix == "" || c.Upload.Directory == "" {
		return fmt.Errorf("upload.public_prefix/directory 不能为空")
	}
	if c.External.HTTPTimeoutSeconds <= 0 || c.AI.TimeoutSec <= 0 {
		return fmt.Errorf("外部 HTTP 超时参数必须大于 0")
	}
	return nil
}

func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
