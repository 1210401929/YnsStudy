package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"nys-go-api/internal/config"
)

// IndexNowKeyPath 是公开的密钥文件地址。放在站点根目录，表示该密钥对全站 URL 有效。
const IndexNowKeyPath = "/indexnow.txt"

const indexNowTimeout = 15 * time.Second

type indexNowPayload struct {
	Host        string   `json:"host"`
	Key         string   `json:"key"`
	KeyLocation string   `json:"keyLocation"`
	URLList     []string `json:"urlList"`
}

// IndexNowEnabled 表示是否配置了 IndexNow 密钥。
func (s *Service) IndexNowEnabled() bool {
	return strings.TrimSpace(s.Config.SEO.IndexNowKey) != ""
}

// NotifyBlogChanged 在公开文章发布、修改或删除后通知搜索引擎重新抓取。
// 删除的文章同样需要提交，搜索引擎会据此发现 404 并尽快移出索引。
func (s *Service) NotifyBlogChanged(blogID string) {
	blogID = strings.TrimSpace(blogID)
	if blogID == "" {
		return
	}
	s.SubmitIndexNow("/oneBlog/"+url.PathEscape(blogID), "/archive")
}

// SubmitIndexNow 在后台提交站内路径，不阻塞发布文章等用户请求；失败只记录日志。
func (s *Service) SubmitIndexNow(paths ...string) {
	if !s.IndexNowEnabled() || len(paths) == 0 {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), indexNowTimeout)
		defer cancel()
		if err := s.submitIndexNow(ctx, paths); err != nil {
			log.Printf("IndexNow 推送失败 %v: %v", paths, err)
		}
	}()
}

func (s *Service) submitIndexNow(ctx context.Context, paths []string) error {
	domain := strings.TrimRight(s.Config.External.DomainName, "/")
	site, err := url.Parse(domain)
	if err != nil || site.Host == "" {
		return fmt.Errorf("external.domain_name 无效: %q", domain)
	}
	urls := make([]string, 0, len(paths))
	for _, path := range paths {
		urls = append(urls, domain+"/"+strings.TrimLeft(path, "/"))
	}
	body, err := json.Marshal(indexNowPayload{
		Host:        site.Host,
		Key:         strings.TrimSpace(s.Config.SEO.IndexNowKey),
		KeyLocation: domain + IndexNowKeyPath,
		URLList:     urls,
	})
	if err != nil {
		return err
	}
	endpoint := strings.TrimSpace(s.Config.SEO.IndexNowEndpoint)
	if endpoint == "" {
		endpoint = config.DefaultIndexNowEndpoint
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	client := s.HTTP
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	// 200 表示已接收，202 表示已接收但密钥仍在校验中，两者都算成功。
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusAccepted {
		detail, _ := io.ReadAll(io.LimitReader(response.Body, 512))
		return fmt.Errorf("HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(detail)))
	}
	return nil
}
