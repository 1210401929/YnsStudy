package com.example.pub_api.service.serviceImpl;


import com.example.common_api.bean.ResultBody;
import com.example.pub_api.mapper.SqlMapper;
import com.example.pub_api.service.SqlService;
import org.apache.ibatis.exceptions.PersistenceException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import javax.sql.DataSource;
import java.sql.*;
import java.util.*;

@Component
public class SqlServiceImpl implements SqlService {

    @Autowired
    private SqlMapper sqlMapper;

    @Autowired
    DataSource dataSource;

    @Override
    @Transactional//开启事务
    public ResultBody saveAllTableData(String saveType, String tableName, ArrayList<HashMap<String, Object>> data, String key) {

        String sql = new String();
        if (saveType.equals("add")) {
            for (HashMap<String, Object> oneData : data) {
                Set<String> setDataInfo = oneData.keySet();
                ArrayList<String> dataInfo = new ArrayList<>(setDataInfo);
                ArrayList<String> dataValue = new ArrayList<>();
                for (String info : dataInfo) {
                    dataValue.add("'" + String.valueOf(oneData.get(info)) + "'");
                }
                sql += " insert into " + tableName + " (" + String.join(" , ", dataInfo) + ") values(" + String.join(" , ", dataValue) + "); ";
            }
        } else if (saveType.equals("edit")) {
            for (HashMap<String, Object> oneData : data) {
                Set<String> setDataInfo = oneData.keySet();
                ArrayList<String> dataInfo = new ArrayList<>(setDataInfo);
                ArrayList<String> dataValue = new ArrayList<>();
                for (String info : dataInfo) {
                    dataValue.add(info + " = '" + oneData.get(info) + "'");
                }
                sql += " update " + tableName + " set " + String.join(",", dataValue) + " where guid = '" + oneData.get(key) + "'; ";

            }
        }

        return exeSql(sql);
    }

    @Override
    @Transactional//开启事务
    public ResultBody saveAllTableDataByParams(String saveType, String tableName, ArrayList<HashMap<String, Object>> data, String key) {
        try (Connection conn = dataSource.getConnection()) {
            for (HashMap<String, Object> oneData : data) {
                if ("add".equalsIgnoreCase(saveType)) {
                    // 构建 INSERT SQL
                    List<String> columns = new ArrayList<>(oneData.keySet());
                    String placeholders = String.join(", ", Collections.nCopies(columns.size(), "?"));
                    String sql = "INSERT INTO " + tableName + " (" + String.join(", ", columns) + ") VALUES (" + placeholders + ")";

                    try (PreparedStatement stmt = conn.prepareStatement(sql)) {
                        for (int i = 0; i < columns.size(); i++) {
                            //替换参数  把值放到对应得?上
                            stmt.setObject(i + 1, oneData.get(columns.get(i)));
                        }
                        stmt.executeUpdate();
                    }

                } else if ("edit".equalsIgnoreCase(saveType)) {
                    // 构建 UPDATE SQL
                    List<String> setClauses = new ArrayList<>();
                    List<Object> values = new ArrayList<>();
                    for (Map.Entry<String, Object> entry : oneData.entrySet()) {
                        if (!entry.getKey().equals(key)) {
                            setClauses.add(entry.getKey() + " = ?");
                            values.add(entry.getValue());
                        }
                    }

                    String sql = "UPDATE " + tableName + " SET " + String.join(", ", setClauses) + " WHERE " + key + " = ?";
                    try (PreparedStatement stmt = conn.prepareStatement(sql)) {
                        int i = 1;
                        for (Object val : values) {
                            stmt.setObject(i++, val);
                        }
                        stmt.setObject(i, oneData.get(key)); // 最后一个是 where key = ?
                        stmt.executeUpdate();
                    }
                } else {
                    return ResultBody.createErrorResult("不支持的操作类型: " + saveType);
                }
            }

            return ResultBody.createSuccessResult("操作成功");

        } catch (SQLException e) {
            return ResultBody.createErrorResult("执行失败: " + e.getMessage());
        }
    }

    @Override
    public ResultBody selectList(String sql) {
        if (!checkSql(sql, "select"))
            return ResultBody.createErrorResult("禁止执行此类危险SQL");

        try {
            List<Map<String, Object>> list = new ArrayList<>();
            list = sqlMapper.selectList(sql);
            return ResultBody.createSuccessResult(list);
            // 处理成功的情况
        } catch (PersistenceException e) {
            // 处理 SQL 执行失败的情况
            return ResultBody.createErrorResult("SQL执行失败：" + e.getMessage());
        }
    }

    @Override
    public ResultBody selectListByParams(String sql, List<Object> params) {
        List<Map<String, Object>> resultList = new ArrayList<>();
        // 安全校验：仅允许执行 SELECT 语句
        if (!checkSql(sql, "select")) {
            return ResultBody.createErrorResult("仅允许执行 SELECT 查询");
        }
        try (Connection conn = dataSource.getConnection();
             PreparedStatement stmt = conn.prepareStatement(sql)) {
            // 设置参数（防止 SQL 注入）
            if (params != null) {
                for (int i = 0; i < params.size(); i++) {
                    stmt.setObject(i + 1, params.get(i));
                }
            }
            try (ResultSet rs = stmt.executeQuery()) {
                ResultSetMetaData metaData = rs.getMetaData();
                int columnCount = metaData.getColumnCount();

                while (rs.next()) {
                    Map<String, Object> row = new HashMap<>();
                    for (int i = 1; i <= columnCount; i++) {
                        row.put(metaData.getColumnLabel(i), rs.getObject(i));
                    }
                    resultList.add(row);
                }
            }
            return ResultBody.createSuccessResult(resultList);
        } catch (SQLException e) {
            return ResultBody.createErrorResult("查询失败: " + e.getMessage());
        }
    }

    @Override
    public ResultBody deleteBySql(String sql) {
        if (!checkSql(sql, "delete")) {
            return ResultBody.createErrorResult("禁止执行此类危险SQL");
        }

        try {
            int i = sqlMapper.deleteBySql(sql);
            // 处理成功的情况，返回表示成功的结果
            return ResultBody.createSuccessResult(i);
        } catch (PersistenceException e) {

            return ResultBody.createErrorResult("SQL执行失败：" + e.getMessage());
        }
    }

    @Override
    public ResultBody updateBySql(String sql) {
        if (!checkSql(sql, "update")) {
            return ResultBody.createErrorResult("禁止执行此类危险SQL");
        }

        try {
            int i = sqlMapper.updateBySql(sql);
            // 处理成功的情况，返回表示成功的结果
            return ResultBody.createSuccessResult(i);
        } catch (PersistenceException e) {

            return ResultBody.createErrorResult("SQL执行失败：" + e.getMessage());
        }
    }

    @Override
    public ResultBody exeSql(String sql) {
        if (!checkSql(sql, "exe")) {
            return ResultBody.createErrorResult("禁止执行此类危险SQL");
        }

        try {
            int i = sqlMapper.exeSql(sql);
            // 处理成功的情况，返回表示成功的结果
            return ResultBody.createSuccessResult(i);
        } catch (PersistenceException e) {

            return ResultBody.createErrorResult("SQL执行失败：" + e.getMessage());
        }
    }

    //使用jdbc原生的方式实现内容
    @Override
    public ResultBody exeSqlByParams(String sql, List<Object> params) {
        List<Map<String, Object>> resultList = new ArrayList<>();

        //获取数据库链接
        try (Connection conn = dataSource.getConnection();
             PreparedStatement stmt = conn.prepareStatement(sql)) {

            // 设置参数（防止 SQL 注入）
            if (params != null) {
                for (int i = 0; i < params.size(); i++) {
                    stmt.setObject(i + 1, params.get(i));
                }
            }

            //执行sql
            boolean hasResultSet = stmt.execute();

            if (hasResultSet) {
                try (ResultSet rs = stmt.getResultSet()) {
                    ResultSetMetaData metaData = rs.getMetaData();
                    int columnCount = metaData.getColumnCount();

                    while (rs.next()) {
                        Map<String, Object> row = new HashMap<>();
                        for (int i = 1; i <= columnCount; i++) {
                            row.put(metaData.getColumnLabel(i), rs.getObject(i));
                        }
                        resultList.add(row);
                    }
                }
                return ResultBody.createSuccessResult(resultList);
            } else {
                int updateCount = stmt.getUpdateCount(); // 影响的行数
                return ResultBody.createSuccessResult("执行成功，影响行数：" + updateCount);
            }
        } catch (SQLException e) {
            return ResultBody.createErrorResult("执行失败: " + e.getMessage());
        }
    }

    private boolean checkSql(String sql, String type) {
        String checkSql = sql.toLowerCase();
        if (type != "exe") {
            if (checkSql.indexOf(type) == -1) {
                return false;
            }
        }
        if ((checkSql.indexOf("drop ") != -1) || (checkSql.indexOf("truncate ") != -1))
            return false;
        return true;
    }
}
