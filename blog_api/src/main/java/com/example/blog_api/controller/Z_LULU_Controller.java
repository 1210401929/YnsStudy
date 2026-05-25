package com.example.blog_api.controller;

import com.example.blog_api.service.Z_LULU_Service;
import com.example.common_api.bean.ResultBody;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;
import java.util.List;
import javax.servlet.http.HttpServletRequest;

@RestController
@RequestMapping("blog-api/lulu")
public class Z_LULU_Controller {

    @Autowired
    private Z_LULU_Service luluService;

    @PostMapping("/status")
    public Map<String, Object> getStatus(@RequestBody Map<String, Object> params) {
        Long userNum = Long.valueOf(params.get("userNum").toString());
        return luluService.getAndRefreshPetStatus(userNum);
    }

    @PostMapping("/feed")
    public Map<String, Object> feed(@RequestBody Map<String, Object> params, HttpServletRequest request) {
        Long userNum = Long.valueOf(params.get("userNum").toString());
        return luluService.feedPet(userNum, getClientIp(request), getUserAgent(request));
    }

    @PostMapping("/play")
    public Map<String, Object> play(@RequestBody Map<String, Object> params, HttpServletRequest request) {
        Long userNum = Long.valueOf(params.get("userNum").toString());
        String actionName = params.get("actionName") == null ? "玩耍" : params.get("actionName").toString();
        return luluService.playPet(userNum, actionName, getClientIp(request), getUserAgent(request));
    }

    @PostMapping("/sleep")
    public Map<String, Object> sleep(@RequestBody Map<String, Object> params, HttpServletRequest request) {
        Long userNum = Long.valueOf(params.get("userNum").toString());
        return luluService.toggleSleepPet(userNum, getClientIp(request), getUserAgent(request));
    }

    @PostMapping("/messages")
    public List<Map<String, Object>> getMessages(@RequestBody Map<String, Object> params) {
        Long userNum = Long.valueOf(params.get("userNum").toString());
        return luluService.getMessages(userNum);
    }

    @PostMapping("/message/add")
    public ResultBody addMessage(@RequestBody Map<String, Object> params, HttpServletRequest request) {
        Long userNum = Long.valueOf(params.get("userNum").toString());
        String content = params.get("content") == null ? "" : params.get("content").toString();
        return luluService.addMessage(userNum, content, getClientIp(request));
    }

    @PostMapping("/message/delete")
    public List<Map<String, Object>> deleteMessage(@RequestBody Map<String, Object> params) {
        Long userNum = Long.valueOf(params.get("userNum").toString());
        Long messageId = Long.valueOf(params.get("messageId").toString());
        return luluService.deleteMessage(userNum, messageId);
    }

    @PostMapping("/logs")
    public List<Map<String, Object>> getLogs(@RequestBody Map<String, Object> params) {
        Long userNum = Long.valueOf(params.get("userNum").toString());
        return luluService.getLogs(userNum);
    }

    private String getClientIp(HttpServletRequest request) {
        String[] headerNames = {
                "X-Forwarded-For",
                "X-Real-IP",
                "Proxy-Client-IP",
                "WL-Proxy-Client-IP"
        };

        for (String headerName : headerNames) {
            String ip = request.getHeader(headerName);
            if (ip != null && !ip.isEmpty() && !"unknown".equalsIgnoreCase(ip)) {
                return ip.split(",")[0].trim();
            }
        }
        return request.getRemoteAddr();
    }

    private String getUserAgent(HttpServletRequest request) {
        String userAgent = request.getHeader("User-Agent");
        return userAgent == null ? "" : userAgent;
    }
}
