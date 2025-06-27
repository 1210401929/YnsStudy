package com.example.blog_api.service.serviceImpl;

import com.example.blog_api.service.UserInformationService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.service.CallService;
import com.example.common_api.util.FunToUrlUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Component
public class UserInformationServiceImpl implements UserInformationService {

    @Autowired
    CallService callService;

    @Override
    public ResultBody getBlogAndResourceByUserCode(String userCode) {
        Map<String,Object> data = new HashMap<>();
        ResultBody result = getBlogByUserCode(userCode);
        data.put("blog",result.result);
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
        String sql = "select * from blogInfo where USERCODE = ?  order by create_time DESC";
        List<Object> listParam = Arrays.asList(userCode);
        Map<String,Object> params_ = new HashMap<>();
        params_.put("sql",sql);
        params_.put("params",listParam);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params_);
        return result;
    }
}
