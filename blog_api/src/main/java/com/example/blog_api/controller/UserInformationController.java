package com.example.blog_api.controller;

import com.example.blog_api.service.UserInformationService;
import com.example.common_api.bean.ResultBody;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import javax.servlet.http.HttpSession;
import java.util.Map;

@Controller
@RequestMapping("blog-api/userInformation")
public class UserInformationController {


    @Autowired
    UserInformationService userInformationService;
    @RequestMapping("/followUser")
    @ResponseBody
    public ResultBody followUser(@RequestBody Map<String,String> params, HttpSession session){
        String followUserCode = params.get("followUserCode");
        String followUserName = params.get("followUserName");
        return userInformationService.followUser(followUserCode,followUserName,session);
    }
    @RequestMapping("/noFollowUser")
    @ResponseBody
    public ResultBody noFollowUser(@RequestBody Map<String,String> params, HttpSession session){
        String followUserCode = params.get("followUserCode");
        return userInformationService.noFollowUser(followUserCode,session);
    }
    @RequestMapping("/getFollowUser")
    @ResponseBody
    public ResultBody getFollowUser(@RequestBody Map<String,String> params){
        String userCode = params.get("userCode");
        return userInformationService.getFollowUser(userCode);
    }
    @RequestMapping("/getBlogAndResourceByUserCode")
    @ResponseBody
    public ResultBody getBlogAndResourceByUserCode(@RequestBody Map<String,String> params){
        String userCode = params.get("userCode");
        return userInformationService.getBlogAndResourceByUserCode(userCode);
    }

    @RequestMapping("/getResourceByUserCode")
    @ResponseBody
    public ResultBody getResourceByUserCode(@RequestBody Map<String,String> params){
        String userCode = params.get("userCode");
        return userInformationService.getResourceByUserCode(userCode);
    }

    @RequestMapping("/getBlogByUserCode")
    @ResponseBody
    public ResultBody getBlogByUserCode(@RequestBody Map<String,String> params){
        String userCode = params.get("userCode");
        return userInformationService.getBlogByUserCode(userCode);
    }
}
