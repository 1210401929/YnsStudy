<!-- 后台：数据概览 -->
<template>
  <div v-loading="loading" class="admin-dashboard">
    <div class="toolbar">
      <span class="hint">数据实时统计，点击刷新获取最新数字</span>
      <el-button size="small" @click="load">刷新</el-button>
    </div>

    <section class="tiles">
      <div v-for="tile in tiles" :key="tile.label" class="tile">
        <div class="tile-value">{{ show(tile.value) }}</div>
        <div class="tile-label">{{ tile.label }}</div>
        <div v-if="tile.extra" class="tile-extra">{{ tile.extra }}</div>
      </div>
    </section>

    <section class="block">
      <h3 class="block-title">近期新增</h3>
      <el-table :data="recentRows" border size="small" class="admin-table">
        <el-table-column prop="label" label="" width="110"/>
        <el-table-column label="文章"><template #default="{ row }">{{ show(row.articles) }}</template></el-table-column>
        <el-table-column label="随记"><template #default="{ row }">{{ show(row.communityPosts) }}</template></el-table-column>
        <el-table-column label="评论"><template #default="{ row }">{{ show(row.comments) }}</template></el-table-column>
        <el-table-column label="留言"><template #default="{ row }">{{ show(row.guestbook) }}</template></el-table-column>
        <el-table-column label="新用户"><template #default="{ row }">{{ show(row.users) }}</template></el-table-column>
      </el-table>
    </section>

    <div class="two-columns">
      <section class="block">
        <h3 class="block-title">阅读量最高的文章</h3>
        <ol v-if="topArticles.length" class="top-list">
          <li v-for="article in topArticles" :key="article.GUID">
            <a :href="`/oneBlog/${encodeURIComponent(article.GUID)}`" target="_blank" rel="noopener">{{ article.BLOG_TITLE || '未命名文章' }}</a>
            <span class="meta">{{ article.USERNAME }} · {{ article.VIEW_PAGE }} 次阅读</span>
          </li>
        </ol>
        <el-empty v-else description="暂无数据" :image-size="60"/>
      </section>

      <section class="block">
        <h3 class="block-title">最近登录</h3>
        <el-table v-if="recentLogins.length" :data="recentLogins" size="small" max-height="360" class="admin-table">
          <el-table-column prop="USERNAME" label="用户" min-width="100"/>
          <el-table-column prop="LOGINADDRESS" label="地区" min-width="90"/>
          <el-table-column prop="LOGINIP" label="IP" min-width="120"/>
          <el-table-column label="时间" min-width="150">
            <template #default="{ row }">{{ pubFormatDate(row.LOGIN_TIME) }}</template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="暂无登录记录（或登录记录表没有时间字段）" :image-size="60"/>
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { pubFormatDate, sendAxiosRequestChecked } from '@/utils/common.js'

const loading = ref(false)
const data = ref({ totals: {}, recent: {}, topArticles: [], recentLogins: [] })

// 统计失败或缺少字段时后台返回 null
const show = (value) => (value === null || value === undefined ? '无数据' : value)

const tiles = computed(() => {
  const t = data.value.totals || {}
  const comments = (t.blogComments ?? null) === null || (t.communityComments ?? null) === null
      ? null : t.blogComments + t.communityComments
  return [
    { label: '公开文章', value: t.publicArticles, extra: t.privateArticles != null ? `另有私密 ${t.privateArticles} 篇` : '' },
    { label: '总阅读量', value: t.totalViews },
    { label: '注册用户', value: t.users },
    { label: '随记', value: t.communityPosts },
    { label: '留言', value: t.guestbook },
    { label: '评论', value: comments, extra: t.blogComments != null ? `文章 ${t.blogComments} · 随记 ${t.communityComments}` : '' },
    { label: '累计登录', value: t.logins }
  ]
})

const recentRows = computed(() => {
  const r = data.value.recent || {}
  return [
    { label: '最近 7 天', ...(r.days7 || {}) },
    { label: '最近 30 天', ...(r.days30 || {}) }
  ]
})
const topArticles = computed(() => data.value.topArticles || [])
const recentLogins = computed(() => data.value.recentLogins || [])

const load = async () => {
  loading.value = true
  try {
    const result = await sendAxiosRequestChecked('/blog-api/admin/getDashboard', {}, '获取统计数据失败')
    if (result) data.value = result.result
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.hint {
  color: var(--j-muted);
  font-size: 13px;
}

.tiles {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 14px;
}

.tile {
  padding: 16px 18px;
  background: var(--j-paper-warm);
  border: 1px solid var(--j-rule);
}

.tile-value {
  font-family: var(--j-num);
  font-size: 28px;
  color: var(--j-ink);
}

.tile-label {
  margin-top: 4px;
  font-size: 13px;
  color: var(--j-ink-soft);
}

.tile-extra {
  margin-top: 4px;
  font-size: 12px;
  color: var(--j-muted);
}

.block {
  margin-top: 22px;
  min-width: 0;
}

.block-title {
  margin: 0 0 10px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 18px;
  color: var(--j-ink);
}

.two-columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.top-list {
  margin: 0;
  padding-left: 22px;
}

.top-list li {
  padding: 6px 0;
  border-bottom: 1px dashed var(--j-rule);
  line-height: 1.6;
}

.top-list a {
  color: var(--j-pen);
  text-decoration: none;
  overflow-wrap: anywhere;
}

.top-list .meta {
  display: block;
  font-size: 12px;
  color: var(--j-muted);
}

@media (max-width: 900px) {
  .two-columns {
    grid-template-columns: 1fr;
  }
}
</style>
