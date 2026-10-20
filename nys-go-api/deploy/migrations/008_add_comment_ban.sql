-- 评论限制：按用户、IP、网址、邮箱、昵称禁止某些人在本站评论。

-- 限制规则。TYPE 取值：
--   user      注册用户账号（精确匹配）
--   ip        IP 地址（精确匹配）或网段（如 1.2.3.0/24）
--   website   评论填写的网址包含该值即拦截（不区分大小写）
--   email     评论填写的邮箱包含该值即拦截（不区分大小写）
--   nickname  评论昵称包含该值即拦截（不区分大小写）
-- EXPIRE_TIME 为空表示永久，否则到期后自动失效。
CREATE TABLE IF NOT EXISTS commentBan (
    ID          BIGINT       NOT NULL AUTO_INCREMENT,
    TYPE        VARCHAR(20)  NOT NULL,
    VALUE       VARCHAR(255) NOT NULL,
    REMARK      VARCHAR(255) NULL,
    CREATE_USER VARCHAR(100) NULL,
    CREATE_TIME DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    EXPIRE_TIME DATETIME     NULL,
    PRIMARY KEY (ID),
    UNIQUE KEY uk_comment_ban (TYPE, VALUE)
) DEFAULT CHARSET = utf8mb4;

-- 评论记录发表者 IP，用于在后台按 IP 禁止评论。只有执行本迁移之后的新评论才有 IP。
-- 下面两条 ALTER 只需执行一次；重复执行会提示 Duplicate column name 'IP'，忽略即可。
ALTER TABLE blogComment ADD COLUMN IP VARCHAR(64) NULL;
ALTER TABLE communityComment ADD COLUMN IP VARCHAR(64) NULL;
