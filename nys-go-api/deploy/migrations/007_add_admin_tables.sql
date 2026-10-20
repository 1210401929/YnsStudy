-- 超级管理员后台第二批功能需要的两张表，可以重复执行。

-- 违禁词：在后台在线增删，立即生效，不再需要修改 config.yaml 后重启。
-- config.yaml 中原有的 content.banned_words 仍然生效（后台只读显示），两者合并使用。
CREATE TABLE IF NOT EXISTS bannedWord (
    WORD        VARCHAR(100) NOT NULL,
    CREATE_USER VARCHAR(100) NULL,
    CREATE_TIME DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (WORD)
) DEFAULT CHARSET = utf8mb4;

-- 管理员操作日志：封禁、设置管理员、重置密码、删除内容、修改公告等操作都会记录一条。
CREATE TABLE IF NOT EXISTS adminLog (
    ID          BIGINT       NOT NULL AUTO_INCREMENT,
    ACTOR_CODE  VARCHAR(100) NOT NULL,
    ACTOR_NAME  VARCHAR(100) NULL,
    ACTION      VARCHAR(50)  NOT NULL,
    TARGET      VARCHAR(255) NULL,
    DETAIL      VARCHAR(1000) NULL,
    IP          VARCHAR(64)  NULL,
    CREATE_TIME DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (ID),
    KEY idx_admin_log_time (CREATE_TIME),
    KEY idx_admin_log_action (ACTION)
) DEFAULT CHARSET = utf8mb4;
