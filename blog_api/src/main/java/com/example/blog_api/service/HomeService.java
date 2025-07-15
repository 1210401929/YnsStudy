package com.example.blog_api.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;

@Service
public interface HomeService {
    ResultBody getHomeData();

    //获取优质作者
    ResultBody getHigAuthor(String num);
}
