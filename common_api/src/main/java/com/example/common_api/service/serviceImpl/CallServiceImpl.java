package com.example.common_api.service.serviceImpl;

import com.example.common_api.bean.ResultBody;

import com.example.common_api.service.CallService;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.*;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

import javax.servlet.http.HttpServletRequest;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.util.Enumeration;
import java.util.HashMap;
import java.util.Map;

//公共请求网关   统一走这里
@Component
public class CallServiceImpl implements CallService {
    @Autowired
    private RestTemplate restTemplate;

    //网关实例网址
    final static String pubApiUrl = "http://gateway-service";

    @Override
    public ResultBody callFunNoParams(String url) {
        if(url.indexOf(pubApiUrl)==-1){
            url = pubApiUrl + url;
        }
        System.out.println("发送网关,地址:" + url);
        return restTemplate.getForObject(url, ResultBody.class);
    }

    @Override
    public ResultBody callFunOneParams(String url, String paramsName, String paramsValue) {
        if(url.indexOf(pubApiUrl)==-1){
            url = pubApiUrl + url;
        }
        Map<String,Object> params = new HashMap<>();
        params.put(paramsName,paramsValue);
        return callFunWithParams(url,params);
    }

    @Override
    public ResultBody callFunWithParams(String url, Map<String, Object> params) {
        if(url.indexOf(pubApiUrl)==-1){
            url = pubApiUrl + url;
        }
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);  // 设置请求头为 JSON 类型

        // 将 Map 作为请求体
        HttpEntity<Map<String, Object>> request = new HttpEntity<>(params, headers);
        System.out.println("发送网关,地址:" + url);
        System.out.println("参数:" + params);
        System.out.println("请求头:" + headers);
        return restTemplate.postForObject(url, request, ResultBody.class);
    }

    @Override
    public ResponseEntity<Object> proxyToGateWay(HttpServletRequest request, byte[] body) {
        try {
            // 1. 组装目标地址
            String path = request.getRequestURI();
            String targetUrl = pubApiUrl + path;
            System.out.println("准备跳转请求,请求路径:" + targetUrl);
            // 2. 拷贝请求头
            HttpHeaders headers = new HttpHeaders();
            Enumeration<String> headerNames = request.getHeaderNames();
            while (headerNames.hasMoreElements()) {
                String headerName = headerNames.nextElement();
                if (!"host".equalsIgnoreCase(headerName)) { // 避免Host问题
                    headers.add(headerName, request.getHeader(headerName));
                }
            }

            // 3. 封装请求体
            HttpEntity<byte[]> httpEntity = new HttpEntity<>(body, headers);

            // 4. 执行请求
            ResponseEntity<byte[]> gatewayResponse = restTemplate.exchange(
                    URI.create(targetUrl),
                    HttpMethod.resolve(request.getMethod()),
                    httpEntity,
                    byte[].class
            );

            // 5. 返回响应
            HttpHeaders responseHeaders = new HttpHeaders();
            gatewayResponse.getHeaders().forEach((key, value) -> responseHeaders.put(key, value));
            // 打印响应内容
            System.out.println("网关响应内容: " + gatewayResponse);
            ResponseEntity<byte[]> entity = new ResponseEntity<>(gatewayResponse.getBody(), responseHeaders, gatewayResponse.getStatusCode());
            String entityBody = new String(entity.getBody(), StandardCharsets.UTF_8).toString();
            System.out.println("已接受返回结果:" + entity);
            // 动态判断响应类型
            try {
                ObjectMapper objectMapper = new ObjectMapper();
                Object responseObject = objectMapper.readValue(entityBody, Object.class);
                System.out.println("已接受返回具体结果:" + responseObject);
                // 返回 Object，Spring 会自动根据类型转换为 JSON 格式
                return ResponseEntity.ok(responseObject);
            } catch (JsonProcessingException e) {
                e.printStackTrace();
                return ResponseEntity.status(500).body("请求处理失败: " + e.getMessage());
            }

        } catch (Exception e) {
            e.printStackTrace();
            return ResponseEntity.status(500).body(("代理请求失败: " + e.getMessage()).getBytes(StandardCharsets.UTF_8));
        }
    }
}

