package com.example.blog_api.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;

import javax.servlet.http.HttpSession;

@Service
public interface UserInformationService {
    //关注用户
    ResultBody followUser(String followUserCode, String followUserName, HttpSession session);

    ResultBody noFollowUser(String followUserCode, HttpSession session);

    //获取关注用户和粉丝
    ResultBody getFollowUser(String userCode);

    ResultBody getBlogAndResourceByUserCode(String userCode);

    ResultBody getBlogAndCommunityByUserCode(String userCode, int page, int pageSize, String keyWord);

    ResultBody getCommunityByUserCode(String userCode);

    ResultBody getResourceByUserCode(String userCode);

    ResultBody getBlogByUserCode(String userCode);

    //设置用户主页背景图片和背景音乐
    ResultBody setPersonInfo(String userCode, String fieldName, String fieldValue);

    //获取用户主页设置的信息
    ResultBody getPersonInfo(String userCode);
}
