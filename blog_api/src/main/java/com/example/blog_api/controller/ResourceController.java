package com.example.blog_api.controller;

import com.example.blog_api.service.ResourceService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import javax.servlet.http.HttpSession;
import java.util.Map;

@Controller
@RequestMapping("blog-api/resource")
public class ResourceController {

    @Autowired
    ResourceService resourceService;

    @RequestMapping("/addFileInfo")
    @ResponseBody
    public ResultBody addFileInfo(@RequestBody Map<String, Object> params, HttpSession session) {
        Map<String, Object> fileInfo = (Map<String, Object>) params.get("fileInfo");
        return resourceService.addFileInfo(session, fileInfo);
    }

    @RequestMapping("/delFileInfo")
    @ResponseBody
    public ResultBody delFileInfo(@RequestBody Map<String, Object> params) {
        String guid = (String) params.get("guid");
        String url = (String) params.get("url");
        return resourceService.delFileInfo(guid, url);
    }

    @RequestMapping("/getAllFile")
    @ResponseBody
    public ResultBody getAllFile(@RequestBody Map<String, Object> params) {
        int page = (Integer) params.get("page");
        int pageSize = (Integer) params.get("pageSize");
        String keyword = (String) params.get("keyword");
        return resourceService.getAllFile(page, pageSize, keyword);
    }

    @RequestMapping("/getFileByUser")
    @ResponseBody
    public ResultBody getFileByUser(@RequestBody Map<String, String> params) {
        String userCode = params.get("userCode");
        return resourceService.getFileByUser(userCode);
    }
    @RequestMapping("/updateFileInfo")
    @ResponseBody
    public ResultBody updateFileInfo(@RequestBody Map<String, String> params) {
        String guid = params.get("guid");
        String originalFileName = params.get("originalFileName");
        String remark = params.get("remark");
        return resourceService.updateFileInfo(guid,originalFileName,remark);
    }
    @RequestMapping("/setFileDownNum")
    @ResponseBody
    public ResultBody setFileDownNum(@RequestBody Map<String, String> params) {
        String guid = params.get("guid");
        return resourceService.setFileDownNum(guid);
    }
}
