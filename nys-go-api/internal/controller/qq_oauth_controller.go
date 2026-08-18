package controller

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

const qqOAuthStateCookieName = "nys_qq_oauth_state"

type qqLoginMessage struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Payload any    `json:"payload,omitempty"`
}

func (h *Controller) qqAuthorize(c *gin.Context) {
	authorizationURL, state, err := h.service.QQAuthorizationURL(c.Request.Context())
	if err != nil {
		h.renderQQLoginResult(c, model.Failure(err.Error()))
		return
	}
	h.setQQStateCookie(c, state, h.service.Config.QQOAuth.StateExpirationSeconds)
	c.Redirect(http.StatusFound, authorizationURL)
}

func (h *Controller) qqCallback(c *gin.Context) {
	state := strings.TrimSpace(c.Query("state"))
	cookieState, cookieErr := c.Cookie(qqOAuthStateCookieName)
	h.setQQStateCookie(c, "", -1)
	if cookieErr != nil || state == "" || cookieState != state {
		h.renderQQLoginResult(c, model.Failure("QQ登录请求与当前浏览器不匹配，请关闭窗口后重新登录"))
		return
	}
	if oauthError := strings.TrimSpace(c.Query("error")); oauthError != "" {
		description := strings.TrimSpace(c.Query("error_description"))
		if description == "" {
			description = oauthError
		}
		h.renderQQLoginResult(c, model.Failure("QQ授权未完成: "+description))
		return
	}
	result := h.service.CompleteQQLogin(c, c.Query("code"), state)
	h.renderQQLoginResult(c, result)
}

func (h *Controller) setQQStateCookie(c *gin.Context, value string, maxAge int) {
	redirectURI, _ := url.Parse(h.service.Config.QQOAuth.RedirectURI)
	secure := redirectURI != nil && strings.EqualFold(redirectURI.Scheme, "https")
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(qqOAuthStateCookieName, value, maxAge, "/", "", secure, true)
}

func (h *Controller) renderQQLoginResult(c *gin.Context, result model.Result) {
	message := qqLoginMessage{
		Type:    "nys:qq-login",
		Success: !result.IsError,
		Message: result.ErrMsg,
		Payload: result.Result,
	}
	if message.Success {
		message.Message = "QQ登录成功"
	}
	messageJSON, _ := json.Marshal(message)
	targetOriginJSON, _ := json.Marshal(h.service.Config.QQOAuth.FrontendOrigin)

	title := "QQ登录成功"
	statusClass := "success"
	visibleMessage := "授权完成，正在返回网站…"
	if result.IsError {
		title = "QQ登录失败"
		statusClass = "error"
		visibleMessage = result.ErrMsg
	}

	c.Header("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; script-src 'unsafe-inline'")
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>%s</title>
  <style>
    *{box-sizing:border-box}body{margin:0;min-height:100vh;display:grid;place-items:center;background:linear-gradient(145deg,#eef8ff,#f7fbff);font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC",sans-serif;color:#26384a}.card{width:min(360px,calc(100%% - 32px));padding:34px 28px;text-align:center;background:#fff;border:1px solid rgba(18,183,245,.18);border-radius:18px;box-shadow:0 18px 50px rgba(30,96,130,.14)}.mark{width:62px;height:62px;margin:0 auto 18px;border-radius:20px;display:grid;place-items:center;background:#12b7f5;color:#fff;font-size:24px;font-weight:800;box-shadow:0 10px 24px rgba(18,183,245,.26)}h1{margin:0 0 10px;font-size:21px}.message{margin:0;color:#647586;line-height:1.7;word-break:break-word}.error .mark{background:#f56c6c;box-shadow:0 10px 24px rgba(245,108,108,.24)}button{margin-top:22px;border:0;border-radius:10px;padding:10px 24px;background:#eef3f7;color:#435466;cursor:pointer}
  </style>
</head>
<body>
  <main class="card %s">
    <div class="mark">QQ</div>
    <h1>%s</h1>
    <p class="message">%s</p>
    <button type="button" onclick="window.close()">关闭窗口</button>
  </main>
  <script>
    const loginMessage = %s;
    const targetOrigin = %s;
    if (window.opener && !window.opener.closed && targetOrigin) {
      window.opener.postMessage(loginMessage, targetOrigin);
      if (loginMessage.success) window.setTimeout(() => window.close(), 500);
    }
  </script>
</body>
</html>`, html.EscapeString(title), statusClass, html.EscapeString(title), html.EscapeString(visibleMessage), messageJSON, targetOriginJSON)))
}
