-- 噜妹 NPC、稀有外出彩蛋、往来记录与按 IP 长期记忆查询支持。
-- 可重复执行：表和索引均通过 IF NOT EXISTS / information_schema 防重。

CREATE TABLE IF NOT EXISTS z_lulu_npc_event (
    ID BIGINT NOT NULL AUTO_INCREMENT COMMENT '事件主键',
    USER_NUM BIGINT NOT NULL DEFAULT 1 COMMENT '噜噜所属用户编号',
    IP_ADDRESS VARCHAR(64) NOT NULL COMMENT '访客 IP，用于隔离各自的 NPC 事件与记忆',
    EVENT_DATE DATE NOT NULL COMMENT '事件日期',
    EVENT_TYPE VARCHAR(32) NOT NULL COMMENT 'LETTER 来信 / OUTING 外出彩蛋',
    TITLE VARCHAR(120) NOT NULL COMMENT '事件标题',
    CONTENT VARCHAR(1000) NOT NULL COMMENT '事件正文',
    EXPIRES_AT DATETIME NULL COMMENT '前台特殊展示截止时间',
    CREATE_TIME DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (ID),
    UNIQUE KEY uk_lulu_npc_ip_day (USER_NUM, IP_ADDRESS, EVENT_DATE),
    KEY idx_lulu_npc_ip_time (USER_NUM, IP_ADDRESS, CREATE_TIME)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='噜噜与噜妹 NPC 往来事件';

SET @lulu_schema = DATABASE();

SET @lulu_sql = IF(
    (SELECT COUNT(1) FROM information_schema.statistics
     WHERE table_schema = @lulu_schema AND table_name = 'z_lulu_log' AND index_name = 'idx_lulu_log_user_ip_time') = 0,
    'CREATE INDEX idx_lulu_log_user_ip_time ON z_lulu_log (USER_NUM, IP_ADDRESS, CREATE_TIME, ACTION_TYPE)',
    'SELECT 1'
);
PREPARE lulu_stmt FROM @lulu_sql;
EXECUTE lulu_stmt;
DEALLOCATE PREPARE lulu_stmt;

SET @lulu_sql = IF(
    (SELECT COUNT(1) FROM information_schema.statistics
     WHERE table_schema = @lulu_schema AND table_name = 'z_lulu_message' AND index_name = 'idx_lulu_message_user_ip_time') = 0,
    'CREATE INDEX idx_lulu_message_user_ip_time ON z_lulu_message (USER_NUM, IP_ADDRESS, CREATE_TIME)',
    'SELECT 1'
);
PREPARE lulu_stmt FROM @lulu_sql;
EXECUTE lulu_stmt;
DEALLOCATE PREPARE lulu_stmt;
