-- 早期短信注册的账号直接把手机号当作昵称，这些昵称会出现在文章、评论、归档页和搜索结果中。
-- 本脚本把形如 11 位手机号的昵称统一改为“手机用户”+后四位（与现在新注册账号的默认昵称一致）。
-- 昵称在发布内容时会冗余保存到各表的 USERNAME 字段，因此这些表要一起更新。
-- 脚本可以重复执行：改过的昵称不再匹配手机号格式。

-- 执行前可以先看看会影响哪些账号：
SELECT GUID, CODE, NAME, CONCAT('手机用户', RIGHT(NAME, 4)) AS NEW_NAME
FROM userInfo
WHERE NAME REGEXP '^1[0-9]{10}$';

UPDATE userInfo         SET NAME           = CONCAT('手机用户', RIGHT(NAME, 4))           WHERE NAME           REGEXP '^1[0-9]{10}$';
UPDATE blogInfo         SET USERNAME       = CONCAT('手机用户', RIGHT(USERNAME, 4))       WHERE USERNAME       REGEXP '^1[0-9]{10}$';
UPDATE blogComment      SET USERNAME       = CONCAT('手机用户', RIGHT(USERNAME, 4))       WHERE USERNAME       REGEXP '^1[0-9]{10}$';
UPDATE blogGiveLike     SET USERNAME       = CONCAT('手机用户', RIGHT(USERNAME, 4))       WHERE USERNAME       REGEXP '^1[0-9]{10}$';
UPDATE communityInfo    SET USERNAME       = CONCAT('手机用户', RIGHT(USERNAME, 4))       WHERE USERNAME       REGEXP '^1[0-9]{10}$';
UPDATE communityComment SET USERNAME       = CONCAT('手机用户', RIGHT(USERNAME, 4))       WHERE USERNAME       REGEXP '^1[0-9]{10}$';
UPDATE fileInfo         SET USERNAME       = CONCAT('手机用户', RIGHT(USERNAME, 4))       WHERE USERNAME       REGEXP '^1[0-9]{10}$';
UPDATE friendLinkInfo   SET USERNAME       = CONCAT('手机用户', RIGHT(USERNAME, 4))       WHERE USERNAME       REGEXP '^1[0-9]{10}$';
UPDATE userFollow       SET USERNAME       = CONCAT('手机用户', RIGHT(USERNAME, 4))       WHERE USERNAME       REGEXP '^1[0-9]{10}$';
UPDATE userFollow       SET FOLLOWUSERNAME = CONCAT('手机用户', RIGHT(FOLLOWUSERNAME, 4)) WHERE FOLLOWUSERNAME REGEXP '^1[0-9]{10}$';
UPDATE announcementInfo SET userName       = CONCAT('手机用户', RIGHT(userName, 4))       WHERE userName       REGEXP '^1[0-9]{10}$';
