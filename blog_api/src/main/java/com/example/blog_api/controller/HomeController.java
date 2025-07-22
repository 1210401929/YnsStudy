package com.example.blog_api.controller;

import com.example.blog_api.service.HomeService;
import com.example.common_api.bean.ResultBody;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import javax.jws.WebResult;
import java.util.Map;

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

    @RequestMapping("/getWebsiteStatistics")
    @ResponseBody
    public ResultBody getWebsiteStatistics() {
        return homeService.getWebsiteStatistics();
    }

    @RequestMapping("/getHigAuthor")
    @ResponseBody
    public ResultBody getHigAuthor(@RequestBody Map<String,String> params) {
        String num = params.get("num");
        return homeService.getHigAuthor(num);
    }
}
