<template>
  <div class="welcome-wrapper" :class="currentTheme">

    <header class="glass-nav animate-nav">
      <div class="nav-content">
        <div class="brand-logo">
          <span class="logo-text">YnsStudy</span>
          <span class="logo-dot">.</span>
        </div>
        <nav class="nav-links">
          <a v-for="item in menuItems" :key="item.router" @click="menuClick(item)" class="nav-item">
            {{ item.name }}
          </a>
        </nav>
        <div class="nav-action">
          <LoginDialog />
        </div>
      </div>
    </header>

    <main class="hero-section">
      <div class="content-box">
        <h1 class="main-title animate-item delay-1">
          探索知识的边界
        </h1>

        <div class="description-group animate-item delay-2">
          <h2 class="slogan">少一点迷茫，多一点引导</h2>
          <p class="sub-slogan">学习之路不再孤单，永远相信美好的事情即将发生</p>
        </div>

        <div class="action-buttons animate-item delay-3">
          <el-button type="success" @click="buttonClick('ai')" size="large" class="flat-btn accent-btn" round>
            ✨ 智能助手
          </el-button>
          <el-button type="primary" @click="buttonClick('article')" size="large" class="flat-btn primary-btn" round>
            📝 内容社区
          </el-button>
          <el-button @click="buttonClick('aboutWe')" size="large" class="flat-btn info-btn" round>
            关于我们
          </el-button>
        </div>
      </div>
    </main>

    <!-- 底部互动提示 -->
    <footer class="interaction-hint animate-item delay-4">
      <span>按 <kbd>B</kbd> 键切换主题强调色</span>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from "vue-router"
import { useHead } from '@vueuse/head'
import LoginDialog from "@/components/main/LoginDialog.vue"
import * as menuUtil from "@/utils/menu.js"

const router = useRouter()

// 主题切换改为切换“强调色”，保持整体白色扁平基调不变
const themes = ['theme-blue', 'theme-green', 'theme-purple']
const currentTheme = ref(themes[0])
const menuItems = ref(menuUtil.getMenuItems())

// SEO 配置
useHead({
  title: 'YnsStudy - 少一点迷茫，多一点引导',
  meta: [
    { name: 'description', content: 'YnsStudy：少一点迷茫，多一点引导，学习之路不再孤单。永远相信美好的事情即将发生。' },
    { name: 'keywords', content: 'YnsStudy, 学习引导, 内容社区, 技术博客, AI辅助学习, 开发者社区, 编程学习' }
  ]
})

const menuClick = (menu) => router.push({ name: menu.router })

const buttonClick = (type) => {
  const routes = {
    ai: () => window.open(router.resolve({ name: 'YnsStudyAi' }).href, "YnsStudyAi"),
    article: () => router.push({ name: "Home" }),
    aboutWe: () => router.push({ name: "About" })
  }
  routes[type]?.()
}

// 切换强调色
const switchTheme = () => {
  const index = themes.indexOf(currentTheme.value)
  currentTheme.value = themes[(index + 1) % themes.length]
}

const handleKey = (e) => {
  if (e && e.key && e.key.toLowerCase() === 'b') {
    switchTheme()
  }
}

onMounted(() => window.addEventListener('keydown', handleKey))
onBeforeUnmount(() => window.removeEventListener('keydown', handleKey))
</script>

<style scoped>
/* 核心容器：纯白/浅灰扁平基调 */
.welcome-wrapper {
  position: relative;
  min-height: 100vh;
  width: 100%;
  display: flex;
  flex-direction: column;
  background-color: #fafbfc;
  color: #333333;
  overflow-x: hidden;
  scrollbar-gutter: stable;
  transition: background-color 0.6s ease, color 0.6s ease;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* 顶部环境光晕 */
.welcome-wrapper::before {
  content: '';
  position: absolute;
  top: -20vh;
  left: 50%;
  transform: translateX(-50%);
  width: 70vw;
  height: 50vh;
  background: radial-gradient(circle, var(--accent-color) 0%, rgba(255,255,255,0) 70%);
  opacity: 0.06;
  z-index: 0;
  pointer-events: none;
  transition: background 0.8s ease;
}

/* 主题强调色变量 */
.theme-blue { --accent-color: #409eff; --accent-hover: #66b1ff; }
.theme-green { --accent-color: #67c23a; --accent-hover: #85ce61; }
.theme-purple { --accent-color: #8e2de2; --accent-hover: #a04ef6; }

/* 导航栏：修改为相对定位，跟随页面一起滚动 */
.glass-nav {
  position: relative; /* 从 fixed 改为 relative */
  width: 100%;
  z-index: 100;
  /* 依然保留一点半透明玻璃质感，能让背后的环境光晕透出来，显得高级 */
  background: rgba(255, 255, 255, 0.4);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-bottom: 1px solid rgba(0, 0, 0, 0.03);
}

.nav-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 2rem;
}

.brand-logo {
  font-size: 1.5rem;
  font-weight: 800;
  color: #111827;
  letter-spacing: -0.5px;
  cursor: pointer;
}

.logo-dot {
  color: var(--accent-color);
  transition: color 0.5s ease;
}

.nav-links {
  display: flex;
  gap: 2.5rem;
}

/* 导航链接动效优化 */
.nav-item {
  position: relative;
  font-size: 0.95rem;
  font-weight: 500;
  color: #4b5563;
  cursor: pointer;
  transition: color 0.3s ease;
  padding: 0.5rem 0;
}

.nav-item::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  width: 0;
  height: 2px;
  background-color: var(--accent-color);
  transition: width 0.3s ease;
  transform: translateX(-50%);
  border-radius: 2px;
}

.nav-item:hover {
  color: #111827;
}

.nav-item:hover::after {
  width: 100%;
}

/* 主视觉区 */
.hero-section {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  /* 调整：因为导航栏不再悬浮脱离文档流，所以顶部 padding 减小到正常值 */
  padding: 60px 20px 80px;
  z-index: 10;
}

.content-box {
  text-align: center;
  max-width: 800px;
  width: 100%;
}

/* 主标题：深色与主题色的高级渐变 */
.main-title {
  font-size: 4rem;
  font-weight: 800;
  margin-bottom: 1.5rem;
  letter-spacing: -1.5px;
  background: linear-gradient(135deg, #111827 30%, var(--accent-color) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.description-group {
  margin-bottom: 3.5rem;
}

.slogan {
  font-size: 1.5rem;
  color: #374151;
  font-weight: 600;
  margin-bottom: 1rem;
  letter-spacing: 0.5px;
}

.sub-slogan {
  font-size: 1.15rem;
  color: #6b7280;
  font-weight: 400;
}

/* 按钮操作区 */
.action-buttons {
  display: flex;
  justify-content: center;
  gap: 1.2rem;
  flex-wrap: wrap;
}

/* 统一扁平化按钮基础样式 */
.flat-btn {
  padding: 14px 36px !important;
  font-size: 1rem !important;
  font-weight: 600 !important;
  border: none !important;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.02) !important;
  transition: transform 0.3s ease, box-shadow 0.3s ease, background-color 0.3s ease, color 0.3s ease !important;
}

.flat-btn:hover {
  transform: translateY(-3px) !important;
  box-shadow: 0 8px 15px rgba(0, 0, 0, 0.05) !important;
}

.flat-btn:active {
  transform: translateY(-1px) !important;
}

/* 自定义辅助色按钮 */
.info-btn {
  background-color: #f3f4f6 !important;
  color: #374151 !important;
}

.info-btn:hover {
  background-color: #e5e7eb !important;
  color: #111827 !important;
}

/* 底部提示 */
.interaction-hint {
  padding: 2rem;
  text-align: center;
  font-size: 0.85rem;
  color: #9ca3af;
  z-index: 10;
}

kbd {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  box-shadow: 0 2px 0 #e5e7eb;
  padding: 2px 8px;
  border-radius: 6px;
  color: var(--accent-color);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-weight: 600;
  transition: color 0.3s ease;
}

/* --- 入场动画定义 --- */
.animate-item {
  opacity: 0;
  will-change: transform, opacity;
  -webkit-backface-visibility: hidden;
  backface-visibility: hidden;
  transform: translateZ(0);
  animation: fadeInUp 0.8s cubic-bezier(0.2, 0.8, 0.2, 1) forwards;
}

.animate-nav {
  animation: fadeInDown 0.8s ease-out forwards;
}

@keyframes fadeInUp {
  0% {
    opacity: 0;
    transform: translate3d(0, 25px, 0);
  }
  100% {
    opacity: 1;
    transform: translate3d(0, 0, 0);
  }
}

@keyframes fadeInDown {
  0% {
    opacity: 0;
    transform: translate3d(0, -20px, 0);
  }
  100% {
    opacity: 1;
    transform: translate3d(0, 0, 0);
  }
}

.delay-1 { animation-delay: 0.1s; }
.delay-2 { animation-delay: 0.2s; }
.delay-3 { animation-delay: 0.3s; }
.delay-4 { animation-delay: 0.5s; }

/* 响应式调整 */
@media (max-width: 768px) {
  .main-title {
    font-size: 2.5rem;
  }

  .slogan {
    font-size: 1.25rem;
  }

  .nav-links {
    display: none;
  }

  .action-buttons {
    flex-direction: column;
    padding: 0 20px;
    gap: 1rem;
  }

  .flat-btn {
    width: 100%;
    margin-left: 0 !important;
  }

  .welcome-wrapper::before {
    width: 100vw;
    height: 40vh;
  }
}
</style>