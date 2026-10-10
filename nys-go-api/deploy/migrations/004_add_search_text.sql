-- 为文章和社区帖子保存一份纯文本，用于标题/正文搜索。
-- 正文是 HTML，直接 LIKE 会匹配到 span、style 等标签和属性；这一列只存去掉标签后的文字。
-- 已有数据不需要手动处理：Go 服务启动时会自动补齐为空的 SEARCH_TEXT。
-- 可重复执行：通过 information_schema 判断字段是否已存在。

SET @search_schema = DATABASE();

SET @search_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @search_schema AND LOWER(table_name) = 'bloginfo' AND UPPER(column_name) = 'SEARCH_TEXT') = 0,
    'ALTER TABLE blogInfo ADD COLUMN SEARCH_TEXT MEDIUMTEXT NULL DEFAULT NULL COMMENT ''正文纯文本，用于搜索''',
    'SELECT 1'
);
PREPARE search_stmt FROM @search_sql;
EXECUTE search_stmt;
DEALLOCATE PREPARE search_stmt;

SET @search_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @search_schema AND LOWER(table_name) = 'communityinfo' AND UPPER(column_name) = 'SEARCH_TEXT') = 0,
    'ALTER TABLE communityInfo ADD COLUMN SEARCH_TEXT MEDIUMTEXT NULL DEFAULT NULL COMMENT ''帖子纯文本，用于搜索''',
    'SELECT 1'
);
PREPARE search_stmt FROM @search_sql;
EXECUTE search_stmt;
DEALLOCATE PREPARE search_stmt;
