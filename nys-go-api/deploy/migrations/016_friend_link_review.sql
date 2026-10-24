-- 友链审核：普通用户提交的友链先是“待审核”，站长（超级管理员）或管理员通过后才公开显示。
-- 已有的友链全部视为已通过，执行后页面显示不变。
-- 可重复执行。执行后需要重启 Go 服务才会开始审核。

SET @link_schema = DATABASE();

SET @link_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @link_schema AND LOWER(table_name) = 'friendlinkinfo' AND UPPER(column_name) = 'STATUS') = 0,
    'ALTER TABLE friendLinkInfo ADD COLUMN STATUS VARCHAR(16) NOT NULL DEFAULT ''approved'' COMMENT ''审核状态：pending 待审核，approved 已通过''',
    'SELECT 1'
);
PREPARE link_stmt FROM @link_sql;
EXECUTE link_stmt;
DEALLOCATE PREPARE link_stmt;
