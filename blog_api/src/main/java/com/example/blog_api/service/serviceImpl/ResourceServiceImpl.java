package com.example.blog_api.service.serviceImpl;

import com.example.blog_api.service.ResourceService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import com.example.common_api.service.CallService;
import com.example.common_api.util.FunToUrlUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import javax.servlet.http.HttpSession;
import java.util.*;

@Component
public class ResourceServiceImpl implements ResourceService {
    @Autowired
    CallService callService;

    @Override
    public ResultBody addFileInfo(HttpSession session, Map<String, Object> fileInfo) {
        Map<String, Object> params = new HashMap<>();
        List<Map<String, Object>> data = new ArrayList<>();
        data.add(fileInfo);
        params.put("saveType", "add");
        params.put("tableName", "FILEINFO");
        params.put("data", data);
        params.put("key", "GUID");
        return callService.callFunWithParams(FunToUrlUtil.saveAllTableDataByParamsUrl, params);
    }

    @Override
    @Transactional //开启事务
    public ResultBody delFileInfo(String guid,String url) {
        if(url==null){
            ResultBody.createErrorResult("路径不存在或不正确!");
        }
        String sql = "delete from fileInfo where guid = '"+guid+"'";
        ResultBody result = callService.callFunOneParams(FunToUrlUtil.exeSqlUrl,"sql",sql);
        if(result!=null && !result.isError){
            Map<String,Object> params = new HashMap<>();
            params.put("url",url);
            result = callService.callFunWithParams(FunToUrlUtil.deleteFileByUrlUrl,params);
            return result;
        }else{
            return ResultBody.createErrorResult("删除失败!");
        }
    }

    @Override
    public ResultBody getAllFile(int page, int pageSize, String keyWord) {
        // 0. 容错处理
        if (page < 1) page = 1;
        if (pageSize <= 0) pageSize = 10;
        int offset = (page - 1) * pageSize;

        // 1) 主查询 SQL 与参数
        String baseFrom = " FROM fileInfo f ";
        String where = "";
        List<Object> listParams = new ArrayList<>();
        List<Object> countParams = new ArrayList<>();

        // 关键字搜索
        if (keyWord != null && !keyWord.trim().isEmpty()) {
            where = " WHERE (f.ORIGINALFILENAME LIKE ? OR f.REMARK LIKE ?) ";
            String kw = "%" + keyWord.trim() + "%";
            listParams.add(kw);
            listParams.add(kw);
            countParams.add(kw);
            countParams.add(kw);
        }

        String listSql = "SELECT f.* " + baseFrom + where +
                " ORDER BY f.create_time DESC LIMIT ? OFFSET ?";
        listParams.add(pageSize);
        listParams.add(offset);

        // 2) 统计总数 SQL 与参数
        String countSql = "SELECT COUNT(*) AS total " + baseFrom + where;

        // 3) 查询数据列表
        Map<String, Object> params = new HashMap<>();
        params.put("sql", listSql);
        params.put("params", listParams);
        ResultBody listResult = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        if (listResult == null || listResult.isError) {
            return listResult;
        }

        // 4) 查询总数
        params = new HashMap<>();
        params.put("sql", countSql);
        params.put("params", countParams);
        ResultBody countResult = callService.callFunWithParams(FunToUrlUtil.selectListByParamsUrl, params);
        if (countResult == null || countResult.isError) {
            return countResult;
        }

        // 5) 解析总数
        int total = 0;
        List<Map<String, Object>> countData = (List<Map<String, Object>>) countResult.result;
        if (countData != null && !countData.isEmpty() && countData.get(0).get("total") != null) {
            total = Integer.parseInt(countData.get(0).get("total").toString());
        }

        // 6) 组装分页结果
        Map<String, Object> result = new HashMap<>();
        result.put("total", total);
        result.put("data", listResult.result);

        return ResultBody.createSuccessResult(result);
    }


    @Override
    public ResultBody getFileByUser(String userCode) {
        String sql = "select * from fileInfo where userCode = '" + userCode + "' order by create_time desc";
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params);
        return result;
    }

    @Override
    public ResultBody getFileById(String guid) {
        String sql = "select * from fileInfo where guid = '" + guid + "'";
        Map<String, Object> params = new HashMap<>();
        params.put("sql", sql);
        ResultBody result = callService.callFunWithParams(FunToUrlUtil.selectListUrl, params);
        return result;
    }

    @Override
    public ResultBody updateFileInfo(String guid, String originalFileName, String remark) {
        String sql = "update fileInfo set ORIGINALFILENAME = ?,REMARK = ? where guid = ?";
        List<String> listParam = Arrays.asList(originalFileName,remark,guid);
        Map<String,Object> params = new HashMap<>();
        params.put("sql",sql);
        params.put("params",listParam);
        ResultBody resultBody = callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl,params);
        return resultBody;
    }

    @Override
    public ResultBody setFileDownNum(String guid) {
        String sql = "update fileInfo set downNum = downNum+1 where guid = ?";
        List<String> listParam = Arrays.asList(guid);
        Map<String,Object> params = new HashMap<>();
        params.put("sql",sql);
        params.put("params",listParam);
        ResultBody resultBody = callService.callFunWithParams(FunToUrlUtil.exeSqlByParamsUrl,params);
        return resultBody;
    }
}
