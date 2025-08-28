package com.example.blog_api.service.serviceImpl;

import com.example.blog_api.service.AnnouncementService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import com.example.common_api.service.CallService;
import com.example.common_api.util.FunToUrlUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import javax.servlet.http.HttpSession;
import java.util.*;

@Component
public class AnnouncementServiceImpl implements AnnouncementService {

    @Autowired
    CallService callService;

    @Override
    public ResultBody addAnnouncement(Map<String, String> announcement, HttpSession session) {
        if (announcement.get("TEXT") == null) {
            return ResultBody.createErrorResult("未传递有效内容:text");
        }
        UserBean userInfo = (UserBean) session.getAttribute("userInfo");
        if (userInfo == null) {
            return ResultBody.createErrorResult("用户未登录!");
        }
        String userCode = userInfo.getCODE();
        String userName = userInfo.getNAME();
        announcement.put("userCode", userCode);
        announcement.put("userName", userName);
        Map<String, Object> params = new HashMap<>();
        List<Map<String, String>> data = new ArrayList<>();
        data.add(announcement);
        params.put("saveType", "add");
        params.put("tableName", "announcementInfo");
        params.put("data", data);
        params.put("key", "GUID");
        return callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl, params);
    }

    @Override
    public ResultBody editAnnouncement(Map<String, String> announcement) {
        if (announcement.get("TEXT") == null) {
            return ResultBody.createErrorResult("未传递有效内容:text");
        }
        Map<String, Object> params = new HashMap<>();
        List<Map<String, String>> data = new ArrayList<>();
        data.add(announcement);
        params.put("saveType", "edit");
        params.put("tableName", "announcementInfo");
        params.put("data", data);
        params.put("key", "GUID");
        return callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl, params);
    }

    @Override
    public ResultBody deleteAnnouncement(String guid) {
        String sql = "delete from announcementInfo where guid=?";
        List<Object> listParams = Arrays.asList(guid);
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        params.put("params", listParams);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl, params);
        return result;
    }

    @Override
    public ResultBody getAllAnnouncement() {
        String sql = "select * from announcementInfo order by TYPE,CREATE_TIME desc";
        ResultBody result = callService.callFunOneParams(FunToUrlUtil.selectListUrl, "sql", sql);
        return result;
    }

    @Override
    public ResultBody getAnnouncementByType(String type) {
        String sql = "select * from announcementInfo where ISENABLE = '1' and type = '" + type + "' order by TYPE,CREATE_TIME desc";
        ResultBody result = callService.callFunOneParams(FunToUrlUtil.selectListUrl, "sql", sql);
        return result;
    }
}
