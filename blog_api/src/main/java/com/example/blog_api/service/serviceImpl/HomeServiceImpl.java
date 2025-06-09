package com.example.blog_api.service.serviceImpl;

import com.example.blog_api.service.HomeService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.service.CallService;
import com.example.common_api.util.FunToUrlUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import javax.xml.transform.Result;
import java.util.HashMap;
import java.util.Map;

@Component
public class HomeServiceImpl implements HomeService {
    @Autowired
    CallService callService;

    @Override
    public ResultBody getHomeData() {
        Map<String, Object> data = new HashMap<>();
        //获取热门文章
        ResultBody result = getHotBlogData();
        data.put("hotBlogData", result.result);
        result = getHotFileData();
        data.put("hotFileData", result.result);
        return ResultBody.createSuccessResult(data);
    }

    @Override
    public ResultBody getHigAuthor() {
        //根据blogInfo(文章)表  blogComment(评论)表 计算出前三名优质作者
        String sql = "SELECT \n" +
                "    u.USERCODE,\n" +
                "    COALESCE(b.username, c.username) AS USERNAME,\n" +
                "    COALESCE(b.article_count, 0) AS ARTICLE_COUNT,\n" +
                "    COALESCE(c.comment_count, 0) AS COMMENT_COUNT,\n" +
                "    COALESCE(b.article_count, 0) + COALESCE(c.comment_count, 0) AS TOTAL_SCORE\n" +
                "FROM (\n" +
                "    SELECT usercode FROM blogInfo\n" +
                "    UNION\n" +
                "    SELECT usercode FROM blogComment\n" +
                ") u\n" +
                "LEFT JOIN (\n" +
                "    SELECT usercode, username, COUNT(*) AS article_count\n" +
                "    FROM blogInfo\n" +
                "    GROUP BY usercode, username\n" +
                ") b ON u.usercode = b.usercode\n" +
                "LEFT JOIN (\n" +
                "    SELECT usercode, username, COUNT(*) AS comment_count\n" +
                "    FROM blogComment\n" +
                "    GROUP BY usercode, username\n" +
                ") c ON u.usercode = c.usercode\n" +
                "ORDER BY total_score DESC\n" +
                "LIMIT 3;\n";

        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params);
        return result;
    }

    //获取热门下载内容
    private ResultBody getHotFileData() {
        String sql = "select * from fileinfo order by DOWNNUM DESC limit 5";
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params);
        return result;
    }

    //获取热门文章
    private ResultBody getHotBlogData() {
        //该sql根据 (CREATE_TIME创建时间 COMMENT_COUNT评论数 VIEW_PAGE浏览数) 这三个字段计算权重  得出热门文章 最多五篇文章
        String sql =
                "SELECT " +
                        "    b.*, " +
                        "    COUNT(c.guid) AS COMMENT_COUNT, " +
                        "    b.VIEW_PAGE, " +
                        "    b.CREATE_TIME, " +
                        "    ( " +
                        "        (COUNT(c.guid) * 5) + " +
                        "        (b.VIEW_PAGE * 1) - " +
                        "        (UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(b.CREATE_TIME)) / 3600 * 0.2 " +
                        "    ) AS HOT_SCORE " +
                        "FROM blogInfo b " +
                        "LEFT JOIN blogComment c ON b.guid = c.blogId " +
                        "WHERE b.blog_type = 'public' " +
                        "GROUP BY b.guid " +
                        "ORDER BY HOT_SCORE DESC " +
                        "LIMIT 5";
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params);
        return result;
    }
}
