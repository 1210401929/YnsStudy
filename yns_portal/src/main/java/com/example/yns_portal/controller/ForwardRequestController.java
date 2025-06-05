package com.example.yns_portal.controller;

import com.example.common_api.bean.ResultBody;
import com.example.common_api.service.CallService;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import javax.servlet.http.HttpServletRequest;
import java.nio.charset.StandardCharsets;

//拦截前台请求,并转发至gateway网关
@RestController
public class ForwardRequestController {
    @Autowired
    CallService callService;

    @RequestMapping({"/pub-api/**", "/ai-api/**"})
    public ResponseEntity<Object> proxyToGateWay(HttpServletRequest request, @RequestBody(required = false) byte[] body) {
        return callService.proxyToGateWay(request, body);
    }
}