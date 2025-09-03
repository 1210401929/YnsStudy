package com.example.ai_api;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.netflix.eureka.EnableEurekaClient;
import org.springframework.session.data.redis.config.annotation.web.http.EnableRedisHttpSession;

@SpringBootApplication(scanBasePackages = "com.example")
//声明为注册服务
@EnableEurekaClient
//使用redis存储session
@EnableRedisHttpSession
public class AiApiApplication {

    public static void main(String[] args) {
        SpringApplication.run(AiApiApplication.class, args);
    }

}
