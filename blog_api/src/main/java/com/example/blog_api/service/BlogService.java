package com.example.blog_api.service;

import com.example.blog_api.Bean.BlogBean;
import com.example.common_api.bean.ResultBody;

import javax.servlet.http.HttpSession;
import java.util.Map;

public interface BlogService {
    ResultBody addBlog(BlogBean blogBean);

    ResultBody getUserBlog(HttpSession session);

    ResultBody getBlog(HttpSession session,String blogId);

    ResultBody getAllBlog(int page,int pageSize ,String keyWord);

    ResultBody updateBlog(String guid, String title, String content);

    ResultBody deleteBlog(String guid);

    //新增评论
    ResultBody addComment(Map<String, String> blogComment);

    //获取评论
    ResultBody getComment(String blogId);

    //删除评论
    ResultBody deleteComment(String CommentGuid);
}
