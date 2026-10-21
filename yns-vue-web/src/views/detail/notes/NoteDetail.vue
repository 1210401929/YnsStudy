<!-- 单条随记页 /note/:id：左栏作者卡片，右边一张完整的便签 + 评论 + 上一条/下一条 -->
<template>
  <div class="note-detail-page j-desk">
    <div class="note-top-bar">
      <router-link to="/ynsStudy/Notes" class="back-link">← 便签墙</router-link>
      <div class="note-top-title">{{ note ? title : '便签' }}</div>
      <span class="back-link-placeholder"></span>
    </div>

    <main v-if="notFound" class="note-not-found">
      <h1>便签不存在</h1>
      <p>这张便签可能已被撕下（删除），或地址有误。</p>
      <router-link to="/ynsStudy/Notes">回到便签墙</router-link>
    </main>

    <main v-else class="main-body">
      <aside class="side-column">
        <UserInfo v-if="note?.USERCODE" :user="authorInfo" :target-user-code="note.USERCODE" @blog-click="openBlog"/>
      </aside>

      <div class="content-side">
        <p v-if="!note" class="note-loading">正在从墙上取下这张便签…</p>

        <template v-else>
          <article class="note-sheet" :class="type.paper">
            <span v-if="note.ISTOP === '1'" class="note-pin" title="置顶"></span>
            <span v-else class="j-tape j-tape--top note-type-tape"></span>

            <header class="note-head">
              <el-avatar :src="note.AVATAR" :size="40" class="author-avatar" @click="openAuthor"
                         :style="!note.AVATAR ? getAvatarStyle(note.USERNAME) : {}">
                {{ note.USERNAME?.charAt(0) }}
              </el-avatar>
              <div class="head-text">
                <a v-if="note.USERNUM" class="author-name" :href="`/user/${encodeURIComponent(note.USERNUM)}`"
                   @click.prevent="openAuthor">{{ note.USERNAME || '访客' }}</a>
                <span v-else class="author-name">{{ note.USERNAME || '访客' }}</span>
                <div class="head-meta">
                  <span class="note-type-badge"><i class="note-type-dot" :style="{ background: type.color }"></i>{{ type.label }}</span>
                  <time>{{ pubFormatDate(note.CREATE_TIME) }}</time>
                  <span v-if="note.UPDATE_TIME" class="edited" :title="'修改于 ' + pubFormatDate(note.UPDATE_TIME)">已编辑</span>
                  <button type="button" class="meta-comment" title="查看评论" @click="scrollToComments">💬 {{ commentTotal }}</button>
                </div>
              </div>
              <div class="head-tools">
                <button v-if="isSuperAdmin" type="button" class="tool-link" @click="toggleTop">{{ note.ISTOP === '1' ? '取消置顶' : '置顶' }}</button>
                <button v-if="canEdit" type="button" class="tool-link" @click="editVisible = true">编辑</button>
                <button v-if="canDelete" type="button" class="tool-link danger" @click="removeNote">删除</button>
              </div>
            </header>

            <ArticleViewer v-if="note.TEXT" :content="note.TEXT" class="note-text"/>

            <p v-if="note.TAGS?.length" class="note-tags">
              <router-link v-for="tag in note.TAGS" :key="tag" :to="tagHref(tag)" class="note-tag" rel="tag">#{{ tag }}</router-link>
            </p>

            <div v-if="images.length" class="note-gallery" :class="{ single: images.length === 1 }">
              <el-image v-for="(image, index) in images" :key="image" :src="image" :alt="`${title} 配图 ${index + 1}`"
                        :preview-src-list="images" :initial-index="index" fit="cover" lazy class="gallery-item"
                        preview-teleported hide-on-click-modal/>
            </div>
          </article>

          <section ref="commentAnchorRef" class="comment-anchor">
            <CommentSection
                kind="note"
                :target-id="note.GUID"
                :author-code="note.USERCODE || ''"
                :author-name="note.USERNAME || ''"
                title="便签下的讨论"
                @count="commentTotal = $event"
            />
          </section>

          <nav class="note-pager" aria-label="上一条和下一条">
            <a v-if="previous" class="pager-card" :href="notePath(previous.GUID)" @click.prevent="go(previous)">
              <span class="pager-label">← 上一条</span>
              <span class="pager-title">{{ previous.TITLE }}</span>
            </a>
            <span v-else></span>
            <a v-if="next" class="pager-card pager-next" :href="notePath(next.GUID)" @click.prevent="go(next)">
              <span class="pager-label">下一条 →</span>
              <span class="pager-title">{{ next.TITLE }}</span>
            </a>
          </nav>
          <p class="back-wall"><router-link to="/ynsStudy/Notes">回到便签墙 →</router-link></p>

          <el-dialog v-model="editVisible" title="编辑便签" width="min(720px, 94vw)" append-to-body destroy-on-close>
            <NoteComposer :note="note" @saved="onSaved" @cancel="editVisible = false"/>
          </el-dialog>
        </template>
      </div>
    </main>
  </div>
</template>

<script setup>
import { computed, nextTick, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import ArticleViewer from '@/components/detail/ArticleViewer.vue'
import CommentSection from '@/components/detail/comment/CommentSection.vue'
import UserInfo from '@/components/main/UserInfo.vue'
import NoteComposer from '@/components/detail/notes/NoteComposer.vue'
import { useUserStore } from '@/stores/main/user.js'
import {
  ele_confirm, extractPlainTextFromHTML, getCurrentUserAdminObject, getUserInfoByCode,
  pubFormatDate, sendAxiosRequest, sendAxiosRequestChecked
} from '@/utils/common.js'
import { canDeleteNote, canEditNote } from '@/utils/notePermission.js'
import { tagHref } from '@/utils/noteTags.js'
import '@/css/notes.css'
import { getAvatarStyle } from '@/utils/avatarStyle.js'
import { pubOpenOneBlog, pubOpenUser } from '@/utils/blogUtil.js'
import { noteTitle, notePath, noteTypeOf } from '@/utils/noteTypes.js'
import { absoluteUrl, SITE_NAME, useSeo } from '@/utils/seo.js'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
userStore.initFromLocal()

const note = ref(null)
const previous = ref(null)
const next = ref(null)
const notFound = ref(false)
const authorInfo = ref({})
const commentTotal = ref(0)
const commentAnchorRef = ref(null)

const type = computed(() => noteTypeOf(note.value?.TYPE))
const title = computed(() => noteTitle(note.value))
const images = computed(() => Array.isArray(note.value?.IMAGES) ? note.value.IMAGES : [])

const load = async (id) => {
  note.value = null
  notFound.value = false
  const res = await sendAxiosRequest('/blog-api/community/getNote', { noteId: id })
  if (!res || res.isError || !res.result?.note) {
    notFound.value = true
    window.prerenderReady = true
    return
  }
  note.value = res.result.note
  // 用旧编号（一长串字符）打开时，地址换成数字编号的新地址
  if (String(note.value.GUID) !== String(id)) router.replace(notePath(note.value.GUID))
  previous.value = res.result.previous || null
  next.value = res.result.next || null
  authorInfo.value = {}
  if (note.value.USERCODE) {
    getUserInfoByCode(note.value.USERCODE).then(result => {
      if (result && !result.isError) authorInfo.value = result.result || {}
    })
  }
  nextTick(() => { window.prerenderReady = true })
}
watch(() => route.params.id, (id) => {
  // 旧地址换成新地址时内容已经加载过，不再重复请求
  if (!id || String(note.value?.GUID) === String(id)) return
  window.scrollTo(0, 0)
  load(id)
}, { immediate: true })

const go = (item) => router.push(notePath(item.GUID))
const scrollToComments = () => commentAnchorRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
const openAuthor = () => note.value?.USERCODE && pubOpenUser(router, note.value.USERCODE)
const openBlog = (blog) => pubOpenOneBlog(router, blog.GUID)

const isSuperAdmin = computed(() => getCurrentUserAdminObject().adminLevel === 'superAdmin')
const canEdit = computed(() => canEditNote(note.value, userStore.userBean))
const canDelete = computed(() => canDeleteNote(note.value, userStore.userBean))

const editVisible = ref(false)
function onSaved(saved) {
  note.value = { ...note.value, ...saved }
  editVisible.value = false
}

function removeNote() {
  ele_confirm('确定撕下这张便签吗？它下面的评论会一起删除。', async () => {
    const result = await sendAxiosRequestChecked('/blog-api/community/deleteCommunity', { communityGuid: note.value.GUID }, '删除失败')
    if (!result) return
    ElMessage.success('已删除')
    router.push('/ynsStudy/Notes')
  })
}

async function toggleTop() {
  const isTop = note.value.ISTOP === '1' ? '0' : '1'
  const result = await sendAxiosRequestChecked('/blog-api/community/setTopCommunity', { communityGuid: note.value.GUID, isTop }, '置顶操作失败')
  if (!result) return
  note.value.ISTOP = isTop === '1' ? '1' : null
  ElMessage.success(isTop === '1' ? '已置顶' : '已取消置顶')
}

// ---------- SEO：与 Go 输出的首屏保持一致 ----------
useSeo(() => {
  if (notFound.value) return { title: `便签不存在 - ${SITE_NAME}`, noindex: true }
  if (!note.value) return { title: `便签墙 - ${SITE_NAME}`, path: route.path }
  const text = extractPlainTextFromHTML(note.value.TEXT || '').replace(/\s+/g, ' ').trim()
  const description = text ? (text.length > 160 ? `${text.slice(0, 160)}…` : text) : `${note.value.USERNAME || SITE_NAME} 在便签墙贴的一张便签`
  const path = notePath(note.value.GUID)
  const image = images.value[0] ? absoluteUrl(images.value[0]) : undefined
  return {
    title: `${title.value} - 便签墙 - ${SITE_NAME}`,
    description,
    path,
    type: 'article',
    image,
    jsonLd: {
      '@context': 'https://schema.org',
      '@type': 'SocialMediaPosting',
      headline: title.value,
      description,
      url: absoluteUrl(path),
      author: { '@type': 'Person', name: note.value.USERNAME || SITE_NAME },
      ...(image ? { image: [image] } : {})
    }
  }
})
</script>

<style scoped>
.note-detail-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  width: 100%;
  overflow-x: hidden;
  overflow-x: clip;
  box-sizing: border-box;
}

.note-top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 20px;
  background: var(--j-paper);
  border-bottom: 1px solid var(--j-rule);
  box-shadow: 0 6px 14px -12px rgba(60, 50, 30, 0.35);
}

.back-link,
.back-link-placeholder {
  width: 90px;
  flex-shrink: 0;
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-pen);
  text-decoration: none;
}

.note-top-title {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-align: center;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-family: var(--j-hand);
  font-size: 18px;
  color: var(--j-ink);
}

.main-body {
  display: flex;
  align-items: flex-start;
  gap: 28px;
  width: 100%;
  max-width: 1160px;
  margin: 0 auto;
  padding: 32px 24px 72px;
  box-sizing: border-box;
}

.side-column {
  align-self: stretch;
  width: 280px;
  flex-shrink: 0;
}

.side-column :deep(.left-sidebar) {
  position: sticky;
  top: 20px;
}

.content-side {
  flex: 1;
  min-width: 0;
  max-width: 820px;
}

.note-loading {
  padding: 80px 0;
  text-align: center;
  font-family: var(--j-hand);
  font-size: 18px;
  color: var(--j-muted);
}

/* ---------- 便签本体 ---------- */
.note-sheet {
  position: relative;
  padding: 30px 34px 26px;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.1), 0 14px 24px -16px rgba(60, 50, 30, 0.45);
}


.note-head {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 14px;
  margin-bottom: 14px;
  border-bottom: 1px dashed rgba(77, 73, 67, 0.3);
}

.author-avatar {
  flex-shrink: 0;
  cursor: pointer;
}

.head-text {
  flex: 1;
  min-width: 0;
}

.author-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--j-ink);
  text-decoration: none;
  cursor: pointer;
}

.author-name:hover {
  color: var(--j-pen);
}

.head-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-top: 3px;
  font-size: 13px;
  color: var(--j-muted);
}

.edited {
  font-size: 12px;
  color: var(--j-muted);
}

.meta-comment {
  padding: 0;
  border: none;
  background: none;
  font-size: 13px;
  color: var(--j-muted);
  cursor: pointer;
}

.meta-comment:hover {
  color: var(--j-pen);
}

.head-tools {
  display: flex;
  flex-shrink: 0;
  gap: 4px;
}

.tool-link {
  padding: 2px 6px;
  border: none;
  background: none;
  font-size: 13px;
  color: var(--j-pen);
  cursor: pointer;
}

.tool-link.danger {
  color: var(--j-stamp);
}

/* 便签纸上正文的底色透明，透出纸的颜色 */
.note-text :deep(.article-body) {
  padding: 0;
  font-size: 15.5px;
  line-height: 1.85;
}

/* 正文组件默认最小高度 300px，便签通常很短，不需要 */
.note-sheet .note-text {
  min-height: 0;
}

.note-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 12px;
  margin: 16px 0 0;
}

.note-tag {
  padding: 1px 10px;
  border-radius: 10px;
  background: #eef3f8;
  font-size: 13px;
  color: var(--j-pen);
  text-decoration: none;
}

.note-tag:hover {
  background: #dfe9f3;
}

.note-gallery {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  margin-top: 18px;
}

.note-gallery.single {
  grid-template-columns: minmax(0, 1fr);
  max-width: 560px;
}

.gallery-item {
  display: block;
  padding: 5px;
  background: #fff;
  box-shadow: 0 1px 3px rgba(60, 50, 30, 0.25);
  aspect-ratio: 4 / 3;
  cursor: zoom-in;
}

.note-gallery.single .gallery-item {
  aspect-ratio: auto;
  max-height: 520px;
}

.comment-anchor {
  scroll-margin-top: 20px;
}

/* ---------- 上一条 / 下一条：两张索引卡 ---------- */
.note-pager {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 30px 16px;
  margin-top: 30px;
  padding-top: 22px;
}

.pager-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 14px 16px 12px;
  border: 1px solid var(--j-rule);
  border-top: 2px solid rgba(232, 165, 155, 0.65);
  background: var(--j-paper);
  box-shadow: var(--j-shadow);
  color: var(--j-ink);
  text-decoration: none;
  transition: transform 0.2s ease;
}

.pager-card:hover {
  transform: translateY(-2px);
}

.pager-next {
  text-align: right;
}

.pager-label {
  position: absolute;
  top: -24px;
  left: 14px;
  padding: 2px 12px 3px;
  border-radius: 6px 6px 0 0;
  background: #f6e3a1;
  font-family: var(--j-hand);
  font-size: 13px;
}

.pager-next .pager-label {
  left: auto;
  right: 14px;
  background: #c7e2cf;
}

.pager-title {
  font-size: 14px;
  line-height: 1.6;
  overflow-wrap: anywhere;
}

.back-wall {
  margin: 18px 0 0;
  font-size: 14px;
}

.back-wall a {
  color: var(--j-pen);
  text-decoration: none;
}

.note-not-found {
  width: min(560px, calc(100% - 32px));
  margin: 14vh auto;
  padding: 40px 28px 32px;
  box-sizing: border-box;
  text-align: center;
  background: var(--j-note);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.1), 0 16px 28px -16px rgba(60, 50, 30, 0.45);
  transform: rotate(-1deg);
}

.note-not-found h1 {
  margin-top: 0;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 28px;
}

.note-not-found a {
  color: var(--j-pen);
}

@media (max-width: 991px) {
  .main-body {
    flex-direction: column;
    gap: 20px;
  }

  .side-column {
    width: 100%;
    align-self: auto;
  }

  .side-column :deep(.left-sidebar) {
    position: static;
  }

  .content-side {
    width: 100%;
    max-width: none;
  }
}

@media (max-width: 640px) {
  .main-body {
    padding: 16px 12px 56px;
  }

  .note-sheet {
    padding: 24px 18px 20px;
  }

  .note-gallery {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .note-pager {
    grid-template-columns: 1fr;
  }

  .pager-next {
    text-align: left;
  }

  .pager-next .pager-label {
    left: 14px;
    right: auto;
  }

  .back-link-placeholder {
    display: none;
  }
}
</style>
