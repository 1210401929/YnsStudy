package com.example.pub_api.service;

import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import org.springframework.stereotype.Service;

@Service
public interface AuthorityService {
    ResultBody getUserMenu(UserBean user);

}
