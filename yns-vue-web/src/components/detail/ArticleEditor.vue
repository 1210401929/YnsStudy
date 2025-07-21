<template>
  <div class="editor-container">
    <!-- 公开 / 私密 -->
    <el-radio-group
        v-if="!isReadOnly"
        v-model="isPublic"
        style="margin-bottom: 10px"
    >
      <el-radio-button :label="true">公开</el-radio-button>
      <el-radio-button :label="false">私密</el-radio-button>
    </el-radio-group>

    <!-- 标题 -->
    <el-input
        v-if="!isReadOnly"
        v-model="localTitle"
        placeholder="请输入文章标题"
        size="large"
        style="margin-bottom: 10px"
    />

    <!-- 编辑器 -->
    <div :style="wrapperStyle" class="editor-wrapper">
      <Toolbar
          v-if="!isReadOnly"
          :editor="editorRef"
          :defaultConfig="toolbarConfig"
          style="border-bottom: 1px solid #ccc"
      />

      <div ref="editorWrapper">
        <Editor
            v-model="localContent"
            :defaultConfig="editorConfig"
            :style="editorStyle"
            @onCreated="handleCreated"
            @customPaste="handleCustomPaste"
        />
      </div>
    </div>

    <!-- 底部按钮 -->
    <div v-if="!isReadOnly" class="bottom-buttons">
      <el-button @click="$emit('cancel')">取消</el-button>
      <el-button type="primary" @click="submit">提交</el-button>
    </div>

    <!-- 图片查看器 -->
    <el-image-viewer
        v-if="showViewer"
        :url-list="[viewerUrl]"
        @close="showViewer = false"
    />
  </div>
</template>

<script setup>
import '@wangeditor/editor/dist/css/style.css'
import {
  ref,
  computed,
  shallowRef,
  watch,
  onMounted,
  onUnmounted,
  onBeforeUnmount
} from 'vue'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import { sendAxiosRequest } from '@/utils/common.js'
import { ElMessage } from 'element-plus'

/* ---------- props / emits ---------- */
const props = defineProps({
  title: String,
  content: String,
  isPublic: Boolean,
  isReadOnly: Boolean,
  saveType: String
})
const emits = defineEmits(['submit', 'cancel'])

/* ---------- 状态 ---------- */
const isPublic = ref(props.isPublic)
const localTitle = ref(props.title || '')
const localContent = ref(props.content || '')
const editorRef = shallowRef(null)
const editorWrapper = ref(null)

/* ---------- 工具栏 ---------- */
const toolbarConfig = { excludeKeys: ['insertVideo', 'uploadVideo'] }

/* ---------- 独立上传函数 ---------- */
async function uploadImage(file) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('spliceUrl', 'editorImage')
  const res = await sendAxiosRequest('/pub-api/upload/uploadFile', formData)
  if (res && !res.isError) return res.result.fileViewUrl
  throw new Error(res?.errMsg || '上传失败')
}

/* ---------- 编辑器基础配置（不再含 customPaste） ---------- */
const editorConfig = {
  placeholder: '请输入正文...',
  readOnly: props.isReadOnly,
  MENU_CONF: {
    uploadImage: {
      async customUpload(file, insertFn) {
        try {
          const url = await uploadImage(file)
          insertFn(url)
        } catch (err) {
          ElMessage.error(err.message)
        }
      }
    },
    lineHeight: {
      lineHeightList: ['0.5', '1', '1.5', '2', '2.5'],
      defaultValue: '0.5'
    }
  }
}

/* ---------- 自定义粘贴处理（事件方式） ---------- */
const handleCustomPaste = async (editor, event, callback) => {
  event.preventDefault()
  const data = event.clipboardData
  if (!data) return callback(false)

  /* 1) 先处理真正的文件图片（截图等） */
  for (const file of Array.from(data.files || [])) {
    if (file.type.startsWith('image/')) {
      try {
        const url = await uploadImage(file)
        editor.insertNode({
          type: 'image',
          src: url,
          alt: file.name,
          children: [{ text: '' }]
        })
      } catch (err) {
        ElMessage.error(err.message || '图片上传失败')
      }
    }
  }

  /* 2) 过滤 file:// 本地路径 img */
  const html = data.getData('text/html') || ''
  const safeHtml = html.replace(
      /<img[^>]+src=["']file:[^"']+["'][^>]*>/gi,
      ''
  )
  if (safeHtml) {
    editor.dangerouslyInsertHtml(safeHtml)
  } else {
    const text = data.getData('text/plain') || ''
    if (text) editor.insertText(text)
  }

  /* 始终阻断默认粘贴 */
  callback(false)
}

/* ---------- 图片预览 ---------- */
const showViewer = ref(false)
const viewerUrl = ref('')
const showImagePreview = (url) => {
  viewerUrl.value = url
  showViewer.value = true
}

/* ---------- 双击查看 ---------- */
const onDblClick = (e) => {
  if (e.target.tagName === 'IMG') showImagePreview(e.target.src)
}

/* ---------- 生命周期 ---------- */
const handleCreated = (editor) => (editorRef.value = editor)

onMounted(() => editorWrapper.value?.addEventListener('dblclick', onDblClick))
onUnmounted(() =>
    editorWrapper.value?.removeEventListener('dblclick', onDblClick)
)
onBeforeUnmount(() => {
  editorRef.value?.destroy()
  editorRef.value = null
})

/* ---------- 提交 ---------- */
function submit() {
  if (!localTitle.value.trim()) {
    ElMessage.error('请输入标题')
    return
  }
  const contentClean = localContent.value.replace(
      /<img[^>]+src=["']file:[^"']+["'][^>]*>/gi,
      ''
  )
  emits('submit', {
    blog_type: isPublic.value ? 'public' : 'privacy',
    title: localTitle.value.trim(),
    content: contentClean.trim()
  })
  if (props.saveType === 'add') {
    localTitle.value = ''
    localContent.value = ''
  }
}

/* ---------- 样式计算 ---------- */
const wrapperStyle = computed(() => ({
  border: '0px solid #ccc',
  marginBottom: '0px',
  minHeight: props.isReadOnly ? '300px' : 'auto'
}))
const editorStyle = computed(() => ({
  minHeight: props.isReadOnly ? '300px' : 'auto',
  height: props.isReadOnly ? '100%' : '420px',
  overflowY: 'auto'
}))

/* ---------- 同步外部变更 ---------- */
watch(() => props.title, (v) => (localTitle.value = v))
watch(() => props.content, (v) => (localContent.value = v))
</script>

<style scoped>
.editor-container {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}
.editor-wrapper {
  flex: 1;
  height: 100%;
  overflow-y: auto;
}
.bottom-buttons {
  margin-top: auto;
  text-align: right;
  padding: 12px 0 0 0;
  background-color: #fff;
}
:deep(img) {
  cursor: zoom-in;
}
</style>
