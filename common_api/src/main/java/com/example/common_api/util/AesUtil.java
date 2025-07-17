package com.example.common_api.util;

import javax.crypto.Cipher;
import javax.crypto.spec.IvParameterSpec;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.util.Base64;

public final class AesUtil {

    // 16 字节 key 和 IV，必须与前端保持一致（UTF-8 编码）
    private static final String KEY_STR = "1234567890123456";
    private static final String IV_STR  = "6543210987654321";

    private static final SecretKeySpec KEY =
            new SecretKeySpec(KEY_STR.getBytes(StandardCharsets.UTF_8), "AES");
    private static final IvParameterSpec IV =
            new IvParameterSpec(IV_STR.getBytes(StandardCharsets.UTF_8));

    private AesUtil() {}

    // 解密 Base64 -> 明文
    public static String decrypt(String base64CipherText) {
        try {
            byte[] cipherBytes = Base64.getDecoder().decode(base64CipherText);
            Cipher cipher = Cipher.getInstance("AES/CBC/PKCS5Padding");
            cipher.init(Cipher.DECRYPT_MODE, KEY, IV);
            byte[] plainBytes = cipher.doFinal(cipherBytes);
            return new String(plainBytes, StandardCharsets.UTF_8);
        } catch (Exception e) {
            throw new IllegalStateException("AES decrypt error", e);
        }
    }

    // 加密 明文 -> Base64
    public static String encrypt(String plainText) {
        try {
            Cipher cipher = Cipher.getInstance("AES/CBC/PKCS5Padding");
            cipher.init(Cipher.ENCRYPT_MODE, KEY, IV);
            byte[] encrypted = cipher.doFinal(plainText.getBytes(StandardCharsets.UTF_8));
            return Base64.getEncoder().encodeToString(encrypted);
        } catch (Exception e) {
            throw new IllegalStateException("AES encrypt error", e);
        }
    }
}
