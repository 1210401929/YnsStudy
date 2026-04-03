import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
    const isDev = mode === 'development';

    // 公共配置（开发和生产通用）
    const baseConfig = {
        plugins: [vue()],
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
        }
    };

    // 如果是开发环境，添加 server 配置
    if (isDev) {
        baseConfig.server = {
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
