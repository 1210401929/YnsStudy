package com.example.common_api.util;

import com.example.common_api.config.LoginCfg;
import io.jsonwebtoken.*;

import java.nio.charset.StandardCharsets;
import java.util.Date;

public class JwtUtils {
    private static final String SECRET = LoginCfg.PUB_SECRET;
    private static final long EXPIRATION = LoginCfg.PUB_EXPIRATION;

    /**
     * 生成 Token
     */
    public static String generateToken(String userId) {
        return Jwts.builder()
                .setSubject(userId)
                .setIssuedAt(new Date())
                .setExpiration(new Date(System.currentTimeMillis() + EXPIRATION))
                // 【修改点1】：统一使用 UTF-8 字节数组进行签名，避免 Base64 解析歧义
                .signWith(SignatureAlgorithm.HS256, SECRET.getBytes(StandardCharsets.UTF_8))
                .compact();
    }

    /**
     * 验证并解析 Token 获取 UserId
     */
    public static String getUserIdFromToken(String token) {
        try {
            Claims claims = Jwts.parser()
                    // 【修改点2】：统一使用 UTF-8 字节数组进行验证
                    .setSigningKey(SECRET.getBytes(StandardCharsets.UTF_8))
                    .parseClaimsJws(token)
                    .getBody();
            return claims.getSubject();
        } catch (ExpiredJwtException e) {
            System.err.println("❌ Token解析失败: Token已过期");
            return null;
        } catch (SignatureException e) {
            System.err.println("❌ Token解析失败: 签名不匹配 (秘钥被改动或Token被篡改)");
            return null;
        } catch (MalformedJwtException e) {
            System.err.println("❌ Token解析失败: Token格式错误 (可能是传了带Bearer的字符串: " + token + ")");
            return null;
        } catch (Exception e) {
            System.err.println("❌ Token解析失败: 发生了未知的异常");
            // 【最关键的一步】：把完整报错打印出来
            e.printStackTrace();
            return null;
        }
    }
}