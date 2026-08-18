package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
	"nys-go-api/internal/security"
)

const qqOAuthStatePrefix = "oauth:qq:state:"

type qqAccessTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Error            int    `json:"error"`
	ErrorDescription string `json:"error_description"`
	Code             int    `json:"code"`
	Message          string `json:"msg"`
}

type qqOpenIDResponse struct {
	ClientID         string `json:"client_id"`
	OpenID           string `json:"openid"`
	Error            int    `json:"error"`
	ErrorDescription string `json:"error_description"`
	Code             int    `json:"code"`
	Message          string `json:"msg"`
}

type qqUserInfoResponse struct {
	Ret          int    `json:"ret"`
	Message      string `json:"msg"`
	Nickname     string `json:"nickname"`
	FigureURL    string `json:"figureurl"`
	FigureURL1   string `json:"figureurl_1"`
	FigureURL2   string `json:"figureurl_2"`
	FigureURLQQ1 string `json:"figureurl_qq_1"`
	FigureURLQQ2 string `json:"figureurl_qq_2"`
}

func (s *Service) QQAuthorizationURL(ctx context.Context) (authorizationURL, state string, err error) {
	cfg := s.Config.QQOAuth
	if !cfg.Enabled {
		return "", "", fmt.Errorf("QQ登录尚未启用，请先完善 config.yaml 中的 qq_oauth 配置")
	}

	state, err = randomURLSafeValue(32)
	if err != nil {
		return "", "", fmt.Errorf("生成QQ登录状态码: %w", err)
	}
	ttl := time.Duration(cfg.StateExpirationSeconds) * time.Second
	if err := s.Cache.Set(ctx, qqOAuthStatePrefix+state, "pending", ttl); err != nil {
		return "", "", fmt.Errorf("保存QQ登录状态码: %w", err)
	}

	endpoint, err := url.Parse(cfg.AuthorizeURL)
	if err != nil {
		return "", "", fmt.Errorf("QQ授权地址配置错误: %w", err)
	}
	query := endpoint.Query()
	query.Set("response_type", "code")
	query.Set("client_id", cfg.AppID)
	query.Set("redirect_uri", cfg.RedirectURI)
	query.Set("state", state)
	if strings.TrimSpace(cfg.Scope) != "" {
		query.Set("scope", cfg.Scope)
	}
	endpoint.RawQuery = query.Encode()
	return endpoint.String(), state, nil
}

func (s *Service) CompleteQQLogin(c *gin.Context, code, state string) model.Result {
	if !s.Config.QQOAuth.Enabled {
		return model.Failure("QQ登录尚未启用")
	}
	code = strings.TrimSpace(code)
	state = strings.TrimSpace(state)
	if code == "" || state == "" {
		return model.Failure("QQ登录回调缺少 code 或 state")
	}

	ctx := contextOf(c)
	stateValue, err := s.Cache.Get(ctx, qqOAuthStatePrefix+state)
	if err != nil || stateValue != "pending" {
		return model.Failure("QQ登录请求已失效，请关闭窗口后重新登录")
	}
	_ = s.Cache.Delete(ctx, qqOAuthStatePrefix+state)

	oauthContext, cancel := context.WithTimeout(ctx, time.Duration(s.Config.QQOAuth.RequestTimeoutSeconds)*time.Second)
	defer cancel()
	accessToken, err := s.exchangeQQAccessToken(oauthContext, code)
	if err != nil {
		return model.Failure("获取QQ授权凭证失败: " + err.Error())
	}
	openID, err := s.fetchQQOpenID(oauthContext, accessToken)
	if err != nil {
		return model.Failure("获取QQ用户标识失败: " + err.Error())
	}
	profile, err := s.fetchQQUserInfo(oauthContext, accessToken, openID)
	if err != nil {
		return model.Failure("获取QQ用户信息失败: " + err.Error())
	}

	user, err := s.getOrCreateQQUser(ctx, openID, profile)
	if err != nil {
		return model.Failure("创建QQ登录账号失败: " + err.Error())
	}
	return s.completeLogin(c, user)
}

func (s *Service) exchangeQQAccessToken(ctx context.Context, code string) (string, error) {
	cfg := s.Config.QQOAuth
	endpoint, err := url.Parse(cfg.TokenURL)
	if err != nil {
		return "", err
	}
	query := endpoint.Query()
	query.Set("grant_type", "authorization_code")
	query.Set("client_id", cfg.AppID)
	query.Set("client_secret", cfg.AppKey)
	query.Set("code", code)
	query.Set("redirect_uri", cfg.RedirectURI)
	query.Set("fmt", "json")
	endpoint.RawQuery = query.Encode()

	body, err := s.qqGET(ctx, endpoint.String())
	if err != nil {
		return "", err
	}
	response, err := parseQQAccessToken(body)
	if err != nil {
		return "", err
	}
	if response.AccessToken == "" {
		return "", qqResponseError(response.Error, response.Code, response.ErrorDescription, response.Message)
	}
	return response.AccessToken, nil
}

func (s *Service) fetchQQOpenID(ctx context.Context, accessToken string) (string, error) {
	endpoint, err := url.Parse(s.Config.QQOAuth.OpenIDURL)
	if err != nil {
		return "", err
	}
	query := endpoint.Query()
	query.Set("access_token", accessToken)
	query.Set("fmt", "json")
	endpoint.RawQuery = query.Encode()

	body, err := s.qqGET(ctx, endpoint.String())
	if err != nil {
		return "", err
	}
	var response qqOpenIDResponse
	if err := decodeQQJSON(body, &response); err != nil {
		return "", err
	}
	if response.OpenID == "" {
		return "", qqResponseError(response.Error, response.Code, response.ErrorDescription, response.Message)
	}
	if response.ClientID != "" && response.ClientID != s.Config.QQOAuth.AppID {
		return "", fmt.Errorf("QQ返回的 client_id 与当前应用不一致")
	}
	return response.OpenID, nil
}

func (s *Service) fetchQQUserInfo(ctx context.Context, accessToken, openID string) (qqUserInfoResponse, error) {
	endpoint, err := url.Parse(s.Config.QQOAuth.UserInfoURL)
	if err != nil {
		return qqUserInfoResponse{}, err
	}
	query := endpoint.Query()
	query.Set("access_token", accessToken)
	query.Set("oauth_consumer_key", s.Config.QQOAuth.AppID)
	query.Set("openid", openID)
	query.Set("format", "json")
	endpoint.RawQuery = query.Encode()

	body, err := s.qqGET(ctx, endpoint.String())
	if err != nil {
		return qqUserInfoResponse{}, err
	}
	var response qqUserInfoResponse
	if err := decodeQQJSON(body, &response); err != nil {
		return qqUserInfoResponse{}, err
	}
	if response.Ret != 0 {
		return qqUserInfoResponse{}, qqResponseError(response.Ret, 0, "", response.Message)
	}
	return response, nil
}

func (s *Service) qqGET(ctx context.Context, endpoint string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	response, err := s.HTTP.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("QQ接口返回 %s", response.Status)
	}
	return body, nil
}

func (s *Service) getOrCreateQQUser(ctx context.Context, openID string, profile qqUserInfoResponse) (model.User, error) {
	userCode := qqAccountCode(s.Config.QQOAuth.AppID, openID)
	rows, err := s.Repo.Query(ctx, "SELECT * FROM userInfo WHERE CODE = ? LIMIT 1", userCode)
	if err != nil {
		return model.User{}, err
	}
	if len(rows) > 0 {
		return model.UserFromRow(rows[0]), nil
	}

	guid, err := randomHexValue(16)
	if err != nil {
		return model.User{}, err
	}
	randomPassword, err := randomURLSafeValue(32)
	if err != nil {
		return model.User{}, err
	}
	salt, err := security.NewSalt()
	if err != nil {
		return model.User{}, err
	}
	hashedPassword, err := security.HashPassword(randomPassword, salt)
	if err != nil {
		return model.User{}, err
	}

	name := truncateQQNickname(strings.TrimSpace(profile.Nickname), 50)
	if name == "" {
		name = "QQ用户"
	}
	avatar := firstNonEmpty(profile.FigureURLQQ2, profile.FigureURLQQ1, profile.FigureURL2, profile.FigureURL1, profile.FigureURL)
	_, insertErr := s.Repo.Exec(ctx, `INSERT INTO userInfo
(GUID, CODE, NAME, PASSWORD, PASSWORDSALT, ROLE, AVATAR, REMARK)
VALUES (?, ?, ?, ?, ?, '1', ?, ?)`, guid, userCode, name, hashedPassword, salt, avatar, "通过QQ登录创建")
	if insertErr != nil {
		// 两个回调同时到达时，另一个请求可能已经完成创建。
		rows, queryErr := s.Repo.Query(ctx, "SELECT * FROM userInfo WHERE CODE = ? LIMIT 1", userCode)
		if queryErr != nil || len(rows) == 0 {
			return model.User{}, insertErr
		}
		return model.UserFromRow(rows[0]), nil
	}

	rows, err = s.Repo.Query(ctx, "SELECT * FROM userInfo WHERE CODE = ? LIMIT 1", userCode)
	if err != nil || len(rows) == 0 {
		if err == nil {
			err = fmt.Errorf("QQ账号创建后未查询到用户")
		}
		return model.User{}, err
	}
	return model.UserFromRow(rows[0]), nil
}

func parseQQAccessToken(body []byte) (qqAccessTokenResponse, error) {
	var response qqAccessTokenResponse
	if decodeQQJSON(body, &response) == nil && (response.AccessToken != "" || response.Error != 0 || response.Code != 0) {
		return response, nil
	}
	values, err := url.ParseQuery(strings.TrimSpace(string(body)))
	if err != nil {
		return response, fmt.Errorf("解析QQ Access Token响应: %w", err)
	}
	response.AccessToken = values.Get("access_token")
	response.RefreshToken = values.Get("refresh_token")
	response.ExpiresIn, _ = strconv.Atoi(values.Get("expires_in"))
	response.Error, _ = strconv.Atoi(values.Get("error"))
	response.ErrorDescription = values.Get("error_description")
	if response.AccessToken == "" && response.Error == 0 {
		return response, fmt.Errorf("QQ Access Token响应格式异常")
	}
	return response, nil
}

func decodeQQJSON(body []byte, target any) error {
	value := strings.TrimSpace(string(body))
	if start, end := strings.IndexByte(value, '{'), strings.LastIndexByte(value, '}'); start >= 0 && end >= start {
		value = value[start : end+1]
	}
	if err := json.Unmarshal([]byte(value), target); err != nil {
		return fmt.Errorf("解析QQ接口响应: %w", err)
	}
	return nil
}

func qqResponseError(errorCode, code int, description, message string) error {
	detail := firstNonEmpty(strings.TrimSpace(description), strings.TrimSpace(message), "QQ接口未返回详细错误")
	if errorCode != 0 {
		return fmt.Errorf("%s (error=%d)", detail, errorCode)
	}
	if code != 0 {
		return fmt.Errorf("%s (code=%d)", detail, code)
	}
	return fmt.Errorf("%s", detail)
}

func qqAccountCode(appID, openID string) string {
	sum := sha256.Sum256([]byte(appID + "\x00" + openID))
	return "$userQQ" + hex.EncodeToString(sum[:16])
}

func randomURLSafeValue(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func randomHexValue(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func truncateQQNickname(value string, maximum int) string {
	runes := []rune(value)
	if len(runes) <= maximum {
		return value
	}
	return string(runes[:maximum])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
