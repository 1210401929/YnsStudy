<!-- 文章底部的站内链接：上一篇、下一篇、同作者的其他文章和文章归档 -->
<template>
  <nav v-if="loaded" class="article-links" aria-label="更多文章">
    <div v-if="previous || next" class="article-pager">
      <a v-if="previous" class="pager-link" :href="articleHref(previous)">
        <span class="pager-label">← 上一篇</span>
        <span class="pager-title">{{ articleTitle(previous) }}</span>
      </a>
      <span v-else></span>
      <a v-if="next" class="pager-link pager-next" :href="articleHref(next)">
        <span class="pager-label">下一篇 →</span>
        <span class="pager-title">{{ articleTitle(next) }}</span>
      </a>
    </div>

    <section v-if="related.length" class="related">
      <h2 class="related-title">{{ author || '作者' }} 的其他文章</h2>
      <ul>
        <li v-for="item in related" :key="item.GUID">
          <a :href="articleHref(item)">{{ articleTitle(item) }}</a>
        </li>
      </ul>
    </section>

    <a class="archive-link" href="/archive">浏览全部文章 →</a>
  </nav>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { sendAxiosRequest } from '@/utils/common.js'

const props = defineProps({
  blogId: { type: [String, Number], required: true },
  author: { type: String, default: '' }
})

const loaded = ref(false)
const previous = ref(null)
const next = ref(null)
const relatedRaw = ref([])

// 已经作为上一篇/下一篇出现的文章不再重复列出
const related = computed(() => {
  const shown = new Set([previous.value?.GUID, next.value?.GUID].filter(Boolean))
  return relatedRaw.value.filter(item => item.GUID && !shown.has(item.GUID))
})

// 使用普通链接整页跳转：文章页首屏由服务端输出，搜索引擎也能直接抓取这些地址
const articleHref = (item) => `/oneBlog/${encodeURIComponent(item.GUID)}`
const articleTitle = (item) => item.BLOG_TITLE || '未命名文章'

const load = async (blogId) => {
  loaded.value = false
  try {
    const res = await sendAxiosRequest('/blog-api/blog/getArticleLinks', { blogId })
    if (res && !res.isError && res.result) {
      previous.value = res.result.previous || null
      next.value = res.result.next || null
      relatedRaw.value = Array.isArray(res.result.related) ? res.result.related : []
    }
  } catch (e) {
    console.error('加载相关文章失败', e)
  } finally {
    loaded.value = true
  }
}

watch(() => props.blogId, (blogId) => blogId && load(blogId), { immediate: true })
</script>

<style scoped>
.article-links {
  margin-top: 28px;
  padding-top: 20px;
  border-top: 1px dashed var(--j-rule-strong);
}

/* 上一篇 / 下一篇：两张索引卡，顶上露出彩色书签 */
.article-pager {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 30px 16px;
  padding-top: 22px;
}

.pager-link {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 14px 16px 12px;
  border: 1px solid var(--j-rule);
  border-top: 2px solid rgba(232, 165, 155, 0.65);
  background: var(--j-paper);
  box-shadow: var(--j-shadow);
  color: var(--j-ink);
  text-decoration: none;
  transition: transform 0.2s ease;
}

.pager-link:hover {
  transform: translateY(-2px);
}

.pager-next {
  text-align: right;
}

.pager-label {
  position: absolute;
  top: -24px;
  left: 14px;
  padding: 2px 12px 3px;
  border-radius: 6px 6px 0 0;
  background: #f6e3a1;
  font-family: var(--j-hand);
  font-size: 13px;
  color: var(--j-ink);
}

.pager-next .pager-label {
  left: auto;
  right: 14px;
  background: #c7e2cf;
}

.pager-title {
  font-size: 14px;
  line-height: 1.6;
  overflow-wrap: anywhere;
}

.related {
  margin-top: 22px;
}

.related-title {
  margin: 0 0 8px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 18px;
  color: var(--j-ink);
}

/* 同作者文章：楷体手写清单，实心圆点 */
.related ul {
  margin: 0;
  padding-left: 20px;
  list-style: disc;
}

.related li {
  padding: 3px 0;
  font-family: var(--j-hand);
  font-size: 15px;
  line-height: 1.7;
}

.related li::marker {
  color: var(--j-pen);
}

.related a,
.archive-link {
  color: var(--j-pen);
  text-decoration: none;
}

.related a:hover,
.archive-link:hover {
  text-decoration: underline;
  text-underline-offset: 4px;
}

.archive-link {
  display: inline-block;
  margin-top: 18px;
  font-size: 14px;
}

@media (max-width: 640px) {
  .article-pager {
    grid-template-columns: 1fr;
  }

  .pager-next {
    text-align: left;
  }

  .pager-next .pager-label {
    left: 14px;
    right: auto;
  }
}
</style>
