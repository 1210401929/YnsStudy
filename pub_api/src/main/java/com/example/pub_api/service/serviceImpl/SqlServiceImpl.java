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
        // 用于存放最终查询出来的完整数据库记录
        List<Map<String, Object>> resultDataList = new ArrayList<>();

        try (Connection conn = dataSource.getConnection()) {
            for (HashMap<String, Object> oneData : data) {
                Object primaryKeyValue = null; // 用于记录当前操作数据的主键值

                if ("add".equalsIgnoreCase(saveType)) {
                    // 构建 INSERT SQL
                    List<String> columns = new ArrayList<>(oneData.keySet());
                    String placeholders = String.join(", ", Collections.nCopies(columns.size(), "?"));
                    String sql = "INSERT INTO " + tableName + " (" + String.join(", ", columns) + ") VALUES (" + placeholders + ")";

                    // 判断传入的数据中是否已经包含了主键
                    boolean hasKey = oneData.containsKey(key) && oneData.get(key) != null;

                    // 重点：使用 Statement.RETURN_GENERATED_KEYS 以便获取数据库生成的自增主键
                    try (PreparedStatement stmt = conn.prepareStatement(sql, Statement.RETURN_GENERATED_KEYS)) {
                        for (int i = 0; i < columns.size(); i++) {
                            stmt.setObject(i + 1, oneData.get(columns.get(i)));
                        }
                        stmt.executeUpdate();

                        if (hasKey) {
                            primaryKeyValue = oneData.get(key);
                        } else {
                            // 如果没有传主键，说明可能是自增主键，从 ResultSet 获取
                            try (ResultSet generatedKeys = stmt.getGeneratedKeys()) {
                                if (generatedKeys.next()) {
                                    primaryKeyValue = generatedKeys.getObject(1);
                                }
                            }
                        }
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
                        primaryKeyValue = oneData.get(key); // 更新操作必定有主键
                        stmt.setObject(i, primaryKeyValue);
                        stmt.executeUpdate();
                    }
                } else {
                    return ResultBody.createErrorResult("不支持的操作类型: " + saveType);
                }

                // 核心新增逻辑：拿到主键后，去数据库反查这条完整的数据
                if (primaryKeyValue != null) {
                    String selectSql = "SELECT * FROM " + tableName + " WHERE " + key + " = ?";
                    try (PreparedStatement selectStmt = conn.prepareStatement(selectSql)) {
                        selectStmt.setObject(1, primaryKeyValue);
                        try (ResultSet rs = selectStmt.executeQuery()) {
                            if (rs.next()) {
                                ResultSetMetaData metaData = rs.getMetaData();
                                int columnCount = metaData.getColumnCount();
                                Map<String, Object> rowData = new HashMap<>();
                                // 遍历列，将一整条数据组装成 Map
                                for (int i = 1; i <= columnCount; i++) {
                                    rowData.put(metaData.getColumnLabel(i), rs.getObject(i));
                                }
                                resultDataList.add(rowData);
                            }
                        }
                    }
                } else {
                    // 兜底方案：如果因为某些原因没有获取到主键，则直接返回传入的参数
                    resultDataList.add(oneData);
                }
            }

            // 修改返回结果：将包含完整数据库记录的 List 返回（请确保您的 ResultBody 有接收数据对象的方法，例如传入 resultDataList）
            return ResultBody.createSuccessResult(resultDataList);

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
            System.out.println("即将执行的 SQL: " + stmt.toString());
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

    // 执行多条 SQL，支持事务
    @Override
    @Transactional(rollbackFor = Exception.class) // 显式指定异常回滚
    public ResultBody exeSqlListByParams(List<String> sqls, List<List<Object>> paramsList) {
        if (sqls == null || paramsList == null || sqls.size() != paramsList.size()) {
            return ResultBody.createErrorResult("SQL 语句与参数列表数量不匹配");
        }

        // 获取数据库链接（Spring 会管理这里的事务连接）
        try (Connection conn = dataSource.getConnection()) {
            int totalUpdateCount = 0;

            for (int i = 0; i < sqls.size(); i++) {
                String sql = sqls.get(i);
                List<Object> params = paramsList.get(i);

                // 简单的安全校验
                if (!checkSql(sql, "exe")) {
                    throw new SQLException("禁止执行危险 SQL: " + sql);
                }

                try (PreparedStatement stmt = conn.prepareStatement(sql)) {
                    // 设置参数
                    if (params != null) {
                        for (int j = 0; j < params.size(); j++) {
                            stmt.setObject(j + 1, params.get(j));
                        }
                    }

                    // 执行更新操作
                    totalUpdateCount += stmt.executeUpdate();
                }
            }
            return ResultBody.createSuccessResult("批量执行成功，总影响行数：" + totalUpdateCount);
        } catch (SQLException e) {
            // 在 @Transactional 下，抛出 RuntimeException 会触发回滚
            throw new RuntimeException("数据库执行失败，事务已回滚: " + e.getMessage());
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
