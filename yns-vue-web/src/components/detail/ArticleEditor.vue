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
      <!-- 左侧 AI 按钮 -->
      <div class="left-buttons">
        <el-button type="success" round @click="aiButtonMh">Ai帮写</el-button>
        <el-button type="primary" round @click="aiButtonXx">Ai续写</el-button>
      </div>

      <!-- 右侧 取消/提交 按钮 -->
      <div class="right-buttons">
        <el-button @click="$emit('cancel')">取消</el-button>
        <el-button type="primary" @click="submit">提交</el-button>
      </div>
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
import {Editor, Toolbar} from '@wangeditor/editor-for-vue'
import {sendAxiosRequest, encrypt, pubLoading, getSendAxiosUrl} from '@/utils/common.js'
import {ElMessage} from 'element-plus'
import {built_in_preface, normalizeAiHtml} from "@/utils/formatHtml.js";

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
const toolbarConfig = {excludeKeys: ['insertVideo', 'uploadVideo']}

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
      lineHeightList: ['1', '1.4', '1.6', '1.8', '2'],
      defaultValue: '1'   // 关键：别用 0.5
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
          children: [{text: ''}]
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

/* ---------- ai按钮组 ---------- */
//从头写
function aiButtonMh() {

  let preface = localTitle.value;
  if (!preface) {
    ElMessage.error("告诉ai写点什么吧,写在文章标题区域!");
    return false;
  }
  if (preface.length > 50) {
    ElMessage.error(`内容过多,尽量简短且控制在50个字以内`);
    return false;
  }
  preface = `帮我写博客,直接说内容` + preface;
  editorRef.value.clear()
  localContent.value = "";
  aiHelp(preface);
}

//续写
function aiButtonXx() {
  let preface = `下面是我的博客内容(帮我续写,直接说内容)`;
  aiHelp(preface);
}

function aiHelp(preface = "") {

  let content = localContent.value;
  content = preface + content;

  try {
    editorRef.value?.focus(true)
  } catch {
  }
  pubLoading("start", {text: "正在使用 AI 协助写作，发现格式存在问题时，请对其进行少量修改与优化。"})
  const source = new EventSource(getSendAxiosUrl(`/ai-api/ai/qWenStream?message=${encrypt(content)}`));
  let buffer = '';

  const flush = (raw) => {
    const safe = normalizeAiHtml(raw);
    if (safe) {
      try {
        editorRef.value?.focus(true)
      } catch {
      }
      editorRef.value?.dangerouslyInsertHtml(safe);
    }
  };

  source.onmessage = (event) => {
    buffer += event.data;

    const chunks = [];
    let idx = 0;

    // 优先完整代码块
    while ((idx = buffer.indexOf('</pre>')) !== -1) {
      chunks.push(buffer.slice(0, idx + 6));
      buffer = buffer.slice(idx + 6);
    }
    // 其次段落
    while ((idx = buffer.indexOf('</p>')) !== -1) {
      chunks.push(buffer.slice(0, idx + 4));
      buffer = buffer.slice(idx + 4);
    }
    // 再次空行
    const nl2 = buffer.indexOf('\n\n');
    if (chunks.length === 0 && nl2 !== -1) {
      chunks.push(buffer.slice(0, nl2 + 2));
      buffer = buffer.slice(nl2 + 2);
    }

    for (const raw of chunks) flush(raw);
  };

  source.onerror = (err) => {
    console.error('后台调用AI 出错:', err);
    pubLoading("close");
    source.close();
    if (buffer.trim()) flush(buffer);
    buffer = '';
  };
}

/* ---------- 提交 ---------- */
function submit() {
  if (!localTitle.value.trim()) {
    ElMessage.error('请输入标题')
    return
  }
  //清除本地图片路径内容,例如 file:///C:/Users/EDY/Desktop/pic.png
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
  width: '100%',
  overflowY: 'auto'
}))

/* ---------- 同步外部变更 ---------- */
watch(() => props.title, (v) => (localTitle.value = v))
watch(() => props.content, (v) => (localContent.value = v))

</script>
<style scoped>
/* 给文本容器加左右内边距，避免内容贴边被裁 */
:deep(.w-e-text-container) {
  padding: 0 16px; /* 或者 padding-left: 16px; */
  box-sizing: border-box;
  overflow-x: auto; /* 防止代码块等超长内容被裁掉 */
}

/* 1) 真正的内容容器是 .w-e-text —— 在这里加“护边”最稳 */
:deep(.w-e-text) {
  padding-left: 0px;  /*左侧安全区 */
  padding-right: 0px; /* 顺便给右侧也加一点 */
  box-sizing: border-box;
}

/* 2) 任何被注入的左缩进/负缩进一律清零（覆盖内联样式） */
:deep(.w-e-text > *) {
  margin-left: 0 !important;
  text-indent: 0 !important;
}

/* 3) 列表缩进用 padding 来实现，避免项目符号被“顶”出可视区 */
:deep(ul), :deep(ol) {
  margin: 0.6em 0 !important;
  padding-left: 1.6em !important; /* 统一列表缩进 */
}

/* 统一段落上下间距，避免内联 margin 干扰视觉 */
:deep(p) {
  margin: 0.6em 0 !important;
}

/* 列表缩进，避免项目符号跑到可视区外 */
:deep(ul), :deep(ol) {
  margin: 0.6em 0 !important;
  padding-left: 1.2em !important;
}

/* 代码块常见溢出 */
:deep(pre), :deep(code) {
  white-space: pre-wrap;
  word-break: break-word;
}


.editor-container {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow-x: visible;
  overflow-y: hidden;
}

.editor-wrapper {
  flex: 1;
  height: 100%;
  overflow-y: auto;
}

.bottom-buttons {
  display: flex;
  justify-content: space-between; /* 左右分布 */
  align-items: center;
  margin-top: auto;
  padding: 12px 0 0 0;
  background-color: #fff;
}

.left-buttons {
  flex-shrink: 0;
}

.right-buttons {
  flex-shrink: 0;
}

:deep(img) {
  cursor: zoom-in;
}

/* 移动端适配 */
@media (max-width: 768px) {
  /* 缩短内边距 */
  :deep(.w-e-text-container) {
    padding: 0px!important; /* 或者 padding-left: 16px; */
    box-sizing: border-box;
    overflow-x: auto; /* 防止代码块等超长内容被裁掉 */
  }
}

</style>
