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
