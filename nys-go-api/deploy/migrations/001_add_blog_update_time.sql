-- 为 sitemap.xml 和 BlogPosting.dateModified 保存真实的文章修改时间。
-- 这个字段不会使用 ON UPDATE CURRENT_TIMESTAMP，避免文章阅读量变化时误改 lastmod。
ALTER TABLE blogInfo
    ADD COLUMN UPDATE_TIME DATETIME NULL DEFAULT NULL AFTER CREATE_TIME;

-- 已有文章没有历史修改时间，因此以首次发布时间作为当前 lastmod。
UPDATE blogInfo
SET UPDATE_TIME = CREATE_TIME
WHERE UPDATE_TIME IS NULL;

-- 新发布文章自动以发布时间作为初始更新时间；后续正文编辑由 Go 显式写入 NOW()。
ALTER TABLE blogInfo
    MODIFY COLUMN UPDATE_TIME DATETIME NULL DEFAULT CURRENT_TIMESTAMP;
