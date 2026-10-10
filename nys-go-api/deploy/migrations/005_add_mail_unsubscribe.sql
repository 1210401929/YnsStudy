-- 评论邮件通知的退订名单：邮件底部点“退订”后记录在这里，之后不再给该邮箱发通知。
-- 可重复执行。
CREATE TABLE IF NOT EXISTS mailUnsubscribe (
    EMAIL VARCHAR(191) NOT NULL COMMENT '退订的邮箱（小写）',
    CREATE_TIME DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '退订时间',
    PRIMARY KEY (EMAIL)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件通知退订名单';
