import { fileURLToPath, URL } from 'node:url'
import { createReadStream, statSync } from 'node:fs'
import { extname, isAbsolute, relative, resolve } from 'node:path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

const luluAssetsDirectory = fileURLToPath(new URL('../lulu', import.meta.url))

const luluAssetMimeTypes = {
    '.png': 'image/png',
    '.webp': 'image/webp',
    '.jpg': 'image/jpeg',
    '.jpeg': 'image/jpeg',
    '.gif': 'image/gif',
    '.svg': 'image/svg+xml',
    '.avif': 'image/avif'
}

const serveLuluAsset = (request, response) => {
    let requestPath
    try {
        requestPath = decodeURIComponent((request.url || '/').split('?')[0])
    } catch {
        response.statusCode = 400
        response.end('Invalid asset path')
        return
    }

    const assetPath = resolve(luluAssetsDirectory, requestPath.replace(/^[/\\]+/, ''))
    const relativePath = relative(luluAssetsDirectory, assetPath)
    if (relativePath.startsWith('..') || isAbsolute(relativePath)) {
        response.statusCode = 403
        response.end('Forbidden')
        return
    }

    let assetStat
    try {
        assetStat = statSync(assetPath)
    } catch {
        response.statusCode = 404
        response.end('Lulu asset not found')
        return
    }

    if (!assetStat.isFile()) {
        response.statusCode = 404
        response.end('Lulu asset not found')
        return
    }

    response.setHeader('Content-Type', luluAssetMimeTypes[extname(assetPath).toLowerCase()] || 'application/octet-stream')
    response.setHeader('Content-Length', assetStat.size)
    response.setHeader('Last-Modified', assetStat.mtime.toUTCString())
    response.setHeader('Cache-Control', 'no-cache')

    if (request.method === 'HEAD') {
        response.statusCode = 200
        response.end()
        return
    }

    const stream = createReadStream(assetPath)
    stream.on('error', (error) => response.destroy(error))
    stream.pipe(response)
}

const luluExternalAssetsPlugin = () => ({
    name: 'lulu-external-assets',
    configureServer(server) {
        server.middlewares.use('/picture/lulu', serveLuluAsset)
    },
    configurePreviewServer(server) {
        server.middlewares.use('/picture/lulu', serveLuluAsset)
    }
})

export default defineConfig(({ mode }) => {
    const isDev = mode === 'development';

    // 公共配置（开发和生产通用）
    const baseConfig = {
        plugins: [
            vue(),
            // Element Plus 组件按需引入；样式仍由 main.js 全量引入，保证手账主题覆盖顺序不变
            Components({
                resolvers: [ElementPlusResolver({ importStyle: false })],
                dts: false
            }),
            luluExternalAssetsPlugin()
        ],
        resolve: {
            alias: {
                '@': fileURLToPath(new URL('./src', import.meta.url))
            }
        },
        //打包混淆
        esbuild: {
            pure: isDev ? [] : ['console.log', 'console.info'],
            drop: isDev ? [] : ['debugger'],
            legalComments: 'none', // 移除所有的备注/版权注释
        },
        build: {
            sourcemap: false, // 生产环境务必关闭，防止源码泄露
            minify: 'esbuild', // 确保开启压缩
            chunkSizeWarningLimit: 1500, // 优化打包体验
            rollupOptions: {
                output: {
                    // 第三方库单独成包，业务代码更新时浏览器可继续使用缓存。
                    // Element Plus 不再合成一个大包：组件已按需引入，交给打包工具按页面拆分，
                    // 每个页面只下载自己用到的组件（首屏 JS 约少 25%）。
                    manualChunks(id) {
                        if (!id.includes('node_modules')) return
                        if (id.includes('@wangeditor')) return 'vendor-editor'
                        if (/node_modules\/(vue|@vue|vue-router|pinia|@vueuse)\//.test(id)) return 'vendor-vue'
                    }
                }
            }
        }
    };

    // 如果是开发环境，添加 server 配置
    if (isDev) {
        baseConfig.server = {
            host: '127.0.0.1',
            port: 8080,
            open: false,
            proxy: {
                '/pub-api': {
                    target: 'http://localhost:8889',
                    changeOrigin: true,
                    rewrite: path => path
                },
                '/blog-api': {
                    target: 'http://localhost:8889',
                    changeOrigin: true,
                    rewrite: path => path
                },
                '/ai-api': {
                    target: 'http://localhost:8889',
                    changeOrigin: true,
                    rewrite: path => path
                }
            }
        };
    }

    // 返回配置
    return baseConfig;
});
