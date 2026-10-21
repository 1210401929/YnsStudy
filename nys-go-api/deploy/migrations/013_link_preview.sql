-- 便签墙的链接卡片：随记正文里第一个网址的标题、简介、配图、网站名和图标，发布或编辑时由服务端抓取后保存，
-- 之后打开便签墙不再重复抓取。JSON 格式，例如：
-- {"url":"https://go.dev/","title":"The Go Programming Language","description":"...","image":"https://...","siteName":"go.dev","icon":"https://go.dev/favicon.ico"}
-- 可重复执行。
SET @note_schema = DATABASE();

SET @note_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @note_schema AND LOWER(table_name) = 'communityinfo' AND UPPER(column_name) = 'LINK_PREVIEW') = 0,
    'ALTER TABLE communityInfo ADD COLUMN LINK_PREVIEW TEXT NULL DEFAULT NULL COMMENT ''正文第一个网址的链接卡片信息，JSON''',
    'SELECT 1'
);
PREPARE note_stmt FROM @note_sql;
EXECUTE note_stmt;
DEALLOCATE PREPARE note_stmt;
