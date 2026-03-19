package com.example.blog_api.service;

import com.example.blog_api.Bean.BlogBean;
import com.example.common_api.bean.ResultBody;

import javax.servlet.http.HttpSession;
import java.util.Map;

public interface BlogService {
    //-------------------------------------博客相关----------------------------------------------------------------
    ResultBody addBlog(BlogBean blogBean);

    //获取用户所有博客内容(当前登录用户)
    ResultBody getUserBlog(HttpSession session);

    //获取用户所有博客内容（指定用户）
    ResultBody getUserBlogByUserCode(String userCode,String blogType);

    //获取用户所有博客的标题和唯一标识
    ResultBody getUserBlogTitle(HttpSession session);

    //获取用户博客数量
    ResultBody getUserBlogNum(HttpSession session);

    //根据分类id获取分类下的博客标题和唯一标识
    ResultBody getBlogTitleByCatId(String catId);

    //修改文章的分类
    ResultBody updateBlogCatId(String blogId, String catId);

    //根据博客唯一标识获取博客内容
    ResultBody getBlog(HttpSession session, String blogId);

    //获取系统所有博客内容
    ResultBody getAllBlog(int page, int pageSize, String keyWord);

    ResultBody updateBlog(String guid, String title, String content, String blog_type);

    ResultBody deleteBlog(String guid);

    //-------------------------------------博客分类相关----------------------------------------------------------------
    ResultBody addBlogCat(Map<String, String> blogCat);

    ResultBody updateBlogCat(Map<String, String> blogCat);

    ResultBody deleteBlogCat(String guid);

    ResultBody getUserBlogCat(String userCode);

    //-------------------------------------评论相关----------------------------------------------------------------
    //新增评论
    ResultBody addComment(Map<String, String> blogComment);

    //获取评论
    ResultBody getComment(String blogId);

    //删除评论
    ResultBody deleteComment(String CommentGuid);

    //---------------------------------------点赞收藏相关-----------------------------------------------------------------
    //点赞调用
    ResultBody giveLikeBlog(String blogId, HttpSession session);

    //取消点赞调用
    ResultBody noGiveLikeBlog(String blogId, HttpSession session);

    //获取点赞
    ResultBody getGiveLikeByBlogId(String blogId);

    //收藏调用
    ResultBody collectBlog(String blogId, HttpSession session);

    //取消收藏调用
    ResultBody noCollectBlog(String blogId, HttpSession session);

    //获取收藏  根据博客
    ResultBody getCollectByBlogId(String blogId);

    //获取收藏  根据用户
    ResultBody getCollectByUserCode(String userCode);

    //获取收藏和点赞  根据博客ID
    ResultBody getLikeAndCollectByBlogId(String blogId);

    //获取收藏和点赞  根据用户
    ResultBody getLikeAndCollectByUserCode(String userCode, String isCountOnly, String type);

}
