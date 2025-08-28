package com.example.blog_api.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;

import javax.servlet.http.HttpSession;
import java.util.Map;

@Service
public interface AnnouncementService {

    ResultBody addAnnouncement(Map<String,String> announcement, HttpSession session);

    ResultBody editAnnouncement(Map<String,String> announcement);

    ResultBody deleteAnnouncement(String guid);

    ResultBody getAllAnnouncement();

    ResultBody getAnnouncementByType(String type);
}
