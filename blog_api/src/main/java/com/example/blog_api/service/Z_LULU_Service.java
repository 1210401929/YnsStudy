package com.example.blog_api.service;

import java.util.Map;

public interface Z_LULU_Service {

    Map<String, Object> getAndRefreshPetStatus(Long userNum);

    Map<String, Object> feedPet(Long userNum);

    Map<String, Object> playPet(Long userNum);

    Map<String, Object> toggleSleepPet(Long userNum);
}
