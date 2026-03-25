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
        Map<String, Object> blogData = BeanUtil.beanToMap(blogBean);
        //如果没有传递GUID,则根据记录表(bloginfo_guid_sequence) 记录最大条数 +1 后赋值给guid
        if (blogData.get("GUID") == null) {
            String selectSql = "select TABLE_NAME,CURRENT_VALUE from bloginfo_guid_sequence";
            Map<String, Object> params_ = new HashMap<>();
            params_.put("sql", selectSql);
            ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params_);
            // 1. 先安全地拿到 List
            List<Map<String, Object>> list = (List<Map<String, Object>>) result.result;
            if (list != null && !list.isEmpty()) {
                // 2. 拿到 CURRENT_VALUE 的 Object 值
                Object currentValueObj = list.get(0).get("CURRENT_VALUE");
                // 3. 转换为 long 并 +1 (用 String.valueOf 兜底，防止直接强转报错)
                long nextGuid = Long.parseLong(String.valueOf(currentValueObj)) + 1L;
                // 4. 放入 blogData
                blogData.put("GUID", nextGuid);
            }
            //把记录表的记录值加1
            String updateSql = "UPDATE bloginfo_guid_sequence SET current_value = current_value + 1 WHERE table_name = 'bloginfo'";
            //对于使用访问网关  推荐先接收返回结果,再返回,否则会有未知问题
            result = callService.callFunOneParams(FunToUrlUtil.exeSqlUrl, "sql", updateSql);
        }
        //通过封装的bean映射,把bean转换为map类型
        data.add(blogData);
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
            ResultBody result = getUserBlogByUserCode(userBean.getCODE(), null);
            return result;
        }
    }

    @Override
    public ResultBody getUserBlogByUserCode(String userCode, String blogType) {
        Boolean isAll = blogType == null ? true : false;
        if (userCode == null) {
            return ResultBody.createErrorResult("用户编码出错!");
        } else {
            String sql = "select GUID,BLOG_TITLE,BLOG_TYPE,USERCODE,USERNAME,VIEW_PAGE,CREATE_TIME,CAT_ID from blogInfo " +
                    "where userCode = '" + userCode + "' ";
            //如果不是查询该用户所有文章,则根据文章类型过滤    私密/公开
            if (!isAll) {
                sql += "and blog_type = 'public'";
            }
            //排序
            sql += "order by (case BLOG_TYPE when 'public' then 0 when 'privacy' then 1 else 2 end), create_time DESC";
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
    public ResultBody getUserBlogTitle(HttpSession session) {
        return null;
    }

    @Override
    public ResultBody getUserBlogNum(HttpSession session) {
        return null;
    }

    @Override
    public ResultBody getBlogTitleByCatId(String catId) {
        return null;
    }

    @Override
    public ResultBody updateBlogCatId(String blogId, String catId) {
        String updateSql = "update blogInfo set CAT_ID = ? where GUID = ?";
        List<Object> listParmas = Arrays.asList(catId, blogId);
        Map<String, Object> params = new HashMap<>();
        params.put("sql", updateSql);
        params.put("params", listParmas);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody getBlog(HttpSession session, String blogId) {
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        String userCode = userBean == null ? "" : userBean.getCODE();
        //这里需要判断  只允许查到该用户的文章,或者文章是公开的
        //String sql = "select * from blogInfo where guid = '" + blogId + "' and (userCode ='" + userCode + "' or blog_type = 'public')";
        String sql = "SELECT b.*, u.AVATAR " +
                "FROM blogInfo b " +
                "LEFT JOIN userinfo u ON b.userCode = u.code " +
                "WHERE b.guid = ? AND (b.userCode = ? OR b.blog_type = 'public')";
        List<Object> listParam = Arrays.asList(blogId, userCode);
        Map<String, Object> params_ = new HashMap<>();
        params_.put("sql", sql);
        params_.put("params", listParam);
        //对于使用访问网关  推荐先接收返回结果,再返回,否则会有未知问题
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params_);
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
            listSql = "SELECT b.*, u.AVATAR FROM blogInfo b " +
                    "LEFT JOIN userinfo u ON b.USERCODE = u.code " +
                    "WHERE (b.blog_title LIKE ? OR b.mainText LIKE ?) and b.blog_type = 'public'" +
                    "ORDER BY b.create_time DESC LIMIT ? OFFSET ?";
            listParams.add("%" + keyWord + "%");
            listParams.add("%" + keyWord + "%");
        } else {
            listSql = "SELECT b.*, u.AVATAR FROM blogInfo b " +
                    "LEFT JOIN userinfo u ON b.USERCODE = u.code where b.blog_type = 'public'" +
                    "ORDER BY b.create_time DESC LIMIT ? OFFSET ?";
        }
        listParams.add(pageSize);
        listParams.add(offset);

        // 2. 构建总数查询 SQL 和参数
        String countSql;
        List<Object> countParams = new ArrayList<>();
        if (keyWord != null && !keyWord.trim().isEmpty()) {
            countSql = "SELECT COUNT(*) AS total FROM blogInfo b " +
                    "LEFT JOIN userinfo u ON b.USERCODE = u.code " +
                    "WHERE (b.blog_title LIKE ? OR b.mainText LIKE ?) and b.blog_type = 'public'";
            countParams.add("%" + keyWord + "%");
            countParams.add("%" + keyWord + "%");
        } else {
            countSql = "SELECT COUNT(*) AS total FROM blogInfo where blog_type = 'public'";
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

        // 5. 获取总数
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
    public ResultBody updateBlog(String guid, String title, String content, String blog_type) {
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
            System.out.println("修改文章内容=>需要删除的图片:" + removedUrls);
            Map<String, Object> params = new HashMap<>();
            params.put("urls", removedUrls);
            //发送网关请求,删除图片
            result = callService.callFunWithParams(FunToUrlUtil.deleteFileByUrlsUrl, params);
            //删除图片后,修改文章主体内容
            if (result != null && !result.isError) {
                sql = "update blogInfo set blog_title = ? , mainText = ?,blog_type = ? where guid = ?";
                List<Object> listParams = Arrays.asList(title, content, blog_type, guid);
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
                    //根据博客ID,删除关于该博客的所有点赞和收藏
                    deleteLikeAndCollectByBlogId(guid);
                    return ResultBody.createSuccessResult(result.result);
                } else {
                    return ResultBody.createErrorResult("删除异常!");
                }
            }
        }
        return ResultBody.createErrorResult("删除异常!");
    }

    @Override
    public ResultBody addBlogCat(Map<String, String> blogCat) {
        Map<String, Object> params = new HashMap<>();
        List<Map<String, String>> data = new ArrayList<>();
        data.add(blogCat);
        params.put("saveType", "add");
        params.put("tableName", "BLOGCATINFO");
        params.put("data", data);
        params.put("key", "GUID");
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody updateBlogCat(Map<String, String> blogCat) {
        Map<String, Object> params = new HashMap<>();
        List<Map<String, String>> data = new ArrayList<>();
        data.add(blogCat);
        params.put("saveType", "edit");
        params.put("tableName", "BLOGCATINFO");
        params.put("data", data);
        params.put("key", "GUID");
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl, params);
        return result;
    }

    public ResultBody deleteBlogCat(String guid) {
        // 准备 SQL 列表
        List<String> sqls = Arrays.asList(
                //删除前，先把属于该分类以及下级分类的文章的分类id清除
                "update blogInfo set cat_id = null where cat_id in (select guid from BLOGCATINFO where guid = ? or superguid = ?)",
                //删除分类sql
                "delete from BLOGCATINFO where guid = ? or superguid = ?"
        );

        // 准备对应的参数列表
        List<List<Object>> paramsList = Arrays.asList(
                Arrays.asList(guid, guid), // 第一条 SQL 的参数
                Arrays.asList(guid, guid)  // 第二条 SQL 的参数
        );

        // 封装成 Map 传给网关
        Map<String, Object> payload = new HashMap<>();
        payload.put("sqls", sqls);
        payload.put("params", paramsList);

        // 调用服务端新增的方法
        return callService.callFunWithParams(FunToUrlUtil.exeSqlListByParamsUrl, payload);
    }

    @Override
    public ResultBody getUserBlogCat(String userCode) {

        String sql = "select * from blogCatInfo " +
                "where userCode = '" + userCode + "' " +
                "order by orderNo ";
        Map<String, Object> params_ = new HashMap<>();
        params_.put("sql", sql);
        //对于使用访问网关  推荐先接收返回结果,再返回,否则会有未知问题
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params_);
        if (result != null && !result.isError) {
            return result;
        } else {
            return ResultBody.createErrorResult("查询异常!");
        }

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
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody getComment(String blogId) {
        // 通过 LEFT JOIN 关联同一张表，获取父评论（主评论）的时间信息用于排序
        String sql = "SELECT \n" +
                "  bc.*, \n" +
                "  ui.AVATAR \n" +
                "FROM \n" +
                "  BLOGCOMMENT bc\n" +
                "LEFT JOIN \n" +
                "  userinfo ui \n" +
                "ON \n" +
                "  bc.usercode = ui.code\n" +
                "LEFT JOIN \n" +
                "  BLOGCOMMENT p \n" + // 关联父评论
                "ON \n" +
                "  bc.SUPERGUID = p.GUID\n" +
                "WHERE \n" +
                "  bc.blogId = '" + blogId + "'\n" +
                "ORDER BY \n" +
                // 1. 把主评论和它的所有子评论作为一个“组”，按主评论的时间倒序排（最新的话题在最上）
                "  COALESCE(p.create_time, bc.create_time) DESC, \n" +
                // 2. 确保在一个“组”内，主评论永远排在第一位（0在前，1在后）
                "  CASE WHEN bc.SUPERGUID IS NULL OR bc.SUPERGUID = '' THEN 0 ELSE 1 END ASC, \n" +
                // 3. 子评论之间，按照它们自己的发布时间正序排（像正常聊天一样从上往下读）
                "  bc.create_time ASC";

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

    @Override
    public ResultBody giveLikeBlog(String blogId, HttpSession session) {
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        List<Map<String, Object>> data = new ArrayList<>();
        Map<String, Object> map = new HashMap<>();
        map.put("BLOGID", blogId);
        map.put("USERCODE", userBean.getCODE());
        map.put("USERNAME", userBean.getNAME());
        map.put("TYPE", "like");
        data.add(map);
        Map<String, Object> params = new HashMap<>();
        params.put("saveType", "add");
        params.put("tableName", "blogGiveLike");
        params.put("data", data);
        params.put("key", "GUID");
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody noGiveLikeBlog(String blogId, HttpSession session) {
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        String userCode = userBean.getCODE();
        String sql = "delete from blogGiveLike where blogId = ? and userCode = ? and type = 'like'";
        List<Object> listParams = Arrays.asList(blogId, userCode);
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        params.put("params", listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody getGiveLikeByBlogId(String blogId) {
        String sql = "select * from blogGiveLike where blogId = ? and type = 'like'";
        List<Object> listParams = Arrays.asList(blogId);
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        params.put("params", listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody collectBlog(String blogId, HttpSession session) {
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        List<Map<String, Object>> data = new ArrayList<>();
        Map<String, Object> map = new HashMap<>();
        map.put("BLOGID", blogId);
        map.put("USERCODE", userBean.getCODE());
        map.put("USERNAME", userBean.getNAME());
        map.put("TYPE", "collect");
        data.add(map);
        Map<String, Object> params = new HashMap<>();
        params.put("saveType", "add");
        params.put("tableName", "blogGiveLike");
        params.put("data", data);
        params.put("key", "GUID");
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody noCollectBlog(String blogId, HttpSession session) {
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        String userCode = userBean.getCODE();
        String sql = "delete from blogGiveLike where blogId = ? and userCode = ? and type = 'collect'";
        List<Object> listParams = Arrays.asList(blogId, userCode);
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        params.put("params", listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody getCollectByBlogId(String blogId) {
        String sql = "select * from blogGiveLike where blogId = ? and type = 'collect'";
        List<Object> listParams = Arrays.asList(blogId);
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        params.put("params", listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody getCollectByUserCode(String userCode) {
        String sql = "select * from blogGiveLike where userCode = ? and type = 'collect'";
        List<Object> listParams = Arrays.asList(userCode);
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        params.put("params", listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody getLikeAndCollectByBlogId(String blogId) {
        String sql = "select * from blogGiveLike where blogId = ?";
        List<Object> listParams = Arrays.asList(blogId);
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        params.put("params", listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody getLikeAndCollectByUserCode(String userCode, String isCountOnly, String type) {
        Map<String, Object> params = new HashMap<>();
        List<Object> listParams = new ArrayList<>();
        listParams.add(userCode);

        // 判断是否仅查询数字
        boolean countOnly = "true".equalsIgnoreCase(isCountOnly);

        if (countOnly) {
            // ==========================================================
            // 模式 1：仅获取数量（极速响应，不连表）
            // ==========================================================
            // 直接根据 TYPE (通常是 'like' 和 'collect') 分组统计数量
            String sql = "SELECT TYPE, COUNT(1) AS TOTAL FROM blogGiveLike WHERE userCode = ? GROUP BY TYPE";
            params.put("sql", sql);
            params.put("params", listParams);
            ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
            // 返回的结果通常是一个包含 TYPE 和 TOTAL 的 List 集合，例如：
            // [{"TYPE": "like", "TOTAL": 15}, {"TYPE": "collect", "TOTAL": 8}]
            return result;

        } else {
            // ==========================================================
            // 模式 2：获取完整内容（原逻辑保持不变）
            // ==========================================================
            String sql = "SELECT\n" +
                    "  b.GUID,b.BLOG_TITLE,b.BLOG_TYPE,b.USERCODE,b.USERNAME,b.VIEW_PAGE,b.CREATE_TIME,g.TYPE\n" +
                    "FROM\n" +
                    "  bloginfo b\n" +
                    "  JOIN blogGiveLike g ON g.BLOGID = b.GUID\n" + // 这里最好用 JOIN，因为是从 g 表过滤的
                    "WHERE\n" +
                    "  g.userCode = ?";
            if (type != null) {
                sql += " AND g.TYPE = ?";
                listParams.add(type);
            }
            params.put("sql", sql);
            params.put("params", listParams);
            ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
            return result;
        }
    }

    //根据博客ID,删除关于该博客的所有点赞和收藏
    private ResultBody deleteLikeAndCollectByBlogId(String blogId) {
        String sql = "delete from blogGiveLike where blogId = ?";
        List<Object> listParams = Arrays.asList(blogId);
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
