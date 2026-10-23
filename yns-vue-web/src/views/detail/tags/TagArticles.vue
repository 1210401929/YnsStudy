<!-- 文章标签页 /tag/标签名：这个标签下的公开文章。首屏由 Go 输出（seo_tag.go），这里接管后提供分页和跳转 -->
<template>
  <div class="tag-page j-desk">
    <div class="tag-layout">
      <main class="tag-paper">
        <nav class="tag-crumb" aria-label="面包屑">
          <router-link to="/ynsStudy/Home">最新文章</router-link> / 标签
        </nav>

        <header class="tag-head">
          <h1 class="tag-title">#{{ tagName }}</h1>
          <p v-if="!notFound" class="tag-summary">
            <template v-if="loaded">共 {{ total }} 篇公开文章</template>
            <template v-else>正在翻找…</template>
            · <router-link :to="{ name: 'Home', query: { tag: tagName } }">在首页筛选</router-link>
          </p>
        </header>

        <p v-if="notFound" class="tag-empty">这个标签下还没有公开文章。<router-link to="/ynsStudy/Home">看看最新文章 →</router-link></p>

        <ol v-else class="tag-list">
          <li v-for="article in articles" :key="article.GUID" class="tag-item">
            <time class="tag-date" :datetime="article.CREATE_TIME">{{ formatDay(article.CREATE_TIME) }}</time>
            <div class="tag-body">
              <h2 class="tag-item-title">
                <router-link :to="{ name: 'oneBlog', params: { g: article.GUID } }">{{ article.BLOG_TITLE || '未命名文章' }}</router-link>
              </h2>
              <p v-if="article.EXCERPT" class="tag-excerpt">{{ article.EXCERPT }}</p>
              <div class="tag-meta">
                <router-link v-if="article.USERNUM" :to="`/user/${encodeURIComponent(article.USERNUM)}`" class="tag-author">{{ article.USERNAME }}</router-link>
                <span v-else>{{ article.USERNAME }}</span>
                <router-link
                    v-for="tag in article.TAGS || []"
                    :key="tag"
                    :to="blogTagHref(tag)"
                    class="j-tag tag-chip"
                    :class="{ 'is-active': tag.toLowerCase() === tagName.toLowerCase() }"
                    rel="tag"
                >#{{ tag }}</router-link>
              </div>
            </div>
            <img
                v-if="article.FIRST_IMAGE"
                :src="article.FIRST_IMAGE"
                :alt="`${article.BLOG_TITLE} 的文章配图`"
                class="tag-cover"
                loading="lazy"
                decoding="async"
                @error="article.FIRST_IMAGE = ''"
            />
          </li>
        </ol>

        <el-pagination
            v-if="!notFound && total > pageSize"
            class="tag-pager"
            layout="prev, pager, next"
            :total="total"
            :page-size="pageSize"
            :current-page="page"
            @current-change="goPage"
        />
      </main>

      <aside v-if="otherTags.length" class="tag-side">
        <h2 class="tag-side-title">其他标签</h2>
        <div class="tag-side-list">
          <router-link v-for="tag in otherTags" :key="tag.NAME" :to="blogTagHref(tag.NAME)" class="j-tag" rel="tag">
            #{{ tag.NAME }}<span class="tag-side-count">{{ tag.TOTAL }}</span>
          </router-link>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { sendAxiosRequest } from '@/utils/common.js'
import { blogTagHref, loadBlogTags } from '@/utils/blogTags.js'
import { absoluteUrl, SITE_NAME, useSeo } from '@/utils/seo.js'

// 收录门槛与 Go 端 service.BlogTagIndexableMin 一致：文章太少的标签页不让搜索引擎收录
const INDEXABLE_MIN = 2

const route = useRoute()
const router = useRouter()

const tagName = ref(String(route.params.name || '').trim())
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const articles = ref([])
const loaded = ref(false)
const notFound = ref(false)
const popularTags = ref([])
let requestId = 0

const readPage = () => {
  const value = Number.parseInt(String(route.query.page || '1'), 10)
  return Number.isFinite(value) && value > 0 ? value : 1
}

async function load() {
  const current = ++requestId
  const name = String(route.params.name || '').trim()
  tagName.value = name
  page.value = readPage()
  loaded.value = false
  const res = await sendAxiosRequest('/blog-api/blog/getTagArticles', { tag: name, page: page.value }).catch(() => null)
  if (current !== requestId) return
  const result = res && !res.isError ? res.result : null
  loaded.value = true
  if (!result || !result.tag || !Number(result.total)) {
    notFound.value = true
    total.value = 0
    articles.value = []
    return
  }
  notFound.value = false
  // 大小写以第一次创建的写法为准（go → Go），地址也换成规范写法
  if (result.tag !== name) {
    router.replace({ path: blogTagHref(result.tag), query: route.query })
  }
  tagName.value = result.tag
  total.value = Number(result.total) || 0
  pageSize.value = Number(result.pageSize) || 20
  articles.value = result.data || []
}

watch(() => [route.params.name, route.query.page], () => {
  if (route.name === 'tag') load()
}, { immediate: true })

loadBlogTags().then(list => (popularTags.value = list))
const otherTags = computed(() => popularTags.value
    .filter(tag => tag.NAME && tag.NAME.toLowerCase() !== tagName.value.toLowerCase())
    .slice(0, 20))

function goPage(next) {
  const query = { ...route.query }
  if (next > 1) query.page = String(next)
  else delete query.page
  router.push({ path: route.path, query })
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const formatDay = (value) => String(value || '').slice(0, 10)

// 与 Go 端 renderSEOTagHTML 一致的标题、描述和规范网址
useSeo(() => {
  const name = tagName.value
  const titles = articles.value.slice(0, 3).map(item => `《${item.BLOG_TITLE}》`).filter(item => item !== '《》')
  let description = `${SITE_NAME} 上标签为「${name}」的 ${total.value} 篇公开文章`
  if (titles.length) description += `，包括${titles.join('、')}`
  const path = blogTagHref(name) + (page.value > 1 ? `?page=${page.value}` : '')
  return {
    title: `${name} 相关文章${page.value > 1 ? `（第 ${page.value} 页）` : ''} - ${SITE_NAME}`,
    description: `${description}。`.slice(0, 180),
    path,
    noindex: notFound.value || (loaded.value && total.value < INDEXABLE_MIN),
    jsonLd: {
      '@context': 'https://schema.org',
      '@type': 'CollectionPage',
      name: `${name} 相关文章`,
      url: absoluteUrl(path),
      keywords: name
    }
  }
})
</script>

<style scoped>
.tag-page {
  min-height: 100%;
  padding: 28px 16px 48px;
}

.tag-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 260px;
  gap: 28px;
  max-width: 1100px;
  margin: 0 auto;
  align-items: start;
}

.tag-paper {
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
  padding: 28px 32px 32px;
  min-width: 0;
}

.tag-crumb {
  font-size: 13px;
  color: var(--j-muted);
  margin-bottom: 14px;
}

.tag-crumb a,
.tag-summary a,
.tag-empty a {
  color: var(--j-pen);
  text-decoration: none;
}

.tag-title {
  margin: 0 0 6px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 32px;
  color: var(--j-ink);
}

.tag-summary {
  margin: 0 0 16px;
  font-size: 14px;
  color: var(--j-muted);
}

.tag-empty {
  padding: 32px 0;
  color: var(--j-muted);
}

.tag-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.tag-item {
  display: grid;
  grid-template-columns: 92px minmax(0, 1fr) auto;
  gap: 18px;
  padding: 18px 0;
  border-top: 1px dashed var(--j-rule);
}

.tag-date {
  padding-top: 4px;
  font-size: 13px;
  color: var(--j-muted);
}

.tag-item-title {
  margin: 0 0 6px;
  font-size: 18px;
  line-height: 1.5;
}

.tag-item-title a {
  color: var(--j-ink);
  text-decoration: none;
}

.tag-item-title a:hover {
  color: var(--j-pen);
  text-decoration: underline;
}

.tag-excerpt {
  margin: 0 0 8px;
  font-size: 14px;
  line-height: 1.75;
  color: var(--j-ink-soft);
  overflow-wrap: anywhere;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.tag-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--j-muted);
}

.tag-author {
  color: inherit;
  text-decoration: none;
  margin-right: 4px;
}

.tag-author:hover {
  color: var(--j-pen);
}

.tag-chip {
  padding: 0 6px;
  font-size: 12px;
  line-height: 1.6;
}

.tag-cover {
  width: 120px;
  height: 84px;
  object-fit: cover;
  border: 4px solid #fff;
  box-shadow: 0 2px 6px rgba(60, 50, 30, 0.2);
}

.tag-pager {
  justify-content: center;
  margin-top: 24px;
}

.tag-side {
  background: var(--j-note);
  padding: 20px 18px;
  box-shadow: var(--j-shadow);
  transform: rotate(0.6deg);
}

.tag-side-title {
  margin: 0 0 12px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 20px;
}

.tag-side-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-side-count {
  margin-left: 4px;
  font-size: 11px;
  color: var(--j-muted);
}

@media (max-width: 900px) {
  .tag-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .tag-side {
    transform: none;
  }
}

@media (max-width: 640px) {
  .tag-page {
    padding: 14px 10px 32px;
  }

  .tag-paper {
    padding: 20px 16px 24px;
  }

  .tag-title {
    font-size: 26px;
  }

  .tag-item {
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 6px 12px;
  }

  .tag-date {
    grid-column: 1 / -1;
    padding: 0;
  }

  .tag-cover {
    width: 84px;
    height: 64px;
  }
}
</style>
