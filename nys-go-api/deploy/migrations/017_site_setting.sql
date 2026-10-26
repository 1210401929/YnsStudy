-- 网站设置：站长在后台修改、不需要改代码的配置，按“键 → 值”保存，值一般是 JSON。
-- 目前用于关于页的“站长推荐”（about_featured：推荐的文章和系列），以后其他可配置项也放在这里。
-- 可重复执行。执行后需要重启 Go 服务才会开始使用。

CREATE TABLE IF NOT EXISTS siteSetting (
    SETTING_KEY   VARCHAR(64) NOT NULL COMMENT '设置项，例如 about_featured',
    SETTING_VALUE TEXT        NULL COMMENT '设置内容，一般是 JSON',
    UPDATE_USER   VARCHAR(100) NULL DEFAULT NULL COMMENT '最后修改人账号',
    UPDATE_TIME   DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (SETTING_KEY)
) DEFAULT CHARSET = utf8mb4;
