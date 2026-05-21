package com.example.blog_api.controller;

import com.example.blog_api.service.HomeService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.config.LoginCfg;
import com.example.common_api.service.CallService;
import com.example.common_api.util.FunToUrlUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import javax.jws.WebResult;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.io.PrintWriter;
import java.time.LocalDate;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Controller
@RequestMapping("blog-api/home")
public class HomeController {

    @Autowired
    HomeService homeService;

    @Autowired
    CallService callService;

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

    @GetMapping(value = "/sitemap_blog.xml", produces = "application/xml;charset=UTF-8")
    public void sitemapBlogXml(HttpServletResponse response) throws IOException {
        // 强制设置响应头
        response.setContentType("application/xml");
        response.setCharacterEncoding("UTF-8");

        // 直接获取 PrintWriter 进行流式写入，极大地节省 JVM 内存
        PrintWriter out = response.getWriter();

        // 建议：这个方法内部应该优化为只 SELECT GUID, CREATE_TIME FROM blog_table
        ResultBody result = homeService.getAllBlogGuidAndTime();

        out.print("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n");
        out.print("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n");

        for (Map<String,Object> oneBlog : (List<Map<String,Object>>)result.result) {
            // XML 特殊字符简易转义 (重点处理 &)
            String rawLoc = LoginCfg.domainName + "/oneBlog/" + oneBlog.get("GUID");
            String safeLoc = rawLoc.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;");

            out.print("  <url>\n");
            out.print("    <loc>" + safeLoc + "</loc>\n");

            // 时间格式化处理
            String lastmod = "";
            Object timeObj = oneBlog.get("CREATE_TIME");
            if (timeObj instanceof java.time.LocalDateTime) {
                lastmod = ((java.time.LocalDateTime) timeObj).format(java.time.format.DateTimeFormatter.ISO_LOCAL_DATE);
            } else if (timeObj instanceof java.util.Date) {
                lastmod = new java.text.SimpleDateFormat("yyyy-MM-dd").format((java.util.Date) timeObj);
            } else if (timeObj != null) {
                String timeStr = timeObj.toString();
                lastmod = timeStr.length() >= 10 ? timeStr.substring(0, 10) : timeStr;
            }

            if (!lastmod.isEmpty()) {
                out.print("    <lastmod>" + lastmod + "</lastmod>\n");
            }

            out.print("    <changefreq>weekly</changefreq>\n"); // 文章如果有评论，建议改为 weekly
            out.print("    <priority>0.8</priority>\n");
            out.print("  </url>\n");

            // 可选：定期 flush()，将缓冲区数据推向前端，防止 Nginx 响应超时
            // out.flush();
        }

        out.print("</urlset>");
        out.flush();
    }

    @GetMapping(value = "/sitemap_user.xml", produces = "application/xml;charset=UTF-8")
    public void sitemapUserXml(HttpServletResponse response) throws IOException {
        // 1. 强制设置响应头，确保浏览器和爬虫正确识别 XML
        response.setContentType("application/xml");
        response.setCharacterEncoding("UTF-8");
        // 2. 拿到 PrintWriter，改为流式输出，边查边写，不占内存
        PrintWriter out = response.getWriter();
        // 以后有空了，在底层专门写个没有分页、只 SELECT USERNUM 的极简 SQL。
        Map<String,Object> params = new HashMap<>();
        params.put("page", 1);
        params.put("pageSize", 10000);
        params.put("keyword", null);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.getAllUserInfoUrl, params);
        out.print("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n");
        out.print("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n");
        // 严谨一点，加上非空判断，防止空指针异常导致整个 XML 崩溃
        if (result != null && result.result != null) {
            Map<String, Object> resultMap = (Map<String, Object>) result.result;
            List<Map<String, Object>> userList = (List<Map<String, Object>>) resultMap.get("data");
            if (userList != null) {
                for (Map<String, Object> user : userList) {
                    Object userNumObj = user.get("USERNUM");
                    if (userNumObj == null) continue;
                    // 3. 核心：XML 特殊字符简易转义
                    String rawLoc = LoginCfg.domainName + "/user/" + userNumObj.toString();
                    String safeLoc = rawLoc.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;");
                    out.print("  <url>\n");
                    out.print("    <loc>" + safeLoc + "</loc>\n");
                    // 4. SEO 策略调整
                    // 个人主页资料更新频率一般，设为 weekly 比较合理
                    out.print("    <changefreq>weekly</changefreq>\n");
                    // 个人主页权重 (0.5) 应该略低于核心的文章内容 (0.8)
                    out.print("    <priority>0.5</priority>\n");
                    out.print("  </url>\n");
                }
            }
        }

        out.print("</urlset>");
        // 将缓冲区最后的残余数据强行推向前端
        out.flush();
    }

    @GetMapping(value = "/rss.xml", produces = "application/xml;charset=UTF-8")
    public void rssXml(HttpServletResponse response) throws IOException {
        // 1. 设置编码和内容类型
        response.setContentType("application/xml;charset=UTF-8");
        response.setCharacterEncoding("UTF-8");

        // 2. 获取数据
        ResultBody result = homeService.getBlogDetail5();
        List<Map<String, Object>> blogList = (List<Map<String, Object>>) result.result;
        StringBuilder sb = new StringBuilder();

        // 引入样式
        sb.append("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n");
        sb.append("<rss version=\"2.0\" xmlns:atom=\"http://www.w3.org/2005/Atom\">\n");
        sb.append("  <channel>\n");

        // --- 网站元数据 ---
        sb.append("    <atom:link href=\"" + LoginCfg.domainName + "/api/blog-api/home/rss.xml\" rel=\"self\" type=\"application/rss+xml\"/>\n");
        sb.append("    <title>ynsstudy</title>\n");
        sb.append("    <link>" + LoginCfg.domainName + "</link>\n");
        sb.append("    <description>ynsStudy - 技术分享与记录</description>\n");
        sb.append("    <language>zh-CN</language>\n");

        // --- 新增：lastBuildDate (取最新一篇文章的时间，如果为空则取当前时间) ---
        String lastBuildDate = "";
        if (blogList != null && !blogList.isEmpty()) {
            lastBuildDate = formatToRfc822(blogList.get(0).get("CREATE_TIME"));
        } else {
            // 列表为空时的保底方案：当前时间
            lastBuildDate = java.time.ZonedDateTime.now(java.time.ZoneId.of("Asia/Shanghai"))
                    .format(java.time.format.DateTimeFormatter.RFC_1123_DATE_TIME);
        }
        sb.append("    <lastBuildDate>").append(lastBuildDate).append("</lastBuildDate>\n");

        // --- 遍历文章列表 ---
        if (blogList != null) {
            for (Map<String, Object> oneBlog : blogList) {
                sb.append("    <item>\n");

                // 标题
                String title = oneBlog.containsKey("BLOG_TITLE") ? (String) oneBlog.get("BLOG_TITLE") : "文章 ID: " + oneBlog.get("GUID");
                sb.append("      <title><![CDATA[" + title + "]]></title>\n");

                // 链接与 GUID
                String link = LoginCfg.domainName + "/oneBlog/" + oneBlog.get("GUID");
                sb.append("      <link>" + link + "</link>\n");
                sb.append("      <guid>" + link + "</guid>\n");

                // 时间（直接调用抽离出去的方法）
                String pubDate = formatToRfc822(oneBlog.get("CREATE_TIME"));
                sb.append("      <pubDate>").append(pubDate).append("</pubDate>\n");

                // 摘要
                String content = oneBlog.containsKey("MAINTEXT") ? getPureText((String) oneBlog.get("MAINTEXT"), 50) : "";
                sb.append("      <description><![CDATA[" + content + "]]></description>\n");
                sb.append("    </item>\n");
            }
        }

        sb.append("  </channel>\n");
        sb.append("</rss>");

        response.getWriter().write(sb.toString());
    }

    /**
     * 新增的私有辅助方法：将数据库查出的各种时间对象统一格式化为 RSS 要求的 RFC 822 格式
     */
    private String formatToRfc822(Object timeObj) {
        if (timeObj == null) {
            return "";
        }
        try {
            java.time.LocalDateTime localDateTime = null;
            if (timeObj instanceof java.time.LocalDateTime) {
                localDateTime = (java.time.LocalDateTime) timeObj;
            } else if (timeObj instanceof java.sql.Timestamp) {
                localDateTime = ((java.sql.Timestamp) timeObj).toLocalDateTime();
            } else {
                String timeStr = timeObj.toString().replace(" ", "T");
                if (timeStr.length() > 19) {
                    timeStr = timeStr.substring(0, 19);
                }
                localDateTime = java.time.LocalDateTime.parse(timeStr);
            }
            // 强制指定东八区
            return localDateTime.atZone(java.time.ZoneId.of("Asia/Shanghai"))
                    .format(java.time.format.DateTimeFormatter.RFC_1123_DATE_TIME);
        } catch (Exception e) {
            // 解析失败时回退到原始字符串
            return timeObj.toString();
        }
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
