package com.example.blog_api.service;

import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import org.springframework.stereotype.Service;

import javax.servlet.http.HttpSession;
import java.util.Map;

@Service
public interface ResourceService {

    ResultBody addFileInfo(HttpSession session, Map<String, Object> fileInfo);

    ResultBody delFileInfo(String guid,String url);

    ResultBody getAllFile();

    ResultBody getFileByUser(UserBean userBean);

    ResultBody setFileDownNum(String guid);
}
