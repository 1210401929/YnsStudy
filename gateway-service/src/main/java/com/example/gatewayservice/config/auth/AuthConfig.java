package com.example.gatewayservice.config.auth;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
//读取application.yml文件的auth下的配置信息,用于验证接口token
@ConfigurationProperties(prefix = "auth")
public class AuthConfig {

    // 变量名必须和 yaml 里的 whitelist 保持一致
    private List<String> whitelist;

    // 必须有 getter 和 setter 方法，Spring 才能成功注入
    public List<String> getWhitelist() {
        return whitelist;
    }

    public void setWhitelist(List<String> whitelist) {
        this.whitelist = whitelist;
    }
}