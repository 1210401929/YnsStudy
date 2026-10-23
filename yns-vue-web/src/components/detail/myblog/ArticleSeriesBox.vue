<!-- 文章所属系列：顶部显示“本文属于系列《X》第 n / m 篇”并可展开目录；底部显示系列里的上一篇、下一篇 -->
<template>
  <aside v-if="series && placement === 'top'" class="series-box" aria-label="所属系列">
    <div class="series-head">
      <span class="series-badge">系列</span>
      <span class="series-text">
        本文属于《<router-link :to="seriesHref(series.ID)" class="series-name">{{ series.TITLE }}</router-link>》
        <template v-if="series.INDEX">第 {{ series.INDEX }} / {{ series.TOTAL }} 篇</template>
        <template v-else>共 {{ series.TOTAL }} 篇</template>
      </span>
      <button type="button" class="series-toggle" :aria-expanded="expanded" @click="expanded = !expanded">
        {{ expanded ? '收起目录' : '展开目录' }}
      </button>
    </div>
    <ol v-if="expanded" class="series-items">
      <li v-for="(item, index) in series.ITEMS" :key="item.GUID" :class="{ 'is-current': item.GUID === String(currentGuid) }">
        <span class="series-no">{{ index + 1 }}.</span>
        <span v-if="item.GUID === String(currentGuid)" class="series-item-title">{{ item.BLOG_TITLE || '未命名文章' }}（本文）</span>
        <router-link v-else :to="{ name: 'oneBlog', params: { g: item.GUID } }" class="series-item-title">{{ item.BLOG_TITLE || '未命名文章' }}</router-link>
        <span v-if="item.BLOG_TYPE !== 'public'" class="series-private">私密</span>
      </li>
    </ol>
  </aside>

  <nav v-else-if="series && placement === 'bottom' && (series.PREV || series.NEXT)" class="series-pager" aria-label="系列里的上一篇和下一篇">
    <router-link v-if="series.PREV" :to="{ name: 'oneBlog', params: { g: series.PREV.GUID } }" class="series-pager-link">
      <small>系列上一篇</small>← {{ series.PREV.BLOG_TITLE || '未命名文章' }}
    </router-link>
    <span v-else></span>
    <router-link v-if="series.NEXT" :to="{ name: 'oneBlog', params: { g: series.NEXT.GUID } }" class="series-pager-link is-next">
      <small>系列下一篇</small>{{ series.NEXT.BLOG_TITLE || '未命名文章' }} →
    </router-link>
  </nav>
</template>

<script setup>
import { ref } from 'vue'
import { seriesHref } from '@/utils/blogSeries.js'

defineProps({
  // 后端 getBlog 返回的 SERIES：ID、TITLE、TOTAL、INDEX、PREV、NEXT、ITEMS
  series: { type: Object, default: null },
  currentGuid: { type: [String, Number], default: '' },
  // top：标题下的系列信息和目录；bottom：文末的系列上一篇、下一篇
  placement: { type: String, default: 'top' }
})

const expanded = ref(false)
</script>

<style scoped>
.series-box {
  margin: 0 0 20px;
  padding: 12px 16px;
  background: var(--j-note);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.12);
  font-size: 14px;
}

.series-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 10px;
}

.series-badge {
  padding: 0 6px;
  border: 1px solid var(--j-stamp);
  color: var(--j-stamp);
  font-size: 12px;
  line-height: 1.6;
  transform: rotate(-3deg);
}

.series-text {
  flex: 1;
  min-width: 0;
  color: var(--j-ink-soft);
}

.series-name {
  color: var(--j-pen);
  text-decoration: none;
  font-weight: 600;
}

.series-name:hover {
  text-decoration: underline;
}

.series-toggle {
  border: none;
  background: none;
  padding: 0;
  color: var(--j-pen);
  font-size: 13px;
  cursor: pointer;
}

.series-items {
  list-style: none;
  margin: 10px 0 0;
  padding: 10px 0 0;
  border-top: 1px dashed rgba(60, 50, 30, 0.25);
}

.series-items li {
  display: flex;
  align-items: baseline;
  gap: 6px;
  padding: 3px 0;
  line-height: 1.6;
}

.series-no {
  flex-shrink: 0;
  width: 2em;
  color: var(--j-muted);
  font-family: var(--j-num);
}

.series-item-title {
  color: var(--j-ink);
  text-decoration: none;
  overflow-wrap: anywhere;
}

a.series-item-title:hover {
  color: var(--j-pen);
  text-decoration: underline;
}

.is-current .series-item-title {
  color: var(--j-stamp);
  font-weight: 600;
}

.series-private {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--j-muted);
}

.series-pager {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  margin: 0 0 18px;
}

.series-pager-link {
  display: flex;
  flex-direction: column;
  max-width: 48%;
  padding: 10px 14px;
  border: 1px dashed var(--j-rule-strong);
  background: var(--j-paper-warm);
  color: var(--j-pen);
  text-decoration: none;
  font-size: 14px;
  line-height: 1.6;
  overflow-wrap: anywhere;
}

.series-pager-link small {
  color: var(--j-muted);
  font-size: 12px;
}

.series-pager-link.is-next {
  margin-left: auto;
  text-align: right;
}

.series-pager-link:hover {
  border-color: var(--j-pen);
}

@media (max-width: 640px) {
  .series-pager {
    flex-direction: column;
  }

  .series-pager-link {
    max-width: none;
  }
}
</style>
