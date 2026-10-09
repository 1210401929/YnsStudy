<template>
  <div class="welcome j-desk">
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
      <div class="cover" role="img" aria-label="YnsStudy 学习手账封面">
        <span class="cover-spine"></span>
        <span class="cover-band"></span>
        <div class="cover-label">
          <span class="j-tape j-tape--top"></span>
          <p class="cover-kicker">学习手账 · {{ year }}</p>
          <h1 class="cover-title">YnsStudy</h1>
          <p class="cover-slogan">少一点迷茫，多一点引导</p>
          <p class="cover-owner">属于每一个正在学习的人</p>
        </div>
      </div>

      <!-- 贴在旁边的便签，就是入口 -->
      <div class="notes">
        <p class="notes-intro">学习之路不再孤单，<br>永远相信美好的事情即将发生。</p>
        <a href="/ynsStudy/Home" class="note note-yellow" @click.prevent="buttonClick('article')">
          <span class="note-title">翻开看看</span>
          <span class="note-desc">文章、资源和社区都在里面</span>
        </a>
        <a href="/YnsStudyAi" class="note note-green" @click.prevent="buttonClick('ai')">
          <span class="note-title">问问智能助手</span>
          <span class="note-desc">卡住的时候，找它聊聊</span>
        </a>
        <a href="/ynsStudy/About" class="note note-pink" @click.prevent="buttonClick('aboutWe')">
          <span class="note-title">认识一下我们</span>
          <span class="note-desc">这个小站是怎么来的</span>
        </a>
      </div>
    </main>
  </div>
</template>

<script setup>
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

const year = new Date().getFullYear()

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
  background-color: #3f5a57;
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
  background: #b5473d;
  box-shadow: 1px 0 2px rgba(0, 0, 0, 0.3), inset -2px 0 2px rgba(0, 0, 0, 0.15);
}

.cover-label {
  position: absolute;
  top: 120px;
  left: 50px;
  right: 72px;
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
  font-size: 44px;
  line-height: 1.1;
}

.cover-slogan {
  margin: 16px 0 0;
  padding-top: 14px;
  border-top: 1px dashed var(--j-rule-strong);
  font-family: var(--j-hand);
  font-size: 18px;
  color: var(--j-ink-soft);
}

.cover-owner {
  margin: 18px 0 0;
  font-size: 12px;
  color: var(--j-muted);
  letter-spacing: 0.05em;
}

/* ============ 便签入口 ============ */
.notes {
  display: flex;
  flex-direction: column;
  gap: 22px;
  width: 300px;
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

.note-yellow { background: var(--j-note); transform: rotate(-1.5deg); }
.note-green { background: #d4ead9; transform: rotate(1deg); margin-left: 18px; }
.note-pink { background: #f6d9d6; transform: rotate(-0.8deg); }

.note-title {
  font-family: var(--j-hand);
  font-size: 20px;
}

.note-desc {
  font-size: 13px;
  color: var(--j-ink-soft);
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
    font-size: 36px;
  }

  .cover-slogan {
    font-size: 15px;
  }

  .notes {
    width: 100%;
    max-width: 340px;
  }

  .note-green {
    margin-left: 0;
  }
}
</style>
