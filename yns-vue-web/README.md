# yns-vue-web

This template should help get you started developing with Vue 3 in Vite.

## Recommended IDE Setup

[VSCode](https://code.visualstudio.com/) + [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur).

## Customize configuration

See [Vite Configuration Reference](https://vitejs.dev/config/).

## Project Setup

```sh
npm install
```

### Compile and Hot-Reload for Development

```sh
npm run dev
```

### Compile and Minify for Production

```sh
npm run build
```

## 噜噜素材独立部署

噜噜图片存放在与 Vue 项目同级的 `../lulu`，不再放在 `public` 中。
开发和预览时，Vite 会把该目录映射到原来的 `/picture/lulu/` 地址；执行
`npm run build` 时则不会把它复制进 `dist`。

生产环境目录：

```text
/var/www/
├── vue_dist/  # 上传 dist 中的前端文件
└── lulu/      # 单独上传 ../lulu 中的素材
    ├── benti/
    ├── npc/
    └── scenes/
```

把 `deploy/nginx-lulu-assets.conf.example` 中的 `location` 加入网站现有的
Nginx `server` 块，执行 `nginx -t` 并重载 Nginx。以后只替换噜噜图片时，
直接上传到 `/var/www/lulu` 对应位置即可，不需要重新构建或上传 Vue 包。

部署噜噜共享场景时，先执行 Go 项目的 `deploy/migrations/019_add_lulu_scene_state.sql`，
上传 `lulu/scenes/scene-winter-courtyard.webp` 和 `lulu/scenes/scene-winter-fireplace.webp` 到素材目录，
再发布 Vue 构建产物和 Go API。数据库已执行过该迁移时，新增场景无需再次迁移。
首次启用时全站使用原始背景，之后以数据库中保存的场景为准。

部署双人衣柜时，先执行 Go 项目的 `deploy/migrations/020_add_lulu_outfit_state.sql`，
再把 `lulu/benti/` 和 `lulu/npc/` 下本次新增的 8 张冬装 WebP 上传到对应素材目录，
并依次更新 Go API、Vue 构建产物。已保存的噜噜穿搭会按旧衣柜的顺序读取，直到首次使用新衣柜保存。

部署 Lv.70 婚礼时，先执行 `nys-go-api/deploy/migrations/021_add_lulu_wedding.sql`，
把 `lulu-wedding-photo-formal.webp`、`lulu-wedding-ceremony.webp` 和 `lulu-wedding-preparation.webp`
上传到 `/var/www/lulu/npc/`，并把 `scene-wedding-room.webp` 上传到 `/var/www/lulu/scenes/`，
再更新 Go API 与 Vue 构建产物。已经达到 Lv.70 的噜噜会在婚礼手账首次读取时自动补记一次婚礼。
