package service

// 便签墙的链接卡片：抓取网址的标题、简介、配图、网站名和图标。
//
// 这里会由服务器去访问用户填写的网址，必须防止被用来探测服务器所在的内网（SSRF）：
//   - 只允许 http/https，只允许默认端口 80/443，不允许网址里带账号密码；
//   - 连接时检查实际解析出的 IP（在拨号阶段检查，DNS 改指向内网也拦得住），内网、本机、保留地址一律拒绝；
//   - 最多跳转 3 次，每次跳转都重新检查；整个请求 8 秒超时，只读取前 1MB，只解析 HTML。

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"

	"nys-go-api/internal/model"
)

// LinkPreview 是一张链接卡片的内容
type LinkPreview struct {
	URL         string `json:"url"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
	SiteName    string `json:"siteName,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

const (
	linkPreviewMaxBytes    = 1 << 20
	linkPreviewCacheTTL    = 7 * 24 * time.Hour
	linkPreviewUserAgent   = "Mozilla/5.0 (compatible; YnsStudyLinkPreview/1.0; +https://ynsstudy.cn)"
	maxLinkPreviewTitle    = 120
	maxLinkPreviewDescribe = 200
)

var errLinkBlocked = errors.New("这个网址不能生成链接卡片")

// 正文里的链接（Markdown 转出的 HTML，裸网址也会被转成 <a>）
var firstLinkPattern = regexp.MustCompile(`(?i)<a\s[^>]*href="(https?://[^"\s]+)"`)

// cgnat 是运营商级 NAT 的地址段，也属于内网
var cgnat = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

// isPublicIP 只有公网地址才允许访问
func isPublicIP(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() || ip.IsMulticast() {
		return false
	}
	if v4 := ip.To4(); v4 != nil {
		// 0.0.0.0/8、100.64.0.0/10、192.0.0.0/24、198.18.0.0/15（测试网段）、240.0.0.0/4（保留）
		if v4[0] == 0 || cgnat.Contains(v4) || (v4[0] == 192 && v4[1] == 0 && v4[2] == 0) ||
			(v4[0] == 198 && (v4[1] == 18 || v4[1] == 19)) || v4[0] >= 240 {
			return false
		}
	}
	return true
}

// validatePreviewURL 检查网址本身是否允许抓取，返回规范化后的网址
func validatePreviewURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 2048 {
		return nil, errLinkBlocked
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.User != nil {
		return nil, errLinkBlocked
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, errLinkBlocked
	}
	if port := parsed.Port(); port != "" && port != "80" && port != "443" {
		return nil, errLinkBlocked
	}
	// 直接写 IP 的网址在这里先检查一次，域名在连接时检查解析出的 IP
	if ip := net.ParseIP(parsed.Hostname()); ip != nil && !isPublicIP(ip) {
		return nil, errLinkBlocked
	}
	if host := strings.ToLower(parsed.Hostname()); host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return nil, errLinkBlocked
	}
	parsed.Fragment = ""
	return parsed, nil
}

var linkPreviewClient = newLinkPreviewClient()

func newLinkPreviewClient() *http.Client {
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
		// 在真正连接前检查解析出的 IP：防止域名指向内网，或者解析结果在检查之后被改掉（DNS rebinding）
		Control: func(network, address string, _ syscall.RawConn) error {
			host, port, err := net.SplitHostPort(address)
			if err != nil || (port != "80" && port != "443") || !isPublicIP(net.ParseIP(host)) {
				return errLinkBlocked
			}
			return nil
		},
	}
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 6 * time.Second,
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
	}
	return &http.Client{
		Timeout:   8 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return errors.New("跳转次数过多")
			}
			if _, err := validatePreviewURL(req.URL.String()); err != nil {
				return err
			}
			return nil
		},
	}
}

// FetchLinkPreview 抓取网址的链接卡片信息，结果缓存 7 天
func (s *Service) FetchLinkPreview(ctx context.Context, rawURL string) (*LinkPreview, error) {
	target, err := validatePreviewURL(rawURL)
	if err != nil {
		return nil, err
	}
	sum := sha1.Sum([]byte(target.String()))
	cacheKey := "link_preview:" + hex.EncodeToString(sum[:])
	if s.Cache != nil {
		if cached, err := s.Cache.Get(ctx, cacheKey); err == nil && cached != "" {
			var preview LinkPreview
			if json.Unmarshal([]byte(cached), &preview) == nil {
				return &preview, nil
			}
		}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, errLinkBlocked
	}
	request.Header.Set("User-Agent", linkPreviewUserAgent)
	request.Header.Set("Accept", "text/html,application/xhtml+xml")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.6")
	response, err := linkPreviewClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("打开网址失败: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("网址返回 %d", response.StatusCode)
	}
	contentType := response.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "html") {
		return nil, errors.New("不是网页")
	}
	// 按网页声明的编码（例如 GBK）转成 UTF-8
	body, err := charset.NewReader(io.LimitReader(response.Body, linkPreviewMaxBytes), contentType)
	if err != nil {
		return nil, errors.New("无法识别网页编码")
	}
	preview := parseLinkPreview(body, response.Request.URL)
	preview.URL = target.String()

	if s.Cache != nil {
		if encoded, err := json.Marshal(preview); err == nil {
			_ = s.Cache.Set(ctx, cacheKey, string(encoded), linkPreviewCacheTTL)
		}
	}
	return preview, nil
}

// parseLinkPreview 从网页 <head> 里读取 Open Graph、标题、描述和图标
func parseLinkPreview(body io.Reader, pageURL *url.URL) *LinkPreview {
	meta := make(map[string]string)
	var title, icon string
	tokenizer := html.NewTokenizer(body)
	inTitle := false
parse:
	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			break parse
		case html.StartTagToken, html.SelfClosingTagToken:
			name, hasAttr := tokenizer.TagName()
			tag := string(name)
			if tag == "body" {
				break parse
			}
			attrs := map[string]string{}
			for hasAttr {
				var key, value []byte
				key, value, hasAttr = tokenizer.TagAttr()
				attrs[strings.ToLower(string(key))] = string(value)
			}
			switch tag {
			case "title":
				inTitle = true
			case "meta":
				key := strings.ToLower(attrs["property"])
				if key == "" {
					key = strings.ToLower(attrs["name"])
				}
				if key != "" && attrs["content"] != "" && meta[key] == "" {
					meta[key] = attrs["content"]
				}
			case "link":
				rel := strings.ToLower(attrs["rel"])
				if icon == "" && strings.Contains(rel, "icon") && attrs["href"] != "" {
					icon = attrs["href"]
				}
			}
		case html.TextToken:
			if inTitle && title == "" {
				title = string(tokenizer.Text())
			}
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			if string(name) == "title" {
				inTitle = false
			} else if string(name) == "head" {
				break parse
			}
		}
	}

	pick := func(keys ...string) string {
		for _, key := range keys {
			if value := collapseSpaces(meta[key]); value != "" {
				return value
			}
		}
		return ""
	}
	preview := &LinkPreview{
		Title:       truncateText(firstNonEmpty(pick("og:title", "twitter:title"), collapseSpaces(title)), maxLinkPreviewTitle),
		Description: truncateText(pick("og:description", "twitter:description", "description"), maxLinkPreviewDescribe),
		SiteName:    truncateText(pick("og:site_name", "application-name"), 60),
		Image:       safeHTTPSURL(pageURL, pick("og:image", "og:image:url", "twitter:image")),
		Icon:        safeHTTPSURL(pageURL, firstNonEmpty(icon, "/favicon.ico")),
	}
	if preview.SiteName == "" && pageURL != nil {
		preview.SiteName = strings.TrimPrefix(pageURL.Hostname(), "www.")
	}
	return preview
}

// safeHTTPSURL 把相对地址补全；只保留 https 图片（本站是 https，http 图片会被浏览器拦截）
func safeHTTPSURL(base *url.URL, value string) string {
	value = strings.TrimSpace(value)
	if value == "" || base == nil || len(value) > 1024 {
		return ""
	}
	ref, err := url.Parse(value)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(ref)
	if resolved.Scheme == "http" && resolved.Host == base.Host && base.Scheme == "https" {
		resolved.Scheme = "https"
	}
	if resolved.Scheme != "https" || resolved.Host == "" {
		return ""
	}
	return resolved.String()
}

func collapseSpaces(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

// truncateText 见 admin_log_service.go，firstNonEmpty 见 qq_oauth_service.go

// firstLinkInHTML 返回正文里的第一个 http/https 链接
func firstLinkInHTML(text string) string {
	if match := firstLinkPattern.FindStringSubmatch(text); match != nil {
		return html.UnescapeString(match[1])
	}
	return ""
}

// refreshNoteLinkPreview 发布或编辑随记后在后台抓取正文第一个链接的卡片信息并保存；
// 网址没变且已经抓到过就不重复抓。抓取失败时只保存网址和域名，页面照样显示一张简单的卡片。
func (s *Service) refreshNoteLinkPreview(guid, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if !s.tableHasColumn(ctx, "communityInfo", "LINK_PREVIEW") {
		return
	}
	link := firstLinkInHTML(text)
	if link == "" {
		if _, err := s.Repo.Exec(ctx, "UPDATE communityInfo SET LINK_PREVIEW = NULL WHERE GUID = ?", guid); err != nil {
			log.Printf("清除链接卡片失败: %v", err)
		}
		return
	}
	if rows, err := s.Repo.Query(ctx, "SELECT LINK_PREVIEW FROM communityInfo WHERE GUID = ? LIMIT 1", guid); err == nil && len(rows) > 0 {
		var existing LinkPreview
		if json.Unmarshal([]byte(model.StringValue(rows[0], "LINK_PREVIEW")), &existing) == nil && existing.URL == link && existing.Title != "" {
			return
		}
	}
	preview, err := s.FetchLinkPreview(ctx, link)
	if err != nil {
		log.Printf("抓取链接卡片失败 %s: %v", link, err)
		if errors.Is(err, errLinkBlocked) {
			preview = nil
		} else if parsed, parseErr := url.Parse(link); parseErr == nil {
			preview = &LinkPreview{URL: link, SiteName: strings.TrimPrefix(parsed.Hostname(), "www.")}
		}
	}
	var value any
	if preview != nil {
		encoded, _ := json.Marshal(preview)
		value = string(encoded)
	}
	if _, err := s.Repo.Exec(ctx, "UPDATE communityInfo SET LINK_PREVIEW = ? WHERE GUID = ?", value, guid); err != nil {
		log.Printf("保存链接卡片失败: %v", err)
	}
}

// PreviewLink 发布框里实时预览链接卡片。只有登录用户能用，并限制每人每分钟 20 次，防止被拿来批量访问别的网站。
func (s *Service) PreviewLink(c *gin.Context, rawURL string) model.Result {
	actor, failure := s.requireLogin(c)
	if failure != nil {
		return *failure
	}
	ctx := contextOf(c)
	if s.Cache != nil {
		count, err := s.Cache.Increment(ctx, "link_preview_limit:"+actor.Code(), time.Minute)
		if err == nil && count > 20 {
			return model.Failure("预览太频繁了，请稍后再试")
		}
	}
	preview, err := s.FetchLinkPreview(ctx, rawURL)
	if err != nil {
		if errors.Is(err, errLinkBlocked) {
			return model.Failure(errLinkBlocked.Error())
		}
		return model.Failure("暂时无法读取这个网址的信息")
	}
	return model.Success(preview)
}
