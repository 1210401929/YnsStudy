<template>
  <div class="app-container">
    <!-- 顶部区域 -->
    <header class="header-container">
      <!-- 左侧 Logo -->
      <div class="logo">YnsStudy</div>

      <!-- 中间菜单栏（桌面端） -->
      <nav class="menu-wrapper desktop-only">
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

      <!-- 移动端菜单按钮 -->
      <div class="mobile-menu-btn mobile-only" @click="drawerVisible = true">
        <el-icon><Menu /></el-icon>
      </div>

      <!-- 登录组件（桌面端显示） -->
      <div class="login-wrapper desktop-only">
        <LoginDialog />
      </div>
    </header>

    <!-- 抽屉菜单（移动端） -->
    <el-drawer
        v-model="drawerVisible"
        direction="ltr"
        size="70%"
        class="mobile-only"
    >
      <div class="drawer-header">
        <div class="drawer-logo">YnsStudy</div>
      </div>
      <el-menu
          :default-active="activeMenu"
          @select="(name) => { navigateTo(name); drawerVisible = false }"
          mode="vertical"
      >
        <el-menu-item
            v-for="item in menuItems"
            :key="item.router"
            :index="item.router"
        >
          {{ item.name }}
        </el-menu-item>
      </el-menu>

      <!-- 登录组件（移动端） -->
      <div class="login-mobile">
        <LoginDialog />
      </div>
    </el-drawer>

    <!-- 主内容区域 -->
    <main class="content-container">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import LoginDialog from '@/components/main/LoginDialog.vue';
import * as menuUtil from '@/utils/menu.js';
import { Menu } from '@element-plus/icons-vue';

const router = useRouter();
const route = useRoute();
const activeMenu = ref(route.name);
const menuItems = ref(menuUtil.getMenuItems());
const drawerVisible = ref(false);

const navigateTo = (routerName) => {
  activeMenu.value = routerName;
  const target = menuItems.value.find(item => item.router === routerName);
  if (target?.path) {
    router.push({ name: routerName });
  }
};

onMounted(() => {
  const path = window.location.pathname.split('/')[2];
  activeMenu.value = path;
});

//监听路由  如果改变,则修改菜单栏选中内容
//因路由名称  和菜单栏选中名称对应并完全一致,所以可以直接使用
watch(()=>route.name,(newValue)=>{
  const routerNames = menuItems.value.map(item=>item.router);
  if(routerNames.indexOf(newValue)!=-1){
    activeMenu.value = newValue;
  }
})

</script>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  background-color: #f4f6f8;
}

.header-container {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  background: #ffffff;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.06);
  flex-shrink: 0;
  position: relative;
  z-index: 10;
}

.logo {
  font-size: 22px;
  font-weight: 700;
  color: #0077b6;
  font-style: oblique;
}

/* 桌面菜单 */
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

.login-wrapper {
  display: flex;
  align-items: center;
}

/* 主内容 */
.content-container {
  flex-grow: 1;
  overflow-y: auto;
}

/* 手机端抽屉菜单样式 */
.drawer-header {
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.drawer-logo {
  font-size: 20px;
  font-weight: bold;
  color: #0077b6;
}

.login-mobile {
  padding: 16px;
}

/* 响应式样式 */
.desktop-only {
  display: flex;
}

.mobile-only {
  display: none !important;
}

.mobile-menu-btn {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 0 8px;
  font-size: 22px;
}

/* 屏幕宽度小于 768px 时的样式 */
@media (max-width: 768px) {
  .desktop-only {
    display: none !important;
  }

  .mobile-only {
    display: block !important;
  }

  .login-wrapper {
    display: none;
  }
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
