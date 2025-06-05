package com.example.common_api.util;

import java.lang.reflect.Field;
import java.util.HashMap;
import java.util.Map;

public class BeanUtil {
    public static   Map<String, Object> beanToMap(Object bean) {
        Map<String, Object> map = new HashMap<>();
        if (bean == null) return map;

        Field[] fields = bean.getClass().getDeclaredFields();
        try {
            for (Field field : fields) {
                field.setAccessible(true); // 允许访问私有字段
                Object value = field.get(bean);
                if (value != null) {
                    map.put(field.getName(), value);
                }
            }
        } catch (IllegalAccessException e) {
            e.printStackTrace(); // 可根据需求改成日志记录
        }
        return map;
    }
}
