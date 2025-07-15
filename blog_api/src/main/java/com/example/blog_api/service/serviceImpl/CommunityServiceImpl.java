package com.example.blog_api.service.serviceImpl;

import com.example.blog_api.service.CommunityService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import com.example.common_api.service.CallService;
import com.example.common_api.util.FunToUrlUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import javax.servlet.http.HttpSession;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Component
public class CommunityServiceImpl implements CommunityService {
    @Autowired
    CallService callService;

    @Override
    public ResultBody addCommunity(Map<String, Object> community, HttpSession session) {
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        if (userBean == null) {
            return ResultBody.createErrorResult("用户未登录!");
        }
        Map<String, Object> params = new HashMap<>();
        List<Map<String, Object>> data = new ArrayList<>();
        data.add(community);
        params.put("saveType", "add");
        params.put("tableName", "communityInfo");
        params.put("data", data);
        params.put("key", "GUID");
        return callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl, params);
    }

    @Override
    public ResultBody setTopCommunity(String CommentGuid, String isTop) {
        String sql;
        if (isTop != null && isTop.equals("0")) {
            sql = "update communityInfo set isTop = null where GUID = '" + CommentGuid + "'";
        } else if (isTop != null && isTop.equals("1")) {
            sql = "update communityInfo set isTop = '1' where GUID = '" + CommentGuid + "'";
        } else {
            sql = "update communityInfo set isTop = isTop where GUID = '" + CommentGuid + "'";
        }
        ResultBody result = callService.callFunOneParams(FunToUrlUtil.exeSqlUrl, "sql", sql);
        return result;
    }

    @Override
    @Transactional//开启事务
    public ResultBody deleteCommunity(String CommentGuid) {
        //删除主内容
        String delInfoSql = "delete from communityinfo where guid = '" + CommentGuid + "'";
        //删除相关评论
        String sqlCommentSql = "delete from communitycomment where communityId = '" + CommentGuid + "'";
        callService.callFunOneParams(FunToUrlUtil.exeSqlUrl, "sql", delInfoSql);
        callService.callFunOneParams(FunToUrlUtil.exeSqlUrl, "sql", sqlCommentSql);
        return ResultBody.createSuccessResult("删除成功");
    }

    @Override
    public ResultBody getCommunity(String guid) {
        return null;
    }

    @Override
    public ResultBody getAllCommunity(int page, int pageSize, String keyWord) {
        int offset = (page - 1) * pageSize;

        String listSql;
        List<Object> listParams = new ArrayList<>();

        if (keyWord != null && !keyWord.trim().isEmpty()) {
            listSql = "SELECT b.*, u.AVATAR FROM communityInfo b " +
                    "LEFT JOIN userinfo u ON b.USERCODE = u.code " +
                    "WHERE b.title LIKE ? OR b.text LIKE ? " +
                    "ORDER BY b.ISTOP DESC, b.create_time DESC LIMIT ? OFFSET ?";
            listParams.add("%" + keyWord + "%");
            listParams.add("%" + keyWord + "%");
        } else {
            listSql = "SELECT b.*, u.AVATAR FROM communityInfo b " +
                    "LEFT JOIN userinfo u ON b.USERCODE = u.code " +
                    "ORDER BY b.ISTOP DESC, b.create_time DESC LIMIT ? OFFSET ?";
        }
        listParams.add(pageSize);
        listParams.add(offset);

        // 查询总数 SQL
        String countSql;
        List<Object> countParams = new ArrayList<>();
        if (keyWord != null && !keyWord.trim().isEmpty()) {
            countSql = "SELECT COUNT(*) AS total FROM communityInfo b " +
                    "LEFT JOIN userinfo u ON b.USERCODE = u.code " +
                    "WHERE b.TITLE LIKE ? OR b.TEXT LIKE ?";
            countParams.add("%" + keyWord + "%");
            countParams.add("%" + keyWord + "%");
        } else {
            countSql = "SELECT COUNT(*) AS total FROM communityInfo";
        }

        // 查询列表
        Map<String, Object> params = new HashMap<>();
        params.put("sql", listSql);
        params.put("params", listParams);
        ResultBody listResult = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        if (listResult == null || listResult.isError) {
            return listResult;
        }

        // 查询总数
        params = new HashMap<>();
        params.put("sql", countSql);
        params.put("params", countParams);
        ResultBody countResult = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        if (countResult == null || countResult.isError) {
            return countResult;
        }

        // 解析总数
        int total = 0;
        List<Map<String, Object>> countData = (List<Map<String, Object>>) countResult.result;
        if (!countData.isEmpty() && countData.get(0).get("total") != null) {
            total = Integer.parseInt(countData.get(0).get("total").toString());
        }

        // 组装结果
        Map<String, Object> result = new HashMap<>();
        result.put("total", total);
        result.put("data", listResult.result);

        return ResultBody.createSuccessResult(result);
    }


    @Override
    public ResultBody addComment(Map<String, String> communityComment, HttpSession session) {
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        if (userBean == null) {
            return ResultBody.createErrorResult("用户未登录!");
        }
        Map<String, Object> params = new HashMap<>();
        List<Map<String, String>> data = new ArrayList<>();
        data.add(communityComment);
        params.put("saveType", "add");
        params.put("tableName", "communityComment");
        params.put("data", data);
        params.put("key", "GUID");
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.saveAllTableDataUrl, params);
        return result;
    }

    @Override
    public ResultBody getComment(String communityId) {
        String sql = "SELECT \n" +
                "  bc.*, \n" +
                "  ui.AVATAR \n" +
                "FROM \n" +
                "  communityComment bc\n" +
                "LEFT JOIN \n" +
                "  userinfo ui \n" +
                "ON \n" +
                "  bc.usercode = ui.code\n" +
                "WHERE \n" +
                "  bc.communityId = '" + communityId + "'\n" +
                "ORDER BY \n" +
                "  bc.create_time DESC\n";
        Map<String, Object> params_ = new HashMap<>();
        params_.put("sql", sql);
        //对于使用访问网关  推荐先接收返回结果,再返回,否则会有未知问题
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params_);
        return result;
    }

    @Override
    public ResultBody deleteComment(String CommentGuid) {
        return null;
    }
}
