package com.example.pub_api.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;

import java.io.IOException;

@Service
public interface ApiService {
    ResultBody getCurrentCity() throws IOException;

    ResultBody getClientIpAddress();
}
