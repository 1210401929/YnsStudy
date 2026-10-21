<!-- 便签墙顶部的“贴一张便签”：选类型、写 Markdown、加图片（最多 9 张，上传前压缩） -->
<template>
  <section class="composer" :class="[currentType.paper, { 'is-open': open }]">
    <span class="j-tape j-tape--white composer-tape"></span>

    <div v-if="!userStore.userBean?.code" class="composer-guest">
      ✎ 登录后就可以在墙上贴便签了（右上角登录）
    </div>

    <template v-else>
      <div v-if="!open" class="composer-placeholder" @click="open = true">
        ✎ 贴一张便签：随手记下一个技巧、一条踩坑、一个好链接…
      </div>

      <form v-else class="composer-form" @submit.prevent="submit">
        <div class="type-row" role="radiogroup" aria-label="便签类型">
          <button v-for="item in NOTE_TYPES" :key="item.value" type="button"
                  class="type-chip" :class="[item.paper, { 'is-on': form.type === item.value }]"
                  role="radio" :aria-checked="form.type === item.value"
                  :title="item.hint" @click="form.type = item.value">
            {{ item.label }}
          </button>
          <span class="type-hint">{{ currentType.hint }}</span>
        </div>

        <el-input
            v-if="!preview"
            v-model="form.text"
            type="textarea"
            :autosize="{ minRows: 4, maxRows: 14 }"
            placeholder="写点什么吧（支持 Markdown，例如 **加粗**、`代码`、[链接](https://...)）"
            class="composer-text"
            @keyup.enter.ctrl="submit"
        />
        <div v-else class="composer-preview" v-html="previewHtml"></div>

        <div v-if="images.length" class="image-list">
          <span v-for="(image, index) in images" :key="image.key" class="image-item">
            <img :src="image.url || image.local" alt="待发布的图片">
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
            <span class="hint-text">Ctrl + Enter 发布</span>
            <el-button size="small" @click="cancel">收起</el-button>
            <el-button size="small" type="primary" native-type="submit" :loading="saving" :disabled="uploading">贴上墙</el-button>
          </span>
        </div>
      </form>
    </template>
  </section>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { marked } from 'marked'
import { useUserStore } from '@/stores/main/user.js'
import { getGuid, sanitizeHtml, sendAxiosRequest } from '@/utils/common.js'
import { compressImage, IMAGE_PRESETS } from '@/utils/imageCompress.js'
import { NOTE_TYPES, noteTypeOf } from '@/utils/noteTypes.js'

const MAX_IMAGES = 9
const emit = defineEmits(['posted'])
const userStore = useUserStore()

const open = ref(false)
const preview = ref(false)
const saving = ref(false)
const form = reactive({ type: 'note', text: '' })
// { key, local: 本地预览地址, url: 上传后的地址, uploading }
const images = ref([])

const currentType = computed(() => noteTypeOf(form.type))
const previewHtml = computed(() => sanitizeHtml(marked.parse(form.text || '')))
const uploading = computed(() => images.value.some(image => image.uploading))

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
  preview.value = false
}

function cancel() {
  open.value = false
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
    const html = text ? sanitizeHtml(marked.parse(text)) : ''
    const res = await sendAxiosRequest('/blog-api/community/addCommunity', {
      community: { GUID: getGuid(), TEXT: html, TYPE: form.type, IMAGES: urls }
    })
    if (!res || res.isError) {
      ElMessage.error(res?.errMsg || '发布失败')
      return
    }
    const user = userStore.userBean
    emit('posted', {
      GUID: res.result?.GUID,
      TEXT: html,
      TYPE: res.result?.TYPE || form.type,
      IMAGES: res.result?.IMAGES || urls,
      USERCODE: user.code,
      USERNAME: user.name,
      AVATAR: user.avatar || '',
      CREATE_TIME: '刚刚',
      COMMENTNUM: 0
    })
    ElMessage.success('已贴上墙')
    reset()
    open.value = false
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.composer {
  position: relative;
  padding: 22px 20px 16px;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.1), 0 10px 18px -12px rgba(60, 50, 30, 0.4);
  transition: background-color 0.2s ease;
}

.composer.paper-note { background: var(--j-note); }
.composer.paper-tip { background: #fffdf8; border: 1px solid var(--j-rule); }
.composer.paper-share { background: var(--j-kraft); }
.composer.paper-question { background: #f8dcdc; }

.composer-tape {
  top: -10px;
  left: 50%;
  margin-left: -42px;
  transform: rotate(-2deg);
}

.composer-guest,
.composer-placeholder {
  font-family: var(--j-hand);
  font-size: 17px;
  color: var(--j-ink-soft);
}

.composer-placeholder {
  cursor: text;
}

.composer-placeholder:hover {
  color: var(--j-ink);
}

.type-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.type-chip {
  padding: 3px 14px;
  border: 1px solid rgba(77, 73, 67, 0.25);
  border-radius: 14px;
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-ink-soft);
  cursor: pointer;
  opacity: 0.75;
}

.type-chip.paper-note { background: var(--j-note); }
.type-chip.paper-tip { background: #fffdf8; }
.type-chip.paper-share { background: var(--j-kraft); }
.type-chip.paper-question { background: #f8dcdc; }

.type-chip.is-on {
  border-color: var(--j-ink-soft);
  color: var(--j-ink);
  opacity: 1;
  box-shadow: 0 1px 3px rgba(60, 50, 30, 0.25);
}

.type-hint {
  font-size: 12px;
  color: var(--j-muted);
}

.composer-text :deep(.el-textarea__inner) {
  background: rgba(255, 255, 255, 0.6);
  line-height: 1.75;
}

.composer-preview {
  min-height: 96px;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.6);
  line-height: 1.75;
  overflow-wrap: anywhere;
}

.composer-preview :deep(p) {
  margin: 0 0 0.5em;
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
  border: 1px dashed rgba(77, 73, 67, 0.35);
  border-radius: 3px;
  background: rgba(255, 255, 255, 0.5);
  font-size: 13px;
  color: var(--j-ink-soft);
  cursor: pointer;
}

.tool-btn:hover {
  color: var(--j-pen);
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
  .hint-text {
    display: none;
  }
}
</style>
