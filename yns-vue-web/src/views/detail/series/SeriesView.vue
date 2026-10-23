<!-- 系列页 /series/编号：系列介绍和按顺序排列的文章目录。首屏由 Go 输出（seo_series.go）。
     作者本人（或能管理它的管理员）在这里改名、写简介、调整顺序、删除系列，还能看到系列里的私密文章。 -->
<template>
  <div class="series-page j-desk">
    <main class="series-paper">
      <nav class="series-crumb" aria-label="面包屑">
        <router-link to="/ynsStudy/Home">最新文章</router-link> /
        <router-link :to="SERIES_LIST_PATH">文章系列</router-link>
      </nav>

      <p v-if="notFound" class="series-empty">
        这个系列可能已删除，或者还没有公开文章。<router-link :to="SERIES_LIST_PATH">看看其他系列 →</router-link>
      </p>
      <p v-else-if="!series" class="series-empty">正在翻开这个系列…</p>

      <template v-else>
        <header class="series-head">
          <span class="series-stamp">系列</span>
          <h1 class="series-title">{{ series.TITLE }}</h1>
          <p class="series-meta">
            <router-link v-if="series.USERNUM" :to="`/user/${encodeURIComponent(series.USERNUM)}`">{{ series.USERNAME }}</router-link>
            <span v-else>{{ series.USERNAME }}</span>
            · 共 {{ publicCount }} 篇<template v-if="canManage && articles.length > publicCount">（另有 {{ articles.length - publicCount }} 篇私密文章只有你能看到）</template>
          </p>
          <p v-if="series.DESCRIPTION" class="series-desc">{{ series.DESCRIPTION }}</p>
          <div class="series-actions">
            <router-link v-if="firstPublic" :to="{ name: 'oneBlog', params: { g: firstPublic.GUID } }" class="j-button series-start">从第一篇开始读 →</router-link>
            <template v-if="canManage">
              <button type="button" class="tool-link" @click="openEdit">修改名称和简介</button>
              <button v-if="!sorting && articles.length > 1" type="button" class="tool-link" @click="startSort">调整顺序</button>
              <button type="button" class="tool-link danger" @click="removeSeries">删除系列</button>
            </template>
          </div>
        </header>

        <p v-if="sorting" class="sort-tip">
          用每篇文章右边的 ↑ ↓ 调整顺序，调好后点
          <button type="button" class="tool-link" :disabled="saving" @click="saveSort">保存顺序</button>
          或 <button type="button" class="tool-link" @click="cancelSort">取消</button>
        </p>

        <p v-if="!articles.length" class="series-empty">
          系列里还没有文章。写文章时在“所属系列”里选择这个系列，文章就会出现在这里。
        </p>
        <ol v-else class="series-list">
          <li v-for="(article, index) in displayArticles" :key="article.GUID" class="series-item">
            <span class="series-no">{{ index + 1 }}</span>
            <div class="series-body">
              <h2 class="series-item-title">
                <router-link :to="{ name: 'oneBlog', params: { g: article.GUID } }">{{ article.BLOG_TITLE || '未命名文章' }}</router-link>
                <span v-if="article.BLOG_TYPE !== 'public'" class="series-private">私密</span>
              </h2>
              <p v-if="article.EXCERPT" class="series-excerpt">{{ article.EXCERPT }}</p>
              <span class="series-date">{{ formatDay(article.CREATE_TIME) }}</span>
            </div>
            <div v-if="sorting" class="sort-buttons">
              <button type="button" :disabled="index === 0" :aria-label="`把《${article.BLOG_TITLE}》上移`" @click="move(index, -1)">↑</button>
              <button type="button" :disabled="index === displayArticles.length - 1" :aria-label="`把《${article.BLOG_TITLE}》下移`" @click="move(index, 1)">↓</button>
            </div>
          </li>
        </ol>
      </template>
    </main>

    <el-dialog v-model="editVisible" title="修改系列" width="min(460px, 92vw)">
      <el-form label-position="top" @submit.prevent>
        <el-form-item label="系列名称">
          <el-input v-model="editForm.title" :maxlength="MAX_SERIES_TITLE" show-word-limit/>
        </el-form-item>
        <el-form-item label="系列简介（选填，会显示在系列页和搜索结果里）">
          <el-input v-model="editForm.description" type="textarea" :rows="4" :maxlength="MAX_SERIES_DESCRIPTION" show-word-limit/>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { sendAxiosRequest, sendAxiosRequestChecked } from '@/utils/common.js'
import { loadMySeries, MAX_SERIES_DESCRIPTION, MAX_SERIES_TITLE, SERIES_LIST_PATH, seriesHref } from '@/utils/blogSeries.js'
import { absoluteUrl, SITE_NAME, useSeo } from '@/utils/seo.js'

// 与 Go 端 service.SeriesIndexableMin 一致
const INDEXABLE_MIN = 2

const route = useRoute()
const router = useRouter()

const series = ref(null)
const articles = ref([])
const canManage = ref(false)
const notFound = ref(false)
const sorting = ref(false)
const sortDraft = ref([])
let requestId = 0

const publicArticles = computed(() => articles.value.filter(item => item.BLOG_TYPE === 'public'))
const publicCount = computed(() => publicArticles.value.length)
const firstPublic = computed(() => publicArticles.value[0] || null)

async function load() {
  const current = ++requestId
  const id = String(route.params.id || '')
  sorting.value = false
  const res = await sendAxiosRequest('/blog-api/series/getSeries', { id }).catch(() => null)
  if (current !== requestId) return
  if (!res || res.isError || !res.result?.series) {
    notFound.value = true
    series.value = null
    articles.value = []
    return
  }
  notFound.value = false
  series.value = res.result.series
  articles.value = res.result.articles || []
  canManage.value = !!res.result.canManage
}

watch(() => route.params.id, () => {
  if (route.name === 'series') load()
}, { immediate: true })

const formatDay = (value) => String(value || '').slice(0, 10)

// ================= 管理 =================
const saving = ref(false)
const editVisible = ref(false)
const editForm = reactive({ title: '', description: '' })

function openEdit() {
  editForm.title = series.value?.TITLE || ''
  editForm.description = series.value?.DESCRIPTION || ''
  editVisible.value = true
}

async function saveEdit() {
  if (!editForm.title.trim()) {
    ElMessage.warning('系列名称不能为空')
    return
  }
  saving.value = true
  const result = await sendAxiosRequestChecked('/blog-api/series/updateSeries', {
    id: series.value.ID, title: editForm.title, description: editForm.description
  }, '保存失败')
  saving.value = false
  if (!result) return
  series.value.TITLE = editForm.title.trim()
  series.value.DESCRIPTION = editForm.description.trim()
  editVisible.value = false
  loadMySeries({ refresh: true })
  ElMessage.success('已保存')
}

async function removeSeries() {
  try {
    await ElMessageBox.confirm(`删除系列《${series.value.TITLE}》？系列里的文章不会被删除，只是不再属于这个系列。`, '删除系列', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    })
  } catch {
    return
  }
  const result = await sendAxiosRequestChecked('/blog-api/series/deleteSeries', { id: series.value.ID }, '删除失败')
  if (!result) return
  loadMySeries({ refresh: true })
  ElMessage.success('系列已删除，文章都还在')
  router.replace(SERIES_LIST_PATH)
}

// 调整顺序：在副本上移动，保存时一次提交（sorting、sortDraft 在上面声明，load 会用到）
const displayArticles = computed(() => (sorting.value ? sortDraft.value : articles.value))

function startSort() {
  sortDraft.value = [...articles.value]
  sorting.value = true
}

function cancelSort() {
  sorting.value = false
}

function move(index, step) {
  const list = [...sortDraft.value]
  const target = index + step
  if (target < 0 || target >= list.length) return
  ;[list[index], list[target]] = [list[target], list[index]]
  sortDraft.value = list
}

async function saveSort() {
  saving.value = true
  const result = await sendAxiosRequestChecked('/blog-api/series/sortSeries', {
    id: series.value.ID, guids: sortDraft.value.map(item => item.GUID)
  }, '保存顺序失败')
  saving.value = false
  if (!result) return
  articles.value = [...sortDraft.value]
  sorting.value = false
  ElMessage.success('顺序已保存')
}

// ================= SEO（与 Go 端 renderSEOSeriesHTML 一致） =================
useSeo(() => {
  if (!series.value) return { noindex: notFound.value }
  const name = series.value.TITLE
  const path = seriesHref(series.value.ID)
  const titles = publicArticles.value.slice(0, 3).map(item => `《${item.BLOG_TITLE}》`)
  const description = (series.value.DESCRIPTION ||
      `${series.value.USERNAME}的系列文章《${name}》，共 ${publicCount.value} 篇：${titles.join('、')}`).slice(0, 180)
  return {
    title: `${name}（系列） - ${SITE_NAME}`,
    description,
    path,
    noindex: publicCount.value < INDEXABLE_MIN,
    jsonLd: {
      '@context': 'https://schema.org',
      '@type': 'CreativeWorkSeries',
      name,
      description,
      url: absoluteUrl(path),
      hasPart: publicArticles.value.map((item, index) => ({
        '@type': 'BlogPosting',
        headline: item.BLOG_TITLE,
        url: absoluteUrl(`/oneBlog/${item.GUID}`),
        position: index + 1
      }))
    }
  }
})
</script>

<style scoped>
.series-page {
  min-height: 100%;
  padding: 28px 16px 48px;
}

.series-paper {
  max-width: 880px;
  margin: 0 auto;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
  padding: 28px 36px 36px;
}

.series-crumb {
  font-size: 13px;
  color: var(--j-muted);
  margin-bottom: 18px;
}

.series-crumb a,
.series-meta a,
.series-empty a {
  color: var(--j-pen);
  text-decoration: none;
}

.series-head {
  position: relative;
  padding-bottom: 18px;
  border-bottom: 1px dashed var(--j-rule-strong);
}

.series-stamp {
  display: inline-block;
  padding: 0 8px;
  border: 1.5px solid var(--j-stamp);
  color: var(--j-stamp);
  font-size: 13px;
  transform: rotate(-4deg);
}

.series-title {
  margin: 8px 0 6px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 32px;
  line-height: 1.35;
  color: var(--j-ink);
}

.series-meta {
  margin: 0 0 10px;
  font-size: 14px;
  color: var(--j-muted);
}

.series-desc {
  margin: 0 0 14px;
  line-height: 1.8;
  color: var(--j-ink-soft);
  white-space: pre-line;
}

.series-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px 16px;
}

.series-start {
  text-decoration: none;
}

.tool-link {
  border: none;
  background: none;
  padding: 0;
  color: var(--j-pen);
  font-size: 13px;
  cursor: pointer;
}

.tool-link.danger {
  color: var(--j-stamp);
}

.tool-link:disabled {
  color: var(--j-muted);
  cursor: default;
}

.sort-tip {
  margin: 14px 0 0;
  padding: 8px 12px;
  background: var(--j-note);
  font-size: 13px;
  color: var(--j-ink-soft);
}

.series-empty {
  padding: 28px 0;
  color: var(--j-muted);
}

.series-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.series-item {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr) auto;
  gap: 14px;
  padding: 18px 0;
  border-bottom: 1px dashed var(--j-rule);
}

.series-no {
  font-family: var(--j-num);
  font-size: 26px;
  line-height: 1.2;
  color: var(--j-rule-strong);
  text-align: right;
}

.series-item-title {
  margin: 0 0 6px;
  font-size: 18px;
  line-height: 1.5;
}

.series-item-title a {
  color: var(--j-ink);
  text-decoration: none;
}

.series-item-title a:hover {
  color: var(--j-pen);
  text-decoration: underline;
}

.series-private {
  margin-left: 8px;
  font-size: 12px;
  font-weight: normal;
  color: var(--j-muted);
}

.series-excerpt {
  margin: 0 0 6px;
  font-size: 14px;
  line-height: 1.75;
  color: var(--j-ink-soft);
  overflow-wrap: anywhere;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.series-date {
  font-size: 12px;
  color: var(--j-muted);
}

.sort-buttons {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.sort-buttons button {
  width: 32px;
  height: 28px;
  border: 1px solid var(--j-rule-strong);
  background: var(--j-paper-warm);
  color: var(--j-ink);
  cursor: pointer;
}

.sort-buttons button:disabled {
  color: var(--j-rule-strong);
  cursor: default;
}

@media (max-width: 640px) {
  .series-page {
    padding: 14px 10px 32px;
  }

  .series-paper {
    padding: 20px 16px 24px;
  }

  .series-title {
    font-size: 26px;
  }

  .series-item {
    grid-template-columns: 30px minmax(0, 1fr) auto;
    gap: 10px;
  }

  .series-no {
    font-size: 20px;
  }
}
</style>
