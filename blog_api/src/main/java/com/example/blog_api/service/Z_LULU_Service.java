package com.example.blog_api.service;

import java.util.Map;
import java.util.List;
import com.example.common_api.bean.ResultBody;

public interface Z_LULU_Service {

    Map<String, Object> getAndRefreshPetStatus(Long userNum);

    Map<String, Object> feedPet(Long userNum, String clientIp, String userAgent);

    Map<String, Object> playPet(Long userNum, String actionName, String clientIp, String userAgent);

    Map<String, Object> toggleSleepPet(Long userNum, String clientIp, String userAgent);

    List<Map<String, Object>> getMessages(Long userNum);

    ResultBody addMessage(Long userNum, String content, String clientIp);

    List<Map<String, Object>> deleteMessage(Long userNum, Long messageId);

    List<Map<String, Object>> getLogs(Long userNum);

    Map<String, Object> getFunState(Long userNum);

    Map<String, Object> advanceMission(Long userNum, String missionType, int amount);

    Map<String, Object> changeClothes(Long userNum);
}
