<template>
  <Announcement v-for="al in topAlert" :key="al.GUID" :TEXT="al.TEXT" :URL="al.URL" :URLNAME="al.URLNAME"/>

  <el-container class="main-container">
    <el-main class="res-main">
      <div class="left-panel">
        <div class="panel-header">
          <div class="header-title">
            <el-icon class="title-icon"><FolderOpened /></el-icon>
            公共资源库
          </div>
          <el-input
              v-model="searchKeyword"
              placeholder="搜索文件名、备注或上传者..."
              :prefix-icon="Search"
              clearable
              @clear="resetAndLoad"
              @input="debouncedSearch"
              class="search-input"
          />
        </div>

        <el-tabs v-model="activeCategory" class="category-tabs" @tab-click="handleTabClick">
          <el-tab-pane
              v-for="cat in fileCategories"
              :key="cat.key"
              :name="cat.key"
          >
            <template #label>
              <span class="custom-tab-label">
                <el-icon><component :is="cat.icon" /></el-icon>
                <span>{{ cat.label }}</span>
                <el-badge v-if="getCategoryCount(cat.key) > 0" :value="getCategoryCount(cat.key)" class="tab-badge" type="info"/>
              </span>
            </template>
          </el-tab-pane>
        </el-tabs>

        <div class="scroll-panel" v-loading="loading">
          <transition-group name="list" tag="div" class="file-grid">
            <div
                v-for="file in filteredArticles"
                :key="file.GUID"
                class="file-card-v2"
            >
              <div class="file-card-top">
                <div class="file-type-icon">
                  <el-icon :size="40" :color="getFileColor(file.ORIGINALFILENAME)">
                    <component :is="getFileIcon(file.ORIGINALFILENAME)" />
                  </el-icon>
                </div>
                <div class="file-details">
                  <div class="file-title-row">
                    <span class="file-title" :title="file.ORIGINALFILENAME">{{ file.ORIGINALFILENAME }}</span>
                  </div>

                  <div v-if="file.REMARK" class="file-remark-v2">
                    <el-icon><Warning /></el-icon>
                    {{ file.REMARK }}
                  </div>

                  <div class="file-meta-v2">
                    <span class="meta-item"><el-icon><User/></el-icon>{{ file.USERNAME }}</span>
                    <span class="meta-item"><el-icon><Clock/></el-icon>{{ pubFormatDate(file.CREATE_TIME) }}</span>
                    <span class="meta-item" style="color:#409EFF;"><el-icon><Download/></el-icon>{{ file.DOWNNUM }} 次</span>
                  </div>
                </div>
              </div>

              <div class="file-actions-v2">
                <el-button-group>
                  <el-button type="success" :icon="Download" size="small" plain @click="downloadFile(file)">下载</el-button>
                  <el-tooltip content="复制下载链接" placement="top">
                    <el-button type="primary" :icon="CopyDocument" size="small" plain @click="copyFileUrl(file)" />
                  </el-tooltip>
                  <el-button v-if="getCurrentUserAdminObject().isAdmin" type="danger" :icon="Close" size="small" plain @click="deleteFile(file)" />
                </el-button-group>
              </div>
            </div>
          </transition-group>

          <el-empty v-if="!filteredArticles.length && !loading" description="暂无该分类文件"/>

          <div class="list-footer">
            <el-button
                v-if="!noMore && !loading"
                type="primary"
                plain
                round
                @click="fetchArticles"
                class="load-more-btn"
            >
              加载更多
            </el-button>
            <div v-if="noMore && articles.length > 0" class="end-text-v2">- 到底了 -</div>
          </div>
        </div>
      </div>

      <div v-if="userStore.userBean.code" class="right-panel">
        <div class="panel-header">
          <div class="header-title small">
            <el-icon class="title-icon"><UserFilled /></el-icon>
            我的上传
          </div>
        </div>

        <el-upload
            v-if="myFiles.length < 5 || getCurrentUserAdminObject().isAdmin"
            drag
            :http-request="customUploadRequest"
            :before-upload="beforeUploadCheck"
            :show-file-list="false"
            :on-progress="handleUploadProgress"
            class="upload-area-v2"
        >
          <el-icon class="el-icon--upload"><upload-filled /></el-icon>
          <div class="el-upload__text">
            拖拽或 <em>点击上传</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">单个文件不超过 50MB</div>
          </template>

          <el-progress
              v-if="isUploading"
              :percentage="uploadProgress"
              status="success"
              :text-inside="true"
              :stroke-width="18"
              class="upload-progress"
          />
          <p v-if="showProcessing" class="processing-text">服务器处理中,请稍候...</p>
        </el-upload>

        <el-alert
            v-else
            title="已达到上传限制"
            type="warning"
            description="普通用户最多上传 5 个文件。请删除旧文件后再上传。"
            show-icon
            :closable="false"
            class="limit-alert"
        />

        <div class="my-files-list scroll-panel small">
          <transition-group name="list" tag="div">
            <el-card
                v-for="file in myFiles"
                :key="file.GUID"
                shadow="hover"
                class="my-file-card compact-card"
            >
              <div class="my-file-header">
                <el-icon :size="18" :color="getFileColor(file.ORIGINALFILENAME)" class="small-file-icon">
                  <component :is="getFileIcon(file.ORIGINALFILENAME)" />
                </el-icon>
                <span class="file-title" :title="file.ORIGINALFILENAME">{{ file.ORIGINALFILENAME }}</span>
              </div>

              <div v-if="file.REMARK" class="file-remark-v2 muted" style="margin: 5px 0;">
                <el-icon><Warning /></el-icon>
                {{ file.REMARK }}
              </div>

              <div class="card-meta">
                <span>时间: {{ pubFormatDate(file.CREATE_TIME) }}</span>
                <span>下载: {{ file.DOWNNUM }}次</span>
              </div>
              <div class="my-file-actions">
                <el-button type="success" :icon="Download" size="small" circle @click="downloadFile(file)" title="下载"></el-button>
                <el-button type="primary" :icon="CopyDocument" size="small" circle @click="copyFileUrl(file)" title="复制链接"></el-button>
                <el-button type="warning" :icon="EditPen" size="small" circle @click="openEdit(file)" title="编辑信息"></el-button>
                <el-button type="danger" :icon="Close" size="small" circle @click="deleteFile(file)" title="删除"></el-button>
              </div>
            </el-card>
          </transition-group>
        </div>
      </div>

      <div
          v-if="!userStore.userBean.code"
          class="right-panel empty-login"
      >
        <el-empty
            description="登录后可管理个人上传文件"
            :image-size="100"
        >
          <el-button type="primary">去登录</el-button>
        </el-empty>
      </div>
    </el-main>
  </el-container>

  <el-dialog
      v-model="editVisible"
      title="编辑文件信息"
      width="500px"
      @closed="resetEdit"
      class="custom-dialog"
      destroy-on-close
  >
    <el-form :model="editForm" :rules="editRules" ref="editRef" label-position="top">
      <el-form-item label="文件名称（不含后缀）" prop="ORIGINALFILENAME">
        <el-input v-model.trim="editForm.ORIGINALFILENAME" maxlength="150" show-word-limit clearable placeholder="请输入新文件名"/>
      </el-form-item>
      <el-form-item label="备注 / 文件详情" prop="REMARK">
        <el-input
            v-model.trim="editForm.REMARK"
            type="textarea"
            :autosize="{minRows:4,maxRows:8}"
            maxlength="300"
            show-word-limit
            placeholder="添加备注，方便他人了解文件内容（可选）"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingEdit" @click="submitEdit">保存修改</el-button>
      </div>
    </template>
  </el-dialog>

</template>

<script setup>
import {ref, computed,onMounted,watch } from 'vue'
import {useRoute} from "vue-router";
import {ElMessage} from 'element-plus'
// 导入更多图标
import {
  User, Clock, Close, CopyDocument, Download, Search, EditPen,
  FolderOpened, UserFilled, UploadFilled, Warning, Collection,
  Document, Picture, Suitcase, More, VideoPlay, Headset,
  Files, Tickets, DataBoard, Monitor
} from '@element-plus/icons-vue';

import {
  downloadFileByUrl,
  ele_confirm,
  getGuid,
  sendAxiosRequest,
  pubFormatDate,
  uploadFileWithProgress, getCurrentUserAdminObject
} from "@/utils/common.js";
import {useUserStore} from "@/stores/main/user.js";
import Announcement from "@/components/detail/Announcement.vue";
import {getAnnouncementByRouterName} from "@/utils/blogUtil.js";
import debounce from "lodash/debounce.js";

const route = useRoute();
const userStore = useUserStore();

const uploadProgress = ref(0)
const isUploading = ref(false)

// ---------------------------------------------------------
// [新增] 文件分类相关数据定义
// ---------------------------------------------------------
const activeCategory = ref('all') // 当前激活的 Tab

// 定义分类标准和图标
const fileCategories = [
  { key: 'all', label: '全部', icon: Collection, exts: [] },
  { key: 'doc', label: '文档', icon: Document, exts: ['doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'pdf', 'txt', 'md', 'csv'] },
  { key: 'image', label: '图片', icon: Picture, exts: ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg'] },
  { key: 'video', label: '视频', icon: VideoPlay, exts: ['mp4', 'avi', 'mov', 'wmv', 'flv', 'mkv', 'webm'] },
  { key: 'audio', label: '音乐', icon: Headset, exts: ['mp3', 'wav', 'flac', 'aac', 'ogg', 'm4a'] },
  { key: 'archive', label: '压缩包', icon: Suitcase, exts: ['zip', 'rar', '7z', 'tar', 'gz'] },
  { key: 'other', label: '其他', icon: More, exts: [] }
]

/**
 * [新增] 辅助函数：获取文件后缀名
 */
const getFileExt = (filename) => {
  if (!filename || !filename.includes('.')) return '';
  return filename.split('.').pop().toLowerCase();
}

/**
 * [新增] 辅助函数：判断文件属于哪个大类
 */
const getFileCategoryKey = (filename) => {
  const ext = getFileExt(filename);
  if (!ext) return 'other';

  for (const cat of fileCategories) {
    if (cat.key !== 'all' && cat.key !== 'other' && cat.exts.includes(ext)) {
      return cat.key;
    }
  }
  return 'other';
}

/**
 * [新增] 计算属性：依据当前选择的 Tab 过滤左侧文件列表
 * 这里采用前端筛选已加载数据的方式。
 */
const filteredArticles = computed(() => {
  if (activeCategory.value === 'all') {
    return articles.value;
  }
  return articles.value.filter(file => {
    return getFileCategoryKey(file.ORIGINALFILENAME) === activeCategory.value;
  });
});

/**
 * [新增] 辅助函数：获取统计数量
 */
const getCategoryCount = (key) => {
  if (key === 'all') return articles.value.length;
  return articles.value.filter(file => getFileCategoryKey(file.ORIGINALFILENAME) === key).length;
}

const handleTabClick = (tab) => {
  // 纯前端筛选，切换 tab 不需要重新请求后端
  console.log('切换分类:', tab.props.name);
}

/**
 * 依据后缀名获取 Element Plus 图标组件
 */
const getFileIcon = (filename) => {
  const ext = getFileExt(filename);
  const map = {
    doc: Document, docx: Document, txt: Document, md: Document,
    xls: Tickets, xlsx: Tickets, csv: Tickets, // 用 Tickets 代表表格类
    ppt: DataBoard, pptx: DataBoard,           // 用 DataBoard 代表演示文稿
    pdf: Document,
    zip: Suitcase, rar: Suitcase, '7z': Suitcase, tar: Suitcase, gz: Suitcase,
    html: Monitor, css: Monitor, js: Monitor,
    jpg: Picture, jpeg: Picture, png: Picture, gif: Picture, webp: Picture, svg: Picture,
    mp3: Headset, wav: Headset, flac: Headset, aac: Headset,
    mp4: VideoPlay, avi: VideoPlay, mov: VideoPlay, wmv: VideoPlay
  };

  return map[ext] || Files; // 匹配不到则返回默认的 Files 图标
}

/**
 * 依据后缀名获取图标颜色（模仿各类文件的经典配色）
 */
const getFileColor = (filename) => {
  const ext = getFileExt(filename);
  const colorMap = {
    doc: '#409EFF', docx: '#409EFF',          // Word 经典蓝
    xls: '#67C23A', xlsx: '#67C23A', csv: '#67C23A', // Excel 经典绿
    ppt: '#E6A23C', pptx: '#E6A23C',          // PPT 经典橙
    pdf: '#F56C6C',                           // PDF 经典红
    txt: '#909399', md: '#303133',            // 文本 灰色/黑色
    zip: '#8A2BE2', rar: '#8A2BE2', '7z': '#8A2BE2', // 压缩包 紫色
    jpg: '#E6A23C', jpeg: '#E6A23C', png: '#E6A23C', // 图片 橙黄色
    mp4: '#409EFF', avi: '#409EFF',           // 视频 蓝色
    mp3: '#67C23A', wav: '#67C23A'            // 音频 绿色
  };

  return colorMap[ext] || '#909399'; // 默认返回灰色
}

// ---------------------------------------------------------
// 原有逻辑保持不变
// ---------------------------------------------------------

onMounted(() => {
  let fileGuid = route.query.g || "";
  if(fileGuid){
    setFilesById(fileGuid);
  }else{
    fetchArticles();
  }
})

async function setFilesById(guid){
  loading.value = true;
  let result = await sendAxiosRequest("/blog-api/resource/getFileById",{guid});
  loading.value = false;
  if(result&&!result.isError){
    articles.value = result.result;
    noMore.value = true; // 单个文件查找，直接设为到底
  }
}

// 文件列表数据
const articles = ref([])
const page = ref(1)
const pageSize = 15 // 稍微调大一点，配合卡片布局
const loading = ref(false)
const noMore = ref(false)

// 搜索文件
const searchKeyword = ref('')

const resetAndLoad = () => {
  page.value = 1
  noMore.value = false
  articles.value = []
  fetchArticles()
}

const debouncedSearch = debounce(() => {
  resetAndLoad()
}, 1000)

const fetchArticles = async () => {
  if (loading.value || noMore.value) return
  loading.value = true
  try {
    const res = await sendAxiosRequest('/blog-api/resource/getAllFile', {
      page: page.value,
      pageSize,
      keyword: searchKeyword.value
    })
    const newData = res.result.data;
    if (newData.length < pageSize) {
      noMore.value = true
    }
    articles.value.push(...newData)
    page.value++
  } catch (e) {
    console.error('获取文件失败', e)
    ElMessage.error("获取文件列表失败")
  } finally {
    loading.value = false
  }
}

const myFiles = ref([]);

//获取用户本身上传的附件
const setMyFileData = async () => {
  if(!userStore.userBean.code) return false;
  let result = await sendAxiosRequest("/blog-api/resource/getFileByUser",{userCode:userStore.userBean.code});
  if (result && !result.isError) {
    myFiles.value = result.result;
  }
}

watch(
    () => userStore.userBean && userStore.userBean.code,
    (code) => {
      if (code) {
        setMyFileData();
      } else {
        myFiles.value = []; // 注销时清空
      }
    },
    { immediate: true }
)


const beforeUploadCheck = (file) => {
  let isSuccess = true;
  const maxSizeMB = 50;
  const isLtMaxSize = file.size / 1024 / 1024 < maxSizeMB;
  if (!isLtMaxSize) {
    ElMessage.error(`上传文件不能超过 ${maxSizeMB}MB`);
    isSuccess = false;
  }
  if (isSuccess) {
    ElMessage.info(`正在准备上传...`);
  }
  return isSuccess;
}

const handleUploadProgress = (event) => {
  isUploading.value = true;
  uploadProgress.value = Math.floor(event.percent);
}

const showProcessing = ref(false);

const customUploadRequest = ({file, onProgress, onSuccess, onError}) => {
  isUploading.value = true;
  uploadProgress.value = 0;
  showProcessing.value = false;

  uploadFileWithProgress({
    url: "/pub-api/upload/uploadFile",
    file,
    fieldName: "file",
    extraData: {spliceUrl: "resourcesFile"},
    onProgress: (percent) => {
      uploadProgress.value = percent;
      // onProgress({percent}); // Element Plus 原有回调可能有抖动，这里自定义处理
      if (percent >= 100) {
        showProcessing.value = true; // 显示后端处理中
      }
    },
    onSuccess: (res) => {
      isUploading.value = false;
      uploadProgress.value = 0;
      showProcessing.value = false;
      handleUploadSuccess(res, file)
    },
    onError: (err) => {
      isUploading.value = false;
      uploadProgress.value = 0;
      showProcessing.value = false;
      ElMessage.error(err || "上传失败");
      onError(err);
    },
  });
};

const handleUploadSuccess = (res, file) => {
  ElMessage.success(`《${file.name}》 上传成功！`)
  let fileInfo = {
    GUID: getGuid(),
    ORIGINALFILENAME: res.result.originalFileName,
    SAVEDFILENAME: res.result.savedFileName,
    FILEREALURL: res.result.fileRealUrl,
    FILEVIEWURL: res.result.fileViewUrl,
    USERCODE: userStore.userBean.code,
    USERNAME: userStore.userBean.name,
    REMARK: '', // 初始化备注为空
  }
  // 提交到后端保存记录
  sendAxiosRequest("/blog-api/resource/addFileInfo", {fileInfo});

  // 乐观更新：添加到“我的上传”列表头部
  myFiles.value.unshift({
    ...fileInfo,
    CREATE_TIME: new Date().getTime(), // 使用当前时间戳，格式化由 pubFormatDate 处理
    DOWNNUM: '0'
  })
}

const downloadFile = (file) => {
  file.DOWNNUM++;
  ElMessage.info(`开始下载...`);
  downloadFileByUrl(file.FILEVIEWURL, file.ORIGINALFILENAME)
  sendAxiosRequest("/blog-api/resource/setFileDownNum", {guid: file.GUID});
}

const deleteFile = (file) => {
  ele_confirm(`确定要删除文件《${file.ORIGINALFILENAME}》吗？删除后不可恢复！`, () => {
    //排除右侧我的文件
    myFiles.value = myFiles.value.filter(item => item.GUID !== file.GUID);
    //排除左侧所有文件
    articles.value = articles.value.filter(item => item.GUID !== file.GUID);
    ElMessage.success("删除成功！");
    sendAxiosRequest("/blog-api/resource/delFileInfo", {guid: file.GUID, url: file.FILEVIEWURL});
  });
}

const copyFileUrl = (file) => {
  const domain = window.location.origin;
  const fullUrl = domain + file.FILEVIEWURL;

  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(fullUrl).then(() => {
      ElMessage.success("下载链接已复制到剪贴板");
    }).catch(err => {
      fallbackCopy(fullUrl);
    });
  } else {
    fallbackCopy(fullUrl);
  }
}

// 兼容性的复制方法
const fallbackCopy = (text) => {
  const input = document.createElement('input');
  input.value = text;
  document.body.appendChild(input);
  input.select();
  try {
    document.execCommand('copy');
    ElMessage.success("复制成功");
  } catch (err) {
    ElMessage.error("复制失败，请手动复制");
  }
  document.body.removeChild(input);
}

//编辑附件区域---------------------------------------------------------------------------
// 编辑弹窗相关
const editVisible = ref(false)
const savingEdit = ref(false)
const editRef = ref()
const editingItem = ref(null) // 当前编辑的那条记录引用

const editForm = ref({
  GUID: '',
  ORIGINALFILENAME: '',
  REMARK: ''
})

const editRules = {
  ORIGINALFILENAME: [
    { required: true, message: '请输入文件名称', trigger: 'blur' },
    { min: 1, max: 150, message: '长度在 1 到 150 个字符', trigger: 'blur' }
  ],
  REMARK: [
    { max: 300, message: '最多 300 个字符', trigger: 'blur' }
  ]
}

function openEdit(file) {
  editingItem.value = file

  // 处理文件名，去掉后缀显示给用户编辑
  const full_name = file.ORIGINALFILENAME || '';
  const last_dot = full_name.lastIndexOf('.');
  let name_no_ext = full_name;
  let ext = '';
  if(last_dot > 0) {
    name_no_ext = full_name.substring(0, last_dot);
    ext = full_name.substring(last_dot); // 保存后缀
  }
  editingItem.value._ext = ext; // 临时保存后缀，提交时拼接

  editForm.value.GUID = file.GUID
  editForm.value.ORIGINALFILENAME = name_no_ext
  editForm.value.REMARK = file.REMARK || ''
  editVisible.value = true
}

function resetEdit() {
  editingItem.value = null
  editRef.value?.clearValidate?.()
}

async function submitEdit() {
  if(!editRef.value) return;
  try {
    await editRef.value.validate()
  } catch (fields) {
    return
  }

  savingEdit.value = true

  // 拼接回原后缀
  const newFileNameWithExt = editForm.value.ORIGINALFILENAME + (editingItem.value?._ext || '');

  try {
    // 后端接口调用
    const payload = {
      guid: editForm.value.GUID,
      originalFileName: newFileNameWithExt, // 传带后缀的全名
      remark: editForm.value.REMARK
    }
    const res = await sendAxiosRequest('/blog-api/resource/updateFileInfo', payload)

    if (res && !res.isError) {
      // 乐观更新：更新本地 myFiles 列表中的数据
      if (editingItem.value) {
        editingItem.value.ORIGINALFILENAME = newFileNameWithExt
        editingItem.value.REMARK = editForm.value.REMARK
      }

      // 同步更新 articles 列表中的对应数据（如果存在）
      const articleIdx = articles.value.findIndex(x => x.GUID === editForm.value.GUID);
      if(articleIdx > -1) {
        articles.value[articleIdx].ORIGINALFILENAME = newFileNameWithExt;
        articles.value[articleIdx].REMARK = editForm.value.REMARK;
      }

      ElMessage.success('信息更新成功!')
      editVisible.value = false
    } else {
      ElMessage.error(res?.errMsg || '保存失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('网络请求失败')
  } finally {
    savingEdit.value = false
  }
}
//公告横幅内容
const topAlert = ref([]);
const setTopAlert = async () => {
  topAlert.value = await getAnnouncementByRouterName("Resources");
}
setTopAlert();
</script>

<style scoped>
/* ---------------------------------------------------------
   整体布局与容器
   --------------------------------------------------------- */
.main-container {
  background-color: #f5f7fa; /* 浅灰色背景，突出卡片 */
  min-height: 100vh;
}

.res-main {
  display: flex;
  gap: 24px; /* 加大间距 */
  padding: 24px;
  box-sizing: border-box;
  align-items: flex-start; /* 顶部对齐，防止右侧被拉长 */
}

/* 统一的面板样式 */
.left-panel, .right-panel {
  background: #fff;
  border-radius: 16px; /* 加大圆角 */
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.05); /* 柔和阴影 */
  display: flex;
  flex-direction: column;
  overflow: hidden; /* 保持圆角 */
  transition: all 0.3s ease;
}

.left-panel {
  flex: 8;
}

.right-panel {
  flex: 2;
  position: sticky; /* 右侧悬浮 */
  top: 20px;
  max-height: calc(100vh - 40px);
}

/* 区域头部样式 */
.panel-header {
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 15px;
}

.header-title {
  font-size: 20px;
  font-weight: 600;
  color: #1f2f3d;
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-title.small {
  font-size: 17px;
}

.title-icon {
  color: #409EFF;
}

.search-input {
  width: 300px;
}
.search-input :deep(.el-input__wrapper) {
  border-radius: 20px; /* 圆角搜索框 */
}

/* ---------------------------------------------------------
   Tabs 分类样式
   --------------------------------------------------------- */
.category-tabs {
  padding: 0 24px;
  margin-top: 10px;
}

.category-tabs :deep(.el-tabs__header) {
  margin: 0;
  border-bottom: none;
}

.category-tabs :deep(.el-tabs__nav-wrap::after) {
  display: none; /* 去掉底部分割线 */
}

.category-tabs :deep(.el-tabs__item) {
  height: 50px;
  font-size: 15px;
}

.custom-tab-label {
  display: flex;
  align-items: center;
  gap: 6px;
}

.tab-badge {
  margin-left: 4px;
}
.tab-badge :deep(.el-badge__content) {
  background-color: #f4f4f5;
  color: #909399;
  border: none;
}
.el-tabs__item.is-active .tab-badge :deep(.el-badge__content) {
  background-color: #409EFF;
  color: #fff;
}

/* ---------------------------------------------------------
   左侧文件卡片 V2版 (网格布局)
   --------------------------------------------------------- */
.scroll-panel {
  overflow-y: auto;
  padding: 15px 24px 24px 24px;
  flex-grow: 1;
}

.left-panel .scroll-panel {
  height: calc(100vh - 180px); /* 视情况调整高度 */
}

/* 新增网格容器：一行两列 */
.file-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.file-card-v2 {
  display: flex;
  flex-direction: column; /* 垂直排列 */
  justify-content: space-between; /* 撑开上下，按钮靠底 */
  padding: 16px;
  background-color: #fff;
  border: 1px solid #e9e9e9;
  border-radius: 12px;
  transition: all 0.3s ease;
  position: relative;
  height: 100%;
  box-sizing: border-box;
}

.file-card-v2:hover {
  border-color: #c6e2ff;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.1);
  transform: translateY(-2px);
}

.file-card-top {
  display: flex;
  align-items: flex-start;
  width: 100%;
}

.file-type-icon {
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  margin-right: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.file-type-icon img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.file-details {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.file-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.file-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  word-break: break-all;
}

/* 备注 V2 醒目样式 */
.file-remark-v2 {
  font-size: 13px;
  color: #e6a23c;
  background-color: #fdf6ec;
  padding: 4px 8px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 5px;
  width: fit-content;
  max-width: 100%;

  /* 多行文本省略 */
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-all;
}

.file-remark-v2.muted {
  color: #909399;
  background-color: #f4f4f5;
}

.file-meta-v2 {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  font-size: 13px;
  color: #909399;
  margin-top: 2px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 5px;
}

/* 动作按钮放置在右下角 */
.file-actions-v2 {
  margin-top: 15px;
  padding-top: 12px;
  border-top: 1px dashed #ebeef5; /* 虚线分割 */
  display: flex;
  justify-content: flex-end; /* 靠右对齐 */
  width: 100%;
}

/* 列表底部区域 */
.list-footer {
  margin-top: 20px;
  text-align: center;
  padding-bottom: 10px;
  /* 占满整行 */
  grid-column: 1 / -1;
}

.load-more-btn {
  width: 120px;
}

.end-text-v2 {
  color: #c0c4cc;
  font-size: 13px;
}

/* Loading 遮罩层圆角 */
:deep(.el-loading-mask) {
  border-radius: 0 0 16px 16px;
}

/* ---------------------------------------------------------
   右侧上传与个人中心 V2版
   --------------------------------------------------------- */
.upload-area-v2 {
  padding: 20px 24px 10px 24px;
}

.upload-area-v2 :deep(.el-upload-dragger) {
  padding: 20px;
  border-radius: 12px;
  background-color: #fafafa;
}

.upload-area-v2 :deep(.el-icon--upload) {
  font-size: 48px;
  margin-bottom: 10px;
  color: #a0cfff;
}

.processing-text {
  font-size: 13px;
  color: #e6a23c;
  margin-top: 8px;
  text-align: center;
}

.limit-alert {
  margin: 20px 24px 10px 24px;
  border-radius: 8px;
}

.my-files-list {
  padding: 10px 24px 20px 24px;
  height: calc(100vh - 400px);
}

.my-file-card {
  border-radius: 10px;
  margin-bottom: 10px;
  border: 1px solid #ebeef5;
  transition: all 0.3s;
}
.my-file-card:hover {
  border-color: #dcdfe6;
  background-color: #fafafa;
}

.my-file-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.small-file-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.my-file-card .file-title {
  font-size: 14px;
  font-weight: 600;
}

.my-file-card .card-meta {
  font-size: 12px;
  color: #909399;
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.my-file-actions {
  display: flex;
  justify-content: flex-end;
  gap: 5px;
  border-top: 1px solid #f0f0f0;
  padding-top: 8px;
}

.empty-login {
  justify-content: center;
  align-items: center;
  padding: 40px;
}

/* ---------------------------------------------------------
   弹窗与列表动画
   --------------------------------------------------------- */
/* 列表过渡动画 */
.list-enter-active,
.list-leave-active {
  transition: all 0.4s ease;
}
.list-enter-from,
.list-leave-to {
  opacity: 0;
  transform: translateX(30px);
}

/* 弹窗样式 */
.custom-dialog :deep(.el-dialog__header) {
  margin-right: 0;
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
}
.custom-dialog :deep(.el-dialog__title) {
  font-weight: 600;
}
.custom-dialog :deep(.el-dialog__body) {
  padding: 24px;
}
.custom-dialog :deep(.el-dialog__footer) {
  padding: 10px 24px 20px;
  border-top: 1px solid #f0f0f0;
}

/* ---------------------------------------------------------
   响应式：手机端适配
   --------------------------------------------------------- */
@media (max-width: 992px) {
  .res-main {
    flex-direction: column;
    padding: 15px;
    gap: 15px;
  }

  .left-panel, .right-panel {
    width: 100% !important;
    flex: none !important;
    position: static !important; /* 取消悬浮 */
    max-height: none !important;
  }

  .panel-header {
    padding: 15px;
  }

  .search-input {
    width: 100%;
  }

  .scroll-panel {
    padding: 10px 15px 15px 15px;
  }

  /* 手机端屏幕较窄，强制一行显示一个文件 */
  .file-grid {
    grid-template-columns: 1fr;
  }
}
</style>