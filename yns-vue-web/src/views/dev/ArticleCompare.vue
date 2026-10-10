<!--
  开发环境专用：对比文章的两种展示方式（左：wangEditor 只读，右：直接渲染 HTML）
  只在 npm run dev 时注册路由，打包后不存在。访问 /dev/article-compare
-->
<template>
  <div class="compare-page">
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索文章标题" clearable style="width: 220px" @keyup.enter="loadList"/>
      <el-button @click="loadList">搜索</el-button>
      <el-select v-model="currentId" filterable placeholder="选择一篇文章" style="width: 360px" @change="pickArticle">
        <el-option v-for="item in articles" :key="item.GUID" :label="item.BLOG_TITLE" :value="item.GUID"/>
      </el-select>
      <el-button :disabled="!articles.length" @click="step(-1)">上一篇</el-button>
      <el-button :disabled="!articles.length" @click="step(1)">下一篇</el-button>
      <el-button @click="showPaste = !showPaste">{{ showPaste ? '收起' : '粘贴 HTML' }}</el-button>
      <span class="count">共 {{ articles.length }} 篇</span>
    </div>

    <el-input
        v-if="showPaste"
        v-model="html"
        type="textarea"
        :rows="6"
        placeholder="直接粘贴一段正文 HTML 对比"
        class="paste-box"
    />

    <div class="columns">
      <section class="column">
        <h3>wangEditor 只读（旧）</h3>
        <ArticleEditor :key="renderKey" :isReadOnly="true" :content="html"/>
      </section>
      <section class="column">
        <h3>直接渲染 HTML（新）</h3>
        <ArticleViewer :content="html"/>
      </section>
    </div>
  </div>
</template>

<script setup>
import {defineAsyncComponent, onMounted, ref, watch} from 'vue'
import {ElMessage} from 'element-plus'
import {sendAxiosRequest} from '@/utils/common.js'
import ArticleViewer from '@/components/detail/ArticleViewer.vue'

const ArticleEditor = defineAsyncComponent(() => import('@/components/detail/ArticleEditor.vue'))

const keyword = ref('')
const articles = ref([])
const currentId = ref('')
const html = ref('')
const renderKey = ref(0)
const showPaste = ref(false)

// 编辑器只在创建时读取内容，换文章时重新创建一次
watch(html, () => renderKey.value++)

async function loadList() {
  const res = await sendAxiosRequest('/blog-api/blog/getAllBlog', {page: 1, pageSize: 200, keyword: keyword.value})
  if (!res || res.isError) {
    ElMessage.error(res?.errMsg || '获取文章失败，确认本地后台已启动')
    return
  }
  articles.value = res.result?.data || []
  if (articles.value.length) pickArticle(articles.value[0].GUID)
}

async function pickArticle(guid) {
  currentId.value = guid
  const res = await sendAxiosRequest('/blog-api/blog/getBlog', {blogId: guid})
  html.value = res?.result?.[0]?.MAINTEXT || ''
}

function step(offset) {
  const index = articles.value.findIndex(item => item.GUID === currentId.value)
  const next = articles.value[(index + offset + articles.value.length) % articles.value.length]
  if (next) pickArticle(next.GUID)
}

onMounted(loadList)
</script>

<style scoped>
.compare-page {
  padding: 20px;
  background: var(--j-desk);
  min-height: 100vh;
  box-sizing: border-box;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  margin-bottom: 12px;
}

.count {
  color: var(--j-muted);
  font-size: 13px;
}

.paste-box {
  margin-bottom: 12px;
}

.columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.column {
  min-width: 0;
  padding: 16px 20px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
}

.column h3 {
  margin: 0 0 12px;
  font-family: var(--j-hand);
  font-weight: normal;
  color: var(--j-muted);
}
</style>
