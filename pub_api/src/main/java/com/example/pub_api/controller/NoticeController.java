package com.example.pub_api.controller;

import com.example.common_api.bean.ResultBody;
import com.example.pub_api.service.NoticeService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import java.util.Map;

@Controller
@RequestMapping("pub-api/notice")
public class NoticeController {
    @Autowired
    NoticeService noticeService;

    @RequestMapping("/addNotice")
    @ResponseBody
    public ResultBody addNotice(@RequestBody Map<String, String> params) {
        String sendUserCode = params.get("sendUserCode");
        String receiverUserCode = params.get("receiverUserCode");
        String type = params.get("type");
        String execute = params.get("execute");
        String remark = params.get("remark");
        return noticeService.addNotice(sendUserCode, receiverUserCode, type, execute, remark);
    }

    @RequestMapping("/getNotice")
    @ResponseBody
    public ResultBody getNotice(@RequestBody Map<String, String> params) {
        String userCode = params.get("userCode");
        return noticeService.getNotice(userCode);
    }

    @RequestMapping("/readNotice")
    @ResponseBody
    public ResultBody readNotice(@RequestBody Map<String, String> params) {
        String guid = params.get("guid");
        return noticeService.readNotice(guid);
    }

    @RequestMapping("/allReadNotice")
    @ResponseBody
    public ResultBody allReadNotice(@RequestBody Map<String, String> params) {
        String userCode = params.get("userCode");
        return noticeService.allReadNotice(userCode);
    }
}
