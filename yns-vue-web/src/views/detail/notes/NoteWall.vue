<!-- 便签墙（原“社区”）：随记一张张贴在墙上，可按类型筛选、只看自己的、搜索 -->
<template>
  <Announcement v-for="al in topAlert" :key="al.GUID" :TEXT="al.TEXT" :URL="al.URL" :URLNAME="al.URLNAME"/>

  <div class="note-wall-page">
    <header class="wall-heading">
      <div class="heading-label">
        <span class="j-tape heading-tape"></span>
        <h1>便签墙</h1>
        <p class="page-intro">随手记下的技巧、踩坑、分享和想法，一张张贴在这里。</p>
        <p class="page-stats">墙上共有 {{ total }} 张便签</p>
      </div>
    </header>

    <!-- 两栏：随记是大家贴的内容，留言板在下一版开放 -->
    <nav class="wall-tabs" aria-label="便签墙栏目">
      <span class="wall-tab is-on">随记</span>
      <el-tooltip content="留言板即将开放" placement="top">
        <span class="wall-tab is-disabled">留言</span>
      </el-tooltip>
    </nav>

    <NoteComposer class="wall-composer" @posted="onPosted"/>

    <div class="wall-toolbar">
      <div class="filter-chips" role="tablist" aria-label="按类型筛选">
        <button type="button" class="filter-chip" :class="{ 'is-on': query.type === '' }" @click="setType('')">全部</button>
        <button v-for="item in NOTE_TYPES" :key="item.value" type="button"
                class="filter-chip" :class="{ 'is-on': query.type === item.value }"
                @click="setType(item.value)"><i class="note-type-dot" :style="{ background: item.color }"></i>{{ item.label }}</button>
      </div>
      <div class="toolbar-right">
        <el-checkbox v-if="userStore.userBean?.code" v-model="query.mine" @change="reload">只看我的</el-checkbox>
        <el-select v-model="query.sort" size="small" class="sort-select" @change="reload">
          <el-option label="最新" value="latest"/>
          <el-option label="最多讨论" value="hot"/>
        </el-select>
        <el-input v-model="query.keyword" size="small" clearable placeholder="搜索便签" class="search-input"
                  @keyup.enter="reload" @clear="reload">
          <template #prefix><el-icon><Search/></el-icon></template>
        </el-input>
        <el-button size="small" plain @click="userSearchVisible = true">
          <el-icon class="el-icon--left"><User/></el-icon>找人
        </el-button>
      </div>
    </div>

    <!-- 切换筛选时先保留旧内容（变淡），新内容到了再替换，页面高度不会突然塌下去 -->
    <div ref="wallRef" class="wall" :class="{ 'is-refreshing': refreshing }" :style="{ '--cols': columnCount }">
      <div v-for="(column, ci) in columns" :key="ci" class="wall-column">
        <NoteCard v-for="item in column" :key="item.note.GUID" :note="item.note"
                  @delete="removeNote" @top="toggleTop" @edit="openEdit"/>
      </div>
    </div>

    <p v-if="!loading && !refreshing && notes.length === 0" class="wall-empty">
      {{ query.mine ? '你还没有贴过便签，在上面写下第一张吧' : (query.keyword || query.type ? '没有找到符合条件的便签' : '墙上还空着，来贴第一张便签吧') }}
    </p>
    <div class="wall-more">
      <span v-if="loading && !refreshing" class="more-text">正在翻找…</span>
      <button v-else-if="!noMore" type="button" class="j-button" @click="fetchNotes()">再翻一页</button>
      <span v-else-if="notes.length" class="more-text">墙上就这些了</span>
    </div>

    <el-dialog v-model="userSearchVisible" title="找人" width="min(720px, 92vw)" append-to-body>
      <UserList/>
    </el-dialog>

    <el-dialog v-model="editVisible" title="编辑便签" width="min(720px, 94vw)" append-to-body destroy-on-close>
      <NoteComposer v-if="editingNote" :note="editingNote" @saved="onSaved" @cancel="editVisible = false"/>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, defineAsyncComponent, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, User } from '@element-plus/icons-vue'
import Announcement from '@/components/detail/Announcement.vue'
import NoteCard from '@/components/detail/notes/NoteCard.vue'
import NoteComposer from '@/components/detail/notes/NoteComposer.vue'
import { useUserStore } from '@/stores/main/user.js'
import { ele_confirm, sendAxiosRequest, sendAxiosRequestChecked } from '@/utils/common.js'
import { getAnnouncementByRouterName } from '@/utils/blogUtil.js'
import { NOTE_TYPES } from '@/utils/noteTypes.js'
import '@/css/notes.css'

// 找人只在点开时才加载
const UserList = defineAsyncComponent(() => import('@/components/detail/UserList.vue'))

const PAGE_SIZE = 12
const userStore = useUserStore()

const query = reactive({ type: '', sort: 'latest', keyword: '', mine: false })
const notes = ref([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const refreshing = ref(false)
const noMore = ref(false)
const userSearchVisible = ref(false)
// 每次请求的序号：快速连续切换筛选时，只采用最后一次的结果
let requestSeq = 0

// replace 为 true 时用新结果替换整面墙（切换筛选、搜索），否则追加下一页
const fetchNotes = async (replace = false) => {
  if (!replace && (loading.value || noMore.value)) return
  const seq = ++requestSeq
  const targetPage = replace ? 1 : page.value
  loading.value = true
  refreshing.value = replace
  try {
    const res = await sendAxiosRequest('/blog-api/community/getAllCommunity', {
      page: targetPage, pageSize: PAGE_SIZE, ...query
    })
    if (seq !== requestSeq) return
    if (!res || res.isError) {
      ElMessage.error(res?.errMsg || '获取便签失败')
      return
    }
    const rows = res.result?.data || []
    notes.value = replace ? rows : [...notes.value, ...rows]
    total.value = res.result?.total || 0
    noMore.value = rows.length < PAGE_SIZE
    page.value = targetPage + 1
  } finally {
    if (seq === requestSeq) {
      loading.value = false
      refreshing.value = false
    }
  }
}

const reload = () => fetchNotes(true)

const setType = (type) => {
  query.type = type
  reload()
}

// ---------- 瀑布流：按“哪列最短放哪列”分配，大致保持新的在前 ----------
const wallRef = ref(null)
const columnCount = ref(3)
let resizeObserver = null

const updateColumns = () => {
  const width = wallRef.value?.clientWidth || window.innerWidth
  columnCount.value = width >= 1000 ? 3 : width >= 620 ? 2 : 1
}

// 渲染前无法知道真实高度，按文字长度和图片数量估算
const estimateHeight = (note) => {
  const textLength = String(note.TEXT || '').replace(/<[^>]+>/g, '').length
  const images = Array.isArray(note.IMAGES) ? note.IMAGES.length : 0
  return 110 + Math.min(textLength, 220) * 0.9 + (images ? 170 : 0)
}

const columns = computed(() => {
  const result = Array.from({ length: columnCount.value }, () => [])
  const heights = new Array(columnCount.value).fill(0)
  notes.value.forEach((note, index) => {
    const target = heights.indexOf(Math.min(...heights))
    result[target].push({ note, index })
    heights[target] += estimateHeight(note)
  })
  return result
})

onMounted(() => {
  updateColumns()
  if (typeof ResizeObserver !== 'undefined' && wallRef.value) {
    resizeObserver = new ResizeObserver(updateColumns)
    resizeObserver.observe(wallRef.value)
  }
  // 旋转屏幕等情况下 ResizeObserver 可能不触发，窗口尺寸变化时也重新计算
  window.addEventListener('resize', updateColumns)
  fetchNotes()
})
onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  window.removeEventListener('resize', updateColumns)
})

// ---------- 发布、编辑、删除、置顶 ----------
const editVisible = ref(false)
const editingNote = ref(null)

async function openEdit(note) {
  // 列表里没有 Markdown 原文，编辑前取一次完整的随记
  const res = await sendAxiosRequest('/blog-api/community/getNote', { noteId: note.GUID })
  editingNote.value = res && !res.isError && res.result?.note ? res.result.note : note
  editVisible.value = true
}

function onSaved(saved) {
  const index = notes.value.findIndex(item => item.GUID === saved.GUID)
  if (index !== -1) notes.value[index] = { ...notes.value[index], ...saved }
  editVisible.value = false
}

function onPosted(note) {
  // 筛选条件和新便签不符时也放到最前面，方便作者确认已经贴上
  notes.value.unshift(note)
  total.value++
}

function removeNote(note) {
  ele_confirm('确定撕下这张便签吗？它下面的评论会一起删除。', async () => {
    const result = await sendAxiosRequestChecked('/blog-api/community/deleteCommunity', { communityGuid: note.GUID }, '删除失败')
    if (!result) return
    notes.value = notes.value.filter(item => item.GUID !== note.GUID)
    total.value = Math.max(0, total.value - 1)
    ElMessage.success('已删除')
  })
}

async function toggleTop(note) {
  const isTop = note.ISTOP === '1' ? '0' : '1'
  const result = await sendAxiosRequestChecked('/blog-api/community/setTopCommunity', { communityGuid: note.GUID, isTop }, '置顶操作失败')
  if (!result) return
  note.ISTOP = isTop === '1' ? '1' : null
  ElMessage.success(isTop === '1' ? '已置顶，刷新后排到最前' : '已取消置顶')
}

// 公告沿用原“社区”栏目的设置
const topAlert = ref([])
getAnnouncementByRouterName('Community').then(list => { topAlert.value = list || [] }).catch(() => {})
</script>

<style scoped>
.note-wall-page {
  max-width: 1180px;
  margin: 0 auto;
  padding: 28px 20px 72px;
  box-sizing: border-box;
}

.wall-heading {
  margin-bottom: 22px;
}

.heading-label {
  position: relative;
  display: inline-block;
  padding: 18px 28px 14px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
  transform: rotate(-0.6deg);
}

.heading-tape {
  top: -11px;
  left: 50%;
  margin-left: -42px;
  transform: rotate(-3deg);
}

.heading-label h1 {
  margin: 0;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 30px;
  color: var(--j-ink);
}

.page-intro {
  margin: 6px 0 0;
  font-size: 14px;
  color: var(--j-ink-soft);
}

.page-stats {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--j-muted);
}

/* 栏目：像本子上露出来的索引标签 */
.wall-tabs {
  display: flex;
  gap: 6px;
  margin-bottom: 16px;
  border-bottom: 1px solid var(--j-rule-strong);
}

.wall-tab {
  padding: 6px 18px 8px;
  border: 1px solid transparent;
  border-radius: 6px 6px 0 0;
  font-family: var(--j-hand);
  font-size: 16px;
  color: var(--j-ink-soft);
  transform: translateY(1px);
}

.wall-tab.is-on {
  background: var(--j-paper);
  border-color: var(--j-rule-strong);
  border-bottom-color: var(--j-paper);
  color: var(--j-ink);
}

.wall-tab.is-disabled {
  color: var(--j-muted);
  cursor: not-allowed;
}

.wall-composer {
  margin-bottom: 22px;
}

.wall-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 22px;
}

.filter-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 14px;
  border: 1px solid var(--j-rule-strong);
  border-radius: 14px;
  background: var(--j-paper);
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-ink-soft);
  cursor: pointer;
  transition: background-color 0.15s, color 0.15s, border-color 0.15s;
}

.filter-chip:hover {
  border-color: var(--j-ink-soft);
}

.filter-chip.is-on {
  border-color: var(--j-ink);
  background: var(--j-ink);
  color: #fff;
}

.toolbar-right {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.sort-select {
  width: 110px;
}

.search-input {
  width: 180px;
}

/* 瀑布流 */
.wall {
  display: grid;
  grid-template-columns: repeat(var(--cols), minmax(0, 1fr));
  gap: 26px 22px;
  align-items: start;
}

.wall.is-refreshing {
  opacity: 0.5;
  pointer-events: none;
  transition: opacity 0.15s;
}

.wall-column {
  display: flex;
  flex-direction: column;
  gap: 26px;
  min-width: 0;
}

.wall-empty {
  padding: 48px 0 12px;
  text-align: center;
  font-family: var(--j-hand);
  font-size: 17px;
  color: var(--j-muted);
}

.wall-more {
  display: flex;
  justify-content: center;
  margin-top: 30px;
}

.more-text {
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-muted);
}

@media (max-width: 640px) {
  .note-wall-page {
    padding: 20px 14px 56px;
  }

  .heading-label {
    transform: none;
  }

  .toolbar-right,
  .search-input {
    width: 100%;
  }

  .search-input {
    flex: 1;
  }
}
</style>
