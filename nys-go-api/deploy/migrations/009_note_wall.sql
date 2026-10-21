-- 社区改造为“便签墙”：随记增加类型和图片，随记评论补齐与文章评论相同的字段。
-- 可重复执行：通过 information_schema 判断字段是否已存在，已有的字段会跳过。
-- 旧数据不需要处理：TYPE 为空的随记按“随笔”显示。

SET @note_schema = DATABASE();

-- ---------- communityInfo：随记 ----------

-- 类型：note 随笔、tip 技巧/踩坑、share 分享、question 提问
SET @note_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @note_schema AND LOWER(table_name) = 'communityinfo' AND UPPER(column_name) = 'TYPE') = 0,
    'ALTER TABLE communityInfo ADD COLUMN TYPE VARCHAR(16) NULL DEFAULT NULL COMMENT ''随记类型：note/tip/share/question''',
    'SELECT 1'
);
PREPARE note_stmt FROM @note_sql;
EXECUTE note_stmt;
DEALLOCATE PREPARE note_stmt;

-- 图片：JSON 数组，保存本站上传的图片地址，最多 9 张
SET @note_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @note_schema AND LOWER(table_name) = 'communityinfo' AND UPPER(column_name) = 'IMAGES') = 0,
    'ALTER TABLE communityInfo ADD COLUMN IMAGES TEXT NULL DEFAULT NULL COMMENT ''随记图片地址，JSON 数组''',
    'SELECT 1'
);
PREPARE note_stmt FROM @note_sql;
EXECUTE note_stmt;
DEALLOCATE PREPARE note_stmt;

-- ---------- communityComment：随记评论（与 blogComment 一致，支持匿名评论和回复通知） ----------

SET @note_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @note_schema AND LOWER(table_name) = 'communitycomment' AND UPPER(column_name) = 'USEREMAIL') = 0,
    'ALTER TABLE communityComment ADD COLUMN USEREMAIL VARCHAR(255) NULL DEFAULT NULL COMMENT ''匿名评论者邮箱，用于回复通知''',
    'SELECT 1'
);
PREPARE note_stmt FROM @note_sql;
EXECUTE note_stmt;
DEALLOCATE PREPARE note_stmt;

SET @note_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @note_schema AND LOWER(table_name) = 'communitycomment' AND UPPER(column_name) = 'USERWEBSITE') = 0,
    'ALTER TABLE communityComment ADD COLUMN USERWEBSITE VARCHAR(255) NULL DEFAULT NULL COMMENT ''匿名评论者网址''',
    'SELECT 1'
);
PREPARE note_stmt FROM @note_sql;
EXECUTE note_stmt;
DEALLOCATE PREPARE note_stmt;

SET @note_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @note_schema AND LOWER(table_name) = 'communitycomment' AND UPPER(column_name) = 'RECEIVE_USERCODE') = 0,
    'ALTER TABLE communityComment ADD COLUMN RECEIVE_USERCODE VARCHAR(100) NULL DEFAULT NULL COMMENT ''被回复人账号''',
    'SELECT 1'
);
PREPARE note_stmt FROM @note_sql;
EXECUTE note_stmt;
DEALLOCATE PREPARE note_stmt;

SET @note_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @note_schema AND LOWER(table_name) = 'communitycomment' AND UPPER(column_name) = 'RECEIVE_USERNAME') = 0,
    'ALTER TABLE communityComment ADD COLUMN RECEIVE_USERNAME VARCHAR(100) NULL DEFAULT NULL COMMENT ''被回复人昵称''',
    'SELECT 1'
);
PREPARE note_stmt FROM @note_sql;
EXECUTE note_stmt;
DEALLOCATE PREPARE note_stmt;

SET @note_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @note_schema AND LOWER(table_name) = 'communitycomment' AND UPPER(column_name) = 'IP') = 0,
    'ALTER TABLE communityComment ADD COLUMN IP VARCHAR(64) NULL',
    'SELECT 1'
);
PREPARE note_stmt FROM @note_sql;
EXECUTE note_stmt;
DEALLOCATE PREPARE note_stmt;

-- ---------- 新随记使用数字编号（地址形如 /note/12），和文章共用编号表，各自计数 ----------
-- 旧随记的编号是一长串字符，保持不变，旧地址照常可用。
INSERT INTO blogInfo_guid_sequence (table_name, current_value)
SELECT 'communityinfo', 0 FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM blogInfo_guid_sequence WHERE table_name = 'communityinfo');
