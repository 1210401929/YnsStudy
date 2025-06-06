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
        return resourceService.delFileInfo(guid,url);
    }

    @RequestMapping("/getAllFile")
    @ResponseBody
    public ResultBody getAllFile() {
        return resourceService.getAllFile();
    }

    @RequestMapping("/getFileByUser")
    @ResponseBody
    public ResultBody getFileByUser(HttpSession session) {
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        if (userBean == null) {
            ResultBody.createErrorResult("用户未登录");
        }
        return resourceService.getFileByUser(userBean);
    }

    @RequestMapping("/setFileDownNum")
    @ResponseBody
    public ResultBody setFileDownNum(@RequestBody Map<String,String> params){
        String guid = params.get("guid");
        return resourceService.setFileDownNum(guid);
    }
}
