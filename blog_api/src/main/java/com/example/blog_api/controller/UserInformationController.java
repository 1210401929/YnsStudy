package com.example.blog_api.controller;

import com.example.blog_api.service.UserInformationService;
import com.example.common_api.bean.ResultBody;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import java.util.Map;

@Controller
@RequestMapping("blog-api/userInformation")
public class UserInformationController {


    @Autowired
    UserInformationService userInformationService;


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
