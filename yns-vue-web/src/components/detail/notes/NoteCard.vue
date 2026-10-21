<!-- 便签墙上的一张便签：不同类型用不同的纸，点开进入单条随记页 -->
<template>
  <article class="note-card" :class="[type.paper, { 'is-top': note.ISTOP === '1' }]" @click="openNote">
    <span v-if="note.ISTOP === '1'" class="note-pin" title="置顶"></span>
    <span v-else class="j-tape note-tape" :class="tapeClass"></span>

    <header class="note-head">
      <span class="note-type">{{ type.label }}</span>
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
import { adminUserCode } from '@/config/vue-config.js'
import { getCurrentUserAdminObject, pubFormatDate, sanitizeHtml } from '@/utils/common.js'
import { getAvatarStyle } from '@/utils/avatarStyle.js'
import { pubOpenUser } from '@/utils/blogUtil.js'
import { noteTitle, notePath, noteTypeOf } from '@/utils/noteTypes.js'

const props = defineProps({
  note: { type: Object, required: true },
  index: { type: Number, default: 0 }
})
defineEmits(['delete', 'top'])

const router = useRouter()
const userStore = useUserStore()

const type = computed(() => noteTypeOf(props.note.TYPE))
const title = computed(() => noteTitle(props.note))
const href = computed(() => notePath(props.note.GUID))
const images = computed(() => Array.isArray(props.note.IMAGES) ? props.note.IMAGES : [])
const safeHtml = computed(() => sanitizeHtml(props.note.TEXT || ''))
const tapeClass = computed(() => ['j-tape--green', 'j-tape--pink', '', 'j-tape--blue'][props.index % 4])

// 正文较长时截断，渐隐提示还有内容
const clamped = computed(() => (props.note.TEXT || '').length > 220 || (props.note.TEXT || '').split('</p>').length > 7)

const isSuperAdmin = computed(() => getCurrentUserAdminObject().adminLevel === 'superAdmin')
// 删除规则与后台一致：置顶的只有超级管理员能删；非置顶的作者本人可删，管理员可删普通用户的
const canDelete = computed(() => {
  const admin = getCurrentUserAdminObject()
  if (admin.adminLevel === 'superAdmin') return true
  if (props.note.ISTOP === '1') return false
  return (userStore.userBean?.code && userStore.userBean.code === props.note.USERCODE) ||
      (admin.isAdmin && props.note.USERCODE !== adminUserCode)
})

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
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.1), 0 10px 18px -12px rgba(60, 50, 30, 0.4);
  color: var(--j-ink);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  break-inside: avoid;
}

.note-card:hover {
  transform: translateY(-2px) rotate(-0.3deg);
  box-shadow: 0 2px 4px rgba(60, 50, 30, 0.12), 0 16px 24px -14px rgba(60, 50, 30, 0.45);
}

/* ---------- 四种纸 ---------- */
.paper-note {
  background: var(--j-note);
}

/* 索引卡：顶部一道红线 + 浅蓝横线 */
.paper-tip {
  background-color: #fffdf8;
  background-image:
      linear-gradient(rgba(232, 165, 155, 0.7), rgba(232, 165, 155, 0.7)),
      repeating-linear-gradient(transparent 0 27px, rgba(165, 195, 225, 0.35) 27px 28px);
  background-size: 100% 1px, 100% 100%;
  background-position: 0 44px, 0 44px;
  background-repeat: no-repeat, repeat;
  border: 1px solid var(--j-rule);
}

.paper-share {
  background: var(--j-kraft);
  border: 1px solid #d8c49d;
}

.paper-question {
  background: #f8dcdc;
}

.note-tape {
  top: -10px;
  left: 50%;
  width: 70px;
  margin-left: -35px;
  transform: rotate(-3deg);
}

/* 置顶：一枚红色图钉 */
.note-pin {
  position: absolute;
  top: -8px;
  left: 50%;
  z-index: 2;
  width: 16px;
  height: 16px;
  margin-left: -8px;
  border-radius: 50%;
  background: radial-gradient(circle at 35% 35%, #f08a80 0 25%, var(--j-stamp) 60%, #8f2f28 100%);
  box-shadow: 0 3px 4px rgba(60, 30, 20, 0.35);
}

.note-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 12px;
  color: var(--j-muted);
}

.note-type {
  padding: 0 8px;
  border: 1px solid rgba(77, 73, 67, 0.25);
  border-radius: 10px;
  font-family: var(--j-hand);
  font-size: 13px;
  color: var(--j-ink-soft);
  background: rgba(255, 255, 255, 0.45);
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

.note-body :deep(img) {
  max-width: 100%;
}

.note-body :deep(pre) {
  overflow: auto;
  padding: 8px 10px;
  background: rgba(255, 255, 255, 0.55);
  font-size: 12px;
}

.note-body :deep(code) {
  padding: 0 4px;
  background: rgba(255, 255, 255, 0.55);
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
  border-top: 1px dashed rgba(77, 73, 67, 0.25);
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
