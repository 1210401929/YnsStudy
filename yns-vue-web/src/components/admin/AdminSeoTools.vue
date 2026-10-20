<!-- 后台：SEO 小工具（sitemap 概况、手动推送 IndexNow） -->
<template>
  <div v-loading="loading" class="admin-seo">
    <section class="block">
      <h4 class="sub">sitemap 概况</h4>
      <div class="stats">
        <span>文章 <b>{{ show(status.sitemap?.articles) }}</b></span>
        <span>作者主页 <b>{{ show(status.sitemap?.authors) }}</b></span>
        <span>栏目页 <b>{{ show(status.sitemap?.pages) }}</b></span>
        <span>合计 <b>{{ sitemapTotal }}</b></span>
      </div>
      <div class="links">
        <a v-for="link in status.links || []" :key="link" :href="link" target="_blank" rel="noopener">{{ link }}</a>
      </div>
    </section>

    <section class="block">
      <h4 class="sub">手动推送给 Bing（IndexNow）</h4>
      <p class="hint">
        发布、修改、删除公开文章时已经会自动推送。修改了首页、栏目页或作者资料后，可以在这里手动推送，让 Bing 尽快重新抓取。
        每行一个地址，可以写站内路径（如 <code>/user/101</code>）或本站完整地址，一次最多 100 个。
      </p>
      <el-alert
          v-if="status.indexNowEnabled === false"
          type="warning"
          :closable="false"
          show-icon
          title="服务器 config.yaml 没有配置 seo.indexnow_key，无法推送。"
          class="alert"
      />
      <el-input v-model="urlText" type="textarea" :rows="5" placeholder="/&#10;/archive&#10;/user/101"/>
      <div class="actions">
        <el-button size="small" @click="fillCommon">填入常用地址</el-button>
        <el-button size="small" type="primary" :disabled="!status.indexNowEnabled || !urlText.trim()" :loading="submitting" @click="submit">推送</el-button>
      </div>
      <p v-if="lastResult" class="result">{{ lastResult }}</p>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { sendAxiosRequestChecked } from '@/utils/common.js'
import { SITE_OWNER } from '@/utils/seo.js'

const loading = ref(false)
const submitting = ref(false)
const status = ref({})
const urlText = ref('')
const lastResult = ref('')

const show = (value) => (value === null || value === undefined ? '无数据' : value)
const sitemapTotal = computed(() => {
  const s = status.value.sitemap || {}
  if ([s.articles, s.authors, s.pages].some(v => v === null || v === undefined)) return '无数据'
  return s.articles + s.authors + s.pages
})

const load = async () => {
  loading.value = true
  try {
    const result = await sendAxiosRequestChecked('/blog-api/admin/getSeoStatus', {}, '获取 SEO 状态失败')
    if (result) status.value = result.result
  } finally {
    loading.value = false
  }
}

const fillCommon = () => {
  urlText.value = ['/', '/archive', '/ynsStudy/Home', '/ynsStudy/About', SITE_OWNER.profilePath].join('\n')
}

const submit = async () => {
  const urls = urlText.value.split('\n').map(s => s.trim()).filter(Boolean)
  submitting.value = true
  try {
    const result = await sendAxiosRequestChecked('/blog-api/admin/submitIndexNow', { urls }, '推送失败')
    if (!result) return
    const submitted = result.result.submitted || []
    lastResult.value = `${result.result.time} 已推送 ${submitted.length} 个地址：${submitted.join('、')}`
    ElMessage.success('推送成功')
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.block + .block {
  margin-top: 24px;
}

.sub {
  margin: 0 0 10px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 16px;
}

.stats {
  display: flex;
  flex-wrap: wrap;
  gap: 18px;
  padding: 10px 12px;
  background: var(--j-paper-warm);
  font-size: 13px;
}

.links {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-top: 10px;
  font-size: 13px;
}

.links a {
  color: var(--j-pen);
  text-decoration: none;
}

.hint {
  margin: 0 0 10px;
  color: var(--j-ink-soft);
  font-size: 13px;
  line-height: 1.8;
}

.alert {
  margin-bottom: 10px;
}

.actions {
  display: flex;
  gap: 10px;
  margin-top: 10px;
}

.result {
  margin-top: 10px;
  color: var(--j-ink-soft);
  font-size: 13px;
  overflow-wrap: anywhere;
}
</style>
