<!-- 后台“评论与留言”：全站的文章评论、便签墙随记评论和留言板留言 -->
<template>
  <div class="admin-comments">
    <div class="filters">
      <el-select v-model="query.source" size="small" style="width: 150px" @change="reload">
        <el-option label="全部评论与留言" value="all"/>
        <el-option label="文章评论" value="blog"/>
        <el-option label="随记评论" value="community"/>
        <el-option label="留言板" value="guestbook"/>
      </el-select>
      <el-input
          v-model="query.keyword"
          size="small"
          clearable
          placeholder="搜索内容或昵称"
          style="width: 220px"
          @keyup.enter="reload"
          @clear="reload"
      />
      <el-checkbox v-model="query.anonymousOnly" @change="reload">只看匿名评论</el-checkbox>
      <el-button size="small" type="primary" @click="reload">查询</el-button>
    </div>

    <el-table v-loading="loading" :data="rows" border size="small" class="admin-table" empty-text="没有符合条件的评论或留言">
      <el-table-column label="时间" width="150">
        <template #default="{ row }">{{ pubFormatDate(row.CREATE_TIME) }}</template>
      </el-table-column>
      <el-table-column label="来源" width="80">
        <template #default="{ row }">
          <el-tag size="small" :type="sourceTag[row.SOURCE]?.type">{{ sourceTag[row.SOURCE]?.label || row.SOURCE }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="所在内容" min-width="180">
        <template #default="{ row }">
          <a :href="targetHref(row)" target="_blank" rel="noopener" class="target-link">{{ row.TARGETTITLE || '（已删除或无标题）' }}</a>
        </template>
      </el-table-column>
      <el-table-column label="评论者" min-width="140">
        <template #default="{ row }">
          <div>{{ row.USERNAME || '匿名' }}
            <el-tag v-if="row.ISTOP === '1'" size="small" type="danger">置顶</el-tag>
            <el-tag v-if="row.ISANONYMOUS" size="small" type="info">匿名</el-tag>
            <el-tag v-if="row.ISREPLY" size="small" type="warning">回复</el-tag>
          </div>
          <div v-if="row.USEREMAIL" class="sub">{{ row.USEREMAIL }}</div>
          <div v-if="row.USERWEBSITE" class="sub">{{ row.USERWEBSITE }}</div>
          <div v-if="row.IP" class="sub">IP {{ row.IP }}</div>
        </template>
      </el-table-column>
      <el-table-column label="内容" min-width="260">
        <template #default="{ row }"><div class="comment-text">{{ row.TEXT }}</div></template>
      </el-table-column>
      <el-table-column label="操作" width="190" fixed="right">
        <template #default="{ row }">
          <div class="row-actions">
            <el-button link type="danger" size="small" @click="remove(row)">删除</el-button>
            <el-button v-if="row.SOURCE === 'guestbook' && !row.ISREPLY" link type="primary" size="small" @click="toggleTop(row)">
              {{ row.ISTOP === '1' ? '取消置顶' : '置顶' }}
            </el-button>
            <el-dropdown v-if="banOptions(row).length" trigger="click" @command="(option) => openBan(option)">
              <el-button link type="warning" size="small">禁止评论</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item v-for="option in banOptions(row)" :key="option.type" :command="option">{{ option.label }}</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
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

    <CommentBanDialog v-model="banVisible" :preset-type="banPreset.type" :preset-value="banPreset.value" lock-type/>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { ele_confirm, pubFormatDate, sendAxiosRequestChecked } from '@/utils/common.js'
import CommentBanDialog from './CommentBanDialog.vue'
import { websiteHost } from './commentBanTypes.js'

const query = reactive({ page: 1, pageSize: 20, keyword: '', source: 'all', anonymousOnly: false })
const rows = ref([])
const total = ref(0)
const loading = ref(false)

const sourceTag = {
  blog: { label: '文章', type: 'primary' },
  community: { label: '随记', type: 'success' },
  guestbook: { label: '留言', type: 'warning' }
}

const targetHref = (row) => ({
  blog: `/oneBlog/${encodeURIComponent(row.TARGETID)}`,
  community: `/note/${encodeURIComponent(row.TARGETID)}`,
  guestbook: '/ynsStudy/Notes?tab=guestbook'
})[row.SOURCE] || '/ynsStudy/Notes'

const load = async () => {
  loading.value = true
  try {
    const result = await sendAxiosRequestChecked('/blog-api/admin/getComments', { ...query }, '获取评论失败')
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

// 根据评论里有的信息给出可选的限制方式
const banVisible = ref(false)
const banPreset = reactive({ type: 'ip', value: '' })
const banOptions = (row) => {
  const options = []
  if (row.USERCODE) options.push({ type: 'user', value: row.USERCODE, label: `禁止该用户（${row.USERNAME}）` })
  if (row.IP) options.push({ type: 'ip', value: row.IP, label: `禁止该 IP（${row.IP}）` })
  if (row.USERWEBSITE) options.push({ type: 'website', value: websiteHost(row.USERWEBSITE), label: `禁止该网址（${websiteHost(row.USERWEBSITE)}）` })
  if (row.USEREMAIL) options.push({ type: 'email', value: row.USEREMAIL, label: `禁止该邮箱（${row.USEREMAIL}）` })
  return options
}
const openBan = (option) => {
  banPreset.type = option.type
  banPreset.value = option.value
  banVisible.value = true
}

// 留言板的留言可以置顶（只有超级管理员能进后台）
const toggleTop = async (row) => {
  const isTop = row.ISTOP === '1' ? '0' : '1'
  const result = await sendAxiosRequestChecked('/blog-api/guestbook/setTopMessage', { guid: row.GUID, isTop }, '置顶操作失败')
  if (!result) return
  row.ISTOP = isTop === '1' ? '1' : null
  ElMessage.success(isTop === '1' ? '已置顶' : '已取消置顶')
}

const remove = (row) => {
  const tip = row.ISREPLY ? '确定删除这条回复吗？' : '确定删除这条评论吗？它下面的回复会一起删除。'
  ele_confirm(tip, async () => {
    const result = await sendAxiosRequestChecked('/blog-api/admin/deleteComment', { source: row.SOURCE, guid: row.GUID }, '删除失败')
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

/* 下拉菜单是 inline-flex，和普通按钮放一起基线不同，用 flex 统一垂直居中 */
.row-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.row-actions .el-button + .el-button {
  margin-left: 0;
}

.target-link {
  color: var(--j-pen);
  text-decoration: none;
  overflow-wrap: anywhere;
}

.sub {
  font-size: 12px;
  color: var(--j-muted);
  overflow-wrap: anywhere;
}

.comment-text {
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  line-height: 1.6;
}

.pager {
  margin-top: 14px;
  justify-content: flex-end;
}
</style>
