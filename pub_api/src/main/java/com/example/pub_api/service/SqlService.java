package com.example.pub_api.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

@Service
public interface SqlService {
    ResultBody saveAllTableData(String saveType, String tableName, ArrayList<HashMap<String, Object>> data, String key);
    ResultBody saveAllTableDataByParams(String saveType, String tableName, ArrayList<HashMap<String, Object>> data, String key);
    ResultBody selectList(String sql);
    ResultBody selectListByParams(String sql, List<Object> params);
    ResultBody deleteBySql(String sql);

    ResultBody updateBySql(String sql);

    ResultBody exeSql(String sql);

    //参数形式的exeSql  更安全 更完善
    ResultBody exeSqlByParams(String sql, List<Object> params);
}
