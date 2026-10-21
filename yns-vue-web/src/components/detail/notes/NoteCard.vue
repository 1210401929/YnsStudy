<!-- 便签墙上的一张便签：类型用胶带和圆点的颜色区分，点开进入单条随记页 -->
<template>
  <article class="note-card" :class="type.paper" @click="openNote">
    <span v-if="note.ISTOP === '1'" class="note-pin" title="置顶"></span>
    <span v-else class="j-tape note-tape note-type-tape"></span>

    <header class="note-head">
      <span class="note-type-badge"><i class="note-type-dot" :style="{ background: type.color }"></i>{{ type.label }}</span>
      <time class="note-time">{{ pubFormatDate(note.CREATE_TIME) }}</time>
    </header>

    <div class="note-body" :class="{ 'is-clamped': clamped }" v-html="safeHtml"></div>

    <div v-if="images.length" class="note-images" :class="'count-' + Math.min(images.length, 3)">
      <span v-for="(image, index) in images.slice(0, 3)" :key="image" class="note-thumb">
        <img :src="image" :alt="`${title} 配图 ${index + 1}`" loading="lazy" decoding="async">
        <span v-if="index === 2 && images.length > 3" class="note-more">+{{ images.length - 3 }}</span>
      </span>
    </div>

    <footer class="note-foot">
      <span class="note-author" @click.stop="openAuthor">
        <el-avatar :src="note.AVATAR" :size="22" :style="!note.AVATAR ? getAvatarStyle(note.USERNAME) : {}">
          {{ note.USERNAME?.charAt(0) }}
        </el-avatar>
        <span class="note-author-name">{{ note.USERNAME || '访客' }}</span>
      </span>
      <span class="note-actions" @click.stop>
        <button v-if="isSuperAdmin" type="button" class="note-link" @click="$emit('top', note)">
          {{ note.ISTOP === '1' ? '取消置顶' : '置顶' }}
        </button>
        <button v-if="canEdit" type="button" class="note-link" @click="$emit('edit', note)">编辑</button>
        <button v-if="canDelete" type="button" class="note-link danger" @click="$emit('delete', note)">删除</button>
        <!-- 真实链接：搜索引擎和“在新标签页打开”都能用 -->
        <a class="note-link note-comments" :href="href" @click.prevent="openNote">💬 {{ note.COMMENTNUM || 0 }}</a>
      </span>
    </footer>
  </article>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/main/user.js'
import { getCurrentUserAdminObject, pubFormatDate, sanitizeHtml } from '@/utils/common.js'
import { getAvatarStyle } from '@/utils/avatarStyle.js'
import { pubOpenUser } from '@/utils/blogUtil.js'
import { noteTitle, notePath, noteTypeOf } from '@/utils/noteTypes.js'
import { canDeleteNote, canEditNote } from '@/utils/notePermission.js'
import '@/css/notes.css'

const props = defineProps({
  note: { type: Object, required: true }
})
defineEmits(['delete', 'top', 'edit'])

const router = useRouter()
const userStore = useUserStore()

const type = computed(() => noteTypeOf(props.note.TYPE))
const title = computed(() => noteTitle(props.note))
const href = computed(() => notePath(props.note.GUID))
const images = computed(() => Array.isArray(props.note.IMAGES) ? props.note.IMAGES : [])
const safeHtml = computed(() => sanitizeHtml(props.note.TEXT || ''))

// 正文较长时截断，渐隐提示还有内容
const clamped = computed(() => (props.note.TEXT || '').length > 220 || (props.note.TEXT || '').split('</p>').length > 7)

const isSuperAdmin = computed(() => getCurrentUserAdminObject().adminLevel === 'superAdmin')
const canEdit = computed(() => canEditNote(props.note, userStore.userBean))
const canDelete = computed(() => canDeleteNote(props.note, userStore.userBean))

function openNote(event) {
  // 正文里的链接照常打开，不跳到随记页
  if (event?.target?.closest?.('.note-body a')) return
  router.push(href.value)
}

function openAuthor() {
  if (props.note.USERCODE) pubOpenUser(router, props.note.USERCODE)
}
</script>

<style scoped>
.note-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 22px 18px 12px;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.06), 0 10px 20px -14px rgba(60, 50, 30, 0.35);
  color: var(--j-ink);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.note-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 4px rgba(60, 50, 30, 0.08), 0 16px 26px -14px rgba(60, 50, 30, 0.4);
}

.note-tape {
  top: -10px;
  left: 50%;
  width: 70px;
  margin-left: -35px;
  transform: rotate(-3deg);
}

.note-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 12px;
  color: var(--j-muted);
}

.note-body {
  position: relative;
  font-size: 14px;
  line-height: 1.75;
  overflow-wrap: anywhere;
}

.note-body.is-clamped {
  max-height: 12.2em;
  overflow: hidden;
  -webkit-mask-image: linear-gradient(#000 70%, transparent);
  mask-image: linear-gradient(#000 70%, transparent);
}

.note-body :deep(p) {
  margin: 0 0 0.4em;
}

.note-body :deep(h1),
.note-body :deep(h2),
.note-body :deep(h3) {
  margin: 0 0 0.4em;
  font-size: 16px;
}

.note-body :deep(img) {
  max-width: 100%;
}

.note-body :deep(pre) {
  overflow: auto;
  padding: 8px 10px;
  background: #f6f1e6;
  font-size: 12px;
}

.note-body :deep(code) {
  padding: 0 4px;
  background: #f6f1e6;
  font-size: 0.9em;
}

.note-body :deep(a) {
  color: var(--j-pen);
}

/* 缩略图：一张大图、两张并排、三张一大两小 */
.note-images {
  display: grid;
  gap: 6px;
}

.note-images.count-1 {
  grid-template-columns: 1fr;
}

.note-images.count-2 {
  grid-template-columns: 1fr 1fr;
}

.note-images.count-3 {
  grid-template-columns: 2fr 1fr;
  grid-template-rows: 1fr 1fr;
}

.note-images.count-3 .note-thumb:first-child {
  grid-row: span 2;
}

.note-thumb {
  position: relative;
  display: block;
  overflow: hidden;
  padding: 4px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.2);
  aspect-ratio: 4 / 3;
}

.note-images.count-1 .note-thumb {
  aspect-ratio: 16 / 10;
}

.note-images.count-3 .note-thumb:first-child {
  aspect-ratio: auto;
}

.note-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.note-more {
  position: absolute;
  inset: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(43, 42, 39, 0.45);
  color: #fff;
  font-size: 18px;
}

.note-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding-top: 8px;
  border-top: 1px dashed var(--j-rule-strong);
  font-size: 12px;
}

.note-author {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.note-author-name {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  color: var(--j-ink-soft);
}

.note-author:hover .note-author-name {
  color: var(--j-pen);
}

.note-actions {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
  gap: 8px;
}

.note-link {
  padding: 0;
  border: none;
  background: none;
  font-size: 12px;
  color: var(--j-muted);
  cursor: pointer;
  text-decoration: none;
}

.note-link:hover {
  color: var(--j-pen);
}

.note-link.danger:hover {
  color: var(--j-stamp);
}

.note-comments {
  color: var(--j-ink-soft);
}
</style>
