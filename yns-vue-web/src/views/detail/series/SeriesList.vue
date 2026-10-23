<!-- 全部系列 /series：至少有一篇公开文章的系列，按最近更新排列。首屏由 Go 输出（seo_series.go） -->
<template>
  <div class="series-list-page j-desk">
    <main class="list-paper">
      <nav class="list-crumb" aria-label="面包屑">
        <router-link to="/ynsStudy/Home">最新文章</router-link> / 文章系列
      </nav>
      <header class="list-head">
        <h1 class="list-title">文章系列</h1>
        <p class="list-desc">按顺序编排的专题文章，适合从头读到尾。</p>
        <button v-if="isLoggedIn" type="button" class="j-button" @click="createSeries">＋ 新建我的系列</button>
      </header>

      <p v-if="loaded && !list.length" class="list-empty">还没有系列。作者写文章时可以在“所属系列”里新建一个。</p>
      <ul v-else class="series-cards">
        <li v-for="item in list" :key="item.ID" class="series-card">
          <router-link :to="seriesHref(item.ID)" class="card-link">
            <span class="card-count">{{ item.ARTICLE_COUNT }}<small>篇</small></span>
            <span class="card-body">
              <span class="card-title">{{ item.TITLE }}</span>
              <span v-if="item.DESCRIPTION" class="card-desc">{{ item.DESCRIPTION }}</span>
              <span class="card-meta">{{ item.USERNAME }} · 更新于 {{ formatDay(item.LAST_MODIFIED) }}</span>
            </span>
          </router-link>
        </li>
      </ul>
    </main>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { sendAxiosRequest } from '@/utils/common.js'
import { promptCreateSeries, SERIES_LIST_PATH, seriesHref } from '@/utils/blogSeries.js'
import { SITE_NAME, useSeo } from '@/utils/seo.js'
import { useUserStore } from '@/stores/main/user.js'

const router = useRouter()
const userStore = useUserStore()
const isLoggedIn = computed(() => !!userStore.userBean?.code)

const list = ref([])
const loaded = ref(false)
sendAxiosRequest('/blog-api/series/getSeriesList', { sort: 'latest', limit: 200 })
    .then(res => (list.value = res && !res.isError && Array.isArray(res.result) ? res.result : []))
    .catch(() => (list.value = []))
    .finally(() => (loaded.value = true))

const formatDay = (value) => String(value || '').slice(0, 10)

// 新系列还没有公开文章，不会出现在这个列表里，建好后直接打开它的页面
async function createSeries() {
  const created = await promptCreateSeries()
  if (created?.ID) router.push(seriesHref(created.ID))
}

useSeo(() => ({
  title: `文章系列 - ${SITE_NAME}`,
  description: `${SITE_NAME} 的文章系列：按顺序编排的专题文章合集，共 ${list.value.length} 个系列。`,
  path: SERIES_LIST_PATH,
  noindex: loaded.value && !list.value.length
}))
</script>

<style scoped>
.series-list-page {
  min-height: 100%;
  padding: 28px 16px 48px;
}

.list-paper {
  max-width: 960px;
  margin: 0 auto;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
  padding: 28px 32px 36px;
}

.list-crumb {
  font-size: 13px;
  color: var(--j-muted);
  margin-bottom: 14px;
}

.list-crumb a {
  color: var(--j-pen);
  text-decoration: none;
}

.list-head {
  margin-bottom: 22px;
}

.list-title {
  margin: 0 0 6px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 32px;
}

.list-desc {
  margin: 0 0 12px;
  color: var(--j-muted);
  font-size: 14px;
}

.list-empty {
  padding: 28px 0;
  color: var(--j-muted);
}

.series-cards {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 18px;
}

.card-link {
  display: flex;
  gap: 14px;
  height: 100%;
  padding: 16px 18px;
  border: 1px solid var(--j-rule);
  background: var(--j-paper-warm);
  color: inherit;
  text-decoration: none;
  box-sizing: border-box;
  transition: transform 0.15s, box-shadow 0.15s;
}

.card-link:hover {
  transform: translateY(-2px) rotate(-0.4deg);
  box-shadow: var(--j-shadow);
}

.card-count {
  flex-shrink: 0;
  width: 54px;
  font-family: var(--j-num);
  font-size: 30px;
  line-height: 1.1;
  color: var(--j-stamp);
  text-align: center;
}

.card-count small {
  display: block;
  font-size: 12px;
  color: var(--j-muted);
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.card-title {
  font-family: var(--j-hand);
  font-size: 20px;
  color: var(--j-ink);
  overflow-wrap: anywhere;
}

.card-desc {
  font-size: 13px;
  line-height: 1.7;
  color: var(--j-ink-soft);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-meta {
  font-size: 12px;
  color: var(--j-muted);
}

@media (max-width: 640px) {
  .series-list-page {
    padding: 14px 10px 32px;
  }

  .list-paper {
    padding: 20px 16px 24px;
  }

  .series-cards {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
