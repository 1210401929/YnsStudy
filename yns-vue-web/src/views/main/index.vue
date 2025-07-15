<template>
  <div class="app-container">
    <!-- 顶部区域 -->
    <header class="header-container">
      <!-- 左侧 Logo -->
      <div class="logo">YnsStudy</div>

      <!-- 中间菜单栏 -->
      <nav class="menu-wrapper">
        <el-menu
            class="menu"
            mode="horizontal"
            :default-active="activeMenu"
            @select="navigateTo"
            background-color="transparent"
            text-color="#333"
            active-text-color="#409EFF"
            :ellipsis="false"
        >
          <el-menu-item
              v-for="item in menuItems"
              :key="item.router"
              :index="item.router"
          >
            {{ item.name }}
          </el-menu-item>
        </el-menu>
      </nav>

      <!-- 右侧登录区域 -->
      <div class="login-wrapper">
        <LoginDialog/>
      </div>
    </header>

    <!-- 主内容区域 -->
    <main class="content-container">
      <router-view/>
    </main>
  </div>
</template>

<script setup>
import {ref, onMounted, watch} from 'vue';
import {useRouter, useRoute} from 'vue-router';
import LoginDialog from '@/components/main/LoginDialog.vue';
import * as menuUtil from '@/utils/menu.js';

const router = useRouter();
const route = useRoute();
const activeMenu = ref(route.name);
const menuItems = ref(menuUtil.getMenuItems());

const navigateTo = (routerName) => {

  activeMenu.value = routerName;
  const target = menuItems.value.find(item => item.router === routerName);
  if (target?.path) {
    router.push({name: routerName});
  }
};

onMounted(() => {
  const path = window.location.pathname.split('/')[2];
  activeMenu.value = path;
});

//监听路由  如果改变,则修改菜单栏选中内容
//因路由名称  和菜单栏选中名称对应并完全一致,所以可以直接使用
watch(()=>route.name,(newValue)=>{
  activeMenu.value = newValue;
})
</script>

<style scoped>

.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  background-color: #f4f6f8;
}

/* 顶部 Header */
.header-container {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  background: #ffffff;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.06);
  flex-shrink: 0;
  position: relative;
  z-index: 10;
  flex-shrink: 0;
}

/* 左 Logo 样式 */
.logo {
  font-size: 22px;
  font-weight: 700;
  color: #0077b6;
  font-style:oblique;
}

/* 菜单区域 */
.menu-wrapper {
  flex: 1;
  display: flex;
  justify-content: center;
  overflow: hidden;
}

.menu {
  border-bottom: none;
  background-color: transparent;
  display: inline-block;
}

/* 菜单项样式 */
.el-menu-item {
  padding: 0 18px;
  font-size: 16px;
  font-weight: 500;
  transition: background-color 0.3s ease;
  border-radius: 6px;
}

.el-menu-item:hover {
  background-color: #e6f7ff !important;
}

.el-menu-item.is-active {
  background-color: #d0ecff !important;
}

/* 登录组件容器 */
.login-wrapper {
  display: flex;
  align-items: center;
}

/* 内容区域 */
.content-container {
  flex-grow: 1;
  overflow-y: auto;
}

html, body {
  margin: 0;
  padding: 0;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background-color: #f4f6f8;
}

* {
  box-sizing: border-box;
}
</style>
