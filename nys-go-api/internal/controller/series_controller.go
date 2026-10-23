package controller

import (
	"github.com/gin-gonic/gin"
)

// registerSeriesRoutes 注册文章系列接口。get 开头的接口不需要登录（getMySeries 在服务层要求登录），
// 新建、修改、删除、排序需要登录，权限在服务层校验。
func (h *Controller) registerSeriesRoutes(group *gin.RouterGroup) {
	group.Any("/getSeriesList", func(c *gin.Context) {
		body, _ := readBody(c)
		writeResult(c, h.service.GetSeriesList(c.Request.Context(), stringParam(body, "userCode"), stringParam(body, "sort"), intParam(body, "limit", 50)))
	})
	group.Any("/getSeries", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetSeries(c, stringParam(body, "id")))
		}
	})
	group.Any("/getMySeries", func(c *gin.Context) { writeResult(c, h.service.GetMySeries(c)) })
	group.Any("/addSeries", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.AddSeries(c, stringParam(body, "title"), stringParam(body, "description")))
		}
	})
	group.Any("/updateSeries", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.UpdateSeries(c, stringParam(body, "id"), stringParam(body, "title"), stringParam(body, "description")))
		}
	})
	group.Any("/deleteSeries", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.DeleteSeries(c, stringParam(body, "id")))
		}
	})
	group.Any("/sortSeries", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.SortSeries(c, stringParam(body, "id"), stringSlice(body["guids"])))
		}
	})
}
