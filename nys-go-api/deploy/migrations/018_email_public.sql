-- 邮箱是否在个人主页公开：默认不公开（0），用户可以在个人中心打开（1）。
-- 作者主页任何人都能看、也会被搜索引擎收录，邮箱默认不应该公开。
-- 执行前后都一样不公开（代码按“没有这一列 = 不公开”处理），执行后用户才能自己选择公开。
-- 可重复执行。执行后需要重启 Go 服务。

SET @email_schema = DATABASE();

SET @email_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @email_schema AND LOWER(table_name) = 'userinfo' AND UPPER(column_name) = 'EMAILPUBLIC') = 0,
    'ALTER TABLE userInfo ADD COLUMN EMAILPUBLIC CHAR(1) NOT NULL DEFAULT ''0'' COMMENT ''是否在个人主页公开邮箱：1 公开，0 不公开''',
    'SELECT 1'
);
PREPARE email_stmt FROM @email_sql;
EXECUTE email_stmt;
DEALLOCATE PREPARE email_stmt;
