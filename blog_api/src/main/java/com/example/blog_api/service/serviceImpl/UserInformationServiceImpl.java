package com.example.blog_api.service.serviceImpl;

import com.example.blog_api.service.UserInformationService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import com.example.common_api.service.CallService;
import com.example.common_api.util.FunToUrlUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import javax.servlet.http.HttpSession;
import java.util.*;

@Component
public class UserInformationServiceImpl implements UserInformationService {

    @Autowired
    CallService callService;

    @Override
    public ResultBody followUser(String followUserCode, String followUserName, HttpSession session) {
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        List<Map<String, Object>> data = new ArrayList<>();
        Map<String, Object> map = new HashMap<>();
        map.put("USERCODE", userBean.getCODE());
        map.put("USERNAME", userBean.getNAME());
        map.put("FOLLOWUSERCODE", followUserCode);
        map.put("FOLLOWUSERNAME", followUserName);
        data.add(map);
        Map<String, Object> params = new HashMap<>();
        params.put("saveType", "add");
        params.put("tableName", "userFollow");
        params.put("data", data);
        params.put("key", "GUID");
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody noFollowUser(String followUserCode, HttpSession session) {
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        String userCode = userBean.getCODE();
        String sql = "delete from userFollow where followUserCode = ? and userCode = ?";
        List<Object> listParam = Arrays.asList(followUserCode, userCode);
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        params.put("params", listParam);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody getFollowUser(String userCode) {
        Map<String, Object> res = new HashMap<>();
        //关注用户
        String sql = "SELECT CODE\n" +
                "  ,\n" +
                "  NAME,\n" +
                "  role,\n" +
                "  avatar \n" +
                "FROM\n" +
                "  userInfo u\n" +
                "  LEFT JOIN userFollow f ON u.code = f.FOLLOWUSERCODE \n" +
                "WHERE\n" +
                "  f.USERCODE = ? order by f.create_time";
        List<Object> listParams = Arrays.asList(userCode);
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        params.put("params", listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        if (result != null && !result.isError) {
            res.put("followingUser", result.result);
        }
        //粉丝用户
        sql = "SELECT CODE\n" +
                "  ,\n" +
                "  NAME,\n" +
                "  role,\n" +
                "  avatar \n" +
                "FROM\n" +
                "  userInfo u\n" +
                "  LEFT JOIN userFollow f ON u.code = f.USERCODE \n" +
                "WHERE\n" +
                "  f.FOLLOWUSERCODE = ? order by f.create_time";
        params.put("sql", sql);
        result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        if (result != null && !result.isError) {
            res.put("followersUser", result.result);
        }

        return ResultBody.createSuccessResult(res);
    }

    @Override
    public ResultBody getBlogAndResourceByUserCode(String userCode) {
        Map<String, Object> data = new HashMap<>();
        ResultBody result = getBlogByUserCode(userCode);
        data.put("blog", result.result);
        result = getCommunityByUserCode(userCode);
        data.put("community", result.result);
        result = getResourceByUserCode(userCode);
        data.put("resource", result.result);
        return ResultBody.createSuccessResult(data);
    }

    @Override
    public ResultBody getBlogAndCommunityByUserCode(String userCode, int page, int pageSize, String keyWord) {
        int offset = (page - 1) * pageSize;

        // 1. 构建查询 SQL 和参数
        String sql = "SELECT GUID, BLOG_TITLE, MAINTEXT, USERCODE, USERNAME, CREATE_TIME, 'blog' AS TYPE " +
                "FROM bloginfo WHERE USERCODE = ? and blog_type = 'public' AND (BLOG_TITLE LIKE ? OR MAINTEXT LIKE ?) " +
                "UNION ALL " +
                "SELECT GUID, TITLE AS BLOG_TITLE, TEXT AS MAINTEXT, USERCODE, USERNAME, CREATE_TIME, 'community' AS TYPE " +
                "FROM communityinfo WHERE USERCODE = ? AND (TITLE LIKE ? OR TEXT LIKE ?) " +
                "ORDER BY CREATE_TIME DESC LIMIT ? OFFSET ?";

        List<Object> params = new ArrayList<>();
        params.add(userCode);      // USERCODE for blog
        params.add("%" + keyWord + "%");  // LIKE keyword for blog title
        params.add("%" + keyWord + "%");  // LIKE keyword for blog content
        params.add(userCode);      // USERCODE for community
        params.add("%" + keyWord + "%");  // LIKE keyword for community title
        params.add("%" + keyWord + "%");  // LIKE keyword for community content
        params.add(pageSize);      // LIMIT
        params.add(offset);        // OFFSET

        // 2. 构建总数查询 SQL 和参数
        String countSql = "SELECT COUNT(*) AS total FROM ( " +
                "SELECT 1 FROM bloginfo WHERE USERCODE = ? AND (BLOG_TITLE LIKE ? OR MAINTEXT LIKE ?) " +
                "UNION ALL " +
                "SELECT 1 FROM communityinfo WHERE USERCODE = ? AND (TITLE LIKE ? OR TEXT LIKE ?) " +
                ") AS combined";

        List<Object> countParams = new ArrayList<>();
        countParams.add(userCode);  // USERCODE for blog
        countParams.add("%" + keyWord + "%");  // LIKE keyword for blog title
        countParams.add("%" + keyWord + "%");  // LIKE keyword for blog content
        countParams.add(userCode);  // USERCODE for community
        countParams.add("%" + keyWord + "%");  // LIKE keyword for community title
        countParams.add("%" + keyWord + "%");  // LIKE keyword for community content

        // 3. 查询数据列表
        Map<String, Object> paramsMap = new HashMap<>();
        paramsMap.put("sql", sql);
        paramsMap.put("params", params);
        ResultBody listResult = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, paramsMap);
        if (listResult == null || listResult.isError) {
            return listResult;
        }

        // 4. 查询总数
        Map<String, Object> countParamsMap = new HashMap<>();
        countParamsMap.put("sql", countSql);
        countParamsMap.put("params", countParams);
        ResultBody countResult = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, countParamsMap);
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
    public ResultBody getResourceByUserCode(String userCode) {
        String sql = "select * from fileInfo where USERCODE = ?  order by create_time DESC";
        List<Object> listParam = Arrays.asList(userCode);
        Map<String, Object> params_ = new HashMap<>();
        params_.put("sql", sql);
        params_.put("params", listParam);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params_);
        return result;
    }

    @Override
    public ResultBody getBlogByUserCode(String userCode) {
        String sql = "select * from blogInfo where USERCODE = ? and blog_type = 'public' order by create_time DESC";
        List<Object> listParam = Arrays.asList(userCode);
        Map<String, Object> params_ = new HashMap<>();
        params_.put("sql", sql);
        params_.put("params", listParam);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params_);
        return result;
    }

    @Override
    public ResultBody setPersonInfo(String userCode, String fieldName, String fieldValue) {

        String saveType = "add";
        Map<String, Object> map = new HashMap<>();
        map.put("USERCODE", userCode);
        map.put(fieldName, fieldValue);

        String sql = "select GUID from personInfo where userCode = '" + userCode + "'";
        ResultBody result = callService.callFunOneParams(FunToUrlUtil.selectListUrl, "sql", sql);
        //如果存在记录   update
        if (result != null && !result.isError && ((ArrayList) result.result).size() > 0) {
            saveType = "edit";
            map.put("GUID",((LinkedHashMap) ((ArrayList) result.result).get(0)).get("GUID"));
        }

        List<Map<String, Object>> data = new ArrayList<>();
        data.add(map);
        Map<String, Object> params = new HashMap<>();
        params.put("saveType", saveType);
        params.put("tableName", "personInfo");
        params.put("data", data);
        params.put("key", "GUID");
        result = callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody getPersonInfo(String userCode) {
        String sql = "select * from personInfo where userCode = '" + userCode + "'";
        ResultBody result = callService.callFunOneParams(FunToUrlUtil.selectListUrl, "sql", sql);
        return result;
    }

    @Override
    public ResultBody getCommunityByUserCode(String userCode) {
        String sql = "select * from communityInfo where USERCODE = ?  order by create_time DESC";
        List<Object> listParam = Arrays.asList(userCode);
        Map<String, Object> params_ = new HashMap<>();
        params_.put("sql", sql);
        params_.put("params", listParam);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params_);
        return result;
    }
}
