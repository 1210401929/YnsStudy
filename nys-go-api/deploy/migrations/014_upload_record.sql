-- 上传文件的归属：记录每个上传文件是谁传的。删除头像、文章、资源时只删除作者本人上传的文件，
-- 防止有人把别人的文件地址填进自己的内容里，再借删除操作删掉别人的文件。
-- PATH 是统一后的访问地址，例如 /uploadFile/editorImage/0a1b2c.png（不含域名和参数）。
-- 可重复执行。执行后需要重启 Go 服务才会开始使用这张表。

CREATE TABLE IF NOT EXISTS uploadRecord (
    PATH        VARCHAR(255) NOT NULL COMMENT '文件访问地址，如 /uploadFile/editorImage/xxx.png',
    USERCODE    VARCHAR(100) NOT NULL COMMENT '上传者账号',
    CREATE_TIME DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (PATH),
    KEY idx_upload_record_user (USERCODE)
) DEFAULT CHARSET = utf8mb4;

-- 补齐旧数据的归属（已有记录的不会被覆盖）：
-- 1. 资源页登记过的文件，按登记时间先后，最早登记的人算作上传者
INSERT IGNORE INTO uploadRecord (PATH, USERCODE)
SELECT SUBSTRING_INDEX(SUBSTRING(FILEVIEWURL, LOCATE('/uploadFile/', FILEVIEWURL)), '?', 1), USERCODE
FROM fileInfo
WHERE LOCATE('/uploadFile/', FILEVIEWURL) > 0 AND USERCODE IS NOT NULL AND USERCODE <> ''
ORDER BY CREATE_TIME ASC;

-- 2. 用户当前使用的本站头像
INSERT IGNORE INTO uploadRecord (PATH, USERCODE)
SELECT SUBSTRING_INDEX(SUBSTRING(AVATAR, LOCATE('/uploadFile/', AVATAR)), '?', 1), CODE
FROM userInfo
WHERE LOCATE('/uploadFile/', AVATAR) > 0 AND CODE IS NOT NULL AND CODE <> '';

-- 旧文章里的图片无法用 SQL 可靠地提取，不补记录：这些图片在修改、删除文章时会被保留，不会被误删。
