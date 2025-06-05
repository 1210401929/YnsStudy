package com.example.pub_api.controller;

import com.example.common_api.bean.ResultBody;
import com.example.pub_api.service.ApiService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import java.io.IOException;

@Controller
@RequestMapping("pub-api/api")
public class ApiController {
    @Autowired
    ApiService apiService;

    @RequestMapping("/getCurrentCity")
    @ResponseBody
    public ResultBody getCurrentCity() throws IOException {
        return apiService.getCurrentCity();
    }
    @RequestMapping("/getClientIpAddress")
    @ResponseBody
    public ResultBody getClientIpAddress() {
        return apiService.getClientIpAddress();
    }
}
