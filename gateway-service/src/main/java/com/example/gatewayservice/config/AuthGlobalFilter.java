package com.example.gatewayservice.config;

import com.example.common_api.config.LoginCfg;
import com.example.common_api.util.JwtUtils;
import com.example.gatewayservice.config.auth.AuthConfig;
import org.springframework.cloud.gateway.filter.GatewayFilterChain;
import org.springframework.cloud.gateway.filter.GlobalFilter;
import org.springframework.core.Ordered;
import org.springframework.core.io.buffer.DataBuffer;
import org.springframework.http.HttpStatus;
import org.springframework.http.server.reactive.ServerHttpRequest;
import org.springframework.http.server.reactive.ServerHttpResponse;
import org.springframework.stereotype.Component;
import org.springframework.util.AntPathMatcher;
import org.springframework.web.server.ServerWebExchange;
import reactor.core.publisher.Mono;

import java.nio.charset.StandardCharsets;
import java.util.List;

@Component
public class AuthGlobalFilter implements GlobalFilter, Ordered {

    // 1. 声明刚刚建好的配置类
    private final AuthConfig authConfig;
    private final AntPathMatcher pathMatcher = new AntPathMatcher();

    //万能token,如果传递该token则无视校验
    private final String public_secret = LoginCfg.PUB_PUBLIC_SECRET;
    // 约定微服务互相调用的专属请求头名称
    private final String pubPublicSecretKey = LoginCfg.PUB_PUBLIC_SECRET_KEY;
    // 2. 通过构造器注入配置类 (Spring 推荐方式)
    public AuthGlobalFilter(AuthConfig authConfig) {
        this.authConfig = authConfig;
    }

    @Override
    public Mono<Void> filter(ServerWebExchange exchange, GatewayFilterChain chain) {
        ServerHttpRequest request = exchange.getRequest();
        String path = request.getURI().getPath();

        // 3. 动态从配置类中获取白名单列表
        List<String> whitelist = authConfig.getWhitelist();

        // 2. 检查是不是内部微服务互相调用
        String innerCallSecret = request.getHeaders().getFirst(pubPublicSecretKey);
        // ================= 如果是万能token:放行 =================
        if (public_secret.equals(innerCallSecret)) {
            return chain.filter(exchange);
        }
        // ================= get打头命名的的接口:放行 =================
        int lastSlashIndex = path.lastIndexOf("/");
        if (lastSlashIndex != -1 && lastSlashIndex < path.length() - 1) {
            String methodName = path.substring(lastSlashIndex + 1);
            // 如果方法名是以 "get" 开头，直接无条件放行
            if (methodName.startsWith("get")) {
                return chain.filter(exchange);
            }
        }
        // ================= 白名单:放行 =================
        if (whitelist != null) {
            for (String whiteUrl : whitelist) {
                if (pathMatcher.match(whiteUrl, path)) {
                    return chain.filter(exchange);
                }
            }
        }

        // 2. 获取请求头中的 Token
        String token = request.getHeaders().getFirst("Authorization");

        // 3. 校验 Token 判空
        if (token == null || token.trim().isEmpty()) {
            return unauthorizedResponse(exchange, "未携带Token，请先登录");
        }

        // 去除常见的 "Bearer " 前缀
        if (token.startsWith("Bearer ")) {
            token = token.substring(7);
        }

        // 4. 解析 Token
        String userId = JwtUtils.getUserIdFromToken(token);
        if (userId == null) {
            return unauthorizedResponse(exchange, "Token无效或已过期");
        }

        // 5. 将解析出的 userId 放入请求头，透传给下游微服务
        ServerHttpRequest mutatedRequest = request.mutate()
                .header("X-User-Id", userId)
                .build();

        return chain.filter(exchange.mutate().request(mutatedRequest).build());
    }

    /**
     * 统一处理 401 拦截响应
     */
    private Mono<Void> unauthorizedResponse(ServerWebExchange exchange, String message) {
        ServerHttpResponse response = exchange.getResponse();
        response.setStatusCode(HttpStatus.UNAUTHORIZED);
        response.getHeaders().add("Content-Type", "application/json;charset=UTF-8");

        String jsonResult = String.format("{\"code\": 401, \"message\": \"%s\", \"data\": null}", message);
        DataBuffer buffer = response.bufferFactory().wrap(jsonResult.getBytes(StandardCharsets.UTF_8));
        return response.writeWith(Mono.just(buffer));
    }

    @Override
    public int getOrder() {
        return -200;
    }
}