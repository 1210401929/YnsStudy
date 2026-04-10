package com.example.blog_api.controller;

import com.example.blog_api.service.HomeService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.config.LoginCfg;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import javax.jws.WebResult;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.time.LocalDate;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Controller
@RequestMapping("blog-api/home")
public class HomeController {

    @Autowired
    HomeService homeService;

    @RequestMapping("/getHomeData")
    @ResponseBody
    public ResultBody getHomeData() {
        return homeService.getHomeData();
    }

    @RequestMapping("/getWebsiteStatistics")
    @ResponseBody
    public ResultBody getWebsiteStatistics() {
        return homeService.getWebsiteStatistics();
    }

    @RequestMapping("/getHigAuthor")
    @ResponseBody
    public ResultBody getHigAuthor(@RequestBody Map<String, String> params) {
        String num = params.get("num");
        return homeService.getHigAuthor(num);
    }

    @GetMapping(value = "/sitemap.xml", produces = "application/xml;charset=UTF-8")
    public void sitemapXml(HttpServletResponse response) throws IOException {
        ResultBody result = homeService.getAllBlogGuidAndTime();  // 从数据库获取所有 GUID
        StringBuilder sb = new StringBuilder();

        sb.append("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n");
        sb.append("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n");

        for (Map<String,Object> oneBlog : (List<Map<String,Object>>)result.result) {
            sb.append("  <url>\n");
            sb.append("    <loc>" + LoginCfg.domainName + "/oneBlog/").append(oneBlog.get("GUID")).append("</loc>\n");

            // --- 时间格式化处理开始 ---
            String lastmod = "";
            Object timeObj = oneBlog.get("CREATE_TIME");
            if (timeObj instanceof java.time.LocalDateTime) {
                lastmod = ((java.time.LocalDateTime) timeObj).format(java.time.format.DateTimeFormatter.ISO_LOCAL_DATE);
            } else if (timeObj instanceof java.util.Date) {
                // 兼容 java.sql.Timestamp 和 java.util.Date
                lastmod = new java.text.SimpleDateFormat("yyyy-MM-dd").format((java.util.Date) timeObj);
            } else if (timeObj != null) {
                // 兜底方案：如果已经是类似 "2025-09-15T14:40:34" 的字符串，直接截取前10位
                String timeStr = timeObj.toString();
                lastmod = timeStr.length() >= 10 ? timeStr.substring(0, 10) : timeStr;
            }
            // --- 时间格式化处理结束 ---

            sb.append("    <lastmod>").append(lastmod).append("</lastmod>\n");


            sb.append("    <changefreq>monthly</changefreq>\n");
            sb.append("    <priority>0.8</priority>\n");
            sb.append("  </url>\n");
        }

        sb.append("</urlset>");

        response.getWriter().write(sb.toString());
    }

    @GetMapping(value = "/rss.xml", produces = "application/xml;charset=UTF-8")
    public void rssXml(HttpServletResponse response) throws IOException {
        // 1. 必须在调用 getWriter() 之前设置编码和内容类型
        response.setContentType("application/xml;charset=UTF-8");
        response.setCharacterEncoding("UTF-8");
        // 1. 获取数据
        ResultBody result = homeService.getBlogDetail5();
        List<Map<String, Object>> blogList = (List<Map<String, Object>>) result.result;
        StringBuilder sb = new StringBuilder();
        // 引入样式
        sb.append("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n");
        sb.append("<rss version=\"2.0\" xmlns:atom=\"http://www.w3.org/2005/Atom\">\n");
        sb.append("  <channel>\n");
        // --- 网站元数据 ---
        sb.append("    <title>ynsstudy</title>\n"); // 你的网站标题
        sb.append("    <link>" + LoginCfg.domainName + "</link>\n");
        sb.append("    <description>ynsStudy - 技术分享与记录</description>\n");
        sb.append("    <language>zh-cn</language>\n");
        for (Map<String, Object> oneBlog : blogList) {
            sb.append("    <item>\n");
            // 标题（如果你库里有标题字段，请替换 "BLOG_TITLE"）
            String title = oneBlog.containsKey("BLOG_TITLE") ? (String) oneBlog.get("BLOG_TITLE") : "文章 ID: " + oneBlog.get("GUID");
            sb.append("      <title><![CDATA[" + title + "]]></title>\n");
            // 链接
            String link = LoginCfg.domainName + "/oneBlog/" + oneBlog.get("GUID");
            sb.append("      <link>" + link + "</link>\n");
            sb.append("      <guid>" + oneBlog.get("GUID") + "</guid>\n");
            // --- 时间格式化 (RSS 要求 RFC 822 格式，例如: Wed, 10 Apr 2026 09:40:35 GMT) ---
            // 简单处理也可以用 yyyy-MM-dd，但标准 RSS 客户端更喜欢 RFC 822
            String pubDate = "";
            Object timeObj = oneBlog.get("CREATE_TIME");
            if (timeObj instanceof java.time.LocalDateTime) {
                // 使用标准的 RFC_1123 格式
                pubDate = ((java.time.LocalDateTime) timeObj).atZone(java.time.ZoneId.systemDefault())
                        .format(java.time.format.DateTimeFormatter.RFC_1123_DATE_TIME);
            } else {
                pubDate = oneBlog.get("CREATE_TIME").toString();
            }
            sb.append("      <pubDate>").append(pubDate).append("</pubDate>\n");
            // 摘要
            String content = oneBlog.containsKey("MAINTEXT") ?getPureText((String) oneBlog.get("MAINTEXT"),50)  : "";
            sb.append("<description><![CDATA[" + content + "]]></description>\n");
            sb.append("    </item>\n");
        }
        sb.append("  </channel>\n");
        sb.append("</rss>");

        response.getWriter().write(sb.toString());
    }
    /**
     * 提取 HTML 中的纯文本并截取
     */
    private String getPureText(String html, int limit) {
        if (html == null || html.isEmpty()) {
            return "";
        }
        String pureText = html.replaceAll("<[^>]+>", "");
        pureText = pureText.replaceAll("&nbsp;", " ");
        if (pureText.length() <= limit) {
            return pureText;
        }
        return pureText.substring(0, limit) + "...";
    }
}
