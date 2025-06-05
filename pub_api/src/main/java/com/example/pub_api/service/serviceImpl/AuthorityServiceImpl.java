package com.example.pub_api.service.serviceImpl;

import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import com.example.pub_api.service.AuthorityService;
import com.google.gson.JsonArray;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import org.springframework.stereotype.Component;

import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Component
public class AuthorityServiceImpl implements AuthorityService {

    //读取menu.json文件,获取菜单权限
    @Override
    public ResultBody getUserMenu(UserBean user) {
        // 尝试读取资源文件
        try (InputStream inputStream = AuthorityServiceImpl.class.getClassLoader().getResourceAsStream("menu.json")) {
            if (inputStream == null) {
                throw new IllegalArgumentException("文件没有找到！");
            }
            // 使用InputStreamReader读取输入流
            InputStreamReader reader = new InputStreamReader(inputStream, StandardCharsets.UTF_8);
            // 使用Gson解析JSON文件内容
            JsonElement jsonElement = JsonParser.parseReader(reader);
            // 处理JSON数组并构造成List<HashMap<String, Object>>
            List<HashMap<String, Object>> resultList = new ArrayList<>();
            if (jsonElement.isJsonArray()) {
                JsonArray jsonArray = jsonElement.getAsJsonArray();
                for (JsonElement element : jsonArray) {
                    JsonObject jsonObject = element.getAsJsonObject();
                    HashMap<String, Object> map = parseJsonObject(jsonObject, user);
                    resultList.add(map);
                }
            } else {
                System.out.println("JSON内容不是数组！");
            }
            return ResultBody.createSuccessResult(resultList);
        } catch (Exception e) {
            e.printStackTrace();
        }
        return ResultBody.createSuccessResult("");
    }

    // 解析JsonObject并转换为HashMap<String, Object>
    private static HashMap<String, Object> parseJsonObject(JsonObject jsonObject, UserBean user) {
        HashMap<String, Object> map = new HashMap<>();

        for (Map.Entry<String, JsonElement> entry : jsonObject.entrySet()) {
            String key = entry.getKey();
            JsonElement value = entry.getValue();

            if (value.isJsonArray()) {
                List<Object> list = new ArrayList<>();
                JsonArray jsonArray = value.getAsJsonArray();
                if (key.equals("authority")) {
                    Boolean isHasMenu = false;
                    for (JsonElement arrayElement : jsonArray) {
                        if (arrayElement.getAsString().equals(user.getROLE()) || user.getROLE().equals("*")) {
                            isHasMenu = true;
                        }
                    }
                    if (!isHasMenu) map.put("disabled", true);
                } else {
                    for (JsonElement arrayElement : jsonArray) {
                        if (arrayElement.isJsonPrimitive()) {
                            list.add(arrayElement.getAsString());
                        } else {
                            list.add(parseJsonObject(arrayElement.getAsJsonObject(), user));
                        }
                    }
                }
                map.put(key, list);
            } else if (value.isJsonObject()) {
                map.put(key, parseJsonObject(value.getAsJsonObject(), user));
            } else {
                map.put(key, value.getAsString());
            }
        }
        return map;
    }
}
