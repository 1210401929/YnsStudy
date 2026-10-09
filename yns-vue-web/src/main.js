import './css/main-css/main.css'
import App from './App.vue'
import router from './router'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
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