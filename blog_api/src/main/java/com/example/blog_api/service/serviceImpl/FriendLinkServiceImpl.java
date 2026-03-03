package com.example.blog_api.service.serviceImpl;

import com.example.blog_api.service.FriendLinkService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.service.CallService;
import com.example.common_api.util.FunToUrlUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.*;

@Component
public class FriendLinkServiceImpl implements FriendLinkService {
    @Autowired
    CallService callService;

    @Override
    public ResultBody addFriendLink(HashMap<String, Object> friendLink) {
        List<Map<String, Object>> data = new ArrayList<>();
        data.add(friendLink);
        HashMap<String, Object> params = new HashMap<>();
        params.put("saveType", "add");
        params.put("tableName", "friendLinkInfo");
        params.put("data", data);
        params.put("key", "GUID");
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl,params);
        return result;
    }

    @Override
    public ResultBody updateFriendLink(HashMap<String, Object> friendLink) {
        List<Map<String, Object>> data = new ArrayList<>();
        data.add(friendLink);
        HashMap<String, Object> params = new HashMap<>();
        params.put("saveType", "edit");
        params.put("tableName", "friendLinkInfo");
        params.put("data", data);
        params.put("key", "GUID");
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl,params);
        return result;
    }
    @Override
    public ResultBody deleteFriendLink(String friendLinkId) {
        String deleteSql = "delete from friendLinkInfo where guid = ?";
        List<Object> listParams = Arrays.asList(friendLinkId);
        HashMap<String, Object> params = new HashMap<>();
        params.put("sql", deleteSql);
        params.put("params", listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl, params);
        return result;
    }
    @Override
    public ResultBody getFriendLink(String friendLinkId) {
        String selectSql = "select * from friendLinkInfo where guid = ?";
        List<Object> listParams = Arrays.asList(friendLinkId);
        HashMap<String, Object> params = new HashMap<>();
        params.put("sql", selectSql);
        params.put("params", listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody getFriendLinks() {
        String selectSql = "select * from friendLinkInfo order by create_time";
        List<Object> listParams = Arrays.asList();
        HashMap<String, Object> params = new HashMap<>();
        params.put("sql", selectSql);
        params.put("params", listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        return result;
    }
}
