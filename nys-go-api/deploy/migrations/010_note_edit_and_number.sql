-- 便签墙：支持编辑随记，旧随记改用数字编号。
-- 可重复执行：通过 information_schema 判断字段是否已存在，已有的字段会跳过。
--
-- 执行后不需要手动处理数据：Go 服务启动时会把编号是一长串字符的旧随记依次改成数字编号
-- （评论跟着一起改），旧编号记在 LEGACY_GUID，访问旧地址 /note/旧编号 会 301 跳到新地址。

SET @note_schema = DATABASE();

-- Markdown 原文：编辑时还原成写的时候的样子。旧随记没有原文，编辑时由页面把 HTML 转回 Markdown
SET @note_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @note_schema AND LOWER(table_name) = 'communityinfo' AND UPPER(column_name) = 'SOURCE') = 0,
    'ALTER TABLE communityInfo ADD COLUMN SOURCE MEDIUMTEXT NULL DEFAULT NULL COMMENT ''随记的 Markdown 原文，用于编辑''',
    'SELECT 1'
);
PREPARE note_stmt FROM @note_sql;
EXECUTE note_stmt;
DEALLOCATE PREPARE note_stmt;

-- 最后修改时间：页面显示“已编辑”，站点地图使用
SET @note_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @note_schema AND LOWER(table_name) = 'communityinfo' AND UPPER(column_name) = 'UPDATE_TIME') = 0,
    'ALTER TABLE communityInfo ADD COLUMN UPDATE_TIME DATETIME NULL DEFAULT NULL COMMENT ''最后修改时间''',
    'SELECT 1'
);
PREPARE note_stmt FROM @note_sql;
EXECUTE note_stmt;
DEALLOCATE PREPARE note_stmt;

-- 旧编号：改成数字编号之前的编号，用于旧地址跳转
SET @note_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @note_schema AND LOWER(table_name) = 'communityinfo' AND UPPER(column_name) = 'LEGACY_GUID') = 0,
    'ALTER TABLE communityInfo ADD COLUMN LEGACY_GUID VARCHAR(64) NULL DEFAULT NULL COMMENT ''改为数字编号之前的旧编号'', ADD INDEX idx_community_legacy_guid (LEGACY_GUID)',
    'SELECT 1'
);
PREPARE note_stmt FROM @note_sql;
EXECUTE note_stmt;
DEALLOCATE PREPARE note_stmt;

-- 随记编号计数（迁移 009 已经加过的话这里会跳过）
INSERT INTO blogInfo_guid_sequence (table_name, current_value)
SELECT 'communityinfo', 0 FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM blogInfo_guid_sequence WHERE table_name = 'communityinfo');
