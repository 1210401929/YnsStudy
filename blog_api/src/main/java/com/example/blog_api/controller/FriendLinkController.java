package com.example.blog_api.controller;

import com.example.blog_api.service.FriendLinkService;
import com.example.common_api.bean.ResultBody;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import java.util.HashMap;
import java.util.Map;

@Controller
@RequestMapping("blog-api/friendLink")
public class FriendLinkController {

    @Autowired
    FriendLinkService friendLinkService;

    @RequestMapping("/addFriendLink")
    @ResponseBody
    public ResultBody addFriendLink(@RequestBody Map<String, Object> params) {
        HashMap<String, Object> friendLink = (HashMap<String, Object>) params.get("friendLink");
        return friendLinkService.addFriendLink(friendLink);
    }

    @RequestMapping("/updateFriendLink")
    @ResponseBody
    public ResultBody updateFriendLink(@RequestBody Map<String, Object> params) {
        HashMap<String, Object> friendLink = (HashMap<String, Object>) params.get("friendLink");
        return friendLinkService.updateFriendLink(friendLink);
    }

    @RequestMapping("/deleteFriendLink")
    @ResponseBody
    public ResultBody deleteFriendLink(@RequestBody Map<String, String> params) {
        String friendLinkId = params.get("friendLinkId");
        return friendLinkService.deleteFriendLink(friendLinkId);
    }

    @RequestMapping("/getFriendLink")
    @ResponseBody
    public ResultBody getFriendLink(@RequestBody Map<String, String> params) {
        String friendLinkId = params.get("friendLinkId");
        return friendLinkService.getFriendLink(friendLinkId);
    }

    @RequestMapping("/getFriendLinks")
    @ResponseBody
    public ResultBody getFriendLinks() {
        return friendLinkService.getFriendLinks();
    }
}
