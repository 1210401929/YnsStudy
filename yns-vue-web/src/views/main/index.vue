<template>
  <div class="app-container j-desk">
    <header class="topbar">
      <div class="topbar-inner">
        <a href="/" class="brand" @click.prevent="router.push('/')">
          <span class="brand-name">YnsStudy</span>
          <span class="brand-sub">学习手账</span>
        </a>

        <nav class="nav desktop-only" aria-label="主导航">
          <a
              v-for="item in menuItems"
              :key="item.router"
              :href="item.path"
              class="nav-link"
              :class="{ active: activeMenu === item.router }"
              @click.prevent="navigateTo(item.router)"
          >{{ item.name }}</a>
        </nav>

        <div class="login-wrapper desktop-only">
          <LoginDialog/>
        </div>

        <button type="button" class="menu-button mobile-only" aria-label="打开菜单" @click="drawerVisible = true">
          <el-icon><Menu/></el-icon>
        </button>
      </div>
    </header>

    <el-drawer v-model="drawerVisible" direction="ltr" size="72%" :with-header="false" class="nav-drawer">
      <div class="drawer-brand">YnsStudy <small>学习手账</small></div>
      <nav class="drawer-nav">
        <a
            v-for="item in menuItems"
            :key="item.router"
            :href="item.path"
            class="drawer-link"
            :class="{ active: activeMenu === item.router }"
            @click.prevent="navigateTo(item.router); drawerVisible = false"
        >{{ item.name }}</a>
      </nav>
      <div class="drawer-login">
        <LoginDialog/>
      </div>
    </el-drawer>

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
import {Menu} from '@element-plus/icons-vue';
import { useHead } from '@vueuse/head'; // 1. 引入 useHead

const router = useRouter();
const route = useRoute();
const activeMenu = ref(route.name);
const menuItems = ref(menuUtil.getMenuItems());
const drawerVisible = ref(false);

const navigateTo = (routerName) => {
  activeMenu.value = routerName;
  const target = menuItems.value.find(item => item.router === routerName);
  if (target?.path) {
    router.push({name: routerName});
  }
};
// 2. 定义响应式的 SEO 数据源，给定默认值
const seoTitle = ref('ynsStudy');
const seoDescription = ref('');

useHead({
  title: seoTitle,
  meta: [
    {
      name: 'description',
      content: seoDescription
    }
  ]
});

onMounted(() => {
  const path = window.location.pathname.split('/')[2];
  activeMenu.value = path;
  seoTitle.value = menuItems.value.filter(item => item.router === activeMenu.value)[0]["name"] + " - ynsStudy";
  seoDescription.value = seoTitle.value;
});

//监听路由  如果改变,则修改菜单栏选中内容
//因路由名称  和菜单栏选中名称对应并完全一致,所以可以直接使用
watch(() => route.name, (newValue) => {
  const routerNames = menuItems.value.map(item => item.router);
  if (routerNames.indexOf(newValue) !== -1) {
    activeMenu.value = newValue;
    seoTitle.value = menuItems.value.filter(item => item.router === activeMenu.value)[0]["name"] + " - ynsStudy";
    seoDescription.value = seoTitle.value;
  }
})

</script>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100%;
}

.topbar {
  position: relative;
  z-index: 10;
  flex-shrink: 0;
  background: var(--j-paper);
  border-bottom: 1px solid var(--j-rule);
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.6), 0 6px 14px -12px rgba(60, 50, 30, 0.35);
}

/* 顶栏下沿一条细细的虚线，像本子上撕下来的纸边 */
.topbar::after {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  bottom: 3px;
  border-bottom: 1px dashed var(--j-rule);
}

.topbar-inner {
  max-width: 1168px;
  height: 60px;
  margin: 0 auto;
  padding: 0 24px;
  display: flex;
  align-items: center;
  gap: 32px;
}

.brand {
  display: flex;
  align-items: baseline;
  gap: 8px;
  flex-shrink: 0;
  color: var(--j-ink);
  text-decoration: none;
}

.brand-name {
  font-family: var(--j-hand);
  font-size: 24px;
}

.brand-sub {
  font-family: var(--j-hand);
  font-size: 13px;
  color: var(--j-muted);
}

.nav {
  flex: 1;
  display: flex;
  justify-content: center;
  gap: 6px;
}

.nav-link {
  position: relative;
  padding: 6px 12px;
  font-size: 15px;
  color: var(--j-ink-soft);
  text-decoration: none;
  transition: color 0.15s;
}

/* 当前栏目：荧光笔划一道 */
.nav-link::before {
  content: "";
  position: absolute;
  left: 8px;
  right: 8px;
  bottom: 6px;
  height: 9px;
  background: var(--j-highlight);
  transform: scaleX(0);
  transform-origin: left;
  transition: transform 0.25s ease;
  z-index: -1;
}

.nav-link:hover {
  color: var(--j-ink);
}

.nav-link.active {
  color: var(--j-ink);
}

.nav-link.active::before {
  transform: scaleX(1);
}

.login-wrapper {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.menu-button {
  margin-left: auto;
  border: 1px solid var(--j-rule);
  background: var(--j-paper-warm);
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: var(--j-ink);
  cursor: pointer;
}

.content-container {
  flex-grow: 1;
}

.drawer-brand {
  padding: 4px 0 16px;
  border-bottom: 1px dashed var(--j-rule);
  font-family: var(--j-hand);
  font-size: 24px;
}

.drawer-brand small {
  font-size: 13px;
  color: var(--j-muted);
}

.drawer-nav {
  display: flex;
  flex-direction: column;
  padding: 12px 0;
}

.drawer-link {
  padding: 12px 4px;
  border-bottom: 1px solid var(--j-rule);
  font-size: 16px;
  color: var(--j-ink-soft);
  text-decoration: none;
}

.drawer-link.active {
  color: var(--j-ink);
}

.drawer-link.active::before {
  content: "✓ ";
  font-family: var(--j-hand);
  color: var(--j-stamp);
}

.drawer-login {
  padding-top: 16px;
}

.desktop-only {
  display: flex;
}

.mobile-only {
  display: none !important;
}

@media (max-width: 768px) {
  .topbar-inner {
    padding: 0 16px;
  }

  .desktop-only {
    display: none !important;
  }

  .mobile-only {
    display: flex !important;
  }
}
</style>
