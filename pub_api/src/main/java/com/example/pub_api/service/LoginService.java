package com.example.pub_api.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpSession;

@Service
public interface LoginService {
    //发送短信验证码
    ResultBody sendPhoneCode(String phone);
    //手机号+验证码登录
    ResultBody LoginByPhoneCode(String phone,String code,HttpSession session);
    ResultBody LoginInfo(String userName, String passWord, HttpSession session) throws Exception;
    //验证是否登录
    ResultBody checkUserLogin(HttpSession session);
    //退出登录
    ResultBody logout(HttpSession session);
    //修改密码
    ResultBody changePassWord(HttpSession session,String oldPassWord,String newPassWord,String confirmPassword);
    //注册账号
    ResultBody register(String userName, String userCode, String passWord,String successPassWord, HttpSession session);
}
