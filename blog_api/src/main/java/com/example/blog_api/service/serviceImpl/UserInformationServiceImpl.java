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
        List<Map<String,Object>> data = new ArrayList<>();
        Map<String,Object> map = new HashMap<>();
        map.put("USERCODE",userBean.getCODE());
        map.put("USERNAME",userBean.getNAME());
        map.put("FOLLOWUSERCODE",followUserCode);
        map.put("FOLLOWUSERNAME",followUserName);
        data.add(map);
        Map<String,Object> params = new HashMap<>();
        params.put("saveType","add");
        params.put("tableName","userFollow");
        params.put("data",data);
        params.put("key","GUID");
        ResultBody  result = callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl,params);
        return result;
    }

    @Override
    public ResultBody noFollowUser(String followUserCode, HttpSession session) {
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        String userCode = userBean.getCODE();
        String sql = "delete from userFollow where followUserCode = ? and userCode = ?";
        List<Object> listParam = Arrays.asList(followUserCode,userCode);
        Map<String,Object> params = new HashMap<>();
        params.put("sql",sql);
        params.put("params",listParam);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl,params);
        return result;
    }

    @Override
    public ResultBody getFollowUser(String userCode) {
        Map<String,Object> res = new HashMap<>();
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
        Map<String,Object> params = new HashMap<>();
        params.put("sql",sql);
        params.put("params",listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        if(result!=null && !result.isError){
            res.put("followingUser",result.result);
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
        params.put("sql",sql);
        result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        if(result!=null && !result.isError){
            res.put("followersUser",result.result);
        }

        return  ResultBody.createSuccessResult(res);
    }

    @Override
    public ResultBody getBlogAndResourceByUserCode(String userCode) {
        Map<String,Object> data = new HashMap<>();
        ResultBody result = getBlogByUserCode(userCode);
        data.put("blog",result.result);
        result = getCommunityByUserCode(userCode);
        data.put("community",result.result);
        result = getResourceByUserCode(userCode);
        data.put("resource",result.result);
        return ResultBody.createSuccessResult(data);
    }



    @Override
    public ResultBody getResourceByUserCode(String userCode) {
        String sql = "select * from fileInfo where USERCODE = ?  order by create_time DESC";
        List<Object> listParam = Arrays.asList(userCode);
        Map<String,Object> params_ = new HashMap<>();
        params_.put("sql",sql);
        params_.put("params",listParam);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params_);
        return result;
    }

    @Override
    public ResultBody getBlogByUserCode(String userCode) {
        String sql = "select * from blogInfo where USERCODE = ? and blog_type = 'public' order by create_time DESC";
        List<Object> listParam = Arrays.asList(userCode);
        Map<String,Object> params_ = new HashMap<>();
        params_.put("sql",sql);
        params_.put("params",listParam);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params_);
        return result;
    }
    @Override
    public ResultBody getCommunityByUserCode(String userCode) {
        String sql = "select * from communityInfo where USERCODE = ?  order by create_time DESC";
        List<Object> listParam = Arrays.asList(userCode);
        Map<String,Object> params_ = new HashMap<>();
        params_.put("sql",sql);
        params_.put("params",listParam);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params_);
        return result;
    }
}
