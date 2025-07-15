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
    ResultBody updateBlog(String guid, String title, String content,String blog_type);
    ResultBody deleteBlog(String guid);

    //新增评论
    ResultBody addComment(Map<String, String> blogComment);
    //获取评论
    ResultBody getComment(String blogId);
    //删除评论
    ResultBody deleteComment(String CommentGuid);

    //点赞调用
    ResultBody giveLikeBlog(String blogId,HttpSession session);
    //取消点赞调用
    ResultBody noGiveLikeBlog(String blogId,HttpSession session);
    //获取点赞
    ResultBody getGiveLikeByBlogId(String blogId);

    //收藏调用
    ResultBody collectBlog(String blogId,HttpSession session);
    //取消收藏调用
    ResultBody noCollectBlog(String blogId,HttpSession session);
    //获取收藏  根据博客
    ResultBody getCollectByBlogId(String blogId);
    //获取收藏  根据用户
    ResultBody getCollectByUserCode(String userCode);

    //获取收藏和点赞  根据博客ID
    ResultBody getLikeAndCollectByBlogId(String blogId);
    //获取收藏和点赞  根据用户
    ResultBody getLikeAndCollectByUserCode(String userCode);
}
