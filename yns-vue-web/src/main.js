import './css/main-css/main.css'

import { createApp } from 'vue'
import App from './App.vue'
// 导入路由
import router from './router'
//导入pinia  用于每个组件共享数据
import { createPinia } from 'pinia';
//导入elementPlus
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

const app = createApp(App);
app.use(router);
app.use(createPinia());
app.use(ElementPlus, { size: 'small', zIndex: 3000 })
app.mount('#app');