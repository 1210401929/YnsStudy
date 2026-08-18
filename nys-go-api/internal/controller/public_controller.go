package controller

import (
	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

func (h *Controller) registerPublicRoutes(router *gin.Engine) {
	api := router.Group("/pub-api/api")
	api.Any("/getCurrentCity", func(c *gin.Context) { writeResult(c, h.service.CurrentCity(c)) })
	api.Any("/getClientIpAddress", func(c *gin.Context) { writeResult(c, h.service.ClientIPAddress(c)) })

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

	sql := router.Group("/pub-api/sql")
	sql.Any("/selectList", h.selectList)
	sql.Any("/selectListByParams", h.selectList)
	sql.Any("/deleteBySql", h.executeSQL)
	sql.Any("/updateBySql", h.executeSQL)
	sql.Any("/exeSql", h.executeSQL)
	sql.Any("/exeSqlByParams", h.executeSQL)
	sql.Any("/exeSqlListByParams", h.executeSQLList)
	sql.Any("/exeSqlComposite", h.executeSQLComposite)
	sql.Any("/saveAllTableData", h.saveAll)
	sql.Any("/saveAllTableDataByParams", h.saveAll)

	upload := router.Group("/pub-api/upload")
	upload.POST("/uploadFile", h.uploadFile)
	upload.POST("/deleteFileByUrl", h.deleteFile)
	upload.POST("/deleteFileByUrls", h.deleteFiles)
}

func (h *Controller) sendPhoneCode(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.SendPhoneCode(c.Request.Context(), stringParam(body, "phone")))
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
	writeResult(c, h.service.GetAllUsers(c.Request.Context(), intParam(body, "page", 1), intParam(body, "pageSize", 10), stringParam(body, "keyword")))
}

func (h *Controller) operationUser(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.OperationUser(c.Request.Context(), stringParam(body, "userId"), stringParam(body, "type")))
}

func (h *Controller) addNotice(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.AddNotice(c.Request.Context(), stringParam(body, "sendUserCode"), stringParam(body, "receiverUserCode"), stringParam(body, "type"), stringParam(body, "execute"), stringParam(body, "remark")))
}

func (h *Controller) getNotice(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.GetNotices(c.Request.Context(), stringParam(body, "userCode")))
}

func (h *Controller) readNotice(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.ReadNotice(c.Request.Context(), stringParam(body, "guid")))
}

func (h *Controller) readAllNotices(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.ReadAllNotices(c.Request.Context(), stringParam(body, "userCode")))
}

func (h *Controller) selectList(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.SelectList(c.Request.Context(), stringParam(body, "sql"), anySlice(body["params"])))
}

func (h *Controller) executeSQL(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.ExecuteSQL(c.Request.Context(), stringParam(body, "sql"), anySlice(body["params"])))
}

func (h *Controller) executeSQLList(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.ExecuteSQLBatch(c.Request.Context(), stringSlice(body["sqls"]), nestedAnySlices(body["params"]), false))
}

func (h *Controller) executeSQLComposite(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.ExecuteSQLBatch(c.Request.Context(), stringSlice(body["sqls"]), nestedAnySlices(body["params"]), true))
}

func (h *Controller) saveAll(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.SaveAll(c.Request.Context(), stringParam(body, "saveType"), stringParam(body, "tableName"), mapSlice(body["data"]), stringParam(body, "key")))
}

func (h *Controller) uploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		writeResult(c, model.Failure("上传文件不能为空"))
		return
	}
	writeResult(c, h.service.SaveUploadedFile(file, c.PostForm("spliceUrl")))
}

func (h *Controller) deleteFile(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.DeleteUploadedFile(stringParam(body, "url")))
}

func (h *Controller) deleteFiles(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.DeleteUploadedFiles(stringSlice(body["urls"])))
}
