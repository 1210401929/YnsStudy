package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

func (h *Controller) registerAIRoutes(router *gin.Engine) {
	group := router.Group("/ai-api/ai")
	group.Any("/qWen", func(c *gin.Context) {
		body, ok := requireBody(c)
		if !ok {
			return
		}
		writeResult(c, h.service.AskAI(c.Request.Context(), stringParam(body, "message")))
	})
	group.GET("/qWenStream", h.streamAI)
}

func (h *Controller) streamAI(c *gin.Context) {
	if _, err := h.service.CurrentUser(c); err != nil {
		c.JSON(http.StatusUnauthorized, model.Failure("用户未登录"))
		return
	}
	encrypted := strings.ReplaceAll(strings.TrimSpace(c.Query("message")), " ", "+")
	encrypted = strings.NewReplacer("\r", "", "\n", "", "\t", "").Replace(encrypted)
	plain, err := h.cipher.Decrypt(encrypted)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Failure("AI 请求内容解密失败"))
		return
	}
	if err := h.service.StreamAI(c, string(plain)); err != nil {
		if !c.Writer.Written() {
			c.JSON(http.StatusBadGateway, model.Failure(err.Error()))
		}
	}
}
