package com.example.yns_portal.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpSession;

@Service
public interface LoginService {
    ResultBody LoginInfo(String userName, String passWord, HttpSession session, HttpServletRequest request) throws Exception;

    ResultBody checkUserLogin(HttpSession session);

    ResultBody logout(HttpSession session);

    ResultBody changePassWord(HttpSession session,String oldPassWord,String newPassWord,String confirmPassword);

    ResultBody register(String userName, String userCode, String passWord,String successPassWord, HttpSession session);
}
