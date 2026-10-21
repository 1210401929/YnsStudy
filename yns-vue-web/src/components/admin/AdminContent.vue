<!-- 后台：内容管理（全站文章含私密文章、便签墙随记） -->
<template>
  <div class="admin-content">
    <div class="filters">
      <el-radio-group v-model="kind" size="small" @change="switchKind">
        <el-radio-button value="articles">文章</el-radio-button>
        <el-radio-button value="posts">便签墙随记</el-radio-button>
      </el-radio-group>
      <el-select v-model="query.type" size="small" style="width: 120px" :placeholder="kind === 'articles' ? '全部文章' : '全部帖子'" @change="reload">
        <template v-if="kind === 'articles'">
          <el-option label="全部文章" value=""/>
          <el-option label="公开" value="public"/>
          <el-option label="私密" value="privacy"/>
        </template>
        <template v-else>
          <el-option label="全部帖子" value=""/>
          <el-option label="只看置顶" value="top"/>
        </template>
      </el-select>
      <el-input
          v-model="query.keyword"
          size="small"
          clearable
          :placeholder="kind === 'articles' ? '搜索标题或作者' : '搜索标题、内容或作者'"
          style="width: 220px"
          @keyup.enter="reload"
          @clear="reload"
      />
      <el-button size="small" type="primary" @click="reload">查询</el-button>
    </div>

    <!-- 文章 -->
    <el-table v-if="kind === 'articles'" v-loading="loading" :data="rows" border size="small" class="admin-table" empty-text="没有符合条件的文章">
      <el-table-column label="标题" min-width="240">
        <template #default="{ row }">
          <a :href="`/oneBlog/${encodeURIComponent(row.GUID)}`" target="_blank" rel="noopener" class="title-link">{{ row.BLOG_TITLE || '未命名文章' }}</a>
        </template>
      </el-table-column>
      <el-table-column prop="USERNAME" label="作者" min-width="100"/>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag size="small" :type="row.BLOG_TYPE === 'public' ? 'success' : 'info'">{{ row.BLOG_TYPE === 'public' ? '公开' : '私密' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="VIEW_PAGE" label="阅读" width="80"/>
      <el-table-column label="发布时间" width="160">
        <template #default="{ row }">{{ pubFormatDate(row.CREATE_TIME) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="toggleArticle(row)">{{ row.BLOG_TYPE === 'public' ? '设为私密' : '设为公开' }}</el-button>
          <el-button link type="danger" size="small" @click="removeArticle(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 便签墙随记 -->
    <el-table v-else v-loading="loading" :data="rows" border size="small" class="admin-table" empty-text="没有符合条件的随记">
      <el-table-column label="内容" min-width="260">
        <template #default="{ row }">
          <div class="post-title">
            <el-tag v-if="row.ISTOP === '1'" size="small" type="warning">置顶</el-tag>
            <a :href="`/note/${encodeURIComponent(row.GUID)}`" target="_blank" rel="noopener" class="title-link">{{ row.TITLE || '查看随记' }}</a>
          </div>
          <div class="post-text">{{ row.TEXT }}</div>
        </template>
      </el-table-column>
      <el-table-column prop="USERNAME" label="作者" min-width="100"/>
      <el-table-column prop="COMMENT_COUNT" label="评论" width="70"/>
      <el-table-column label="发布时间" width="160">
        <template #default="{ row }">{{ pubFormatDate(row.CREATE_TIME) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="togglePostTop(row)">{{ row.ISTOP === '1' ? '取消置顶' : '置顶' }}</el-button>
          <el-button link type="danger" size="small" @click="removePost(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
        class="pager"
        layout="total, prev, pager, next"
        :total="total"
        :page-size="query.pageSize"
        v-model:current-page="query.page"
        @current-change="load"
    />
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { ele_confirm, pubFormatDate, sendAxiosRequestChecked } from '@/utils/common.js'

const kind = ref('articles')
const query = reactive({ page: 1, pageSize: 20, keyword: '', type: '' })
const rows = ref([])
const total = ref(0)
const loading = ref(false)

const load = async () => {
  loading.value = true
  try {
    const url = kind.value === 'articles' ? '/blog-api/admin/getArticles' : '/blog-api/admin/getPosts'
    const result = await sendAxiosRequestChecked(url, { ...query }, '获取列表失败')
    if (result) {
      rows.value = result.result.data || []
      total.value = result.result.total || 0
    }
  } finally {
    loading.value = false
  }
}

const reload = () => {
  query.page = 1
  load()
}

// 切换文章/帖子时清空筛选，两者的类型选项不同
const switchKind = () => {
  query.type = ''
  query.keyword = ''
  rows.value = []
  reload()
}

const toggleArticle = (row) => {
  const next = row.BLOG_TYPE === 'public' ? 'privacy' : 'public'
  const tip = next === 'privacy'
      ? `确定把《${row.BLOG_TITLE}》设为私密吗？设为私密后只有作者本人能看到，搜索引擎也会移除这篇文章。`
      : `确定把《${row.BLOG_TITLE}》设为公开吗？`
  ele_confirm(tip, async () => {
    const result = await sendAxiosRequestChecked('/blog-api/admin/setArticleType', { guid: row.GUID, type: next }, '修改失败')
    if (!result) return
    row.BLOG_TYPE = next
    ElMessage.success(result.result || '已修改')
  })
}

const removeArticle = (row) => {
  ele_confirm(`确定删除《${row.BLOG_TITLE}》吗？文章的评论、点赞收藏和正文图片会一起删除，无法恢复。`, async () => {
    const result = await sendAxiosRequestChecked('/blog-api/admin/deleteArticle', { guid: row.GUID }, '删除失败')
    if (!result) return
    ElMessage.success('已删除')
    load()
  })
}

const togglePostTop = (row) => {
  const next = row.ISTOP === '1' ? '0' : '1'
  ele_confirm(next === '1' ? '确定置顶这条帖子吗？' : '确定取消置顶吗？', async () => {
    const result = await sendAxiosRequestChecked('/blog-api/admin/setPostTop', { guid: row.GUID, isTop: next }, '操作失败')
    if (!result) return
    ElMessage.success(next === '1' ? '已置顶' : '已取消置顶')
    load()
  })
}

const removePost = (row) => {
  ele_confirm('确定删除这条帖子吗？帖子下的评论会一起删除，无法恢复。', async () => {
    const result = await sendAxiosRequestChecked('/blog-api/admin/deletePost', { guid: row.GUID }, '删除失败')
    if (!result) return
    ElMessage.success('已删除')
    load()
  })
}

onMounted(load)
</script>

<style scoped>
.filters {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}

.title-link {
  color: var(--j-pen);
  text-decoration: none;
  overflow-wrap: anywhere;
}

.post-title {
  font-weight: 600;
  overflow-wrap: anywhere;
}

.post-text {
  margin-top: 4px;
  color: var(--j-ink-soft);
  font-size: 12px;
  line-height: 1.6;
  overflow-wrap: anywhere;
}

.pager {
  margin-top: 14px;
  justify-content: flex-end;
}
</style>
