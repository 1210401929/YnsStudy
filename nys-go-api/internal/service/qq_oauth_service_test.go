package service

import (
	"context"
	"net/url"
	"testing"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
)

func TestQQAuthorizationURLStoresState(t *testing.T) {
	store := cache.NewMemoryStore()
	service := &Service{
		Config: &config.Config{QQOAuth: config.QQOAuthConfig{
			Enabled:                true,
			AppID:                  "test-app",
			RedirectURI:            "https://example.com/api/pub-api/login/qq/callback",
			AuthorizeURL:           "https://graph.qq.com/oauth2.0/authorize",
			Scope:                  "get_user_info",
			StateExpirationSeconds: 300,
		}},
		Cache: store,
	}

	authorizationURL, generatedState, err := service.QQAuthorizationURL(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(authorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	state := parsed.Query().Get("state")
	if state == "" {
		t.Fatal("授权地址缺少 state")
	}
	if state != generatedState {
		t.Fatal("返回的 state 与授权地址中的 state 不一致")
	}
	if value, err := store.Get(context.Background(), qqOAuthStatePrefix+state); err != nil || value != "pending" {
		t.Fatalf("state 未写入缓存: value=%q err=%v", value, err)
	}
	if parsed.Query().Get("redirect_uri") != service.Config.QQOAuth.RedirectURI {
		t.Fatalf("redirect_uri 不一致: %s", parsed.Query().Get("redirect_uri"))
	}
}

func TestParseQQAccessTokenSupportsJSONAndQueryString(t *testing.T) {
	jsonResponse, err := parseQQAccessToken([]byte(`{"access_token":"json-token","expires_in":3600}`))
	if err != nil || jsonResponse.AccessToken != "json-token" || int(jsonResponse.ExpiresIn) != 3600 {
		t.Fatalf("解析 JSON Token 失败: %#v, %v", jsonResponse, err)
	}
	stringJSONResponse, err := parseQQAccessToken([]byte(`{"access_token":"string-json-token","expires_in":"7776000"}`))
	if err != nil || stringJSONResponse.AccessToken != "string-json-token" || int(stringJSONResponse.ExpiresIn) != 7776000 {
		t.Fatalf("解析字符串有效期的 JSON Token 失败: %#v, %v", stringJSONResponse, err)
	}
	queryResponse, err := parseQQAccessToken([]byte("access_token=query-token&expires_in=3600&refresh_token=refresh"))
	if err != nil || queryResponse.AccessToken != "query-token" || int(queryResponse.ExpiresIn) != 3600 {
		t.Fatalf("解析 QueryString Token 失败: %#v, %v", queryResponse, err)
	}
}

func TestDecodeQQJSONSupportsJSONP(t *testing.T) {
	var response qqOpenIDResponse
	err := decodeQQJSON([]byte(`callback( {"client_id":"app","openid":"openid-1"} );`), &response)
	if err != nil {
		t.Fatal(err)
	}
	if response.ClientID != "app" || response.OpenID != "openid-1" {
		t.Fatalf("解析 JSONP OpenID 失败: %#v", response)
	}
}

func TestQQAccountCodeDoesNotExposeOpenID(t *testing.T) {
	first := qqAccountCode("app", "secret-openid")
	second := qqAccountCode("app", "secret-openid")
	if first != second {
		t.Fatal("相同 OpenID 应生成稳定账号")
	}
	if first == "secret-openid" || len(first) != len("$userQQ")+32 {
		t.Fatalf("QQ账号格式不正确: %s", first)
	}
}
