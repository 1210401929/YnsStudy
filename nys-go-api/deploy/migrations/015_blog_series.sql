-- 文章系列：作者把几篇文章编成一个有顺序的合集（例如《Go 入门》第 1～7 篇），读者可以按顺序阅读。
-- 一篇文章最多属于一个系列；系列只能收录作者自己的文章；删除系列时文章只是退出系列，不会被删除。
-- 可重复执行：表已存在、字段已存在时会跳过。执行后需要重启 Go 服务才会开始使用。

SET @series_schema = DATABASE();

-- 系列表的排序规则跟 blogInfo.USERCODE 一致：线上 blogInfo 是旧表（utf8mb4_unicode_ci），
-- 而新建表默认是 utf8mb4_0900_ai_ci，两者直接比较会报 “Illegal mix of collations”。
SET @series_collation = (
    SELECT COLLATION_NAME FROM information_schema.columns
    WHERE table_schema = @series_schema AND LOWER(table_name) = 'bloginfo' AND UPPER(column_name) = 'USERCODE'
    LIMIT 1
);
SET @series_sql = CONCAT(
    'CREATE TABLE IF NOT EXISTS blogSeries (',
    '  ID          BIGINT       NOT NULL AUTO_INCREMENT,',
    '  USERCODE    VARCHAR(100) NOT NULL COMMENT ''作者账号'',',
    '  TITLE       VARCHAR(60)  NOT NULL COMMENT ''系列名称'',',
    '  DESCRIPTION VARCHAR(300) NULL DEFAULT NULL COMMENT ''系列简介'',',
    '  CREATE_TIME DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,',
    '  UPDATE_TIME DATETIME     NULL DEFAULT NULL COMMENT ''最后修改时间（改名、调整顺序、加入文章）'',',
    '  PRIMARY KEY (ID),',
    '  KEY idx_series_user (USERCODE)',
    ') DEFAULT CHARSET = utf8mb4 COLLATE = ', COALESCE(@series_collation, 'utf8mb4_unicode_ci')
);
PREPARE series_stmt FROM @series_sql;
EXECUTE series_stmt;
DEALLOCATE PREPARE series_stmt;

-- 文章所属系列
SET @series_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @series_schema AND LOWER(table_name) = 'bloginfo' AND UPPER(column_name) = 'SERIES_ID') = 0,
    'ALTER TABLE blogInfo ADD COLUMN SERIES_ID BIGINT NULL DEFAULT NULL COMMENT ''所属系列 blogSeries.ID''',
    'SELECT 1'
);
PREPARE series_stmt FROM @series_sql;
EXECUTE series_stmt;
DEALLOCATE PREPARE series_stmt;

-- 文章在系列里的顺序，从 1 开始
SET @series_sql = IF(
    (SELECT COUNT(1) FROM information_schema.columns
     WHERE table_schema = @series_schema AND LOWER(table_name) = 'bloginfo' AND UPPER(column_name) = 'SERIES_ORDER') = 0,
    'ALTER TABLE blogInfo ADD COLUMN SERIES_ORDER INT NULL DEFAULT NULL COMMENT ''在系列中的顺序''',
    'SELECT 1'
);
PREPARE series_stmt FROM @series_sql;
EXECUTE series_stmt;
DEALLOCATE PREPARE series_stmt;

-- 按系列查文章（系列页、文章页的系列目录）
SET @series_sql = IF(
    (SELECT COUNT(1) FROM information_schema.statistics
     WHERE table_schema = @series_schema AND LOWER(table_name) = 'bloginfo' AND index_name = 'idx_blog_series') = 0,
    'ALTER TABLE blogInfo ADD INDEX idx_blog_series (SERIES_ID, SERIES_ORDER)',
    'SELECT 1'
);
PREPARE series_stmt FROM @series_sql;
EXECUTE series_stmt;
DEALLOCATE PREPARE series_stmt;
