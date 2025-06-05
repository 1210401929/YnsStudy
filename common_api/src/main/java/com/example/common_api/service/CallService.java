package com.example.common_api.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.web.bind.annotation.RequestBody;

import javax.servlet.http.HttpServletRequest;
import java.util.Map;
//调用该访问接口  不能直接返回调用的结果,应该先接收
@Service
public interface CallService {
    //无参数请求
    ResultBody callFunNoParams(String url);
    //单参数请求
    ResultBody callFunOneParams(String url,String params , String pmsValue);
    //多参数请求
    ResultBody callFunWithParams(String url, Map<String, Object> params);
    //转发请求到gateway网关  再由网关下发到下游服务
    ResponseEntity<Object> proxyToGateWay(HttpServletRequest request, @RequestBody(required = false) byte[] body);
}
