package com.example.gatewayservice.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.reactive.UrlBasedCorsConfigurationSource;
import org.springframework.web.cors.reactive.CorsWebFilter;

/**
 * 全局跨域配置
 * 作用：允许前端跨域请求网关，转发到后端服务
 */
@Configuration
public class CorsGlobalConfiguration {

    @Bean
    public CorsWebFilter corsFilter() {
        CorsConfiguration config = new CorsConfiguration();
        // 允许所有来源的跨域请求（也可以改成指定来源）
        config.addAllowedOriginPattern("*");

        // 允许所有请求方法：GET、POST、PUT、DELETE等
        config.addAllowedMethod("*");

        // 允许所有请求头
        config.addAllowedHeader("*");

        // 允许携带 Cookie（比如登录的 Session）
        config.setAllowCredentials(true);

        // 注册跨域配置，拦截所有路径
        UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
        source.registerCorsConfiguration("/**", config);

        return new CorsWebFilter(source);
    }
}