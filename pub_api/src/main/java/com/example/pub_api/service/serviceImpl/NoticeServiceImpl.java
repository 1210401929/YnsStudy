package com.example.pub_api.service.serviceImpl;

import com.example.common_api.bean.ResultBody;
import com.example.pub_api.service.NoticeService;
import com.example.pub_api.service.SqlService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Component
public class NoticeServiceImpl implements NoticeService {
    @Autowired
    SqlService sqlService;

    @Override
    public ResultBody addNotice(String sendUserCode, String receiverUserCode, String type, String execute, String remark) {
        if (sendUserCode == null || receiverUserCode == null || type == null || execute == null) {
            return ResultBody.createErrorResult("传递参数不全,请传递完整参数结构!");
        }
        HashMap<String, Object> mapData = new HashMap<>();
        mapData.put("SENDUSERCODE", sendUserCode);
        mapData.put("RECEIVERUSERCODE", receiverUserCode);
        mapData.put("TYPE", type);
        mapData.put("EXECUTE", execute);
        mapData.put("REMARK", remark);
        ArrayList<HashMap<String, Object>> data = new ArrayList<>();
        data.add(mapData);
        ResultBody result = sqlService.saveAllTableDataByParams("add", "noticeInfo", data, "GUID");
        if (result != null && !result.isError) {
            return ResultBody.createSuccessResult("新增消息成功!");
        } else {
            return ResultBody.createErrorResult("新增消息失败,请检查!");
        }
    }

    @Override
    public ResultBody getNotice(String userCode) {
        String sql = "select * from noticeInfo where RECEIVERUSERCODE = '" + userCode + "' order by CREATE_TIME DESC";
        ResultBody result = sqlService.selectList(sql);
        return result;
    }

    @Override
    public ResultBody readNotice(String guid) {
        String sql = "delete from noticeInfo where guid = '" + guid + "'";
        ResultBody result = sqlService.exeSql(sql);
        return result;
    }

    @Override
    public ResultBody allReadNotice(String userCode) {
        String sql = "delete from noticeInfo where RECEIVERUSERCODE = '" + userCode + "'";
        ResultBody result = sqlService.exeSql(sql);
        return result;
    }
}
