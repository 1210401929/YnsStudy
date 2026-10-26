package controller

import (
	"github.com/gin-gonic/gin"

	"nys-go-api/internal/service"
)

// registerAdminRoutes 注册超级管理员后台接口。权限在 service 层逐个校验（requireSuperAdmin），
// 这里只负责参数解析。
func (h *Controller) registerAdminRoutes(group *gin.RouterGroup) {
	group.Any("/getDashboard", func(c *gin.Context) { writeResult(c, h.service.GetAdminDashboard(c)) })
	group.Any("/getComments", func(c *gin.Context) {
		body, ok := requireBody(c)
		if !ok {
			return
		}
		writeResult(c, h.service.GetAdminComments(c, service.AdminCommentQuery{
			Page:          intParam(body, "page", 1),
			PageSize:      intParam(body, "pageSize", 20),
			Keyword:       stringParam(body, "keyword"),
			Source:        stringParam(body, "source"),
			AnonymousOnly: asBoolValue(body["anonymousOnly"]),
		}))
	})
	group.Any("/deleteComment", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.DeleteAdminComment(c, stringParam(body, "source"), stringParam(body, "guid")))
		}
	})
	group.Any("/getUserDetail", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetAdminUserDetail(c, stringParam(body, "userId")))
		}
	})
	group.Any("/forceLogout", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.ForceLogoutUser(c, stringParam(body, "userId")))
		}
	})
	group.Any("/resetPassword", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.ResetUserPassword(c, stringParam(body, "userId")))
		}
	})

	// 内容管理
	contentQuery := func(body map[string]any) service.AdminContentQuery {
		return service.AdminContentQuery{
			Page:     intParam(body, "page", 1),
			PageSize: intParam(body, "pageSize", 20),
			Keyword:  stringParam(body, "keyword"),
			Type:     stringParam(body, "type"),
		}
	}
	group.Any("/getArticles", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetAdminArticles(c, contentQuery(body)))
		}
	})
	group.Any("/setArticleType", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.SetArticleVisibility(c, stringParam(body, "guid"), stringParam(body, "type")))
		}
	})
	group.Any("/deleteArticle", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.AdminDeleteArticle(c, stringParam(body, "guid")))
		}
	})
	group.Any("/getPosts", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetAdminCommunityPosts(c, contentQuery(body)))
		}
	})
	group.Any("/setPostTop", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.AdminSetTopPost(c, stringParam(body, "guid"), stringParam(body, "isTop")))
		}
	})
	group.Any("/deletePost", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.AdminDeletePost(c, stringParam(body, "guid")))
		}
	})

	// 违禁词
	group.Any("/getBannedWords", func(c *gin.Context) { writeResult(c, h.service.GetBannedWords(c)) })
	group.Any("/addBannedWord", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.AddBannedWord(c, stringParam(body, "word")))
		}
	})
	group.Any("/deleteBannedWord", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.DeleteBannedWord(c, stringParam(body, "word")))
		}
	})

	// SEO 小工具
	group.Any("/getSeoStatus", func(c *gin.Context) { writeResult(c, h.service.GetSeoStatus(c)) })
	group.Any("/submitIndexNow", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.SubmitIndexNowManually(c, stringSlice(body["urls"])))
		}
	})

	// 评论限制
	group.Any("/getCommentBans", func(c *gin.Context) { writeResult(c, h.service.GetCommentBans(c)) })
	group.Any("/addCommentBan", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.AddCommentBan(c, stringParam(body, "type"), stringParam(body, "value"), stringParam(body, "remark"), intParam(body, "days", 0)))
		}
	})
	group.Any("/deleteCommentBan", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.DeleteCommentBan(c, intParam(body, "id", 0)))
		}
	})

	// 操作日志
	// 关于页的站长推荐
	group.Any("/getAboutFeatured", func(c *gin.Context) { writeResult(c, h.service.GetAdminAboutFeatured(c)) })
	group.Any("/setAboutFeatured", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.SetAboutFeatured(c, stringSlice(body["articles"]), stringSlice(body["series"])))
		}
	})

	group.Any("/getLogs", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetAdminLogs(c, intParam(body, "page", 1), intParam(body, "pageSize", 20), stringParam(body, "action")))
		}
	})
}
