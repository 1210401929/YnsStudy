package com.example.blog_api.service;

import com.example.blog_api.Bean.BlogBean;
import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;

import javax.servlet.http.HttpSession;
import java.util.Map;

@Service
public interface CommunityService {
    //新增社区内容
    ResultBody addCommunity(Map<String,Object> community, HttpSession session);
    ResultBody setTopCommunity(String CommentGuid,String isTop);
    //删除社区内容
    ResultBody deleteCommunity(String guid);
    ResultBody getCommunity(String guid);
    ResultBody getAllCommunity(int page,int pageSize ,String keyWord);
    //新增评论
    ResultBody addComment(Map<String, String> communityComment,HttpSession session);
    //获取评论
    ResultBody getComment(String communityId);
    //删除评论
    ResultBody deleteComment(String CommentGuid);
}
