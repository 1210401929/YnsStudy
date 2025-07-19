package com.example.blog_api.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;

@Service
public interface FileService {
    ResultBody consistencyFileCheck(String type);
}
