-- 话题标签：现在用于便签墙的随记，表结构是通用的，以后文章也可以直接使用（TARGET_TYPE = 'blog'）。
-- 可重复执行。

-- 话题本身。名字不区分大小写（utf8mb4 默认排序规则），例如 Go 和 go 是同一个话题
CREATE TABLE IF NOT EXISTS tag (
    ID          BIGINT      NOT NULL AUTO_INCREMENT,
    NAME        VARCHAR(40) NOT NULL,
    CREATE_TIME DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (ID),
    UNIQUE KEY uk_tag_name (NAME)
) DEFAULT CHARSET = utf8mb4;

-- 话题和内容的对应关系。TARGET_TYPE：note 随记（以后可加 blog 文章）
CREATE TABLE IF NOT EXISTS tagRelation (
    TAG_ID      BIGINT      NOT NULL,
    TARGET_TYPE VARCHAR(16) NOT NULL,
    TARGET_ID   VARCHAR(64) NOT NULL,
    CREATE_TIME DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (TAG_ID, TARGET_TYPE, TARGET_ID),
    KEY idx_tag_target (TARGET_TYPE, TARGET_ID)
) DEFAULT CHARSET = utf8mb4;
