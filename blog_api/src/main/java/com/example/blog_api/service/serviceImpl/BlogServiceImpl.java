package com.example.blog_api.service.serviceImpl;

import com.example.blog_api.Bean.BlogBean;
import com.example.blog_api.service.BlogService;
import com.example.common_api.util.FunToUrlUtil;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import com.example.common_api.service.CallService;
import com.example.common_api.util.BeanUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import javax.servlet.http.HttpSession;
import javax.xml.transform.Result;
import java.util.*;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Service
public class BlogServiceImpl implements BlogService {

    @Autowired
    CallService callService;

    @Override
    public ResultBody addBlog(BlogBean blogBean) {
        Map<String, Object> params = new HashMap<>();
        List<Map<String, Object>> data = new ArrayList<>();
        //通过封装的bean映射,把bean转换为map类型
        data.add(BeanUtil.beanToMap(blogBean));
        params.put("saveType", "add");
        params.put("tableName", "BLOGINFO");
        params.put("data", data);
        params.put("key", "GUID");
        return callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl, params);
    }

    @Override
    public ResultBody getUserBlog(HttpSession session) {
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        if (userBean == null) {
            return ResultBody.createErrorResult("用户未登录!");
        } else {
            String userCode = userBean.getCODE();
            String sql = "select * from blogInfo where userCode = '" + userCode + "' order by create_time DESC";
            Map<String, Object> params_ = new HashMap<>();
            params_.put("sql", sql);
            //对于使用访问网关  推荐先接收返回结果,再返回,否则会有未知问题
            ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params_);
            if (result != null && !result.isError) {
                return ResultBody.createSuccessResult(result.result);
            } else {
                return ResultBody.createErrorResult("查询异常!");
            }
        }
    }

    @Override
    public ResultBody getBlog(HttpSession session, String blogId) {
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        String userCode = userBean == null ? "" : userBean.getCODE();
        //这里需要判断  只允许查到该用户的文章,或者文章是公开的
        String sql = "select * from blogInfo where guid = '" + blogId + "' and (userCode ='" + userCode + "' or blog_type = 'public')";
        Map<String, Object> params_ = new HashMap<>();
        params_.put("sql", sql);
        //对于使用访问网关  推荐先接收返回结果,再返回,否则会有未知问题
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params_);
        if (result != null && result.result != null && ((ArrayList) result.result).size() > 0) {
            //如果查到有效数据,给该文章增加1浏览量
            String updateSql = "update blogInfo set view_page = view_page + 1 where guid = ?";
            List<Object> listParams = Arrays.asList(blogId);
            Map<String, Object> params = new HashMap<>();
            params.put("sql", updateSql);
            params.put("params", listParams);
            callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl, params);
        }
        return result;
    }

    @Override
    public ResultBody getAllBlog(int page, int pageSize, String keyWord) {

        int offset = (page - 1) * pageSize;
        // 1. 构建主查询 SQL 和参数
        String listSql;
        List<Object> listParams = new ArrayList<>();
        if (keyWord != null && !keyWord.trim().isEmpty()) {
            listSql = "SELECT * FROM blogInfo " +
                    "WHERE blog_title LIKE ? OR mainText LIKE ? " +
                    "ORDER BY create_time DESC LIMIT ? OFFSET ?";
            listParams.add("%" + keyWord + "%");
            listParams.add("%" + keyWord + "%");
        } else {
            listSql = "SELECT * FROM blogInfo " +
                    "ORDER BY create_time DESC LIMIT ? OFFSET ?";
        }
        listParams.add(pageSize);
        listParams.add(offset);

        // 2. 构建总数查询 SQL 和参数
        String countSql;
        List<Object> countParams = new ArrayList<>();
        if (keyWord != null && !keyWord.trim().isEmpty()) {
            countSql = "SELECT COUNT(*) AS total FROM blogInfo WHERE blog_title LIKE ? OR mainText LIKE ?";
            countParams.add("%" + keyWord + "%");
            countParams.add("%" + keyWord + "%");
        } else {
            countSql = "SELECT COUNT(*) AS total FROM blogInfo";
        }

        // 3. 查询数据列表
        Map<String, Object> params = new HashMap<>();
        params.put("sql", listSql);
        params.put("params", listParams);
        ResultBody listResult = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        if (listResult == null || listResult.isError) {
            return listResult;
        }

        // 4. 查询总数
        params = new HashMap<>();
        params.put("sql", countSql);
        params.put("params", countParams);
        ResultBody countResult = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        if (countResult == null || countResult.isError) {
            return countResult;
        }

        // 5. 获取总数（countResult 返回的是 list，每个 map 有个 "total"）
        int total = 0;
        List<Map<String, Object>> countData = (List<Map<String, Object>>) countResult.result;
        if (!countData.isEmpty() && countData.get(0).get("total") != null) {
            total = Integer.parseInt(countData.get(0).get("total").toString());
        }

        // 6. 组装分页结果
        Map<String, Object> result = new HashMap<>();
        result.put("total", total);
        result.put("data", listResult.result);

        return ResultBody.createSuccessResult(result);
    }


    @Override
    @Transactional//开启事务
    public ResultBody updateBlog(String guid, String title, String content) {
        //查询修改前的数据
        String sql = "select MAINTEXT from blogInfo where guid = '" + guid + "'";
        Map<String, Object> params_ = new HashMap<>();
        params_.put("sql", sql);
        //对于使用访问网关  推荐先接收返回结果,再返回,否则会有未知问题
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params_);
        if (result.result != null && ((ArrayList) result.result).size() > 0) {
            //修改文章第一步,需要找到修改的文章,解析出里面的图片
            List<Map<String, String>> resultList = (List<Map<String, String>>) result.result;
            Map<String, String> firstItem = resultList.get(0);
            String html = firstItem.get("MAINTEXT");
            //修改前  文章包含的路径
            List<String> oldListUrl = analyzeHtml2ImageUrl(html);
            //修改后  文章包含的路径
            List<String> newListUrl = analyzeHtml2ImageUrl(content);
            //挑选出需要删除的数据   (修改前内容中的url  在新文章中不存在的url)
            List<String> removedUrls = new ArrayList<>();
            for (String url : oldListUrl) {
                //判断newListUrl是否包含oldListUrl的子项  不包含则拼接到删除urlList中
                if (!newListUrl.contains(url)) {
                    removedUrls.add(url);
                }
            }
            Map<String, Object> params = new HashMap<>();
            params.put("urls", removedUrls);
            //发送网关请求,删除图片
            result = callService.callFunWithParams(FunToUrlUtil.deleteFileByUrlsUrl, params);
            //删除图片后,修改文章主体内容
            if (result != null && !result.isError) {
                sql = "update blogInfo set blog_title = ? , mainText = ? where guid = ?";
                List<Object> listParams = Arrays.asList(title, content, guid);
                params = new HashMap<>();
                params.put("sql", sql);
                params.put("params", listParams);
                //对于使用访问网关  推荐先接收返回结果,再返回,否则会有未知问题
                result = callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl, params);
                if (result != null && !result.isError) {
                    return ResultBody.createSuccessResult(result.result);
                } else {
                    return ResultBody.createErrorResult("修改异常!");
                }
            }
        }
        return ResultBody.createErrorResult("修改异常!");
    }

    @Override
    @Transactional//开启事务
    public ResultBody deleteBlog(String guid) {
        if (guid == null) {
            return ResultBody.createErrorResult("未传入删除主键!");
        }
        //查询删除前的数据
        String sql = "select MAINTEXT from blogInfo where guid = '" + guid + "'";
        Map<String, Object> params_ = new HashMap<>();
        params_.put("sql", sql);
        //对于使用访问网关  推荐先接收返回结果,再返回,否则会有未知问题
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params_);
        if (result.result != null && ((ArrayList) result.result).size() > 0) {
            //删除文章第一步,需要找到需要删除的文章,解析出里面的图片,并删除
            List<Map<String, String>> resultList = (List<Map<String, String>>) result.result;
            Map<String, String> firstItem = resultList.get(0);
            String html = firstItem.get("MAINTEXT");
            //解析出图片路径
            List<String> urls = analyzeHtml2ImageUrl(html);
            Map<String, Object> params = new HashMap<>();
            params.put("urls", urls);
            //发送网关请求,删除图片
            result = callService.callFunWithParams(FunToUrlUtil.deleteFileByUrlsUrl, params);
            //删除图片后,删除文章主体内容
            if (result != null && !result.isError) {
                sql = "delete from blogInfo where guid = '" + guid + "'";
                //对于使用访问网关  推荐先接收返回结果,再返回,否则会有未知问题
                result = callService.callFunOneParams(FunToUrlUtil.exeSqlUrl, "sql", sql);
                if (result != null && !result.isError) {
                    //删除关于该文章的所有评论
                    deleteCommentByBlogId(guid);
                    return ResultBody.createSuccessResult(result.result);
                } else {
                    return ResultBody.createErrorResult("删除异常!");
                }
            }
        }
        return ResultBody.createErrorResult("删除异常!");
    }

    @Override
    public ResultBody addComment(Map<String, String> blogComment) {
        Map<String, Object> params = new HashMap<>();
        List<Map<String, String>> data = new ArrayList<>();
        data.add(blogComment);
        params.put("saveType", "add");
        params.put("tableName", "BLOGCOMMENT");
        params.put("data", data);
        params.put("key", "GUID");
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.saveAllTableDataUrl, params);
        return result;
    }

    @Override
    public ResultBody getComment(String blogId) {
        String sql = "select * from BLOGCOMMENT where blogId = '" + blogId + "' order by create_time DESC";
        Map<String, Object> params_ = new HashMap<>();
        params_.put("sql", sql);
        //对于使用访问网关  推荐先接收返回结果,再返回,否则会有未知问题
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params_);
        return result;
    }

    @Override
    public ResultBody deleteComment(String CommentGuid) {
        String sql = "delete from BLOGCOMMENT where guid = ? or superguid = ?";
        List<Object> listParams = Arrays.asList(CommentGuid, CommentGuid);
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        params.put("params", listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl, params);
        return result;
    }

    //根据博客ID  删除关于该博客的所有评论
    private ResultBody deleteCommentByBlogId(String blogId) {
        String sql = "delete from BLOGCOMMENT where blogId = ?";
        List<Object> listParams = Arrays.asList(blogId);
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        params.put("params", listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl, params);
        return result;
    }

    //解析html 列出包含的图片地址
    private List<String> analyzeHtml2ImageUrl(String html) {
        List<String> srcList = new ArrayList<>();
        // 提取所有 img 标签
        Pattern imgPattern = Pattern.compile("<img\\b[^>]*?\\bsrc\\s*=\\s*\"([^\"]+)\"", Pattern.CASE_INSENSITIVE);
        Matcher matcher = imgPattern.matcher(html);
        while (matcher.find()) {
            srcList.add(matcher.group(1)); // group(1) 是括号中的内容，即 src 的值
        }
        return srcList;
    }
}
