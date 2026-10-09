package controller

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Controller) registerBlogRoutes(router *gin.Engine) {
	h.registerAnnouncementRoutes(router.Group("/blog-api/sso"))
	h.registerArticleRoutes(router.Group("/blog-api/blog"))
	h.registerCommunityRoutes(router.Group("/blog-api/community"))
	h.registerFriendLinkRoutes(router.Group("/blog-api/friendLink"))
	h.registerResourceRoutes(router.Group("/blog-api/resource"))
	h.registerUserInformationRoutes(router.Group("/blog-api/userInformation"))
	h.registerHomeRoutes(router.Group("/blog-api/home"))
	h.registerLuluRoutes(router.Group("/blog-api/lulu"))
	router.Any("/blog-api/file/consistencyFileCheck", h.consistencyFileCheck)
}

func (h *Controller) registerAnnouncementRoutes(group *gin.RouterGroup) {
	group.Any("/addAnnouncement", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.AddAnnouncement(c, nestedMap(body, "announcement")))
		}
	})
	group.Any("/editAnnouncement", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.EditAnnouncement(c.Request.Context(), nestedMap(body, "announcement")))
		}
	})
	group.Any("/getAllAnnouncement", func(c *gin.Context) { writeResult(c, h.service.GetAllAnnouncements(c.Request.Context())) })
	group.Any("/getAnnouncementByType", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetAnnouncementsByType(c.Request.Context(), stringParam(body, "type")))
		}
	})
	group.Any("/deleteAnnouncement", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.DeleteAnnouncement(c.Request.Context(), stringParam(body, "guid")))
		}
	})
}

func (h *Controller) registerArticleRoutes(group *gin.RouterGroup) {
	group.Any("/addBlog", h.addBlog)
	group.Any("/deleteBlog", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.DeleteBlog(c.Request.Context(), stringParam(body, "guid")))
		}
	})
	group.Any("/updateBlog", h.updateBlog)
	group.Any("/getUserBlog", func(c *gin.Context) { writeResult(c, h.service.GetCurrentUserBlogs(c)) })
	group.Any("/getUserBlogByUserCode", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetBlogsByUser(c.Request.Context(), stringParam(body, "userCode"), true))
		}
	})
	group.Any("/updateBlogCatId", h.updateBlogCategory)
	group.Any("/getBlog", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetBlog(c, stringParam(body, "blogId")))
		}
	})
	group.Any("/getAllBlog", h.getAllBlogs)
	group.Any("/getArticleLinks", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetArticleLinks(c.Request.Context(), stringParam(body, "blogId")))
		}
	})
	group.Any("/addBlogCat", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.AddBlogCategory(c.Request.Context(), nestedMap(body, "blogCat")))
		}
	})
	group.Any("/updateBlogCat", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.UpdateBlogCategoryInfo(c.Request.Context(), nestedMap(body, "blogCat")))
		}
	})
	group.Any("/getUserBlogCat", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetUserBlogCategories(c.Request.Context(), stringParam(body, "userCode")))
		}
	})
	group.Any("/deleteBlogCat", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.DeleteBlogCategory(c.Request.Context(), stringParam(body, "guid")))
		}
	})
	group.Any("/addComment", h.addBlogComment)
	group.Any("/getComment", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetBlogComments(c.Request.Context(), stringParam(body, "blogId")))
		}
	})
	group.Any("/deleteComment", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.DeleteBlogComment(c.Request.Context(), stringParam(body, "blogGuid")))
		}
	})
	group.Any("/giveLikeBlog", func(c *gin.Context) { h.blogReaction(c, "like", false) })
	group.Any("/noGiveLikeBlog", func(c *gin.Context) { h.blogReaction(c, "like", true) })
	group.Any("/collectBlog", func(c *gin.Context) { h.blogReaction(c, "collect", false) })
	group.Any("/noCollectBlog", func(c *gin.Context) { h.blogReaction(c, "collect", true) })
	group.Any("/getGiveLikeByBlogId", func(c *gin.Context) { h.getBlogReactions(c, "like") })
	group.Any("/getLikeAndCollectByBlogId", func(c *gin.Context) { h.getBlogReactions(c, "") })
	group.Any("/getLikeAndCollectByUserCode", h.getUserReactions)
}

func (h *Controller) addBlog(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	blog := nestedMap(body, "blogContent")
	writeResult(c, h.service.AddBlog(c.Request.Context(), blog))
}

func (h *Controller) updateBlog(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.UpdateBlog(c.Request.Context(), stringParam(body, "guid"), stringParam(body, "title"), stringParam(body, "content"), stringParam(body, "blog_type")))
}

func (h *Controller) updateBlogCategory(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.UpdateBlogCategory(c.Request.Context(), stringParam(body, "blogId"), stringParam(body, "catId")))
}

func (h *Controller) getAllBlogs(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.GetAllBlogs(c.Request.Context(), intParam(body, "page", 1), intParam(body, "pageSize", 10), stringParam(body, "keyword")))
}

func (h *Controller) addBlogComment(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.AddBlogComment(c, nestedMap(body, "blogComment")))
}

func (h *Controller) blogReaction(c *gin.Context, reactionType string, remove bool) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	if remove {
		writeResult(c, h.service.RemoveBlogReaction(c, stringParam(body, "blogId"), reactionType))
	} else {
		writeResult(c, h.service.AddBlogReaction(c, stringParam(body, "blogId"), reactionType))
	}
}

func (h *Controller) getBlogReactions(c *gin.Context, reactionType string) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.GetBlogReactions(c.Request.Context(), stringParam(body, "blogId"), reactionType))
}

func (h *Controller) getUserReactions(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.GetUserReactions(c.Request.Context(), stringParam(body, "userCode"), asBoolValue(body["isCountOnly"]), stringParam(body, "type")))
}

func (h *Controller) registerCommunityRoutes(group *gin.RouterGroup) {
	group.Any("/addCommunity", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.AddCommunity(c, nestedMap(body, "community")))
		}
	})
	group.Any("/deleteCommunity", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.DeleteCommunity(c.Request.Context(), stringParam(body, "communityGuid")))
		}
	})
	group.Any("/setTopCommunity", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.SetTopCommunity(c.Request.Context(), stringParam(body, "communityGuid"), stringParam(body, "isTop")))
		}
	})
	group.Any("/getAllCommunity", h.getAllCommunities)
	group.Any("/addComment", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.AddCommunityComment(c, nestedMap(body, "comment")))
		}
	})
	group.Any("/getComment", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetCommunityComments(c.Request.Context(), stringParam(body, "communityId")))
		}
	})
}

func (h *Controller) getAllCommunities(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.GetAllCommunities(c.Request.Context(), intParam(body, "page", 1), intParam(body, "pageSize", 10), stringParam(body, "keyword")))
}

func (h *Controller) registerFriendLinkRoutes(group *gin.RouterGroup) {
	group.Any("/addFriendLink", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.AddFriendLink(c.Request.Context(), nestedMap(body, "friendLink")))
		}
	})
	group.Any("/updateFriendLink", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.UpdateFriendLink(c.Request.Context(), nestedMap(body, "friendLink")))
		}
	})
	group.Any("/deleteFriendLink", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.DeleteFriendLink(c.Request.Context(), stringParam(body, "friendLinkId")))
		}
	})
	group.Any("/getFriendLink", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetFriendLink(c.Request.Context(), stringParam(body, "friendLinkId")))
		}
	})
	group.Any("/getFriendLinks", func(c *gin.Context) { writeResult(c, h.service.GetFriendLinks(c.Request.Context())) })
}

func (h *Controller) registerResourceRoutes(group *gin.RouterGroup) {
	group.Any("/addFileInfo", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.AddFileInfo(c.Request.Context(), nestedMap(body, "fileInfo")))
		}
	})
	group.Any("/delFileInfo", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.DeleteFileInfo(c.Request.Context(), stringParam(body, "guid"), stringParam(body, "url")))
		}
	})
	group.Any("/getAllFile", h.getAllFiles)
	group.Any("/getFileByUser", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetFilesByUser(c.Request.Context(), stringParam(body, "userCode")))
		}
	})
	group.Any("/getFileById", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.GetFileByID(c.Request.Context(), stringParam(body, "guid")))
		}
	})
	group.Any("/updateFileInfo", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.UpdateFileInfo(c.Request.Context(), stringParam(body, "guid"), stringParam(body, "originalFileName"), stringParam(body, "remark")))
		}
	})
	group.Any("/setFileDownNum", func(c *gin.Context) {
		body, ok := requireBody(c)
		if ok {
			writeResult(c, h.service.IncrementFileDownloads(c.Request.Context(), stringParam(body, "guid")))
		}
	})
}

func (h *Controller) getAllFiles(c *gin.Context) {
	body, ok := requireBody(c)
	if !ok {
		return
	}
	writeResult(c, h.service.GetAllFiles(c.Request.Context(), intParam(body, "page", 1), intParam(body, "pageSize", 10), stringParam(body, "keyword")))
}

func (h *Controller) consistencyFileCheck(c *gin.Context) {
	body, _ := readBody(c)
	fileType := stringParam(body, "type")
	if fileType == "" {
		fileType = c.Query("type")
	}
	writeResult(c, h.service.CheckFileConsistency(c.Request.Context(), fileType))
}

func asBoolValue(value any) bool { return strings.EqualFold(fmt.Sprint(value), "true") }
