<!-- 后台：管理员操作日志 -->
<template>
  <div class="admin-logs">
    <div class="filters">
      <el-select v-model="action" size="small" style="width: 160px" placeholder="全部操作" @change="reload">
        <el-option label="全部操作" value=""/>
        <el-option v-for="(label, key) in ACTION_LABELS" :key="key" :label="label" :value="key"/>
      </el-select>
      <el-button size="small" @click="load">刷新</el-button>
    </div>

    <el-alert v-if="errorMessage" type="warning" :closable="false" show-icon :title="errorMessage" class="alert"/>

    <el-table v-loading="loading" :data="rows" border size="small" class="admin-table" empty-text="还没有操作记录">
      <el-table-column label="时间" width="160">
        <template #default="{ row }">{{ pubFormatDate(row.CREATE_TIME) }}</template>
      </el-table-column>
      <el-table-column label="操作人" width="120">
        <template #default="{ row }">{{ row.ACTOR_NAME || row.ACTOR_CODE }}</template>
      </el-table-column>
      <el-table-column label="操作" width="110">
        <template #default="{ row }"><el-tag size="small">{{ ACTION_LABELS[row.ACTION] || row.ACTION }}</el-tag></template>
      </el-table-column>
      <el-table-column label="对象" min-width="160">
        <template #default="{ row }"><span class="wrap">{{ row.TARGET }}</span></template>
      </el-table-column>
      <el-table-column label="详情" min-width="220">
        <template #default="{ row }"><span class="wrap">{{ row.DETAIL }}</span></template>
      </el-table-column>
      <el-table-column prop="IP" label="IP" width="130"/>
    </el-table>

    <el-pagination
        class="pager"
        layout="total, prev, pager, next"
        :total="total"
        :page-size="pageSize"
        v-model:current-page="page"
        @current-change="load"
    />
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { pubFormatDate, sendAxiosRequest } from '@/utils/common.js'

// 与 Go 端 admin_log_service.go 中的 AdminAction* 常量对应
const ACTION_LABELS = {
  setAdmin: '设为管理员',
  removeAdmin: '取消管理员',
  ban: '封禁用户',
  removeBan: '取消封禁',
  forceLogout: '强制下线',
  resetPassword: '重置密码',
  deleteComment: '删除评论',
  deleteArticle: '删除文章',
  articleType: '文章可见性',
  deletePost: '删除随记',
  topPost: '随记置顶',
  announcement: '公告',
  deleteFiles: '清理文件',
  bannedWord: '违禁词',
  indexNow: '推送 IndexNow',
  editArticle: '修改文章',
  deleteResource: '删除资源',
  editResource: '修改资源',
  friendLink: '友链',
  commentBan: '评论限制',
  editPost: '修改随记',
  editSeries: '管理系列',
  aboutFeatured: '关于页推荐'
}

const action = ref('')
const page = ref(1)
const pageSize = 20
const rows = ref([])
const total = ref(0)
const loading = ref(false)
const errorMessage = ref('')

const load = async () => {
  loading.value = true
  try {
    const result = await sendAxiosRequest('/blog-api/admin/getLogs', { page: page.value, pageSize, action: action.value })
    if (result && !result.isError) {
      errorMessage.value = ''
      rows.value = result.result.data || []
      total.value = result.result.total || 0
    } else {
      // 常见原因是还没执行迁移 007，直接显示后台给出的提示
      errorMessage.value = result?.errMsg || '获取操作日志失败'
      rows.value = []
      total.value = 0
    }
  } finally {
    loading.value = false
  }
}

const reload = () => {
  page.value = 1
  load()
}

onMounted(load)
</script>

<style scoped>
.filters {
  display: flex;
  gap: 10px;
  margin-bottom: 14px;
}

.alert {
  margin-bottom: 12px;
}

.wrap {
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.pager {
  margin-top: 14px;
  justify-content: flex-end;
}
</style>
