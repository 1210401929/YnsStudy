package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
	"nys-go-api/internal/service"
)

func (h *Controller) registerLuluRoutes(group *gin.RouterGroup) {
	group.POST("/status", h.petStatus)
	group.POST("/feed", h.feedPet)
	group.POST("/play", h.playPet)
	group.POST("/care", h.smartCarePet)
	group.POST("/visit", h.visitPet)
	group.POST("/sleep", h.sleepPet)
	group.POST("/messages", h.petMessages)
	group.POST("/message/add", h.addPetMessage)
	group.POST("/message/delete", h.deletePetMessage)
	group.POST("/logs", h.petLogs)
	group.POST("/companionship/monthly", h.petMonthlyCompanionship)
	group.POST("/fun-state", h.petFunState)
	group.POST("/mission/advance", h.advancePetMission)
	group.POST("/clothes/change", h.changePetClothes)
}

func (h *Controller) luluBody(c *gin.Context) (map[string]any, int64, bool) {
	body, ok := requireBody(c)
	if !ok {
		return nil, 0, false
	}
	userNum, err := int64Param(body, "userNum")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Failure(err.Error()))
		return nil, 0, false
	}
	return body, userNum, true
}

func (h *Controller) petStatus(c *gin.Context) {
	_, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	value, err := h.service.GetPetStatus(c.Request.Context(), userNum)
	h.writeLulu(c, value, err)
}

func (h *Controller) feedPet(c *gin.Context) {
	_, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	value, err := h.service.FeedPet(c.Request.Context(), userNum, service.ClientIP(c), c.GetHeader("User-Agent"))
	h.writeLulu(c, value, err)
}

func (h *Controller) playPet(c *gin.Context) {
	body, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	value, err := h.service.PlayPet(c.Request.Context(), userNum, stringParam(body, "actionName"), service.ClientIP(c), c.GetHeader("User-Agent"))
	h.writeLulu(c, value, err)
}

func (h *Controller) smartCarePet(c *gin.Context) {
	_, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	value, err := h.service.SmartCarePet(c.Request.Context(), userNum, service.ClientIP(c), c.GetHeader("User-Agent"))
	h.writeLulu(c, value, err)
}

func (h *Controller) visitPet(c *gin.Context) {
	_, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	value, err := h.service.RecordPetVisit(c.Request.Context(), userNum, service.ClientIP(c), c.GetHeader("User-Agent"))
	h.writeLulu(c, value, err)
}

func (h *Controller) sleepPet(c *gin.Context) {
	_, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	value, err := h.service.TogglePetSleep(c.Request.Context(), userNum, service.ClientIP(c), c.GetHeader("User-Agent"))
	h.writeLulu(c, value, err)
}

func (h *Controller) petMessages(c *gin.Context) {
	body, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	value, err := h.service.GetPetMessages(c.Request.Context(), userNum, intParam(body, "page", 1), intParam(body, "pageSize", 8))
	h.writeLulu(c, value, err)
}

func (h *Controller) addPetMessage(c *gin.Context) {
	body, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.AddPetMessage(c.Request.Context(), userNum, stringParam(body, "content"), service.ClientIP(c)))
}

func (h *Controller) deletePetMessage(c *gin.Context) {
	body, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	messageID, err := int64Param(body, "messageId")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Failure(err.Error()))
		return
	}
	value, err := h.service.DeletePetMessage(c.Request.Context(), userNum, messageID)
	h.writeLulu(c, value, err)
}

func (h *Controller) petLogs(c *gin.Context) {
	body, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	value, err := h.service.GetPetLogs(c.Request.Context(), userNum, intParam(body, "page", 1), intParam(body, "pageSize", 10))
	h.writeLulu(c, value, err)
}

func (h *Controller) petMonthlyCompanionship(c *gin.Context) {
	_, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	value, err := h.service.GetMonthlyCompanionship(c.Request.Context(), userNum)
	h.writeLulu(c, value, err)
}

func (h *Controller) petFunState(c *gin.Context) {
	_, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	value, err := h.service.GetPetFunState(c.Request.Context(), userNum)
	h.writeLulu(c, value, err)
}

func (h *Controller) advancePetMission(c *gin.Context) {
	body, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	value, err := h.service.AdvancePetMission(c.Request.Context(), userNum, stringParam(body, "missionType"), intParam(body, "amount", 1))
	h.writeLulu(c, value, err)
}

func (h *Controller) changePetClothes(c *gin.Context) {
	_, userNum, ok := h.luluBody(c)
	if !ok {
		return
	}
	value, err := h.service.ChangePetClothes(c.Request.Context(), userNum)
	h.writeLulu(c, value, err)
}

func (h *Controller) writeLulu(c *gin.Context, value any, err error) {
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Failure(err.Error()))
		return
	}
	c.JSON(http.StatusOK, value)
}
