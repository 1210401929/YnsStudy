<template>
  <div class="editor-container">
    <el-radio-group v-if="!isReadOnly" v-model="isPublic" style="margin-bottom: 10px">
      <el-radio-button :label="true">公开</el-radio-button>
      <el-radio-button :label="false">私密</el-radio-button>
    </el-radio-group>
    <el-input
        v-if="!isReadOnly"
        v-model="localTitle"
        placeholder="请输入文章标题"
        size="large"
        style="margin-bottom: 10px"
    />

    <div :style="wrapperStyle" class="editor-wrapper">
      <!-- 工具栏 -->
      <Toolbar
          v-if="!isReadOnly"
          :editor="editorRef"
          :defaultConfig="toolbarConfig"
          style="border-bottom: 1px solid #ccc"
      />

      <!-- 编辑器包一层容器，用于监听双击 -->
      <div ref="editorWrapper">
        <Editor
            v-model="localContent"
            :defaultConfig="editorConfig"
            :style="editorStyle"
            @onCreated="handleCreated"
        />
      </div>
    </div>

    <!-- 相对于弹窗定位的右下角按钮 -->
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
import {ref, computed, shallowRef, watch, onBeforeUnmount, onMounted, onUnmounted} from 'vue'
import {Editor, Toolbar} from '@wangeditor/editor-for-vue'
import {sendAxiosRequest} from '@/utils/common.js'
import {ElMessage} from 'element-plus'

const props = defineProps({
  title: String,
  content: String,
  isPublic: Boolean,
  isReadOnly: Boolean,
  saveType: String
})

const emits = defineEmits(['submit', 'cancel'])
const isPublic = ref(props.isPublic)
const localTitle = ref(props.title || '')
const localContent = ref(props.content || '')
const editorRef = shallowRef(null)
const editorWrapper = ref(null)

const toolbarConfig = {
  excludeKeys: ['insertVideo', 'uploadVideo']
}

const editorConfig = {
  placeholder: '请输入正文...',
  readOnly: props.isReadOnly,
  MENU_CONF: {
    uploadImage: {
      async customUpload(file, insertFn) {
        const formData = new FormData()
        formData.append('file', file)
        formData.append('spliceUrl', 'editorImage')
        const result = await sendAxiosRequest('/pub-api/upload/uploadFile', formData)
        if (result && !result.isError) {
          insertFn(result.result['fileViewUrl'])
        } else {
          ElMessage.error(result.errMsg)
        }
      }
    },
    lineHeight: {
      lineHeightList: ['0.5', '1', '1.5', '2', '2.5'],
      defaultValue: '0.5' // 默认行高
    }
  }
}

// 图片查看器
const showViewer = ref(false)
const viewerUrl = ref('')

function showImagePreview(url) {
  viewerUrl.value = url
  showViewer.value = true
}

function onDblClick(e) {
  const target = e.target
  if (target.tagName === 'IMG') {
    showImagePreview(target.src)
  }
}

function handleCreated(editor) {
  editorRef.value = editor
}

function submit() {
  if (!localTitle.value.trim()) {
    ElMessage.error('请输入标题')
    return
  }
  emits('submit', {
    blog_type: isPublic.value ? "public" : "privacy",
    title: localTitle.value.trim(),
    content: localContent.value.trim()
  })
  //如果是新增,则提交操作之后,清除新增得内容
  if (props.saveType == "add") {
    localTitle.value = ''
    localContent.value = ''
  }
}

const wrapperStyle = computed(() => ({
  border: '0px solid #ccc',
  marginBottom: '0px',
  minHeight: props.isReadOnly ? '300px' : 'auto',
}))

const editorStyle = computed(() => ({
  minHeight: props.isReadOnly ? '300px' : 'auto',
  height: props.isReadOnly ? '100%' : '420px',
  overflowY: 'auto',
}))

onMounted(() => {
  if (editorWrapper.value) {
    editorWrapper.value.addEventListener('dblclick', onDblClick)
  }
})

onUnmounted(() => {
  if (editorWrapper.value) {
    editorWrapper.value.removeEventListener('dblclick', onDblClick)
  }
})

onBeforeUnmount(() => {
  if (editorRef.value) {
    editorRef.value.destroy()
    editorRef.value = null
  }
})

watch(() => props.title, (val) => {
  localTitle.value = val
})

watch(() => props.content, (val) => {
  localContent.value = val
})

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
