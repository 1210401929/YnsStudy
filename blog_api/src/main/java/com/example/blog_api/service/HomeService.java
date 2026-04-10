package com.example.blog_api.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;

@Service
public interface HomeService {
    ResultBody getHomeData();
    //获取网站统计数据
    ResultBody getWebsiteStatistics();
    //获取优质作者
    ResultBody getHigAuthor(String num);
    //获取所有博客文章的guid和最新时间
    ResultBody getAllBlogGuidAndTime();
    //获取前五条博客文章的信息
    ResultBody getBlogDetail5();
}
