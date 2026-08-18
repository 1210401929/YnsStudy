package controller

import (
	"testing"
	"time"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
	"nys-go-api/internal/security"
	"nys-go-api/internal/service"
	"nys-go-api/internal/session"
)

func TestAllCompatibilityRoutesAreRegistered(t *testing.T) {
	cfg, err := config.Load("../../config/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Server.Mode = "test"
	store := cache.NewMemoryStore()
	jwt := security.NewJWT(cfg.Security.JWTSecret, time.Hour)
	sessions := session.NewManager(store, cfg.Security)
	services := service.New(cfg, nil, store, sessions, jwt)
	cipher, err := security.NewAESCipher(cfg.Security.AESKey, cfg.Security.AESIV)
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(cfg, services, cipher)

	registered := make(map[string]bool)
	for _, route := range router.Routes() {
		registered[route.Path] = true
	}
	expected := []string{
		"/health",
		"/", "/oneBlog/:id", "/sitemap.xml", "/robots.txt", "/rss.xml",
		"/pub-api/api/getCurrentCity", "/pub-api/api/getClientIpAddress",
		"/pub-api/login/sendPhoneCode", "/pub-api/login/loginByPhoneCode", "/pub-api/login/qq/authorize", "/pub-api/login/qq/callback", "/pub-api/login/logout", "/pub-api/login/loginUser", "/pub-api/login/checkUserLogin", "/pub-api/login/changePassWord", "/pub-api/login/register", "/pub-api/login/changeUserInfo", "/pub-api/login/deleteUserAvatarFile", "/pub-api/login/getUserInfoByCode", "/pub-api/login/getUserInfoByNum", "/pub-api/login/getUserInfoByName", "/pub-api/login/getAllUserInfo", "/pub-api/login/operationUser",
		"/pub-api/notice/addNotice", "/pub-api/notice/getNotice", "/pub-api/notice/readNotice", "/pub-api/notice/allReadNotice",
		"/pub-api/sql/selectList", "/pub-api/sql/selectListByParams", "/pub-api/sql/deleteBySql", "/pub-api/sql/updateBySql", "/pub-api/sql/exeSql", "/pub-api/sql/exeSqlByParams", "/pub-api/sql/exeSqlListByParams", "/pub-api/sql/exeSqlComposite", "/pub-api/sql/saveAllTableData", "/pub-api/sql/saveAllTableDataByParams",
		"/pub-api/upload/uploadFile", "/pub-api/upload/deleteFileByUrl", "/pub-api/upload/deleteFileByUrls",
		"/blog-api/sso/addAnnouncement", "/blog-api/sso/editAnnouncement", "/blog-api/sso/getAllAnnouncement", "/blog-api/sso/getAnnouncementByType", "/blog-api/sso/deleteAnnouncement",
		"/blog-api/community/addCommunity", "/blog-api/community/deleteCommunity", "/blog-api/community/setTopCommunity", "/blog-api/community/getAllCommunity", "/blog-api/community/addComment", "/blog-api/community/getComment",
		"/blog-api/friendLink/addFriendLink", "/blog-api/friendLink/updateFriendLink", "/blog-api/friendLink/deleteFriendLink", "/blog-api/friendLink/getFriendLink", "/blog-api/friendLink/getFriendLinks",
		"/blog-api/resource/addFileInfo", "/blog-api/resource/delFileInfo", "/blog-api/resource/getAllFile", "/blog-api/resource/getFileByUser", "/blog-api/resource/getFileById", "/blog-api/resource/updateFileInfo", "/blog-api/resource/setFileDownNum",
		"/blog-api/userInformation/followUser", "/blog-api/userInformation/noFollowUser", "/blog-api/userInformation/getFollowUser", "/blog-api/userInformation/getBlogAndResourceByUserCode", "/blog-api/userInformation/getBlogAndCommunityByUserCode", "/blog-api/userInformation/getResourceByUserCode", "/blog-api/userInformation/getBlogByUserCode", "/blog-api/userInformation/setPersonInfo", "/blog-api/userInformation/getPersonInfo",
		"/blog-api/home/getHomeData", "/blog-api/home/getWebsiteStatistics", "/blog-api/home/getHigAuthor", "/blog-api/home/sitemap_blog.xml", "/blog-api/home/sitemap_user.xml", "/blog-api/home/rss.xml",
		"/blog-api/lulu/status", "/blog-api/lulu/feed", "/blog-api/lulu/play", "/blog-api/lulu/sleep", "/blog-api/lulu/messages", "/blog-api/lulu/message/add", "/blog-api/lulu/message/delete", "/blog-api/lulu/logs", "/blog-api/lulu/fun-state", "/blog-api/lulu/mission/advance", "/blog-api/lulu/clothes/change",
		"/blog-api/blog/addBlog", "/blog-api/blog/deleteBlog", "/blog-api/blog/updateBlog", "/blog-api/blog/getUserBlog", "/blog-api/blog/getUserBlogByUserCode", "/blog-api/blog/updateBlogCatId", "/blog-api/blog/getBlog", "/blog-api/blog/getAllBlog", "/blog-api/blog/addBlogCat", "/blog-api/blog/updateBlogCat", "/blog-api/blog/getUserBlogCat", "/blog-api/blog/deleteBlogCat", "/blog-api/blog/addComment", "/blog-api/blog/getComment", "/blog-api/blog/deleteComment", "/blog-api/blog/giveLikeBlog", "/blog-api/blog/noGiveLikeBlog", "/blog-api/blog/getGiveLikeByBlogId", "/blog-api/blog/getLikeAndCollectByBlogId", "/blog-api/blog/getLikeAndCollectByUserCode", "/blog-api/blog/collectBlog", "/blog-api/blog/noCollectBlog",
		"/blog-api/file/consistencyFileCheck", "/ai-api/ai/qWen", "/ai-api/ai/qWenStream",
	}
	for _, path := range expected {
		if !registered[path] {
			t.Errorf("缺少兼容路由 %s", path)
		}
	}
}
