package service

import (
	"context"
	"net/url"
	"strings"

	"nys-go-api/internal/model"
)

// ProfileArticleLimit 是作者主页服务端首屏列出的文章数量上限；更多文章通过文章归档发现。
const ProfileArticleLimit = 100

// SiteOwner 是网站站长（超级管理员账号）的公开信息，用于首页标题和结构化数据。
type SiteOwner struct {
	Code           string
	Name           string
	AlternateNames []string
	SameAs         []string
	ProfilePath    string // 例如 /user/101；查不到账号时为空
	Avatar         string
}

// GetSiteOwner 读取站长信息。名字、别名、外部主页来自配置（有默认值），
// 主页地址和头像来自站长账号；数据库查询失败时只是缺少这两项，不影响页面输出。
func (s *Service) GetSiteOwner(ctx context.Context) SiteOwner {
	name, alternateNames, sameAs := s.Config.SEO.Owner()
	owner := SiteOwner{Code: s.superAdminCode(), Name: name, AlternateNames: alternateNames, SameAs: sameAs}
	rows, err := s.Repo.Query(ctx, "SELECT USERNUM, AVATAR FROM userInfo WHERE CODE = ? LIMIT 1", owner.Code)
	if err == nil && len(rows) > 0 {
		if number := strings.TrimSpace(model.StringValue(rows[0], "USERNUM")); number != "" {
			owner.ProfilePath = "/user/" + url.PathEscape(number)
		}
		owner.Avatar = strings.TrimSpace(model.StringValue(rows[0], "AVATAR"))
	}
	return owner
}

// PublicProfile 是作者主页服务端首屏需要的数据。
type PublicProfile struct {
	User         model.User
	Articles     []map[string]any
	ArticleTotal int
	IsOwner      bool
}

// GetPublicProfileForSEO 按用户编号读取公开资料和公开文章。found=false 表示用户不存在。
func (s *Service) GetPublicProfileForSEO(ctx context.Context, userNumber string) (PublicProfile, bool, error) {
	userNumber = strings.TrimSpace(userNumber)
	if userNumber == "" {
		return PublicProfile{}, false, nil
	}
	rows, err := s.Repo.Query(ctx, "SELECT * FROM userInfo WHERE USERNUM = ? LIMIT 1", userNumber)
	if err != nil {
		return PublicProfile{}, false, err
	}
	if len(rows) == 0 {
		return PublicProfile{}, false, nil
	}
	user := model.UserFromRow(rows[0]).PublicProfile()
	// 被封禁的账号不对外展示主页
	if user.IsBan == "1" {
		return PublicProfile{}, false, nil
	}
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(*) AS total FROM blogInfo WHERE USERCODE = ? AND BLOG_TYPE = 'public'", user.Code)
	if err != nil {
		return PublicProfile{}, false, err
	}
	articles, err := s.Repo.Query(ctx, `SELECT GUID, BLOG_TITLE, CREATE_TIME
FROM blogInfo
WHERE USERCODE = ? AND BLOG_TYPE = 'public'
ORDER BY CREATE_TIME DESC
LIMIT ?`, user.Code, ProfileArticleLimit)
	if err != nil {
		return PublicProfile{}, false, err
	}
	return PublicProfile{
		User:         user,
		Articles:     articles,
		ArticleTotal: int(firstCount(counts)),
		IsOwner:      user.Code == s.superAdminCode(),
	}, true, nil
}
