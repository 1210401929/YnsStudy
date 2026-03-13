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
}
