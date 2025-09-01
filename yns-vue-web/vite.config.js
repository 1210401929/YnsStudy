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
