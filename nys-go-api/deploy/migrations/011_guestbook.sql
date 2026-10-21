-- 便签墙的留言板：访客不登录也能留言，所有人都可以回复（一层回复，和文章评论一样）。
-- 可重复执行。
CREATE TABLE IF NOT EXISTS guestbook (
    GUID             VARCHAR(64)  NOT NULL,
    SUPERGUID        VARCHAR(64)  NULL COMMENT '回复所属的那条留言；为空表示这是一条留言',
    TEXT             TEXT         NOT NULL,
    USERCODE         VARCHAR(100) NULL COMMENT '登录用户的账号；匿名访客为空',
    USERNAME         VARCHAR(100) NULL,
    USEREMAIL        VARCHAR(255) NULL COMMENT '匿名访客的邮箱，用于回复通知，不公开',
    USERWEBSITE      VARCHAR(255) NULL,
    RECEIVE_USERCODE VARCHAR(100) NULL COMMENT '被回复人账号',
    RECEIVE_USERNAME VARCHAR(100) NULL COMMENT '被回复人昵称',
    IP               VARCHAR(64)  NULL,
    ISTOP            CHAR(1)      NULL COMMENT '1 表示置顶',
    CREATE_TIME      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (GUID),
    KEY idx_guestbook_parent (SUPERGUID),
    KEY idx_guestbook_time (CREATE_TIME)
) DEFAULT CHARSET = utf8mb4;
