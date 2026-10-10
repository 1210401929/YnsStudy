package controller

import (
	"html"
	"net/http"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

func (h *Controller) registerPublicRoutes(router *gin.Engine) {
	api := router.Group("/pub-api/api")
	api.Any("/getCurrentCity", func(c *gin.Context) { writeResult(c, h.service.CurrentCity(c)) })
	api.Any("/getClientIpAddress", func(c *gin.Context) { writeResult(c, h.service.ClientIPAddress(c)) })

	// 通知邮件底部的退订链接（get 开头的接口不需要登录）
	router.GET("/pub-api/mail/getUnsubscribe", h.unsubscribeMail)

	login := router.Group("/pub-api/login")
	login.Any("/sendPhoneCode", h.sendPhoneCode)
	login.Any("/loginByPhoneCode", h.loginByPhoneCode)
	login.GET("/qq/authorize", h.qqAuthorize)
	login.GET("/qq/callback", h.qqCallback)
	login.Any("/logout", func(c *gin.Context) { writeResult(c, h.service.Logout(c)) })
	login.Any("/loginUser", h.loginUser)
	login.Any("/checkUserLogin", func(c *gin.Context) { writeResult(c, h.service.CheckUserLogin(c)) })
	login.Any("/changePassWord", h.changePassword)
	login.Any("/register", h.register)
	login.Any("/changeUserInfo", h.changeUserInfo)
	login.Any("/deleteUserAvatarFile", h.deleteUserAvatar)
	login.Any("/getUserInfoByCode", h.getUserByCode)
	login.Any("/getUserInfoByNum", h.getUserByNum)
	login.Any("/getUserInfoByName", h.getUserByName)
	login.Any("/getAllUserInfo", h.getAllUsers)
	login.Any("/operationUser", h.operationUser)

	notice := router.Group("/pub-api/notice")
	notice.Any("/addNotice", h.addNotice)
	notice.Any("/getNotice", h.getNotice)
	notice.Any("/readNotice", h.readNotice)
	notice.Any("/allReadNotice", h.readAllNotices)

	// 旧系统的 /pub-api/sql/**（任意 SQL）和 /pub-api/upload/deleteFileByUrl(s)（按地址删除任意文件）
	// 已移除：前端没有调用，且任何登录用户都能借此读写整个数据库或删除他人文件。
	upload := router.Group("/pub-api/upload")
	upload.POST("/uploadFile", h.uploadFile)
}

func (h *Controller) sendPhoneCode(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.SendPhoneCode(c, stringParam(body, "phone")))
}

func (h *Controller) loginByPhoneCode(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.LoginByPhoneCode(c, stringParam(body, "phone"), stringParam(body, "code")))
}

func (h *Controller) loginUser(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.Login(c, stringParam(body, "userName"), stringParam(body, "passWord")))
}

func (h *Controller) changePassword(c *gin.Context) {
	body, _ := readBody(c)
	value := func(key string) string {
		if result := stringParam(body, key); result != "" {
			return result
		}
		return c.Query(key)
	}
	writeResult(c, h.service.ChangePassword(c, value("oldPassWord"), value("newPassWord"), value("confirmPassword")))
}

func (h *Controller) register(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.Register(c.Request.Context(), stringParam(body, "userName"), stringParam(body, "userCode"), stringParam(body, "passWord"), stringParam(body, "successPassWord")))
}

func (h *Controller) changeUserInfo(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.ChangeUserInfo(c, nestedMap(body, "userInfo")))
}

func (h *Controller) deleteUserAvatar(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.DeleteUserAvatar(c, stringParam(body, "userCode")))
}

func (h *Controller) getUserByCode(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.GetUserByCode(c.Request.Context(), stringParam(body, "userCode")))
}

func (h *Controller) getUserByNum(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.GetUserByNum(c.Request.Context(), stringParam(body, "userNum")))
}

func (h *Controller) getUserByName(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.GetUsersByName(c.Request.Context(), stringParam(body, "userName")))
}

func (h *Controller) getAllUsers(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.GetAllUsers(c, intParam(body, "page", 1), intParam(body, "pageSize", 10), stringParam(body, "keyword")))
}

func (h *Controller) operationUser(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.OperationUser(c, stringParam(body, "userId"), stringParam(body, "type")))
}

func (h *Controller) addNotice(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.AddNotice(c, stringParam(body, "receiverUserCode"), stringParam(body, "type"), stringParam(body, "execute"), stringParam(body, "remark")))
}

// getNotice 只处理当前登录用户自己的通知，请求中的账号参数不再使用。
func (h *Controller) getNotice(c *gin.Context) {
	writeResult(c, h.service.GetNotices(c))
}

func (h *Controller) readNotice(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.ReadNotice(c, stringParam(body, "guid")))
}

// readAllNotices 只处理当前登录用户自己的通知，请求中的账号参数不再使用。
func (h *Controller) readAllNotices(c *gin.Context) {
	writeResult(c, h.service.ReadAllNotices(c))
}

func (h *Controller) uploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		writeResult(c, model.Failure("上传文件不能为空"))
		return
	}
	writeResult(c, h.service.SaveUploadedFile(file, c.PostForm("spliceUrl")))
}

func (h *Controller) unsubscribeMail(c *gin.Context) {
	message, _ := h.service.Unsubscribe(c.Request.Context(), c.Query("email"), c.Query("token"))
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8">`+
		`<meta name="viewport" content="width=device-width,initial-scale=1"><meta name="robots" content="noindex"><title>邮件退订 - YnsStudy</title></head>`+
		`<body style="margin:0;background:#efe8da;font-family:-apple-system,'PingFang SC','Microsoft YaHei',sans-serif;color:#2b2a27">`+
		`<div style="max-width:480px;margin:18vh auto;padding:32px 24px;background:#fffdf8;border:1px solid #e6dfd1;text-align:center">`+
		`<p style="font-size:16px;line-height:1.8">`+html.EscapeString(message)+`</p>`+
		`<a style="color:#2f5d8a" href="/">返回首页</a></div></body></html>`))
}
