import './css/main-css/main.css'
import App from './App.vue'
import router from './router'
import 'element-plus/dist/index.css'
// 霞鹜文楷屏幕阅读版（简体字形），按字符分片按需加载，用作手账风格的手写字体
import 'lxgw-wenkai-screen-webfont/lxgwwenkaigbscreen.css'
// 手账风格公共样式，放在 Element Plus 之后以覆盖其主题变量
import './css/journal.css'
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createHead } from '@vueuse/head'
import { provideGlobalConfig } from 'element-plus'
import { removeStaticSeoTags } from './utils/seo.js'

const app = createApp(App);

app.use(router);
app.use(createPinia());
app.use(createHead());
// Element Plus 组件由 vite 按需引入，这里只保留原来的全局尺寸和层级配置
provideGlobalConfig({ size: 'small', zIndex: 3000 }, app, true);

// 等首个路由的懒加载组件就绪后再挂载，避免服务端输出的正文被清空后出现一段白屏
router.isReady().then(() => {
    removeStaticSeoTags();
    app.mount('#app');
});