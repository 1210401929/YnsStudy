package com.example.blog_api.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;

@Service
public interface UserInformationService {

    ResultBody getBlogAndResourceByUserCode(String userCode);

    ResultBody getResourceByUserCode(String userCode);

    ResultBody getBlogByUserCode(String userCode);
}
