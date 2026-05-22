package com.example.blog_api.controller;

import com.example.blog_api.service.Z_LULU_Service;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
@RequestMapping("blog-api/lulu")
public class Z_LULU_Controller {

    @Autowired
    private Z_LULU_Service luluService;

    @PostMapping("/status")
    public Map<String, Object> getStatus(@RequestBody Map<String, Object> params) {
        Long userNum = Long.valueOf(params.get("userNum").toString());
        return luluService.getAndRefreshPetStatus(userNum);
    }

    @PostMapping("/feed")
    public Map<String, Object> feed(@RequestBody Map<String, Object> params) {
        Long userNum = Long.valueOf(params.get("userNum").toString());
        return luluService.feedPet(userNum);
    }

    @PostMapping("/play")
    public Map<String, Object> play(@RequestBody Map<String, Object> params) {
        Long userNum = Long.valueOf(params.get("userNum").toString());
        return luluService.playPet(userNum);
    }

    @PostMapping("/sleep")
    public Map<String, Object> sleep(@RequestBody Map<String, Object> params) {
        Long userNum = Long.valueOf(params.get("userNum").toString());
        return luluService.toggleSleepPet(userNum);
    }
}
