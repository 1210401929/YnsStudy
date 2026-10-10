# NYS Go API

这是 `NYS_TOP` Java 多服务后端的 Go 单体版本。原来的 Eureka、Gateway、`pub_api`、`blog_api` 和 `ai_api` 已合并为一个进程，同时保留前端正在使用的接口路径：

- `/pub-api/**`
- `/blog-api/**`
- `/ai-api/**`
- `/uploadFile/**`

## 配置

所有运行参数统一放在 `config/config.yaml`，包括服务端口、MySQL、Redis、JWT、AES、通用密码、QQ 登录、上传目录、短信、地图和 AI 配置。默认从该路径加载；如需使用另一个配置文件，只设置 `NYS_CONFIG` 为文件路径即可。

Docker 部署时 `database.location` 必须保持为 `Asia/Shanghai`。Go 会把数据库时间解析、接口 JSON、RSS、SEO 和进程本地时间统一为北京时间，不依赖容器是否安装系统时区数据。

登录有效期由 `security.jwt_expiration_seconds` 和 `security.session_expiration_seconds` 统一控制，默认都是 `3600` 秒（1 小时）。这是从登录时刻开始计算的绝对期限，检查登录、刷新页面和修改资料都不会续期；过期后必须重新登录。

### QQ 登录配置

在 QQ 互联创建并审核“网站应用”后，填写 `qq_oauth.app_id` 和 `qq_oauth.app_key`，确认 `redirect_uri` 与 QQ 互联后台登记的回调地址完全一致，再将 `qq_oauth.enabled` 改为 `true`。默认生产回调地址为：

```text
https://ynsstudy.cn/api/pub-api/login/qq/callback
```

`frontend_origin` 只填写来源（协议、域名和可选端口），不能包含页面路径。QQ 首次登录会自动创建一个本地用户，QQ OpenID 不会直接作为公开账号保存，而是转换成稳定的摘要账号。

### 邮件通知配置

评论、回复会发邮件给文章作者和被回复的人（匿名访客按评论时留的邮箱），站长还会收到全站的评论、新发表的公开文章、私密改公开的文章、社区帖子和社区评论，方便审核内容。使用 QQ 邮箱时，在 QQ 邮箱“设置 - 账号”里开启 SMTP 服务并生成授权码，然后在 `config/config.yaml` 增加：

```yaml
mail:
  enabled: true
  smtp_host: smtp.qq.com
  smtp_port: 465
  username: 你的QQ号@qq.com   # 发件邮箱
  password: 授权码            # 不是 QQ 密码
  from_name: YnsStudy
  admin_email: 你的QQ号@qq.com # 站长收通知的邮箱，可以和发件邮箱相同
```

执行 `deploy/migrations/005_add_mail_unsubscribe.sql` 创建退订名单表。邮件在后台排队发送，失败只写日志；非站长邮箱每天最多收 10 封（`daily_limit_per_address` 可调整），每封都带退订链接。退订链接默认指向 `external.domain_name` + `/api/pub-api/mail/getUnsubscribe`，接口地址不同时设置 `mail.api_base_url`。

## SEO 部署

文章会持续新增，单纯 Vue CSR 或构建时 Prerender 无法保证新文章的首次 HTML 包含正文。因此将首页 `/`、`/oneBlog/:id`、`/archive`、`/sitemap.xml` 和 `/rss.xml` 交给 Go 动态输出：首页包含站点介绍和最新文章链接，文章页包含真实正文、上一篇/下一篇和同作者文章链接，`/archive`（及 `/archive/page/N`）是按发布时间分页的全部文章归档，其他页面继续由原 Vue + Nginx 提供。

`security.whitelist` 需要包含 `/archive`、`/archive/**` 和 `/indexnow.txt`，否则这些公开页面会被鉴权拦截。

### IndexNow（Bing 等搜索引擎）

在 `seo.indexnow_key` 填写 8-128 位字母、数字或连字符组成的密钥后，Go 会在 `/indexnow.txt` 公开该密钥，并在公开文章发布、修改、改为私密或删除时，后台推送文章地址和 `/archive` 到 `https://api.indexnow.org/indexnow`。推送失败只写日志，不影响发文。本地开发请保持为空，避免把测试数据推送给搜索引擎。Google 不支持 IndexNow，依靠 sitemap 发现新文章。

部署时需要完成以下三项：

1. 将 `seo.frontend_index_file` 改成服务器上 Vue `dist/index.html` 的真实绝对路径。
2. 把 `deploy/nginx-seo.conf.example` 中的 location 合并到网站现有 Nginx `server`，然后执行 `nginx -t` 并重载。
3. 部署或升级时依次执行尚未执行过的 `deploy/migrations/001_add_blog_update_time.sql`、`deploy/migrations/002_add_lulu_paging_indexes.sql`、`deploy/migrations/003_add_lulu_npc_world.sql`、`deploy/migrations/004_add_search_text.sql` 和 `deploy/migrations/005_add_mail_unsubscribe.sql`。它们分别增加文章更新时间字段、噜噜分页统计索引、噜妹 NPC 事件表与按 IP 长期记忆索引、搜索用的纯文本字段，以及邮件退订名单表。

部署后使用 `curl https://ynsstudy.cn/oneBlog/真实文章ID` 检查源代码，应能直接找到文章标题、正文、canonical 和 `application/ld+json`，无需等待 JavaScript。

## 启动

```powershell
go run .
```

服务默认监听 `8889`，与现有 Vue 开发代理保持一致。健康检查地址为 `GET /health`。

## 目录
 
```text
config/                 集中配置
internal/cache/         Redis 与测试用内存缓存
internal/config/        配置加载和校验
internal/controller/    路由、参数绑定及特殊响应
internal/database/      MySQL 连接池
internal/middleware/    CORS、JWT 鉴权、AES 传输加密
internal/model/         API 返回值和公共模型
internal/repository/    通用数据库访问
internal/security/      AES、JWT、密码散列
internal/service/       各业务领域逻辑
internal/session/       Redis 会话
```

## 兼容说明

- 普通业务响应继续使用 `{isError, errMsg, result}`。
- 原网关的白名单、JWT、内部调用密钥和 AES 请求/响应加密已迁移为 Gin 中间件。
- Redis 继续负责会话、短信验证码和评论限流；将 `redis.enabled` 设为 `false` 时会使用进程内缓存，适合本地临时调试。
- 旧系统的 `/pub-api/sql/**`（直接执行 SQL）和 `/pub-api/upload/deleteFileByUrl(s)`（按地址删除文件）已移除：前端没有使用，而任何登录用户都能借此读写整个数据库或删除他人文件。新增业务在 service/repository 中使用参数化 SQL。
- 写操作的权限在后端校验（`internal/service/permission.go`），规则与前端一致：作者本人可以管理自己的内容；管理员（`userInfo.ROLE = 'admin'`）可以管理除超级管理员以外的内容；超级管理员（`security.super_admin_code`，默认 `yulei`，需与前端 `vue-config.js` 的 `adminUserCode` 一致）可以管理全部内容，并独占后台管理（用户管理、公告、文件一致性检查、社区置顶）。作者、发送者等身份字段一律取自当前登录用户，不信任请求体。
