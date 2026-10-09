import './css/main-css/main.css'
import App from './App.vue'
import router from './router'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
// 霞鹜文楷屏幕阅读版（简体字形），按字符分片按需加载，用作手账风格的手写字体
import 'lxgw-wenkai-screen-webfont/lxgwwenkaigbscreen.css'
// 手账风格公共样式，放在 Element Plus 之后以覆盖其主题变量
import './css/journal.css'
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createHead } from '@vueuse/head'

const app = createApp(App);

app.use(router);
app.use(createPinia());
app.use(createHead());
app.use(ElementPlus, { size: 'small', zIndex: 3000 });


app.mount('#app');