package com.example.blog_api.service.serviceImpl;

import com.example.blog_api.service.Z_LULU_Service;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.service.CallService;
import com.example.common_api.util.FunToUrlUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.Duration;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class Z_LULU_ServiceImpl implements Z_LULU_Service {

    @Autowired
    CallService callService;

    private static final String DEFAULT_NAME = "Lulu";
    private static final DateTimeFormatter DB_TIME_FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    @Override
    public Map<String, Object> getAndRefreshPetStatus(Long userNum) {
        String selectSql = "select * from z_pet_status where user_num='" + userNum + "'";
        ResultBody result = callService.callFunOneParams(FunToUrlUtil.selectListUrl, "sql", selectSql);

        List<Map<String, Object>> res = null;
        if (result != null && result.result != null) {
            res = (List<Map<String, Object>>) result.result;
        }

        if (res == null || res.isEmpty()) {
            return createDefaultPet(userNum);
        }

        Map<String, Object> petData = res.get(0);

        // 【核心修复】：兼容旧数据。如果老用户的数据里没有等级和经验，强制赋予初始值，防止后续强转报空指针
        petData.putIfAbsent("LEVEL", 1);
        petData.putIfAbsent("EXP", 0);

        refreshByTime(petData);
        return petData;
    }

    @Override
    public Map<String, Object> feedPet(Long userNum, String clientIp, String userAgent) {
        Map<String, Object> petData = this.getAndRefreshPetStatus(userNum);

        if (petData != null) {
            int currentHunger = ((Number) petData.get("HUNGER")).intValue();
            int currentMood = ((Number) petData.get("MOOD")).intValue();

            petData.put("HUNGER", Math.min(100, currentHunger + 30));
            petData.put("MOOD", Math.min(100, currentMood + 5));
            petData.put("CURRENT_STATE", "IDLE");
            petData.put("LAST_UPDATE_TIME", LocalDateTime.now());

            addExp(petData, 20);
            updatePetStatusInDb(petData);
            insertLog(userNum, "FEED", "喂食", clientIp, userAgent, "饱腹 +30，心情 +5，经验 +20");
        }
        return petData;
    }

    @Override
    public Map<String, Object> playPet(Long userNum, String actionName, String clientIp, String userAgent) {
        Map<String, Object> petData = this.getAndRefreshPetStatus(userNum);
        String cleanActionName = actionName == null || actionName.trim().isEmpty() ? "玩耍" : actionName.trim();

        if (petData != null) {
            String currentState = (String) petData.get("CURRENT_STATE");
            if ("SLEEPING".equals(currentState)) {
                return petData;
            }

            int currentEnergy = ((Number) petData.get("ENERGY")).intValue();
            int currentMood = ((Number) petData.get("MOOD")).intValue();
            int currentHunger = ((Number) petData.get("HUNGER")).intValue();

            if (currentEnergy >= 15) {
                petData.put("ENERGY", Math.max(0, currentEnergy - 15));
                petData.put("HUNGER", Math.max(0, currentHunger - 10));
                petData.put("MOOD", Math.min(100, currentMood + 20));
                petData.put("CURRENT_STATE", "IDLE");
                petData.put("LAST_UPDATE_TIME", LocalDateTime.now());

                addExp(petData, 40);
                updatePetStatusInDb(petData);
                insertLog(userNum, "PLAY", cleanActionName, clientIp, userAgent, cleanActionName + "，心情 +20，经验 +40");
            }
        }
        return petData;
    }

    @Override
    public Map<String, Object> toggleSleepPet(Long userNum, String clientIp, String userAgent) {
        Map<String, Object> petData = this.getAndRefreshPetStatus(userNum);

        if (petData != null) {
            String currentState = (String) petData.get("CURRENT_STATE");
            boolean isWaking = "SLEEPING".equals(currentState);
            petData.put("CURRENT_STATE", isWaking ? "IDLE" : "SLEEPING");
            petData.put("LAST_UPDATE_TIME", LocalDateTime.now());

            updatePetStatusInDb(petData);
            insertLog(userNum, isWaking ? "WAKE" : "SLEEP", isWaking ? "唤醒" : "睡觉", clientIp, userAgent, isWaking ? "把 Lulu 唤醒了" : "让 Lulu 休息");
        }
        return petData;
    }

    @Override
    public List<Map<String, Object>> getMessages(Long userNum) {
        String selectSql = "select * from z_lulu_message where user_num='" + userNum + "' order by create_time desc limit 80";
        ResultBody result = callService.callFunOneParams(FunToUrlUtil.selectListUrl, "sql", selectSql);

        if (result != null && result.result != null) {
            return (List<Map<String, Object>>) result.result;
        }
        return java.util.Collections.emptyList();
    }

    @Override
    public ResultBody addMessage(Long userNum, String content, String clientIp) {
        String cleanContent = content == null ? "" : content.trim();
        if (cleanContent.length() > 500) {
            cleanContent = cleanContent.substring(0, 500);
        }
        if (cleanContent.isEmpty()) {
            return ResultBody.createErrorResult("留言内容不能为空");
        }

        String safeIp = clientIp == null || clientIp.trim().isEmpty() ? "unknown" : clientIp.trim();
        if (countRecentMessagesByIp(safeIp) >= 5) {
            return ResultBody.createErrorResult("留言太频繁了，同一 IP 1 分钟内最多发布 5 条留言");
        }

        LocalDateTime now = LocalDateTime.now();
        insertMessage(userNum, "USER", cleanContent, now, safeIp);

        return ResultBody.createSuccessResult(getMessages(userNum));
    }

    @Override
    public List<Map<String, Object>> deleteMessage(Long userNum, Long messageId) {
        String deleteSql = "delete from z_lulu_message where ID='" + messageId + "' and USER_NUM='" + userNum + "'";
        callService.callFunOneParams(FunToUrlUtil.exeSqlUrl, "sql", deleteSql);
        return getMessages(userNum);
    }

    @Override
    public List<Map<String, Object>> getLogs(Long userNum) {
        String selectSql = "select * from z_lulu_log where USER_NUM='" + userNum + "' order by CREATE_TIME desc limit 120";
        ResultBody result = callService.callFunOneParams(FunToUrlUtil.selectListUrl, "sql", selectSql);

        if (result != null && result.result != null) {
            return (List<Map<String, Object>>) result.result;
        }
        return java.util.Collections.emptyList();
    }

    private Map<String, Object> createDefaultPet(Long userNum) {
        Map<String, Object> petData = new HashMap<>();
        LocalDateTime now = LocalDateTime.now();
        String nowStr = now.format(DB_TIME_FORMATTER);

        petData.put("USER_NUM", userNum);
        petData.put("NAME", DEFAULT_NAME);
        petData.put("HUNGER", 50);
        petData.put("ENERGY", 60);
        petData.put("MOOD", 80);
        petData.put("LEVEL", 1);
        petData.put("EXP", 0);
        petData.put("CURRENT_STATE", "IDLE");
        petData.put("LAST_UPDATE_TIME", now);

        String insertSql = String.format(
                "INSERT INTO z_pet_status (USER_NUM, NAME, HUNGER, ENERGY, MOOD, LEVEL, EXP, CURRENT_STATE, LAST_UPDATE_TIME) " +
                        "VALUES ('%s', '%s', 50, 60, 80, 1, 0, 'IDLE', '%s')",
                userNum, DEFAULT_NAME, nowStr
        );

        callService.callFunOneParams(FunToUrlUtil.exeSqlUrl, "sql", insertSql);
        return petData;
    }

    private void insertMessage(Long userNum, String senderType, String content, LocalDateTime createTime, String ipAddress) {
        String insertSql = String.format(
                "INSERT INTO z_lulu_message (USER_NUM, SENDER_TYPE, CONTENT, CREATE_TIME, IP_ADDRESS) VALUES ('%s', '%s', '%s', '%s', '%s')",
                userNum,
                escapeSql(senderType),
                escapeSql(content),
                createTime.format(DB_TIME_FORMATTER),
                escapeSql(ipAddress)
        );
        callService.callFunOneParams(FunToUrlUtil.exeSqlUrl, "sql", insertSql);
    }

    private void insertLog(Long userNum, String actionType, String actionName, String clientIp, String userAgent, String remark) {
        String safeIp = clientIp == null || clientIp.trim().isEmpty() ? "unknown" : clientIp.trim();
        String safeUserAgent = userAgent == null ? "" : userAgent;
        String browser = parseBrowser(safeUserAgent);
        String deviceModel = parseDeviceModel(safeUserAgent);
        String insertSql = String.format(
                "INSERT INTO z_lulu_log (USER_NUM, ACTION_TYPE, ACTION_NAME, IP_ADDRESS, BROWSER, DEVICE_MODEL, USER_AGENT, REMARK, CREATE_TIME) " +
                        "VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s')",
                userNum,
                escapeSql(actionType),
                escapeSql(actionName),
                escapeSql(safeIp),
                escapeSql(browser),
                escapeSql(deviceModel),
                escapeSql(limitLength(safeUserAgent, 500)),
                escapeSql(remark),
                LocalDateTime.now().format(DB_TIME_FORMATTER)
        );
        callService.callFunOneParams(FunToUrlUtil.exeSqlUrl, "sql", insertSql);
    }

    private String parseBrowser(String userAgent) {
        if (userAgent == null || userAgent.isEmpty()) {
            return "未知浏览器";
        }
        if (userAgent.contains("Edg/")) {
            return "Microsoft Edge";
        }
        if (userAgent.contains("OPR/") || userAgent.contains("Opera")) {
            return "Opera";
        }
        if (userAgent.contains("Firefox/")) {
            return "Firefox";
        }
        if (userAgent.contains("Chrome/") || userAgent.contains("CriOS/")) {
            return "Chrome";
        }
        if (userAgent.contains("Safari/")) {
            return "Safari";
        }
        return "未知浏览器";
    }

    private String parseDeviceModel(String userAgent) {
        if (userAgent == null || userAgent.isEmpty()) {
            return "未知设备";
        }
        if (userAgent.contains("iPhone")) {
            return "iPhone";
        }
        if (userAgent.contains("iPad")) {
            return "iPad";
        }
        int start = userAgent.indexOf("Android");
        if (start >= 0) {
            int end = userAgent.indexOf(")", start);
            String androidInfo = end > start ? userAgent.substring(start, end) : userAgent.substring(start);
            String[] parts = androidInfo.split(";");
            if (parts.length >= 3) {
                return parts[parts.length - 1].trim();
            }
            return "Android";
        }
        if (userAgent.contains("Windows")) {
            return "Windows";
        }
        if (userAgent.contains("Macintosh")) {
            return "Mac";
        }
        return "未知设备";
    }

    private String limitLength(String value, int maxLength) {
        if (value == null || value.length() <= maxLength) {
            return value;
        }
        return value.substring(0, maxLength);
    }

    private int countRecentMessagesByIp(String clientIp) {
        LocalDateTime startTime = LocalDateTime.now().minusMinutes(1);
        String countSql = "select count(1) as MESSAGE_COUNT from z_lulu_message where IP_ADDRESS='" +
                escapeSql(clientIp) + "' and CREATE_TIME>='" + startTime.format(DB_TIME_FORMATTER) + "'";
        ResultBody result = callService.callFunOneParams(FunToUrlUtil.selectListUrl, "sql", countSql);

        if (result == null || result.result == null) {
            return 0;
        }
        List<Map<String, Object>> rows = (List<Map<String, Object>>) result.result;
        if (rows == null || rows.isEmpty()) {
            return 0;
        }
        return getNumber(rows.get(0), "MESSAGE_COUNT", "message_count", 0);
    }

    private int getNumber(Map<String, Object> data, String upperKey, String lowerKey, int defaultValue) {
        Object value = data.get(upperKey);
        if (value == null) {
            value = data.get(lowerKey);
        }
        if (value instanceof Number) {
            return ((Number) value).intValue();
        }
        if (value != null) {
            try {
                return Integer.parseInt(value.toString());
            } catch (NumberFormatException ignored) {
                return defaultValue;
            }
        }
        return defaultValue;
    }

    private String escapeSql(String value) {
        if (value == null) {
            return "";
        }
        return value.replace("\\", "\\\\").replace("'", "''");
    }
    private void refreshByTime(Map<String, Object> petData) {
        LocalDateTime lastUpdateTime = parseDbTime(petData.get("LAST_UPDATE_TIME"));
        LocalDateTime now = LocalDateTime.now();
        // 计算过去了多少分钟
        long minutesPassed = Duration.between(lastUpdateTime, now).toMinutes();

        // 不足 1 分钟直接跳过，不吞时间
        if (minutesPassed <= 0) {
            return;
        }

        int tick = (int) minutesPassed;

        int currentHunger = ((Number) petData.get("HUNGER")).intValue();
        int currentEnergy = ((Number) petData.get("ENERGY")).intValue();
        int currentMood = ((Number) petData.get("MOOD")).intValue();
        String currentState = (String) petData.get("CURRENT_STATE");

        int newHunger;
        int newEnergy;
        int newMood = currentMood;

        if ("SLEEPING".equals(currentState)) {
            // 【睡觉状态】：体力极速恢复！(每分钟 +4 点，25 分钟就能从 0 彻底睡到满血)
            newEnergy = Math.min(100, currentEnergy + (tick * 4));
            // 睡觉也会肚子饿：(每分钟 -1 点)
            newHunger = Math.max(0, currentHunger - tick);

            // 体力满 100，自动精神焕发地醒来
            if (newEnergy >= 100) {
                petData.put("CURRENT_STATE", "IDLE");
            }
        } else {
            // 【清醒待机】：数值掉得飞快！
            // 饥饿度：每分钟 -1 点 (1个多小时不喂就彻底饿扁)
            newHunger = Math.max(0, currentHunger - tick);
            // 体力值：每分钟 -1 点 (这样你很快就能看到它累了，就可以安排它睡觉)
            newEnergy = Math.max(0, currentEnergy - tick);

            // 情绪挂钩：如果它很饿或者很累，心情就会每分钟 -1 点往下掉
            if (newHunger <= 20 || newEnergy <= 20) {
                newMood = Math.max(0, currentMood - tick);
            }
        }

        petData.put("HUNGER", newHunger);
        petData.put("ENERGY", newEnergy);
        petData.put("MOOD", newMood);

        // 精准推移已经结算过的真实分钟数，每一秒都不会浪费
        petData.put("LAST_UPDATE_TIME", lastUpdateTime.plusMinutes(tick));

        updatePetStatusInDb(petData);
    }
    private void addExp(Map<String, Object> petData, int expToAdd) {
        int currentLevel = petData.get("LEVEL") != null ? ((Number) petData.get("LEVEL")).intValue() : 1;
        int currentExp = petData.get("EXP") != null ? ((Number) petData.get("EXP")).intValue() : 0;

        currentExp += expToAdd;

        while (currentExp >= getLevelMaxExp(currentLevel)) {
            currentExp -= getLevelMaxExp(currentLevel);
            currentLevel++;
        }

        petData.put("LEVEL", currentLevel);
        petData.put("EXP", currentExp);
    }

    private int getLevelMaxExp(int level) {
        return level * 50 + 100;
    }

    private void updatePetStatusInDb(Map<String, Object> petData) {
        Long userNum = Long.valueOf(petData.get("USER_NUM").toString());
        int hunger = ((Number) petData.get("HUNGER")).intValue();
        int energy = ((Number) petData.get("ENERGY")).intValue();
        int mood = ((Number) petData.get("MOOD")).intValue();

        // 【核心修复】：极端防空指针处理。确保任何情况下都不会再抛出 NullPointerException
        int level = petData.get("LEVEL") != null ? ((Number) petData.get("LEVEL")).intValue() : 1;
        int exp = petData.get("EXP") != null ? ((Number) petData.get("EXP")).intValue() : 0;
        String currentState = (String) petData.get("CURRENT_STATE");

        LocalDateTime time = parseDbTime(petData.get("LAST_UPDATE_TIME"));
        String timeStr = time.format(DB_TIME_FORMATTER);

        String updateSql = String.format(
                "UPDATE z_pet_status SET HUNGER=%d, ENERGY=%d, MOOD=%d, LEVEL=%d, EXP=%d, CURRENT_STATE='%s', LAST_UPDATE_TIME='%s' WHERE USER_NUM='%s'",
                hunger, energy, mood, level, exp, currentState, timeStr, userNum
        );

        callService.callFunOneParams(FunToUrlUtil.exeSqlUrl, "sql", updateSql);
    }

    private LocalDateTime parseDbTime(Object timeObj) {
        if (timeObj instanceof LocalDateTime) {
            return (LocalDateTime) timeObj;
        } else if (timeObj instanceof java.sql.Timestamp) {
            return ((java.sql.Timestamp) timeObj).toLocalDateTime();
        } else if (timeObj instanceof String) {
            String timeStr = ((String) timeObj).replace("T", " ");
            if (timeStr.contains(".")) {
                timeStr = timeStr.substring(0, timeStr.indexOf("."));
            }
            return LocalDateTime.parse(timeStr, DB_TIME_FORMATTER);
        }
        return LocalDateTime.now();
    }
}
