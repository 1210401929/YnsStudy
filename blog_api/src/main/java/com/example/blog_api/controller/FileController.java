package com.example.blog_api.controller;

import com.example.blog_api.service.FileService;
import com.example.common_api.bean.ResultBody;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;

import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import java.util.Map;


@Controller
@RequestMapping("blog-api/file")
public class FileController {
    @Autowired
    FileService fileService;

    @RequestMapping("/consistencyFileCheck")
    @ResponseBody
    public ResultBody consistencyFileCheck(@RequestBody Map<String,String> params) {
        String type = params.get("type");
        return fileService.consistencyFileCheck(type);
    }
}
