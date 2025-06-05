package com.example.pub_api.controller;

import com.example.common_api.bean.ResultBody;
import com.example.pub_api.service.SqlService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Controller
@RequestMapping("pub-api/sql")
public class SqlController {
    @Autowired
    SqlService sqlService;

    @RequestMapping("/selectList")
    @ResponseBody
    public ResultBody selectList(@RequestBody Map<String, String> params) {
        String sql = params.get("sql");
        return sqlService.selectList(sql);
    }

    @RequestMapping("/selectListByParams")
    @ResponseBody
    public ResultBody selectListByParams(@RequestBody Map<String, Object> params) {
        String sql = (String) params.get("sql");
        List<Object> param = (List<Object>) params.get("params");
        return sqlService.selectListByParams(sql, param);
    }

    @RequestMapping("/deleteBySql")
    @ResponseBody
    public ResultBody deleteBySql(String sql) {
        return sqlService.deleteBySql(sql);
    }

    @RequestMapping("/updateBySql")
    @ResponseBody
    public ResultBody updateBySql(String sql) {
        return sqlService.updateBySql(sql);
    }

    @RequestMapping("/exeSql")
    @ResponseBody
    public ResultBody exeSql(String sql) {
        return sqlService.exeSql(sql);
    }

    @RequestMapping("/exeSqlByParams")
    @ResponseBody
    public ResultBody exeSqlByParams(@RequestBody Map<String, Object> params) {
        String sql = (String) params.get("sql");
        List<Object> param = (List<Object>) params.get("params");
        return sqlService.exeSqlByParams(sql, param);
    }

    @RequestMapping("/saveAllTableData")
    @ResponseBody
    public ResultBody saveAllTableData(@RequestBody Map<String, Object> params) {
        // 从 params 中提取各个字段
        String saveType = (String) params.get("saveType");
        String tableName = (String) params.get("tableName");
        ArrayList<HashMap<String, Object>> data = (ArrayList<HashMap<String, Object>>) params.get("data");
        String key = (String) params.get("key");
        // 调用业务逻辑
        return sqlService.saveAllTableData(saveType, tableName, data, key);
    }

    @RequestMapping("/saveAllTableDataByParams")
    @ResponseBody
    public ResultBody saveAllTableDataByParams(@RequestBody Map<String, Object> params) {
        // 从 params 中提取各个字段
        String saveType = (String) params.get("saveType");
        String tableName = (String) params.get("tableName");
        ArrayList<HashMap<String, Object>> data = (ArrayList<HashMap<String, Object>>) params.get("data");
        String key = (String) params.get("key");
        // 调用业务逻辑
        return sqlService.saveAllTableDataByParams(saveType, tableName, data, key);
    }
}
