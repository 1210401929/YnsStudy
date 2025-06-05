package com.example.pub_api.controller;

import com.example.common_api.bean.ResultBody;
import com.example.pub_api.service.LoginService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpSession;
import java.util.Map;


@Controller
@RequestMapping("pub-api/login")
public class LoginController {
    @Autowired
    LoginService loginService;

    @RequestMapping("/sendPhoneCode")
    @ResponseBody
    public ResultBody sendPhoneCode(@RequestBody Map<String, String> params) {
        String phone = params.get("phone");
        return loginService.sendPhoneCode(phone);
    }

    @RequestMapping("/loginByPhoneCode")
    @ResponseBody
    public ResultBody LoginByPhoneCode(@RequestBody Map<String, String> params, HttpSession session) {
        String phone = params.get("phone");
        String code = params.get("code");
        return loginService.LoginByPhoneCode(phone, code, session);
    }

    @RequestMapping("/logout")
    @ResponseBody
    public ResultBody logout(HttpSession session) {
        ResultBody resultBody = loginService.logout(session);
        return resultBody;
    }

    @RequestMapping("/loginUser")
    @ResponseBody
    public ResultBody loginUser(@RequestBody Map<String, String> paramMap, HttpSession session, HttpServletRequest request) throws Exception {
        String userName = paramMap.get("userName");
        String passWord = paramMap.get("passWord");
        ResultBody resultBody = loginService.LoginInfo(userName, passWord, session);
        return resultBody;
    }

    @RequestMapping("/checkUserLogin")
    @ResponseBody
    public ResultBody checkUserLogin(HttpSession session, HttpServletRequest request) {
        ResultBody resultBody = loginService.checkUserLogin(session);
        return resultBody;
    }

    @RequestMapping("/changePassWord")
    @ResponseBody
    public ResultBody changePassWord(String oldPassWord, String newPassWord, String confirmPassword, HttpSession session) {
        ResultBody resultBody = loginService.changePassWord(session, oldPassWord, newPassWord, confirmPassword);
        return resultBody;
    }

    @RequestMapping("/register")
    @ResponseBody
    public ResultBody register(@RequestBody Map<String, String> params, HttpSession session) {
        String userName = params.get("userName");
        String userCode = params.get("userCode");
        String passWord = params.get("passWord");
        String successPassWord = params.get("successPassWord");
        ResultBody resultBody = loginService.register(userName, userCode, passWord, successPassWord, session);
        return resultBody;
    }
}
