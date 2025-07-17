package com.example.blog_api.controller;

import com.example.blog_api.service.CommunityService;
import com.example.blog_api.service.serviceImpl.CommunityServiceImpl;
import com.example.common_api.bean.ResultBody;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseBody;

import javax.servlet.http.HttpSession;
import java.util.Map;

@Controller
@RequestMapping("blog-api/community")
public class CommunityController {

    @Autowired
    CommunityService communityService;

    @RequestMapping("/addCommunity")
    @ResponseBody
    public ResultBody addCommunity(@RequestBody Map<String, Object> params, HttpSession session) {
        Map<String, Object> community = (Map<String, Object>) params.get("community");
        return communityService.addCommunity(community, session);
    }
    @RequestMapping("/deleteCommunity")
    @ResponseBody
    public ResultBody deleteCommunity(@RequestBody Map<String, Object> params) {
        String communityGuid = (String) params.get("communityGuid");
        return communityService.deleteCommunity(communityGuid);
    }
    @RequestMapping("/setTopCommunity")
    @ResponseBody
    public ResultBody setTopCommunity(@RequestBody Map<String, Object> params) {
        String communityGuid = (String) params.get("communityGuid");
        String isTop = (String) params.get("isTop");
        return communityService.setTopCommunity(communityGuid, isTop);
    }

    @RequestMapping("/getAllCommunity")
    @ResponseBody
    //required:false 表示不是必须需要的参数
    public ResultBody getAllCommunity(@RequestBody Map<String, Object> params) {
        int page = (int)params.get("page");
        int pageSize = (int)params.get("pageSize");
        String keyword = (String) params.get("keyword");
        return communityService.getAllCommunity(page, pageSize, keyword);
    }

    @RequestMapping("/addComment")
    @ResponseBody
    public ResultBody addComment(@RequestBody Map<String, Object> params, HttpSession session) {
        Map<String, String> comment = (Map<String, String>) params.get("comment");
        return communityService.addComment(comment, session);
    }

    @RequestMapping("/getComment")
    @ResponseBody
    public ResultBody getComment(@RequestBody Map<String, Object> params) {
        String communityId = (String) params.get("communityId");
        return communityService.getComment(communityId);
    }

}
