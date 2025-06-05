package com.example.yns_portal.controller;

import com.example.common_api.bean.ResultBody;
import com.example.yns_portal.service.LoginService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpSession;


@Controller
@RequestMapping
public class LoginController {
    @Autowired
    LoginService loginService;

    @RequestMapping("index")
    public String indexHtml() {return "index.html";}

    @RequestMapping({"welcome","/",""})
    public String welcomeHtml() {return "welcome.html";}

    @RequestMapping("logout")
    public ResultBody logout(HttpSession session) {
        ResultBody resultBody = loginService.logout(session);
        return resultBody;
    }

    @RequestMapping("loginUser")
    @ResponseBody
    public ResultBody loginUser(String userName, String passWord, HttpSession session, HttpServletRequest request) throws Exception {
        ResultBody resultBody = loginService.LoginInfo(userName, passWord, session, request);
        return resultBody;
    }
    @RequestMapping("checkUserLogin")
    @ResponseBody
    public ResultBody checkUserLogin(HttpSession session, HttpServletRequest request) {
        ResultBody resultBody = loginService.checkUserLogin(session);
        return resultBody;
    }

    @RequestMapping("changePassWord")
    @ResponseBody
    public ResultBody changePassWord(String oldPassWord, String newPassWord, String confirmPassword, HttpSession session) {
        ResultBody resultBody = loginService.changePassWord(session, oldPassWord, newPassWord, confirmPassword);
        return resultBody;
    }



    @RequestMapping("register")
    @ResponseBody
    public ResultBody register(String userName, String userCode, String passWord, String successPassWord, HttpSession session) {
        ResultBody resultBody = loginService.register(userName, userCode, passWord, successPassWord, session);
        return resultBody;
    }
}
