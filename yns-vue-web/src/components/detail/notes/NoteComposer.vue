<!--
  贴便签 / 编辑便签：选类型、写 Markdown、加图片（最多 9 张，上传前压缩）。
  不传 note 时是便签墙顶部的发布框；传入 note 时是编辑这张便签（放在弹窗里用）。
-->
<template>
  <section class="composer" :class="{ 'is-editing': isEdit }">
    <!-- 收起时：头像 + 像输入框的提示 + 明显的按钮 -->
    <div v-if="!isEdit && !open" class="composer-entry" role="button" tabindex="0"
         @click="startWriting" @keydown.enter.prevent="startWriting">
      <el-avatar :src="userStore.userBean?.avatar" :size="36" class="entry-avatar"
                 :style="!userStore.userBean?.avatar ? getAvatarStyle(userStore.userBean?.name || '访') : {}">
        {{ (userStore.userBean?.name || '访').charAt(0) }}
      </el-avatar>
      <div class="entry-input">
        {{ isLoggedIn ? '写点什么吧：一个技巧、一条踩坑、一个好链接…' : '登录后就可以在墙上贴便签了' }}
      </div>
      <el-button type="primary" class="entry-button" :disabled="!isLoggedIn">
        <el-icon class="el-icon--left"><EditPen/></el-icon>{{ isLoggedIn ? '贴一张便签' : '请先登录' }}
      </el-button>
    </div>

    <form v-else class="composer-form" @submit.prevent="submit">
      <div class="type-row" role="radiogroup" aria-label="便签类型">
        <span class="row-label">类型</span>
        <button v-for="item in NOTE_TYPES" :key="item.value" type="button"
                class="type-chip" :class="{ 'is-on': form.type === item.value }"
                role="radio" :aria-checked="form.type === item.value"
                :title="item.hint" @click="form.type = item.value">
          <i class="note-type-dot" :style="{ background: item.color }"></i>{{ item.label }}
        </button>
        <span class="type-hint">{{ currentType.hint }}</span>
      </div>

      <el-input
          v-if="!preview"
          ref="textRef"
          v-model="form.text"
          type="textarea"
          :autosize="{ minRows: 4, maxRows: 16 }"
          placeholder="写点什么吧（支持 Markdown，例如 **加粗**、`代码`、[链接](https://...)）"
          class="composer-text"
          @keyup.enter.ctrl="submit"
      />
      <div v-else class="composer-preview" v-html="previewHtml || '<p class=&quot;empty&quot;>还没有内容</p>'"></div>

      <NoteTagInput v-model="form.tags" :popular="popularTags" class="composer-tags"/>

      <div v-if="images.length" class="image-list">
        <span v-for="(image, index) in images" :key="image.key" class="image-item">
          <img :src="image.url || image.local" alt="便签图片">
          <span v-if="image.uploading" class="image-mask">上传中…</span>
          <button type="button" class="image-remove" title="移除" @click="removeImage(index)">×</button>
        </span>
      </div>

      <div class="composer-actions">
        <span class="left-actions">
          <label class="tool-btn" :class="{ disabled: images.length >= MAX_IMAGES }">
            <input type="file" accept="image/*" multiple hidden :disabled="images.length >= MAX_IMAGES" @change="pickImages">
            🖼 图片 {{ images.length }}/{{ MAX_IMAGES }}
          </label>
          <button type="button" class="tool-btn" @click="preview = !preview">{{ preview ? '继续编辑' : '预览' }}</button>
        </span>
        <span class="right-actions">
          <span class="hint-text">Ctrl + Enter {{ isEdit ? '保存' : '发布' }}</span>
          <el-button size="small" @click="cancel">{{ isEdit ? '取消' : '收起' }}</el-button>
          <el-button size="small" type="primary" native-type="submit" :loading="saving" :disabled="uploading">
            {{ isEdit ? '保存修改' : '贴上墙' }}
          </el-button>
        </span>
      </div>
    </form>
  </section>
</template>

<script setup>
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { EditPen } from '@element-plus/icons-vue'
import { marked } from 'marked'
import { useUserStore } from '@/stores/main/user.js'
import { getGuid, sanitizeHtml, sendAxiosRequest } from '@/utils/common.js'
import { compressImage, IMAGE_PRESETS } from '@/utils/imageCompress.js'
import { getAvatarStyle } from '@/utils/avatarStyle.js'
import { htmlToMarkdown } from '@/utils/htmlToMarkdown.js'
import { NOTE_TYPES, noteTypeOf } from '@/utils/noteTypes.js'
import { loadPopularTags } from '@/utils/noteTags.js'
import NoteTagInput from './NoteTagInput.vue'
import '@/css/notes.css'

const MAX_IMAGES = 9
const props = defineProps({
  // 传入时为编辑模式
  note: { type: Object, default: null }
})
const emit = defineEmits(['posted', 'saved', 'cancel'])
const userStore = useUserStore()

const isEdit = computed(() => !!props.note)
const isLoggedIn = computed(() => !!userStore.userBean?.code)
const open = ref(false)
const preview = ref(false)
const saving = ref(false)
const textRef = ref(null)
const form = reactive({ type: 'note', text: '', tags: [] })
// 常用话题：打开发布框或编辑时加载一次
const popularTags = ref([])
const fetchPopularTags = async () => {
  popularTags.value = (await loadPopularTags()).map(item => item.NAME)
}
// { key, local: 本地预览地址, url: 上传后的地址, uploading }
const images = ref([])

const currentType = computed(() => noteTypeOf(form.type))
// breaks：单个回车也换行（标准 Markdown 要空一行才分段，写便签时不符合习惯）
const renderMarkdown = (text) => sanitizeHtml(marked.parse(text, { breaks: true, gfm: true }))
const previewHtml = computed(() => renderMarkdown(form.text || ''))
const uploading = computed(() => images.value.some(image => image.uploading))

// 编辑时填入原来的内容；旧随记没有 Markdown 原文，把 HTML 转回 Markdown
watch(() => props.note, (note) => {
  if (!note) return
  form.type = noteTypeOf(note.TYPE).value
  form.text = note.SOURCE || htmlToMarkdown(note.TEXT || '')
  form.tags = Array.isArray(note.TAGS) ? [...note.TAGS] : []
  fetchPopularTags()
  images.value = (Array.isArray(note.IMAGES) ? note.IMAGES : []).map(url => ({ key: url, url, local: '', uploading: false }))
  preview.value = false
}, { immediate: true })

function startWriting() {
  if (!isLoggedIn.value) {
    ElMessage.info('请先点击右上角登录')
    return
  }
  open.value = true
  fetchPopularTags()
  nextTick(() => textRef.value?.focus())
}

async function uploadOne(item, file) {
  try {
    const compressed = await compressImage(file, IMAGE_PRESETS.article)
    const formData = new FormData()
    formData.append('file', compressed)
    // 和文章配图放在同一目录，后台“文件一致性”会把随记图片算作正在使用
    formData.append('spliceUrl', 'editorImage')
    const res = await sendAxiosRequest('/pub-api/upload/uploadFile', formData)
    if (!res || res.isError) throw new Error(res?.errMsg || '上传失败')
    item.url = res.result.fileViewUrl
  } catch (e) {
    ElMessage.error(`图片上传失败：${e.message}`)
    images.value = images.value.filter(image => image !== item)
  } finally {
    item.uploading = false
  }
}

function pickImages(event) {
  const files = [...(event.target.files || [])].filter(file => file.type.startsWith('image/'))
  event.target.value = ''
  const room = MAX_IMAGES - images.value.length
  if (files.length > room) ElMessage.warning(`最多 ${MAX_IMAGES} 张图片，多出的 ${files.length - room} 张没有添加`)
  files.slice(0, room).forEach(file => {
    const item = reactive({ key: getGuid(), local: URL.createObjectURL(file), url: '', uploading: true })
    images.value.push(item)
    uploadOne(item, file)
  })
}

function removeImage(index) {
  const [item] = images.value.splice(index, 1)
  if (item?.local) URL.revokeObjectURL(item.local)
}

function reset() {
  images.value.forEach(item => item.local && URL.revokeObjectURL(item.local))
  images.value = []
  form.text = ''
  form.type = 'note'
  form.tags = []
  preview.value = false
}

function cancel() {
  if (isEdit.value) emit('cancel')
  else open.value = false
}

async function submit() {
  if (saving.value) return
  if (uploading.value) {
    ElMessage.warning('图片还在上传，请稍等')
    return
  }
  const text = form.text.trim()
  const urls = images.value.map(image => image.url).filter(Boolean)
  if (!text && urls.length === 0) {
    ElMessage.warning('写点内容或加一张图片吧')
    return
  }
  saving.value = true
  try {
    const html = text ? renderMarkdown(text) : ''
    const community = { GUID: isEdit.value ? props.note.GUID : getGuid(), TEXT: html, SOURCE: text, TYPE: form.type, IMAGES: urls, TAGS: form.tags }
    const url = isEdit.value ? '/blog-api/community/updateCommunity' : '/blog-api/community/addCommunity'
    const res = await sendAxiosRequest(url, { community })
    if (!res || res.isError) {
      ElMessage.error(res?.errMsg || (isEdit.value ? '保存失败' : '发布失败'))
      return
    }
    // 后端返回保存后的完整随记；字段缺失时用本地内容补上
    const user = userStore.userBean
    const saved = {
      ...(isEdit.value ? props.note : { USERCODE: user.code, USERNAME: user.name, AVATAR: user.avatar || '', CREATE_TIME: '刚刚', COMMENTNUM: 0 }),
      TEXT: html, SOURCE: text, TYPE: form.type, IMAGES: urls, TAGS: [...form.tags],
      ...(res.result || {})
    }
    if (isEdit.value) {
      ElMessage.success('已保存')
      emit('saved', saved)
    } else {
      ElMessage.success('已贴上墙')
      emit('posted', saved)
      reset()
      open.value = false
    }
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.composer {
  position: relative;
  padding: 16px 18px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
}

.composer.is-editing {
  padding: 0;
  background: transparent;
  border: none;
  box-shadow: none;
}

/* ---------- 收起时的入口 ---------- */
.composer-entry {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.entry-avatar {
  flex-shrink: 0;
}

.entry-input {
  flex: 1;
  min-width: 0;
  padding: 9px 14px;
  border: 1px solid var(--j-rule-strong);
  border-radius: 20px;
  background: #fff;
  font-size: 14px;
  color: var(--j-muted);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  transition: border-color 0.15s;
}

.composer-entry:hover .entry-input,
.composer-entry:focus-visible .entry-input {
  border-color: var(--j-pen);
}

.composer-entry:focus-visible {
  outline: none;
}

.entry-button {
  flex-shrink: 0;
}

/* ---------- 展开后 ---------- */
.type-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.row-label {
  font-size: 13px;
  color: var(--j-muted);
}

.type-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 14px 3px 11px;
  border: 1px solid var(--j-rule-strong);
  border-radius: 14px;
  background: #fff;
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-ink-soft);
  cursor: pointer;
}

.type-chip:hover {
  border-color: var(--j-ink-soft);
}

.type-chip.is-on {
  border-color: var(--j-ink);
  background: var(--j-ink);
  color: #fff;
}

.type-hint {
  font-size: 12px;
  color: var(--j-muted);
}

.composer-text :deep(.el-textarea__inner) {
  line-height: 1.75;
}

.composer-preview {
  min-height: 96px;
  padding: 8px 12px;
  border: 1px dashed var(--j-rule-strong);
  background: #fff;
  line-height: 1.75;
  overflow-wrap: anywhere;
}

.composer-preview :deep(p) {
  margin: 0 0 0.5em;
}

.composer-preview :deep(.empty) {
  color: var(--j-muted);
}

.composer-tags {
  margin-top: 10px;
}

.image-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

.image-item {
  position: relative;
  width: 76px;
  height: 76px;
  padding: 3px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.25);
}

.image-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.image-mask {
  position: absolute;
  inset: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.7);
  font-size: 12px;
  color: var(--j-ink-soft);
}

.image-remove {
  position: absolute;
  top: -7px;
  right: -7px;
  width: 20px;
  height: 20px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: var(--j-stamp);
  color: #fff;
  font-size: 14px;
  line-height: 20px;
  cursor: pointer;
}

.composer-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: 12px;
}

.left-actions,
.right-actions {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.tool-btn {
  padding: 3px 10px;
  border: 1px dashed var(--j-rule-strong);
  border-radius: 3px;
  background: #fff;
  font-size: 13px;
  color: var(--j-ink-soft);
  cursor: pointer;
}

.tool-btn:hover {
  color: var(--j-pen);
  border-color: var(--j-pen);
}

.tool-btn.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.hint-text {
  font-size: 12px;
  color: var(--j-muted);
}

@media (max-width: 640px) {
  .hint-text,
  .entry-avatar {
    display: none;
  }

  .entry-input {
    padding: 8px 12px;
  }
}
</style>
