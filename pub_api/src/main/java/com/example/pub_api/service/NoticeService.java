package com.example.pub_api.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;

@Service
public interface NoticeService {
    ResultBody addNotice(String sendUserCode,String receiverUserCode,String type,String execute,String remark);
    ResultBody getNotice(String userCode);
    //已读通知
    ResultBody readNotice(String guid);
    //全部已读
    ResultBody allReadNotice(String userCode);
}
