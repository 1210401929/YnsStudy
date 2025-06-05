package com.example.pub_api.controller;

import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import com.example.pub_api.service.AuthorityService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import javax.servlet.http.HttpSession;

@Controller
@RequestMapping("pub-api/authority")
public class AuthorityController {

    @Autowired
    AuthorityService authorityService;

    @RequestMapping("/getUserMenu")
    @ResponseBody
    public ResultBody getUserMenu(HttpSession session){
        UserBean user = (UserBean) session.getAttribute("userInfo");
        System.out.println("准备返回用户所属菜单");
        return authorityService.getUserMenu(user);
    };
}
