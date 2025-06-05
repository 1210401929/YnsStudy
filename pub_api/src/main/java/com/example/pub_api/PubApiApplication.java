package com.example.pub_api;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.netflix.eureka.EnableEurekaClient;
import org.springframework.session.data.redis.config.annotation.web.http.EnableRedisHttpSession;

//启动时,扫描com.example包下所有内容 寻找带有注解@Service @Controller......等类
@SpringBootApplication(scanBasePackages = "com.example")
//声明为注册服务
@EnableEurekaClient
//使用redis存储session
@EnableRedisHttpSession
//扫描mapper包
@MapperScan("com.example.pub_api.mapper")
public class PubApiApplication {

    public static void main(String[] args) {
        SpringApplication.run(PubApiApplication.class, args);
    }

}
