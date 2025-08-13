package com.example.blog_api.service.serviceImpl;

import com.example.blog_api.service.HomeService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.service.CallService;
import com.example.common_api.util.FunToUrlUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import javax.xml.transform.Result;
import java.util.*;

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
        result = getHigAuthor("4");
        data.put("higAuthor", result.result);
        return ResultBody.createSuccessResult(data);
    }

    @Override
    public ResultBody getWebsiteStatistics() {
        String sql = "SELECT \n" +
                "    (SELECT COUNT(1) FROM blogInfo) AS ARTICLENUM,\n" +
                "    (SELECT COUNT(1) FROM communityInfo) AS COMMUNITYNUM,\n" +
                "    (SELECT SUM(VIEW_PAGE) FROM blogInfo) AS VIEW_PAGE,\n" +
                "    (SELECT COUNT(1) FROM userInfo) AS USERNUM,\n" +
                "    (select COUNT(1) from loginHistory) as USERLOGINNUM;";
        ResultBody result = callService.callFunOneParams(FunToUrlUtil.selectListUrl, "sql", sql);
        if (result != null && !result.isError) {
            Map<String, Integer> map = (Map<String, Integer>) ((ArrayList) result.result).get(0);
            return ResultBody.createSuccessResult(map);
        } else {
            return ResultBody.createErrorResult("获取网站统计失败");
        }
    }

    @Override
    public ResultBody getHigAuthor(String num) {
        //默认4条
        if (num == null) num = "4";
        //根据blogInfo(文章)表  blogComment(评论)表 计算出前三名优质作者
        String sql = "SELECT \n" +
                "    u.USERCODE,\n" +
                "    COALESCE(info.NAME, '') AS USERNAME,\n" +
                "    COALESCE(b.article_count, 0) AS ARTICLE_COUNT,\n" +
                "    COALESCE(follow.follower_count, 0) AS FOLLOWER_COUNT,\n" +
                "    COALESCE(info.AVATAR, '') AS AVATAR,\n" +
                "    -- 加权评分：文章*1 + 被关注数*2\n" +
                "    COALESCE(b.article_count, 0) * 1 +\n" +
                "    COALESCE(follow.follower_count, 0) * 2 AS TOTAL_SCORE\n" +
                "FROM (\n" +
                "    SELECT usercode FROM blogInfo\n" +
                "    UNION\n" +
                "    SELECT followusercode AS usercode FROM userFollow\n" +
                ") u\n" +
                "LEFT JOIN (\n" +
                "    SELECT usercode , COUNT(*) AS article_count\n" +
                "    FROM blogInfo\n" +
                "    GROUP BY usercode\n" +
                ") b ON u.usercode = b.usercode\n" +
                "LEFT JOIN (\n" +
                "    SELECT followusercode, COUNT(*) AS follower_count\n" +
                "    FROM userFollow\n" +
                "    GROUP BY followusercode\n" +
                ") follow ON u.usercode = follow.followusercode\n" +
                "LEFT JOIN userInfo info ON u.usercode = info.CODE\n" +
                "ORDER BY TOTAL_SCORE DESC\n" +
                "LIMIT " + num + ";\n";

        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params);
        return result;
    }

    @Override
    public ResultBody getAllBlogGuids() {
        String sql = "select GUID from blogInfo";
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params);
        List<String> guids = new ArrayList<>();
        for (Map<String, Object> item : (List<Map<String, Object>>) result.result) {
            guids.add((String) item.get("GUID"));
        }
        return ResultBody.createSuccessResult(guids);
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
        String sql = "SELECT \n" +
                "    b.*,\n" +
                "    IFNULL(c.COMMENT_COUNT, 0) AS COMMENT_COUNT,\n" +
                "    IFNULL(gl.LIKE_COUNT, 0) AS LIKE_COUNT,\n" +
                "    IFNULL(gl.COLLECT_COUNT, 0) AS COLLECT_COUNT,\n" +
                "    b.VIEW_PAGE,\n" +
                "    b.CREATE_TIME,\n" +
                "    (\n" +
                "        IFNULL(c.COMMENT_COUNT, 0) * 5\n" +
                "      + b.VIEW_PAGE * 1\n" +
                "      + IFNULL(gl.LIKE_COUNT, 0) * 3\n" +
                "      + IFNULL(gl.COLLECT_COUNT, 0) * 4\n" +
                "      - (UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(b.CREATE_TIME)) / 3600 * 0.2\n" +
                "    ) AS HOT_SCORE\n" +
                "FROM blogInfo b\n" +
                "LEFT JOIN (\n" +
                "    SELECT blogId, COUNT(DISTINCT guid) AS COMMENT_COUNT\n" +
                "    FROM blogComment\n" +
                "    GROUP BY blogId\n" +
                ") c ON b.guid = c.blogId\n" +
                "LEFT JOIN (\n" +
                "    SELECT \n" +
                "        blogid,\n" +
                "        SUM(CASE WHEN type = 'like' THEN 1 ELSE 0 END) AS LIKE_COUNT,\n" +
                "        SUM(CASE WHEN type = 'collect' THEN 1 ELSE 0 END) AS COLLECT_COUNT\n" +
                "    FROM bloggivelike\n" +
                "    GROUP BY blogid\n" +
                ") gl ON b.guid = gl.blogid\n" +
                "WHERE b.blog_type = 'public'\n" +
                "ORDER BY HOT_SCORE DESC\n" +
                "LIMIT 10;\n";
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params);
        return result;
    }
}
