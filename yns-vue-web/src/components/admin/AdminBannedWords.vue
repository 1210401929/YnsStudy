<!-- 后台：违禁词在线管理 -->
<template>
  <div v-loading="loading" class="admin-banned-words">
    <p class="hint">发表文章评论时，内容、昵称、网址中包含以下任意一个词就会被拒绝（不区分大小写）。修改后立即生效。</p>

    <el-alert
        v-if="!tableReady"
        type="warning"
        :closable="false"
        show-icon
        title="数据库中还没有违禁词表，请先执行迁移 007_add_admin_tables.sql；目前只有配置文件中的词生效。"
        class="alert"
    />

    <div class="add-row">
      <el-input
          v-model="newWord"
          size="small"
          maxlength="100"
          placeholder="输入要添加的违禁词，回车添加"
          style="width: 260px"
          :disabled="!tableReady"
          @keyup.enter="add"
      />
      <el-button size="small" type="primary" :disabled="!tableReady || !newWord.trim()" @click="add">添加</el-button>
    </div>

    <h4 class="sub">后台添加的违禁词（{{ custom.length }}）</h4>
    <div v-if="custom.length" class="words">
      <el-tag v-for="item in custom" :key="item.WORD" closable :title="`${item.CREATE_USER || ''} ${pubFormatDate(item.CREATE_TIME)}`" @close="remove(item.WORD)">
        {{ item.WORD }}
      </el-tag>
    </div>
    <el-empty v-else description="还没有添加" :image-size="50"/>

    <h4 class="sub">配置文件中的违禁词（{{ configWords.length }}，只读）</h4>
    <p class="hint small">位于 config.yaml 的 content.banned_words，修改需要重启 Go 服务。</p>
    <div class="words">
      <el-tag v-for="word in configWords" :key="word" type="info">{{ word }}</el-tag>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { ele_confirm, pubFormatDate, sendAxiosRequestChecked } from '@/utils/common.js'

const loading = ref(false)
const tableReady = ref(true)
const custom = ref([])
const configWords = ref([])
const newWord = ref('')

const load = async () => {
  loading.value = true
  try {
    const result = await sendAxiosRequestChecked('/blog-api/admin/getBannedWords', {}, '获取违禁词失败')
    if (!result) return
    tableReady.value = result.result.tableReady !== false
    custom.value = result.result.custom || []
    configWords.value = result.result.config || []
  } finally {
    loading.value = false
  }
}

const add = async () => {
  const word = newWord.value.trim()
  if (!word) return
  const result = await sendAxiosRequestChecked('/blog-api/admin/addBannedWord', { word }, '添加失败')
  if (!result) return
  newWord.value = ''
  ElMessage.success('已添加')
  load()
}

const remove = (word) => {
  ele_confirm(`确定删除违禁词“${word}”吗？`, async () => {
    const result = await sendAxiosRequestChecked('/blog-api/admin/deleteBannedWord', { word }, '删除失败')
    if (!result) return
    ElMessage.success('已删除')
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

.hint.small {
  margin-top: -4px;
  color: var(--j-muted);
  font-size: 12px;
}

.alert {
  margin-bottom: 12px;
}

.add-row {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.sub {
  margin: 18px 0 8px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 16px;
}

.words {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
