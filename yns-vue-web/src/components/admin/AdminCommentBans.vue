<!-- 后台：评论限制规则 -->
<template>
  <div class="admin-comment-bans">
    <p class="hint">
      命中任意一条规则的人无法发表文章评论和随记评论，只会看到“你已被禁止在本站评论”。超级管理员不受限制。
      也可以在“评论与留言”中直接对某条评论或留言的作者、IP、网址或邮箱禁止评论。
    </p>
    <div class="toolbar">
      <el-button size="small" type="danger" @click="openDialog">新增限制</el-button>
      <el-button size="small" @click="load">刷新</el-button>
    </div>

    <el-alert v-if="errorMessage" type="warning" :closable="false" show-icon :title="errorMessage" class="alert"/>

    <el-table v-loading="loading" :data="rows" border size="small" class="admin-table" empty-text="还没有评论限制">
      <el-table-column label="方式" width="90">
        <template #default="{ row }"><el-tag size="small">{{ BAN_TYPES[row.TYPE]?.label || row.TYPE }}</el-tag></template>
      </el-table-column>
      <el-table-column label="内容" min-width="180">
        <template #default="{ row }"><span class="wrap">{{ row.VALUE }}</span></template>
      </el-table-column>
      <el-table-column label="期限" width="170">
        <template #default="{ row }">
          <el-tag v-if="row.EXPIRED == 1" size="small" type="info">已过期</el-tag>
          <span v-else>{{ row.EXPIRE_TIME ? `至 ${pubFormatDate(row.EXPIRE_TIME)}` : '永久' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="备注" min-width="140">
        <template #default="{ row }"><span class="wrap">{{ row.REMARK }}</span></template>
      </el-table-column>
      <el-table-column label="添加时间" width="160">
        <template #default="{ row }">{{ pubFormatDate(row.CREATE_TIME) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="80" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="remove(row)">解除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <CommentBanDialog v-model="dialogVisible" preset-type="ip" @saved="load"/>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { ele_confirm, pubFormatDate, sendAxiosRequest, sendAxiosRequestChecked } from '@/utils/common.js'
import CommentBanDialog from './CommentBanDialog.vue'
import { BAN_TYPES } from './commentBanTypes.js'

const rows = ref([])
const loading = ref(false)
const errorMessage = ref('')
const dialogVisible = ref(false)

const load = async () => {
  loading.value = true
  try {
    const result = await sendAxiosRequest('/blog-api/admin/getCommentBans')
    if (result && !result.isError) {
      errorMessage.value = ''
      rows.value = result.result || []
    } else {
      errorMessage.value = result?.errMsg || '读取评论限制失败'
      rows.value = []
    }
  } finally {
    loading.value = false
  }
}

const openDialog = () => { dialogVisible.value = true }

const remove = (row) => {
  ele_confirm(`确定解除对“${row.VALUE}”的评论限制吗？`, async () => {
    const result = await sendAxiosRequestChecked('/blog-api/admin/deleteCommentBan', { id: row.ID }, '解除失败')
    if (!result) return
    ElMessage.success('已解除')
    load()
  })
}

onMounted(load)
</script>

<style scoped>
.hint {
  margin: 0 0 12px;
  color: var(--j-ink-soft);
  font-size: 13px;
  line-height: 1.8;
}

.toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 14px;
}

.alert {
  margin-bottom: 12px;
}

.wrap {
  overflow-wrap: anywhere;
}
</style>
