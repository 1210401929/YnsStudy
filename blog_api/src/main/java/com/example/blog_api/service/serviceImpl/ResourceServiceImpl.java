package com.example.blog_api.service.serviceImpl;

import com.example.blog_api.service.ResourceService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import com.example.common_api.service.CallService;
import com.example.common_api.util.FunToUrlUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import javax.servlet.http.HttpSession;
import java.util.*;

@Component
public class ResourceServiceImpl implements ResourceService {
    @Autowired
    CallService callService;

    @Override
    public ResultBody addFileInfo(HttpSession session, Map<String, Object> fileInfo) {
        Map<String, Object> params = new HashMap<>();
        List<Map<String, Object>> data = new ArrayList<>();
        data.add(fileInfo);
        params.put("saveType", "add");
        params.put("tableName", "FILEINFO");
        params.put("data", data);
        params.put("key", "GUID");
        return callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl, params);
    }

    @Override
    @Transactional //开启事务
    public ResultBody delFileInfo(String guid,String url) {
        if(url==null){
            ResultBody.createErrorResult("路径不存在或不正确!");
        }
        String sql = "delete from fileInfo where guid = '"+guid+"'";
        ResultBody result = callService.callFunOneParams(FunToUrlUtil.exeSqlUrl,"sql",sql);
        if(result!=null && !result.isError){
            Map<String,Object> params = new HashMap<>();
            params.put("url",url);
            result = callService.callFunWithParams(FunToUrlUtil.deleteFileByUrlUrl,params);
            return result;
        }else{
            return ResultBody.createErrorResult("删除失败!");
        }
    }

    @Override
    public ResultBody getAllFile() {
        String sql = "select * from fileInfo order by create_time desc";
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params);
        return result;
    }

    @Override
    public ResultBody getFileByUser(UserBean userBean) {
        String userCode = userBean.getCODE();
        String sql = "select * from fileInfo where userCode = '" + userCode + "' order by create_time desc";
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params);
        return result;
    }

    @Override
    public ResultBody setFileDownNum(String guid) {
        String sql = "update fileInfo set downNum = downNum+1 where guid = ?";
        List<String> listParam = Arrays.asList(guid);
        Map<String,Object> params = new HashMap<>();
        params.put("sql",sql);
        params.put("params",listParam);
        ResultBody resultBody = callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl,params);
        return resultBody;
    }
}
