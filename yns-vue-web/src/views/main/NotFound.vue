<template>
  <div class="not-found j-desk">
    <main class="not-found-card j-paper">
      <span class="j-tape j-tape--top"></span>
      <p class="not-found-code">404</p>
      <h1 class="j-title">这一页没有找到</h1>
      <p class="not-found-text">地址可能输错了，或者内容已经被移动、删除。</p>
      <nav class="not-found-links" aria-label="常用入口">
        <router-link to="/">回到首页</router-link>
        <router-link to="/ynsStudy/Home">最新文章</router-link>
        <router-link to="/ynsStudy/About">关于本站</router-link>
      </nav>
    </main>
  </div>
</template>

<script setup>
import { useHead } from '@vueuse/head';

// 未匹配的地址由 Nginx 返回 404 状态码，这里再声明 noindex，避免 SPA 内跳转时被当作正常页面
useHead({
  title: '页面不存在 - YnsStudy',
  meta: [{ name: 'robots', content: 'noindex,follow' }]
});
</script>

<style scoped>
.not-found {
  min-height: calc(100vh - 160px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 16px;
  box-sizing: border-box;
}

.not-found-card {
  position: relative;
  width: 100%;
  max-width: 520px;
  padding: 44px 28px 36px;
  text-align: center;
}

.not-found-code {
  margin: 0;
  font-family: var(--j-num);
  font-size: 56px;
  color: var(--j-stamp);
  letter-spacing: 4px;
}

.not-found-card .j-title {
  font-size: 26px;
}

.not-found-text {
  color: var(--j-muted);
  line-height: 1.8;
}

.not-found-links {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 18px;
  margin-top: 22px;
}

.not-found-links a {
  color: var(--j-pen);
  text-decoration: none;
  border-bottom: 1px dashed var(--j-pen);
}
</style>
