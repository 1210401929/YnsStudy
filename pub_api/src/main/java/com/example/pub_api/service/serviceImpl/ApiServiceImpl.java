package com.example.pub_api.service.serviceImpl;

import com.example.common_api.bean.ResultBody;
import com.example.pub_api.service.ApiService;
import com.google.gson.Gson;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import javax.servlet.http.HttpServletRequest;
import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.InetAddress;
import java.net.URL;
import java.net.UnknownHostException;

@Component
public class ApiServiceImpl implements ApiService {

    @Autowired
    private HttpServletRequest request;

    @Override
    public ResultBody getCurrentCity() throws IOException {
        String userAddress = getCurrentLocation((String) getClientIpAddress().result);
        System.out.println("用户地址:"+userAddress);
        return ResultBody.createSuccessResult(userAddress);
    }

    @Override
    public ResultBody getClientIpAddress() {
        // 获取 X-Forwarded-For 头
        String xForwardedForHeader = request.getHeader("X-Forwarded-For");

        if (xForwardedForHeader != null && !xForwardedForHeader.isEmpty()) {
            // 提取第一个 IP 地址，可能是客户端的真实 IP
            String clientIp = xForwardedForHeader.split(",")[0].trim();
            System.out.println("ip (X-Forwarded-For):" + clientIp);

            // 如果获取到的是 Docker 内部 IP 地址，可以尝试返回宿主机的 IP 地址
            if (isDockerInternalIp(clientIp)) {
                clientIp = getHostIp(); // 你可以实现获取宿主机 IP 的方法
                System.out.println("ip (Docker internal):" + clientIp);
            }
            return ResultBody.createSuccessResult(clientIp);
        }
        // 如果没有 X-Forwarded-For 头，直接返回 request.getRemoteAddr()
        String clientIp = request.getRemoteAddr();
        System.out.println("ip (RemoteAddr):" + clientIp);
        // 如果获取到的是 Docker 内部 IP 地址，也返回宿主机的 IP 地址
        if (isDockerInternalIp(clientIp)) {
            clientIp = getHostIp(); // 获取宿主机 IP
            System.out.println("ip (Docker internal):" + clientIp);
        }
        return ResultBody.createSuccessResult(clientIp);
    }

    public String getCurrentLocation(String ipAddress) throws IOException {
        String apiKey = "d95187e862803e5cfd2d43a87b7e6a11"; // 替换为你的高德地图API Key
        String url = "https://restapi.amap.com/v3/ip?output=json&key=" + apiKey + "&ip=" + ipAddress;

        HttpURLConnection connection = (HttpURLConnection) new URL(url).openConnection();
        connection.setRequestMethod("GET");

        BufferedReader reader = new BufferedReader(new InputStreamReader(connection.getInputStream()));
        StringBuilder response = new StringBuilder();
        String line;
        while ((line = reader.readLine()) != null) {
            response.append(line);
        }
        reader.close();

        // 解析JSON格式的响应数据
        String userAddress = parseResponse(response.toString());

        connection.disconnect();

        return userAddress;
    }
    // 判断是否为 Docker 内部 IP 地址
    private boolean isDockerInternalIp(String ip) {
        return ip.startsWith("172.17."); // Docker 默认的 bridge 网络 IP 范围
    }

    // 获取宿主机的 IP 地址
    private String getHostIp() {
        try {
            InetAddress inetAddress = InetAddress.getLocalHost();
            return inetAddress.getHostAddress(); // 获取宿主机的 IP 地址
        } catch (UnknownHostException e) {
            e.printStackTrace();
            return "Unable to determine host IP";
        }
    }
    private static String parseResponse(String response) {
        Gson gson = new Gson();
        ResponseObject responseObject;

        // 尝试解析为单个对象
        try {
            responseObject = gson.fromJson(response, ResponseObject.class);
            return responseObject.city;
        } catch (Exception e) {
            // 解析为单个对象失败，尝试解析为数组
            try {
                ResponseObject[] responseObjectArray = gson.fromJson(response, ResponseObject[].class);
                if (responseObjectArray.length > 0) {
                    return responseObjectArray[0].city;
                } else {
                    return null; // 或者抛出异常，表示无法解析
                }
            } catch (Exception ex) {
                // 解析为数组失败，返回 null 或者抛出异常，表示无法解析
                return null;
            }
        }
    }

    // 定义Java对象来表示JSON中的结构
    private static class ResponseObject {
        String city;
    }
}
