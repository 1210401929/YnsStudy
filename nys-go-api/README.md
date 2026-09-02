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

## SEO 部署

文章会持续新增，单纯 Vue CSR 或构建时 Prerender 无法保证新文章的首次 HTML 包含正文。因此将首页 `/`、`/oneBlog/:id`、`/sitemap.xml` 和 `/rss.xml` 交给 Go 动态输出：首页包含站点介绍和最新文章链接，文章页包含真实正文，其他页面继续由原 Vue + Nginx 提供。

部署时需要完成以下三项：

1. 将 `seo.frontend_index_file` 改成服务器上 Vue `dist/index.html` 的真实绝对路径。
2. 把 `deploy/nginx-seo.conf.example` 中的 location 合并到网站现有 Nginx `server`，然后执行 `nginx -t` 并重载。
3. 部署或升级时依次执行尚未执行过的 `deploy/migrations/001_add_blog_update_time.sql`、`deploy/migrations/002_add_lulu_paging_indexes.sql` 和 `deploy/migrations/003_add_lulu_npc_world.sql`。它们分别增加文章更新时间字段、噜噜分页统计索引，以及噜妹 NPC 事件表与按 IP 长期记忆索引。

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
- `/pub-api/sql/**` 为旧系统兼容接口，能直接执行 SQL，默认仍受鉴权保护。新增业务应优先在 service/repository 中使用参数化 SQL。
