<!-- 后台：关于页的“站长推荐”。挑选几篇公开文章和几个系列，按这里的顺序显示在关于页 -->
<template>
  <div v-loading="loading" class="admin-featured">
    <el-alert
        v-if="loaded && !ready"
        type="warning"
        :closable="false"
        show-icon
        title="网站设置还没有启用：需要先在服务器执行数据库迁移 017，并重启 Go 服务。"
        class="alert"
    />
    <p class="hint">
      推荐的内容会显示在<a href="/ynsStudy/About" target="_blank" rel="noopener">关于页</a>“站长想说”的下面，给第一次来的访客一个“先读什么”的建议。
      只能推荐公开文章和有公开文章的系列，最多 {{ maxArticles }} 篇文章、{{ maxSeries }} 个系列；之后文章改为私密或被删除，会自动不再显示。
    </p>
    <el-alert
        v-if="hiddenArticles > 0 || hiddenSeries > 0"
        type="info"
        :closable="false"
        show-icon
        :title="`之前推荐的内容里有 ${hiddenArticles} 篇文章、${hiddenSeries} 个系列已经不公开或被删除，下面没有列出，保存后会一并清掉。`"
        class="alert"
    />

    <section class="block">
      <h4 class="sub">推荐文章 <span class="count">{{ articles.length }} / {{ maxArticles }}</span></h4>
      <el-select
          v-model="articlePicker"
          filterable
          remote
          clearable
          :remote-method="searchArticles"
          :loading="searching"
          :disabled="articles.length >= maxArticles"
          :placeholder="articles.length >= maxArticles ? `最多推荐 ${maxArticles} 篇` : '输入标题关键词搜索公开文章'"
          class="picker"
          @change="addArticle"
      >
        <el-option
            v-for="item in articleOptions"
            :key="item.GUID"
            :value="item.GUID"
            :label="item.BLOG_TITLE"
            :disabled="articles.some(a => a.GUID === item.GUID)"
        >
          <span class="option-title">{{ item.BLOG_TITLE }}</span>
          <span class="option-meta">{{ item.USERNAME }}</span>
        </el-option>
      </el-select>
      <ol v-if="articles.length" class="picked">
        <li v-for="(item, index) in articles" :key="item.GUID">
          <span class="picked-no">{{ index + 1 }}</span>
          <a :href="`/oneBlog/${item.GUID}`" target="_blank" rel="noopener" class="picked-title">{{ item.BLOG_TITLE }}</a>
          <span class="picked-meta">{{ item.USERNAME }}</span>
          <span class="picked-actions">
            <button type="button" :disabled="index === 0" title="上移" @click="move(articles, index, -1)">↑</button>
            <button type="button" :disabled="index === articles.length - 1" title="下移" @click="move(articles, index, 1)">↓</button>
            <button type="button" class="danger" title="移除" @click="articles.splice(index, 1)">✕</button>
          </span>
        </li>
      </ol>
      <p v-else class="empty">还没有推荐文章</p>
    </section>

    <section class="block">
      <h4 class="sub">推荐系列 <span class="count">{{ series.length }} / {{ maxSeries }}</span></h4>
      <el-select
          v-model="seriesPicker"
          filterable
          clearable
          :disabled="series.length >= maxSeries"
          :placeholder="series.length >= maxSeries ? `最多推荐 ${maxSeries} 个` : '选择系列'"
          class="picker"
          @change="addSeries"
      >
        <el-option
            v-for="item in seriesOptions"
            :key="item.ID"
            :value="String(item.ID)"
            :label="`${item.TITLE}（${item.ARTICLE_COUNT} 篇 · ${item.USERNAME}）`"
            :disabled="series.some(s => String(s.ID) === String(item.ID))"
        />
      </el-select>
      <ol v-if="series.length" class="picked">
        <li v-for="(item, index) in series" :key="item.ID">
          <span class="picked-no">{{ index + 1 }}</span>
          <a :href="`/series/${item.ID}`" target="_blank" rel="noopener" class="picked-title">{{ item.TITLE }}</a>
          <span class="picked-meta">{{ item.ARTICLE_COUNT }} 篇</span>
          <span class="picked-actions">
            <button type="button" :disabled="index === 0" title="上移" @click="move(series, index, -1)">↑</button>
            <button type="button" :disabled="index === series.length - 1" title="下移" @click="move(series, index, 1)">↓</button>
            <button type="button" class="danger" title="移除" @click="series.splice(index, 1)">✕</button>
          </span>
        </li>
      </ol>
      <p v-else class="empty">还没有推荐系列</p>
    </section>

    <div class="actions">
      <el-button type="primary" :loading="saving" :disabled="!ready" @click="save">保存</el-button>
      <el-button @click="load">撤销修改</el-button>
      <a href="/ynsStudy/About" target="_blank" rel="noopener" class="preview-link">去关于页看看 →</a>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { sendAxiosRequest, sendAxiosRequestChecked } from '@/utils/common.js'

const loading = ref(false)
const loaded = ref(false)
const saving = ref(false)
const ready = ref(true)
const maxArticles = ref(6)
const maxSeries = ref(3)
const hiddenArticles = ref(0)
const hiddenSeries = ref(0)

const articles = ref([])
const series = ref([])

async function load() {
  loading.value = true
  const result = await sendAxiosRequestChecked('/blog-api/admin/getAboutFeatured', {}, '读取站长推荐失败')
  loading.value = false
  loaded.value = true
  if (!result?.result) return
  const data = result.result
  ready.value = !!data.ready
  maxArticles.value = data.maxArticles || 6
  maxSeries.value = data.maxSeries || 3
  hiddenArticles.value = Math.max(0, data.hiddenArticles || 0)
  hiddenSeries.value = Math.max(0, data.hiddenSeries || 0)
  articles.value = (data.articles || []).map(({GUID, BLOG_TITLE, USERNAME}) => ({GUID, BLOG_TITLE, USERNAME}))
  series.value = (data.series || []).map(({ID, TITLE, ARTICLE_COUNT}) => ({ID, TITLE, ARTICLE_COUNT}))
}

// 搜索公开文章（首页的文章接口只返回公开文章）
const articlePicker = ref('')
const articleOptions = ref([])
const searching = ref(false)
let searchId = 0
async function searchArticles(keyword) {
  const current = ++searchId
  searching.value = true
  const result = await sendAxiosRequest('/blog-api/blog/getAllBlog', {page: 1, pageSize: 20, keyword: keyword || ''}).catch(() => null)
  if (current !== searchId) return
  searching.value = false
  articleOptions.value = result && !result.isError ? (result.result?.data || []) : []
}

function addArticle(guid) {
  const item = articleOptions.value.find(option => option.GUID === guid)
  articlePicker.value = ''
  if (!item || articles.value.some(a => a.GUID === guid) || articles.value.length >= maxArticles.value) return
  articles.value.push({GUID: item.GUID, BLOG_TITLE: item.BLOG_TITLE, USERNAME: item.USERNAME})
}

// 可选的系列：至少有一篇公开文章的系列
const seriesPicker = ref('')
const seriesOptions = ref([])
async function loadSeriesOptions() {
  const result = await sendAxiosRequest('/blog-api/series/getSeriesList', {sort: 'hot', limit: 200}).catch(() => null)
  seriesOptions.value = result && !result.isError && Array.isArray(result.result) ? result.result : []
}

function addSeries(id) {
  const item = seriesOptions.value.find(option => String(option.ID) === String(id))
  seriesPicker.value = ''
  if (!item || series.value.some(s => String(s.ID) === String(id)) || series.value.length >= maxSeries.value) return
  series.value.push({ID: item.ID, TITLE: item.TITLE, ARTICLE_COUNT: item.ARTICLE_COUNT})
}

function move(list, index, step) {
  const target = index + step
  if (target < 0 || target >= list.length) return
  ;[list[index], list[target]] = [list[target], list[index]]
}

async function save() {
  saving.value = true
  const result = await sendAxiosRequestChecked('/blog-api/admin/setAboutFeatured', {
    articles: articles.value.map(item => item.GUID),
    series: series.value.map(item => String(item.ID))
  }, '保存失败')
  saving.value = false
  if (!result) return
  hiddenArticles.value = 0
  hiddenSeries.value = 0
  ElMessage.success('已保存，关于页会按这个顺序显示')
}

onMounted(() => {
  load()
  searchArticles('')
  loadSeriesOptions()
})
</script>

<style scoped>
.alert {
  margin-bottom: 12px;
}

.hint {
  margin: 0 0 16px;
  color: var(--j-ink-soft);
  font-size: 13px;
  line-height: 1.8;
}

.hint a,
.preview-link {
  color: var(--j-pen);
  text-decoration: none;
}

.block + .block {
  margin-top: 24px;
}

.sub {
  margin: 0 0 10px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 16px;
}

.count {
  margin-left: 6px;
  font-family: var(--j-sans);
  font-size: 12px;
  color: var(--j-muted);
}

.picker {
  width: 420px;
  max-width: 100%;
}

.option-title {
  margin-right: 10px;
}

.option-meta {
  float: right;
  font-size: 12px;
  color: var(--j-muted);
}

.picked {
  list-style: none;
  margin: 10px 0 0;
  padding: 0;
  max-width: 640px;
}

.picked li {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-bottom: 1px dashed var(--j-rule);
  font-size: 14px;
}

.picked-no {
  flex-shrink: 0;
  width: 1.5em;
  font-family: var(--j-num);
  color: var(--j-stamp);
}

.picked-title {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--j-ink);
  text-decoration: none;
}

.picked-title:hover {
  color: var(--j-pen);
}

.picked-meta {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--j-muted);
}

.picked-actions {
  flex-shrink: 0;
  display: flex;
  gap: 4px;
}

.picked-actions button {
  width: 26px;
  height: 24px;
  border: 1px solid var(--j-rule-strong);
  background: var(--j-paper-warm);
  color: var(--j-ink);
  cursor: pointer;
}

.picked-actions button:disabled {
  color: var(--j-rule-strong);
  cursor: default;
}

.picked-actions button.danger {
  color: var(--j-stamp);
}

.empty {
  margin: 10px 0 0;
  font-size: 13px;
  color: var(--j-muted);
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 24px;
}

.preview-link {
  font-size: 13px;
}
</style>
