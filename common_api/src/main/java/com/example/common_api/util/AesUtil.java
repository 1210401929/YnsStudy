package com.example.common_api.util;

import com.example.common_api.bean.ResultBody;

import javax.crypto.Cipher;
import javax.crypto.SecretKey;
import javax.crypto.spec.SecretKeySpec;
import javax.crypto.spec.IvParameterSpec;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;
import java.util.Arrays;
import java.util.Base64;
import java.util.HashMap;
import java.util.Map;

public class AesUtil {
    private static final String ALGORITHM = "AES/CBC/PKCS5Padding";  // 使用 CBC 模式和 PKCS5 填充
    private static final String KEY = "1234567890123456"; // 16 字节密钥
    private static final String IV = "6543210987654321";  // 16 字节初始化向量

    // 加密，支持任何对象类型
    public static String encrypt(Object data) throws Exception {
        // 序列化对象为字节数组
        byte[] dataBytes = serialize(data);

        SecretKeySpec secretKey = new SecretKeySpec(KEY.getBytes(), "AES");
        IvParameterSpec ivParameterSpec = new IvParameterSpec(IV.getBytes());

        Cipher cipher = Cipher.getInstance(ALGORITHM);
        cipher.init(Cipher.ENCRYPT_MODE, secretKey, ivParameterSpec);
        byte[] encryptedData = cipher.doFinal(dataBytes);

        // 使用 Base64 编码加密后的字节数据，方便存储和传输
        return Base64.getEncoder().encodeToString(encryptedData);
    }

    // 解密，支持任何对象类型
    public static Object decrypt(String encryptedData) throws Exception {
        byte[] decodedData = Base64.getDecoder().decode(encryptedData);

        SecretKeySpec secretKey = new SecretKeySpec(KEY.getBytes(), "AES");
        IvParameterSpec ivParameterSpec = new IvParameterSpec(IV.getBytes());

        Cipher cipher = Cipher.getInstance(ALGORITHM);
        cipher.init(Cipher.DECRYPT_MODE, secretKey, ivParameterSpec);
        byte[] decryptedData = cipher.doFinal(decodedData);

        // 反序列化字节数组为对象
        return deserialize(decryptedData);
    }

    // 序列化对象为字节数组
    private static byte[] serialize(Object obj) throws Exception {
        try (ByteArrayOutputStream byteArrayOutputStream = new ByteArrayOutputStream();
             ObjectOutputStream objectOutputStream = new ObjectOutputStream(byteArrayOutputStream)) {
            objectOutputStream.writeObject(obj);
            return byteArrayOutputStream.toByteArray();
        }
    }

    // 反序列化字节数组为对象
    private static Object deserialize(byte[] bytes) throws Exception {
        try (ByteArrayInputStream byteArrayInputStream = new ByteArrayInputStream(bytes);
             ObjectInputStream objectInputStream = new ObjectInputStream(byteArrayInputStream)) {
            return objectInputStream.readObject();
        }
    }

    public static void main(String[] args) throws Exception {
        Map<String,String > map = new HashMap<>();
        map.put("111","111");
        // 示例对象 ResultBody
        ResultBody result = ResultBody.createSuccessResult(map);

        // 加密
        String encryptedData = encrypt(result);
        System.out.println("Encrypted: " + encryptedData);

        // 解密
        ResultBody decryptedResult = (ResultBody) decrypt(encryptedData);
        System.out.println("Decrypted: " + decryptedResult);
    }
}
