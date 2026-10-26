package service

// 关于页的内容：站长想说、功能介绍、建站历程、常见问题、使用技术。
// 前端关于页和 Go 输出的首屏都从这里取，只需要改这一处；以后要改成后台可编辑，也只用换掉这里的数据来源。

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"nys-go-api/internal/model"
)

// AboutItem 是功能、历程、常见问题里的一条
type AboutItem struct {
	Icon  string `json:"icon,omitempty"`
	Date  string `json:"date,omitempty"`
	Title string `json:"title"`
	Text  string `json:"text"`
	Link  string `json:"link,omitempty"`
}

// AboutContent 是关于页的全部文字内容
type AboutContent struct {
	Title        string      `json:"title"`
	Intro        string      `json:"intro"`
	OwnerMessage []string    `json:"ownerMessage"`
	Features     []AboutItem `json:"features"`
	// SiteStartDate 是网站上线日期（YYYY-MM-DD），用来算“已经陪伴了多少天”
	SiteStartDate string `json:"siteStartDate"`
	// ChangelogPath 是站长的《网站更新日志》文章地址，“使用技术”下面放一个入口
	ChangelogPath string      `json:"changelogPath"`
	FAQ           []AboutItem `json:"faq"`
	TechStack     []AboutItem `json:"techStack"`
	ContactEmail  string      `json:"contactEmail"`
	SourceURL     string      `json:"sourceUrl"`
}

// aboutContent 关于页的文字。“站长推荐”不在这里，由站长在后台配置（见 featured_service.go）。
var aboutContent = AboutContent{
	Title: "欢迎来到 YnsStudy",
	Intro: "一个集文章创作、评论互动与资源分享于一体的知识社区",
	OwnerMessage: []string{
		"正如上面的介绍，这个网站从建立之初，就不只是想做成我一个人的博客，而是想为每一个热爱写文章的人提供一个可以自由创作的平台。在这里，每个人都可以发表属于自己的文章。",
		"也正因如此，很多博客聚合网站拒绝收录本站，原因也很简单：这个网站并不算一个纯粹的个人博客。",
		"建站到现在，网站的流量确实不多，也确实没有太多能让路人驻足的理由。因为没有多少流量，也没有多少用户，它慢慢地几乎又变成了我自己的个人博客。",
		"我的技术分享、生活吐槽，还有这些年留下的一点一滴，都记录在了这里。我只希望自己还能坚守当初建站时的本心，让这个网站一直活下去。",
	},
	Features: []AboutItem{
		{Icon: "📝", Title: "发表文章", Text: "富文本编辑，可设为公开或私密；用标签归类，用系列把相关文章编成一套按顺序阅读。", Link: "/ynsStudy/MyBlog"},
		{Icon: "📌", Title: "便签墙", Text: "随手记下技巧、踩坑和想法，支持 Markdown、配图、话题和链接卡片。", Link: "/ynsStudy/Notes"},
		{Icon: "💬", Title: "评论与留言", Text: "文章、随记都能评论，不登录也可以匿名留言；有人回复时会收到邮件提醒。", Link: "/ynsStudy/Notes?tab=guestbook"},
		{Icon: "📂", Title: "资源分享", Text: "上传和下载学习资料、工具和素材，按文档、图片、视频等分类查找。", Link: "/ynsStudy/Resources"},
		{Icon: "🤖", Title: "AI 帮写", Text: "写文章时让 AI 根据标题起草或接着往下写。"},
		{Icon: "📡", Title: "RSS 订阅", Text: "用 RSS 阅读器订阅新文章，不错过更新。", Link: "/rss"},
	},
	// 网站 1.0 上线日期，见《网站更新日志》
	SiteStartDate: "2025-08-23",
	ChangelogPath: "/oneBlog/46140E57DA81405EAB4C48871B7FE790",
	FAQ: []AboutItem{
		{Title: "怎么注册和登录？", Text: "点击右上角的“登录”，可以用账号密码注册，也可以直接用手机验证码或 QQ 登录。"},
		{Title: "怎么发表文章？", Text: "登录后进入导航栏的“发表”，点击右下角的“发布文章”。文章可以设为私密，只有自己能看到。"},
		{Title: "怎么申请友链？", Text: "在友链页点击“加入友链”填写网站信息，站长审核通过后就会显示。友链页的“本站信息”可以直接复制。", Link: "/ynsStudy/FriendLink"},
		{Title: "上传资源有什么限制？", Text: "单个文件不超过 50MB，普通用户最多保留 5 个文件；出于安全考虑，不支持上传 SVG、HTML 这类可以执行脚本的文件。", Link: "/ynsStudy/Resources"},
		{Title: "不登录可以评论吗？", Text: "可以。填写昵称即可匿名评论，留下邮箱的话，有人回复时会收到邮件提醒。"},
	},
	TechStack: []AboutItem{
		{Title: "前端", Text: "Vue 3 + Vite + Element Plus"},
		{Title: "后端", Text: "Go（2026 年 8 月起替代 Spring Cloud + MyBatis）"},
		{Title: "数据", Text: "MySQL + Redis"},
		{Title: "部署", Text: "腾讯云服务器，Docker + Nginx，文件存储在服务器本地"},
	},
	ContactEmail: "yvnansong@qq.com",
	SourceURL:    "https://github.com/1210401929/YnsStudy",
}

// AboutContentData 返回关于页的文字内容
func AboutContentData() AboutContent {
	return aboutContent
}

// AboutOwner 是关于页的站长卡片
type AboutOwner struct {
	Name           string   `json:"name"`
	AlternateNames []string `json:"alternateNames"`
	SameAs         []string `json:"sameAs"`
	ProfilePath    string   `json:"profilePath"`
	Avatar         string   `json:"avatar"`
	Remark         string   `json:"remark"`
}

// GetAboutPage 是关于页接口：文字内容、站长卡片、站长推荐、建站天数和网站统计
func (s *Service) GetAboutPage(ctx context.Context) model.Result {
	owner := s.GetSiteOwner(ctx)
	card := AboutOwner{Name: owner.Name, AlternateNames: owner.AlternateNames, SameAs: owner.SameAs, ProfilePath: owner.ProfilePath, Avatar: owner.Avatar}
	if rows, err := s.Repo.Query(ctx, "SELECT REMARK FROM userInfo WHERE CODE = ? LIMIT 1", owner.Code); err == nil && len(rows) > 0 {
		card.Remark = model.StringValue(rows[0], "REMARK")
	}
	stats, _ := s.SiteStatistics(ctx)
	return model.Success(map[string]any{
		"content":  aboutContent,
		"owner":    card,
		"stats":    stats,
		"featured": s.AboutFeatured(ctx),
		"siteDays": SiteDays(time.Now()),
	})
}

// SiteDays 返回到 now 为止网站上线的天数，上线当天算第 1 天；日期按北京时间计算
func SiteDays(now time.Time) int {
	start, err := time.ParseInLocation("2006-01-02", aboutContent.SiteStartDate, time.Local)
	if err != nil {
		return 0
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	days := int(today.Sub(start).Hours()/24) + 1
	if days < 1 {
		return 0
	}
	return days
}

const siteStatisticsCacheKey = "site_statistics"
const siteStatisticsCacheTTL = 5 * time.Minute

// SiteStatistics 返回网站统计：文章、随记、阅读、用户、登录次数，以及有公开文章的系列和标签数。
// 这些都是整表计数，结果缓存 5 分钟。
func (s *Service) SiteStatistics(ctx context.Context) (map[string]any, error) {
	if s.Cache != nil {
		if cached, err := s.Cache.Get(ctx, siteStatisticsCacheKey); err == nil && cached != "" {
			var stats map[string]any
			if json.Unmarshal([]byte(cached), &stats) == nil {
				return stats, nil
			}
		}
	}
	rows, err := s.Repo.Query(ctx, `SELECT
(SELECT COUNT(1) FROM blogInfo) AS ARTICLENUM,
(SELECT COUNT(1) FROM communityInfo) AS COMMUNITYNUM,
(SELECT COALESCE(SUM(VIEW_PAGE), 0) FROM blogInfo) AS VIEW_PAGE,
(SELECT COUNT(1) FROM userInfo) AS USERNUM,
(SELECT COUNT(1) FROM loginHistory) AS USERLOGINNUM`)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	stats := rows[0]
	stats["SERIESNUM"] = 0
	stats["TAGNUM"] = 0
	if series, err := s.PublicSeriesList(ctx, "", "latest", 200); err == nil {
		stats["SERIESNUM"] = len(series)
	}
	if tags, err := s.PopularBlogTags(ctx, 100); err == nil {
		stats["TAGNUM"] = len(tags)
	}
	if s.Cache != nil {
		if encoded, err := json.Marshal(stats); err == nil {
			if err := s.Cache.Set(ctx, siteStatisticsCacheKey, string(encoded), siteStatisticsCacheTTL); err != nil {
				log.Printf("缓存网站统计失败: %v", err)
			}
		}
	}
	return stats, nil
}
