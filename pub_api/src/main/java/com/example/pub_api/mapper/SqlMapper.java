package com.example.pub_api.mapper;

import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

import java.util.List;
import java.util.Map;

@Mapper
public interface SqlMapper {
    @Select("${sql}")
    List<Map<String, Object>> selectList(String sql);

    @Delete("${sql}")
    int deleteBySql(String sql);

    @Update("${sql}")
    int updateBySql(String sql);

    @Update("${sql}")
    int exeSql(String sql);
}