package com.example.blog_api.controller;

import com.example.blog_api.service.HomeService;
import com.example.common_api.bean.ResultBody;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

@Controller
@RequestMapping("blog-api/home")
public class HomeController {

    @Autowired
    HomeService homeService;

    @RequestMapping("/getHomeData")
    @ResponseBody
    public ResultBody getHomeData() {
        return homeService.getHomeData();
    }

    @RequestMapping("/getHigAuthor")
    @ResponseBody
    public ResultBody getHigAuthor() {
        return homeService.getHigAuthor();
    }
}
