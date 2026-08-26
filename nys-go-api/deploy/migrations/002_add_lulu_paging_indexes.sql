-- 为噜噜留言分页、日志分页和月度陪伴统计补充联合索引。
-- 使用 information_schema 判断，重复执行不会因为索引已存在而失败。

SET @lulu_schema = DATABASE();

SET @lulu_sql = IF(
    (SELECT COUNT(1) FROM information_schema.statistics
     WHERE table_schema = @lulu_schema AND table_name = 'z_lulu_message' AND index_name = 'idx_lulu_message_user_time') = 0,
    'CREATE INDEX idx_lulu_message_user_time ON z_lulu_message (USER_NUM, CREATE_TIME, ID)',
    'SELECT 1'
);
PREPARE lulu_stmt FROM @lulu_sql;
EXECUTE lulu_stmt;
DEALLOCATE PREPARE lulu_stmt;

SET @lulu_sql = IF(
    (SELECT COUNT(1) FROM information_schema.statistics
     WHERE table_schema = @lulu_schema AND table_name = 'z_lulu_log' AND index_name = 'idx_lulu_log_user_time') = 0,
    'CREATE INDEX idx_lulu_log_user_time ON z_lulu_log (USER_NUM, CREATE_TIME, ID)',
    'SELECT 1'
);
PREPARE lulu_stmt FROM @lulu_sql;
EXECUTE lulu_stmt;
DEALLOCATE PREPARE lulu_stmt;

SET @lulu_sql = IF(
    (SELECT COUNT(1) FROM information_schema.statistics
     WHERE table_schema = @lulu_schema AND table_name = 'z_lulu_message' AND index_name = 'idx_lulu_message_ip_time') = 0,
    'CREATE INDEX idx_lulu_message_ip_time ON z_lulu_message (IP_ADDRESS, CREATE_TIME)',
    'SELECT 1'
);
PREPARE lulu_stmt FROM @lulu_sql;
EXECUTE lulu_stmt;
DEALLOCATE PREPARE lulu_stmt;
