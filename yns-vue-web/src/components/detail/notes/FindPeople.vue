<!-- 便签墙“找人”：按昵称、简介或用户编号搜索，只显示公开信息，点开进入对方主页 -->
<template>
  <div class="find-people">
    <el-input v-model="keyword" clearable placeholder="输入昵称、简介关键词，或用户编号" class="find-input"
              @input="searchLater" @clear="search">
      <template #prefix><el-icon><Search/></el-icon></template>
    </el-input>

    <p class="find-tip">{{ keyword.trim() ? `找到 ${total} 位` : '发过文章的作者排在前面' }}</p>

    <ul class="people-list" :class="{ 'is-loading': loading && people.length }">
      <li v-for="person in people" :key="person.CODE">
        <a class="person" :href="profileHref(person)" target="_blank" rel="noopener">
          <el-avatar :src="person.AVATAR" :size="40" :style="!person.AVATAR ? getAvatarStyle(person.NAME) : {}">
            {{ person.NAME?.charAt(0) }}
          </el-avatar>
          <span class="person-text">
            <span class="person-name">
              {{ person.NAME || '用户' }}
              <span v-if="person.ISOWNER" class="j-stamp person-stamp">站长</span>
              <span v-if="person.ISADMIN" class="j-stamp j-stamp--admin person-stamp">管理员</span>
            </span>
            <span class="person-remark">{{ person.REMARK || '这个人很神秘~' }}</span>
          </span>
          <span class="person-count">{{ person.ARTICLES > 0 ? `${person.ARTICLES} 篇文章` : '还没有文章' }}</span>
        </a>
      </li>
    </ul>

    <p v-if="!loading && people.length === 0" class="find-empty">没有找到这个人，换个关键词试试</p>
    <div class="find-more">
      <span v-if="loading" class="more-text">正在找…</span>
      <button v-else-if="!noMore" type="button" class="tool-link" @click="load()">再看看</button>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import debounce from 'lodash/debounce'
import { sendAxiosRequest } from '@/utils/common.js'
import { getAvatarStyle } from '@/utils/avatarStyle.js'

const PAGE_SIZE = 10
const keyword = ref('')
const people = ref([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const noMore = ref(false)
// 快速连续输入时，只采用最后一次搜索的结果
let requestSeq = 0

// replace 为 true 时是新的搜索，否则加载下一页
const load = async (replace = false) => {
  if (!replace && (loading.value || noMore.value)) return
  const seq = ++requestSeq
  const targetPage = replace ? 1 : page.value
  loading.value = true
  try {
    const res = await sendAxiosRequest('/pub-api/login/getPublicUsers', { keyword: keyword.value.trim(), page: targetPage, pageSize: PAGE_SIZE })
    if (seq !== requestSeq) return
    if (!res || res.isError || !res.result) {
      ElMessage.error(res?.errMsg || '搜索用户失败')
      return
    }
    const rows = res.result.data || []
    people.value = replace ? rows : [...people.value, ...rows]
    total.value = res.result.total || 0
    noMore.value = rows.length < PAGE_SIZE
    page.value = targetPage + 1
  } finally {
    if (seq === requestSeq) loading.value = false
  }
}

const search = () => load(true)
const searchLater = debounce(search, 400)

// 有用户编号时直接打开主页地址，在新标签页里看，不打断便签墙
const profileHref = (person) => person.USERNUM ? `/user/${encodeURIComponent(person.USERNUM)}` : '#'

onMounted(search)
</script>

<style scoped>
.find-input :deep(.el-input__wrapper) {
  border-radius: 18px;
}

.find-tip {
  margin: 10px 2px 6px;
  font-size: 12px;
  color: var(--j-muted);
}

.people-list {
  margin: 0;
  padding: 0;
  list-style: none;
  max-height: min(60vh, 520px);
  overflow-y: auto;
}

.people-list.is-loading {
  opacity: 0.6;
}

.person {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 8px;
  border-bottom: 1px dashed var(--j-rule);
  color: var(--j-ink);
  text-decoration: none;
  border-radius: 3px;
}

.person:hover {
  background: rgba(250, 216, 96, 0.16);
}

.person-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.person-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 15px;
  font-weight: 600;
}

.person-stamp {
  font-size: 11px;
  line-height: 15px;
}

.person-remark {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-size: 12px;
  color: var(--j-muted);
}

.person-count {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--j-ink-soft);
}

.find-empty {
  padding: 30px 0 10px;
  text-align: center;
  font-family: var(--j-hand);
  font-size: 16px;
  color: var(--j-muted);
}

.find-more {
  display: flex;
  justify-content: center;
  margin-top: 10px;
}

.more-text {
  font-size: 13px;
  color: var(--j-muted);
}

.tool-link {
  padding: 2px 6px;
  border: none;
  background: none;
  font-size: 13px;
  color: var(--j-pen);
  cursor: pointer;
}
</style>
