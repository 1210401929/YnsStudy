package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nys-go-api/internal/config"
)

func TestSubmitIndexNowSendsPayload(t *testing.T) {
	var received indexNowPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("IndexNow 必须使用 POST，得到 %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("请求体不是合法 JSON: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	s := &Service{
		Config: &config.Config{
			External: config.ExternalConfig{DomainName: "https://ynsstudy.cn/"},
			SEO:      config.SEOConfig{IndexNowKey: "abcdef123456", IndexNowEndpoint: server.URL},
		},
		HTTP: server.Client(),
	}
	if err := s.submitIndexNow(context.Background(), []string{"/oneBlog/442", "archive"}); err != nil {
		t.Fatal(err)
	}
	if received.Host != "ynsstudy.cn" || received.Key != "abcdef123456" {
		t.Fatalf("host 或 key 不正确: %#v", received)
	}
	if received.KeyLocation != "https://ynsstudy.cn/indexnow.txt" {
		t.Fatalf("keyLocation 不正确: %s", received.KeyLocation)
	}
	if len(received.URLList) != 2 || received.URLList[0] != "https://ynsstudy.cn/oneBlog/442" || received.URLList[1] != "https://ynsstudy.cn/archive" {
		t.Fatalf("urlList 不正确: %#v", received.URLList)
	}
}

func TestSubmitIndexNowReportsRejectedKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("key not valid"))
	}))
	defer server.Close()

	s := &Service{
		Config: &config.Config{
			External: config.ExternalConfig{DomainName: "https://ynsstudy.cn"},
			SEO:      config.SEOConfig{IndexNowKey: "abcdef123456", IndexNowEndpoint: server.URL},
		},
		HTTP: server.Client(),
	}
	if err := s.submitIndexNow(context.Background(), []string{"/oneBlog/1"}); err == nil {
		t.Fatal("搜索引擎拒绝时应返回错误")
	}
}

func TestIndexNowDisabledWithoutKey(t *testing.T) {
	s := &Service{Config: &config.Config{}}
	if s.IndexNowEnabled() {
		t.Fatal("未配置密钥时不应启用 IndexNow")
	}
	// 未启用时直接返回，不能启动后台请求。
	s.SubmitIndexNow("/oneBlog/1")
}
