package com.example.common_api.service.serviceImpl;

import com.example.common_api.bean.ResultBody;
import com.example.common_api.config.LoginCfg;
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

    //万能token,微服务之间调用,需要携带该token
    private final String public_secret = LoginCfg.PUB_PUBLIC_SECRET;

    // 约定微服务互相调用的专属请求头名称
    private final String pubPublicSecretKey = LoginCfg.PUB_PUBLIC_SECRET_KEY;

    @Override
    public ResultBody callFunNoParams(String url) {
        if(!url.contains(pubApiUrl)){
            url = pubApiUrl + url;
        }

        // 1. 设置请求头，带上万能 Token
        HttpHeaders headers = new HttpHeaders();
        headers.add(pubPublicSecretKey, public_secret);
        HttpEntity<String> entity = new HttpEntity<>(null, headers);

        System.out.println("发送网关(GET),地址:" + url);

        // 2. GET 请求如果需要传 Header，必须使用 exchange 方法而不是 getForObject
        ResponseEntity<ResultBody> response = restTemplate.exchange(
                url,
                HttpMethod.GET,
                entity,
                ResultBody.class
        );
        return response.getBody();
    }

    @Override
    public ResultBody callFunOneParams(String url, String paramsName, String paramsValue) {
        if(!url.contains(pubApiUrl)){
            url = pubApiUrl + url;
        }
        Map<String,Object> params = new HashMap<>();
        params.put(paramsName,paramsValue);
        return callFunWithParams(url, params);
    }

    @Override
    public ResultBody callFunWithParams(String url, Map<String, Object> params) {
        if(!url.contains(pubApiUrl)){
            url = pubApiUrl + url;
        }
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);  // 设置请求头为 JSON 类型

        // ★ 核心修改：在 POST 请求中加入万能 Token
        headers.add(pubPublicSecretKey, public_secret);

        // 将 Map 作为请求体
        HttpEntity<Map<String, Object>> request = new HttpEntity<>(params, headers);
        System.out.println("发送网关(POST),地址:" + url);
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

            // ★ 核心修改：追加内部调用万能 Token
            headers.add(pubPublicSecretKey, public_secret);

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
            gatewayResponse.getHeaders().forEach(responseHeaders::put);

            // 打印响应内容
            System.out.println("网关响应内容: " + gatewayResponse);
            ResponseEntity<byte[]> entity = new ResponseEntity<>(gatewayResponse.getBody(), responseHeaders, gatewayResponse.getStatusCode());

            if (entity.getBody() == null) {
                return ResponseEntity.status(gatewayResponse.getStatusCode()).build();
            }

            String entityBody = new String(entity.getBody(), StandardCharsets.UTF_8);
            System.out.println("已接受返回结果:" + entity);

            // 动态判断响应类型
            try {
                ObjectMapper objectMapper = new ObjectMapper();
                Object responseObject = objectMapper.readValue(entityBody, Object.class);
                System.out.println("已接受返回具体结果:" + responseObject);
                return ResponseEntity.status(gatewayResponse.getStatusCode()).body(responseObject);
            } catch (JsonProcessingException e) {
                e.printStackTrace();
                // 如果解析 JSON 失败，直接返回字符串
                return ResponseEntity.status(gatewayResponse.getStatusCode()).body(entityBody);
            }

        } catch (Exception e) {
            e.printStackTrace();
            return ResponseEntity.status(500).body("代理请求失败: " + e.getMessage());
        }
    }
}