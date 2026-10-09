<template>
  <div class="welcome j-desk" :class="currentTheme">
    <header class="welcome-nav">
      <span class="welcome-brand">YnsStudy</span>
      <nav class="welcome-links">
        <a v-for="item in menuItems" :key="item.router" :href="item.path" @click.prevent="menuClick(item)">
          {{ item.name }}
        </a>
      </nav>
      <div class="welcome-login">
        <LoginDialog/>
      </div>
    </header>

    <main class="welcome-main">
      <!-- 合着的手账本 -->
      <div class="cover">
        <span class="cover-spine"></span>
        <span class="cover-band"></span>
        <div class="cover-label">
          <span class="j-tape j-tape--top"></span>
          <p class="cover-kicker">YnsStudy</p>
          <h1 class="cover-title">探索知识的边界</h1>
          <h2 class="cover-slogan">少一点迷茫，多一点引导</h2>
        </div>
      </div>

      <!-- 贴在旁边的便签，就是入口 -->
      <div class="notes">
        <p class="notes-intro"><span>学习之路不再孤单，</span><span>永远相信美好的事情即将发生</span></p>
        <a href="/YnsStudyAi" class="note note-green" @click.prevent="buttonClick('ai')">
          <span class="note-title">智能助手</span>
        </a>
        <a href="/ynsStudy/Home" class="note note-yellow" @click.prevent="buttonClick('article')">
          <span class="note-title">内容社区</span>
        </a>
        <a href="/ynsStudy/About" class="note note-pink" @click.prevent="buttonClick('aboutWe')">
          <span class="note-title">关于我们</span>
        </a>
      </div>
    </main>

    <!-- 底部互动提示 -->
    <footer class="interaction-hint">
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
const menuItems = menuUtil.getMenuItems()

// SEO 配置
useHead({
  title: 'YnsStudy - 少一点迷茫，多一点引导',
  meta: [
    { name: 'description', content: 'YnsStudy：少一点迷茫，多一点引导，学习之路不再孤单。永远相信美好的事情即将发生。' },
    { name: 'keywords', content: 'YnsStudy, 学习引导, 内容社区, 技术博客, AI辅助学习, 开发者社区, 编程学习' }
  ]
})

// 按 B 键切换封面的颜色
const themes = ['theme-green', 'theme-blue', 'theme-red']
const currentTheme = ref(themes[0])

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

const menuClick = (menu) => router.push({ name: menu.router })

const buttonClick = (type) => {
  const routes = {
    ai: () => window.open(router.resolve({ name: 'YnsStudyAi' }).href, "YnsStudyAi"),
    article: () => router.push({ name: "Home" }),
    aboutWe: () => router.push({ name: "About" })
  }
  routes[type]?.()
}
</script>

<style scoped>
.theme-green { --cover-color: #3f5a57; --band-color: #b5473d; }
.theme-blue { --cover-color: #34496b; --band-color: #d9a441; }
.theme-red { --cover-color: #7a3b36; --band-color: #3f5a57; }

.welcome {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  color: var(--j-ink);
}

.welcome-nav {
  width: 100%;
  max-width: 1120px;
  margin: 0 auto;
  padding: 20px 24px;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  gap: 32px;
}

.welcome-brand {
  font-family: var(--j-hand);
  font-size: 24px;
}

.welcome-links {
  flex: 1;
  display: flex;
  justify-content: center;
  gap: 24px;
}

.welcome-links a {
  font-size: 15px;
  color: var(--j-ink-soft);
  text-decoration: none;
}

.welcome-links a:hover {
  color: var(--j-ink);
  text-decoration: underline;
  text-decoration-color: var(--j-pen);
  text-underline-offset: 5px;
}

.welcome-main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 72px;
  padding: 32px 24px 80px;
}

/* ============ 封面 ============ */
.cover {
  position: relative;
  flex-shrink: 0;
  width: 380px;
  height: 500px;
  border-radius: 4px 14px 14px 4px;
  background-color: var(--cover-color);
  transition: background-color 0.5s ease;
  /* 布纹 */
  background-image: repeating-linear-gradient(0deg, rgba(255, 255, 255, 0.035) 0 1px, transparent 1px 3px),
  repeating-linear-gradient(90deg, rgba(0, 0, 0, 0.04) 0 1px, transparent 1px 3px);
  box-shadow: inset -6px 0 10px -6px rgba(0, 0, 0, 0.35),
  6px 6px 0 -1px #f3eee2,
  7px 7px 0 -1px var(--j-rule-strong),
  12px 12px 0 -2px #f3eee2,
  13px 13px 0 -2px var(--j-rule-strong),
  0 30px 50px -24px rgba(40, 32, 20, 0.6);
  transform: rotate(-2deg);
}

/* 书脊 */
.cover-spine {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 26px;
  border-radius: 4px 0 0 4px;
  background: linear-gradient(90deg, rgba(0, 0, 0, 0.28), rgba(0, 0, 0, 0.08) 70%, rgba(255, 255, 255, 0.06));
}

/* 绑带 */
.cover-band {
  position: absolute;
  top: -2px;
  bottom: -2px;
  right: 44px;
  width: 12px;
  background: var(--band-color);
  transition: background-color 0.5s ease;
  box-shadow: 1px 0 2px rgba(0, 0, 0, 0.3), inset -2px 0 2px rgba(0, 0, 0, 0.15);
}

.cover-label {
  position: absolute;
  top: 120px;
  left: 44px;
  right: 66px;
  padding: 28px 22px 22px;
  background: var(--j-paper);
  border: 1px solid #e8e0cf;
  text-align: center;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
}

.cover-kicker {
  margin: 0;
  font-family: var(--j-hand);
  font-size: 14px;
  color: var(--j-muted);
  letter-spacing: 0.1em;
}

.cover-title {
  margin: 10px 0 0;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 30px;
  line-height: 1.25;
  white-space: nowrap;
}

.cover-slogan {
  margin: 16px 0 0;
  font-weight: normal;
  padding-top: 14px;
  border-top: 1px dashed var(--j-rule-strong);
  font-family: var(--j-hand);
  font-size: 18px;
  color: var(--j-ink-soft);
}

/* ============ 便签入口 ============ */
.notes {
  display: flex;
  flex-direction: column;
  gap: 22px;
  width: 320px;
}

.notes-intro span {
  display: inline-block;
}

.notes-intro {
  margin: 0 0 8px;
  font-family: var(--j-hand);
  font-size: 19px;
  line-height: 1.7;
  color: var(--j-ink-soft);
}

.note {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 16px 20px 14px;
  color: var(--j-ink);
  text-decoration: none;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.12), 0 10px 18px -12px rgba(60, 50, 30, 0.45);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.note::after {
  content: "→";
  position: absolute;
  right: 18px;
  top: 50%;
  margin-top: -12px;
  font-size: 18px;
  color: rgba(43, 42, 39, 0.45);
  transition: transform 0.2s ease;
}

.note:hover {
  transform: rotate(0deg) translateY(-3px);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.12), 0 16px 22px -12px rgba(60, 50, 30, 0.5);
}

.note:hover::after {
  transform: translateX(4px);
}

.note-green { background: #d4ead9; transform: rotate(-1.5deg); }
.note-yellow { background: var(--j-note); transform: rotate(1deg); margin-left: 18px; }
.note-pink { background: #f6d9d6; transform: rotate(-0.8deg); }

.note-title {
  font-family: var(--j-hand);
  font-size: 20px;
}

.interaction-hint {
  padding: 0 24px 28px;
  text-align: center;
  font-size: 13px;
  color: var(--j-muted);
}

kbd {
  padding: 1px 7px;
  border: 1px solid var(--j-rule-strong);
  border-radius: 3px;
  background: var(--j-paper);
  box-shadow: 0 2px 0 var(--j-rule-strong);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  color: var(--cover-color);
}

@media (max-width: 860px) {
  .welcome-links {
    display: none;
  }

  .welcome-nav {
    justify-content: space-between;
  }

  .welcome-main {
    flex-direction: column;
    gap: 48px;
    padding-top: 16px;
  }

  .cover {
    width: 300px;
    height: 390px;
  }

  .cover-label {
    top: 90px;
    left: 40px;
    right: 62px;
  }

  .cover-title {
    font-size: 24px;
  }

  .cover-slogan {
    font-size: 14px;
    white-space: nowrap;
  }

  .notes {
    width: 100%;
    max-width: 340px;
  }

  .note-yellow {
    margin-left: 0;
  }
}
</style>
