# 项目说明文档

## 🚀 关于启动项目

### 后端启动顺序
依次执行以下 4 个 Spring Boot 项目：
1. `EurekaServerApplication`
2. `GatewayServiceApplication`
3. `PubApiApplication`
4. `BlogApiApplication`

### 前端启动
1. 进入 `yns-vue-web` 目录
2. 执行启动命令：
   `npm run dev`

---

## 🏗️ 关于项目结构

- **eureka-server**: 注册中心
- **gateway-service**: 网关
- **common_api**: 公共模块。包含 `UserBean`、`ResultBody` 类的定义以及通用实现方法，其他微服务通过 `pom` 文件调用该项目。
- **pub_api**: 接口集成服务，用于大多数通用 API。
- **blog_api**: 博客相关服务，包含新增文章、评论等功能。
- **yns-vue-web**: Vue 前端项目。是 `yns_portal` 的迭代版本，实现真正的前后端分离。
- **yns_portal**: 前端项目（**已废弃**）。

---

## ⚙️ 关于配置说明

- **跨域请求**: 由统一网关 (`gateway`) 处理。
- **新增子项目**: 如需增加子项目，需要配置网关配置文件和 Vue 项目的 `vite.config.js` 文件。
- **资源文件映射**: 配置路径为 `pub_api/src/main/java/com/example/pub_api/config/StaticResourceConfig.java`。
- **Nginx 依赖**: 项目**必须**依赖 Nginx 进行静态文件跳转以及端口跳转。

---

## 🗄️ 关于数据库表

- `userInfo` => 用户表
- `userFollow` => 用户关注列表
- `loginHistory` => 用户登录历史表
- `personInfo` => 用户主页表
- `noticeInfo` => 消息通知表
- `fileInfo` => 附件表
- `blogInfo` => 博客文章表
- `blogComment` => 博客评论表
- `blogGiveLike` => 博客点赞收藏表
- `communityInfo` => 社区内容表
- `communityComment` => 社区内容评论表

> **⚠️ 建表规范**: `GUID` 是必须存在的字段，需设置为随机 32 位码。

---

## 📌 其他注意事项

- **Redis 安全**: 因云服务器提示 6379 端口存在对外攻击风险，`pub_api` 和 `blog_api` 的 `yml` 配置文件中的 Redis 密码不再公开。
- **静态文件访问**: 前台静态文件（包括所有图片和上传的资源文件）访问，需要通过 Nginx 转发到网关。
- **文件上传路径**:
  - 富文本编辑器的图片路径: `项目根目录/upload/editorImage`
  - 资源文件路径: `项目根目录/upload/resourcesFile`