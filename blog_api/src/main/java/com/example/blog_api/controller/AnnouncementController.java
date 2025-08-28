package com.example.blog_api.controller;

import com.example.blog_api.service.AnnouncementService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseBody;

import javax.servlet.http.HttpSession;
import java.util.Map;

@Controller
@RequestMapping("blog-api/sso")
public class AnnouncementController {
    @Autowired
    AnnouncementService announcementService;

    @RequestMapping("addAnnouncement")
    @ResponseBody
    public ResultBody addAnnouncement(@RequestBody Map<String, Object> params, HttpSession session) {
        Map<String, String> announcement = (Map<String, String>) params.get("announcement");
        return announcementService.addAnnouncement(announcement, session);
    }

    @RequestMapping("editAnnouncement")
    @ResponseBody
    public ResultBody editAnnouncement(@RequestBody Map<String, Object> params) {
        Map<String, String> announcement = (Map<String, String>) params.get("announcement");
        return announcementService.editAnnouncement(announcement);
    }

    @RequestMapping("getAllAnnouncement")
    @ResponseBody
    public ResultBody getAllAnnouncement() {
        return announcementService.getAllAnnouncement();
    }

    @RequestMapping("getAnnouncementByType")
    @ResponseBody
    public ResultBody getAnnouncementByType(@RequestBody Map<String, String> params) {
        String type = params.get("type");
        return announcementService.getAnnouncementByType(type);
    }

    @RequestMapping("deleteAnnouncement")
    @ResponseBody
    public ResultBody deleteAnnouncement(@RequestBody Map<String, String> params) {
        String guid = params.get("guid");
        return announcementService.deleteAnnouncement(guid);
    }
}
