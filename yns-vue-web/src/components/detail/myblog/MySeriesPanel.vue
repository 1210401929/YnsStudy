<!-- 写作页侧栏的“我的系列”：列出自己的系列和文章数，点开系列页可以改名、调整顺序、删除 -->
<template>
  <section class="my-series">
    <div class="my-series-head">
      <span class="my-series-title">我的系列</span>
      <button type="button" class="my-series-add" @click="create">＋ 新建</button>
    </div>
    <p v-if="loaded && !list.length" class="my-series-empty">
      还没有系列。把相关的文章编成系列，读者可以按顺序阅读。
    </p>
    <ul v-else class="my-series-list">
      <li v-for="item in list" :key="item.ID">
        <router-link :to="seriesHref(item.ID)" class="my-series-link" :title="`打开系列《${item.TITLE}》，可以调整顺序`">
          <span class="my-series-name">{{ item.TITLE }}</span>
          <span class="my-series-count">{{ item.ARTICLE_COUNT || 0 }} 篇</span>
        </router-link>
      </li>
    </ul>
  </section>
</template>

<script setup>
import { ref } from 'vue'
import { loadMySeries, promptCreateSeries, seriesHref } from '@/utils/blogSeries.js'

const list = ref([])
const loaded = ref(false)

async function refresh(force = false) {
  list.value = await loadMySeries({ refresh: force })
  loaded.value = true
}
refresh(true)

async function create() {
  if (await promptCreateSeries()) refresh()
}

defineExpose({ refresh })
</script>

<style scoped>
.my-series {
  margin-top: 16px;
  padding: 14px 16px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
}

.my-series-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.my-series-title {
  font-family: var(--j-hand);
  font-size: 18px;
  color: var(--j-ink);
}

.my-series-add {
  border: none;
  background: none;
  padding: 0;
  color: var(--j-pen);
  font-size: 13px;
  cursor: pointer;
}

.my-series-empty {
  margin: 0;
  font-size: 13px;
  line-height: 1.7;
  color: var(--j-muted);
}

.my-series-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.my-series-link {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  padding: 6px 0;
  border-top: 1px dashed var(--j-rule);
  color: var(--j-ink);
  text-decoration: none;
  font-size: 14px;
}

.my-series-link:hover .my-series-name {
  color: var(--j-pen);
  text-decoration: underline;
}

.my-series-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.my-series-count {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--j-muted);
}
</style>
