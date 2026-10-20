<!-- 后台：清理未被引用的上传文件（先预览，确认后再删除） -->
<template>
  <div class="admin-files">
    <p class="hint">
      只会列出<strong>没有被任何内容引用</strong>、并且上传超过 {{ minAgeDays }} 天的文件。
      7 天内上传的文件可能属于正在编写的文章或浏览器里的草稿，一律跳过。
    </p>

    <div class="types">
      <el-radio-group v-model="fileType" size="small" @change="preview">
        <el-radio-button value="editorImage">文章图片</el-radio-button>
        <el-radio-button value="resourcesFile">资源文件</el-radio-button>
        <el-radio-button value="userAvatar">用户头像</el-radio-button>
      </el-radio-group>
      <el-button size="small" @click="preview">重新检查</el-button>
    </div>

    <div v-loading="loading">
      <div v-if="checked" class="summary">
        找到 <strong>{{ files.length }}</strong> 个可清理文件，共 {{ formatSize(totalSize) }}
        <span v-if="skippedRecent">；另有 {{ skippedRecent }} 个未引用文件因上传不足 {{ minAgeDays }} 天已跳过</span>
      </div>

      <el-table v-if="files.length" :data="files" border size="small" max-height="420" class="admin-table">
        <el-table-column label="文件" min-width="260">
          <template #default="{ row }">
            <a :href="`/uploadFile/${fileType}/${encodeURIComponent(row.name)}`" target="_blank" rel="noopener" class="file-link">{{ row.name }}</a>
          </template>
        </el-table-column>
        <el-table-column label="大小" width="110">
          <template #default="{ row }">{{ formatSize(row.size) }}</template>
        </el-table-column>
        <el-table-column prop="modTime" label="上传时间" width="170"/>
      </el-table>

      <div v-if="files.length" class="actions">
        <el-button type="danger" size="small" @click="confirmDelete">删除这 {{ files.length }} 个文件</el-button>
      </div>
      <el-empty v-else-if="checked" description="没有需要清理的文件" :image-size="60"/>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { ele_confirm, sendAxiosRequestChecked } from '@/utils/common.js'

const fileType = ref('editorImage')
const files = ref([])
const totalSize = ref(0)
const skippedRecent = ref(0)
const minAgeDays = ref(7)
const loading = ref(false)
const checked = ref(false)

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = bytes
  let index = 0
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024
    index++
  }
  return `${value.toFixed(index ? 1 : 0)} ${units[index]}`
}

const run = async (mode) => {
  loading.value = true
  try {
    const result = await sendAxiosRequestChecked('/blog-api/file/consistencyFileCheck', { type: fileType.value, mode }, '检查失败')
    if (!result) return null
    const data = result.result
    files.value = data.files || []
    totalSize.value = data.totalSize || 0
    skippedRecent.value = data.skippedRecent || 0
    minAgeDays.value = data.minAgeDays || 7
    checked.value = true
    return data
  } finally {
    loading.value = false
  }
}

const preview = () => run('preview')

const confirmDelete = () => {
  ele_confirm(`确定删除这 ${files.value.length} 个文件吗？删除后无法恢复。`, async () => {
    // 后台会重新计算待删除文件，不使用这里的列表
    const data = await run('delete')
    if (!data) return
    ElMessage.success(`已删除 ${(data.deleted || []).length} 个文件`)
    preview()
  })
}

onMounted(preview)
</script>

<style scoped>
.hint {
  margin: 0 0 12px;
  color: var(--j-ink-soft);
  font-size: 13px;
  line-height: 1.8;
}

.types {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 14px;
}

.summary {
  margin-bottom: 10px;
  font-size: 14px;
}

.file-link {
  color: var(--j-pen);
  text-decoration: none;
  overflow-wrap: anywhere;
}

.actions {
  margin-top: 12px;
}
</style>
