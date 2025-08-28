<template>
  <!-- 公告横幅 -->
  <Announcement v-for="al in topAlert" :key="al.GUID" :TEXT="al.TEXT" :URL="al.URL" :URLNAME="al.URLNAME" />

  <el-container style="height: 100vh; padding: 20px; box-sizing: border-box;">
    <el-main style="display: flex; gap: 20px;">
      <!-- 左侧：他人上传的文件 -->
      <div style="flex: 7; display: flex; flex-direction: column;">
        <el-input
            v-model="searchKeyword"
            placeholder="搜索他人上传的文件"
            :prefix-icon="Search"
            clearable
            style="margin-bottom: 10px;"
        />
        <div class="scroll-panel">
          <div
              v-for="file in filteredOtherFiles"
              :key="file.GUID"
              class="file-card"
          >
            <div class="card-content">
              <div class="file-title">{{ file.ORIGINALFILENAME }}</div>
              <div class="file-meta">
                <span class="meta-item">
                  <el-icon><User/></el-icon>
                    {{ file.USERNAME }}
                </span>
                <span class="meta-item">
                  <el-icon><Clock/></el-icon>
                    {{ pubFormatDate(file.CREATE_TIME) }}
                </span>
                <span class="meta-item">
                  <el-icon><Download/></el-icon>
                    {{ file.DOWNNUM }} 次
                </span>
              </div>
            </div>
            <div class="card-actions">
              <el-button
                  type="success"
                  :icon="Download"
                  size="small"
                  @click="downloadFile(file)"
              >
                下载
              </el-button>
              <el-button
                  type="primary"
                  :icon="CopyDocument"
                  size="small"
                  @click="copyFileUrl(file)"
              >
                复制下载连接
              </el-button>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧：我上传的文件 -->
      <div v-if="userStore.userBean.code" style="flex: 3; display: flex; flex-direction: column;">
        <el-upload
            v-if="myFiles.length < 5"
            drag
            :http-request="customUploadRequest"
            :before-upload="beforeUploadCheck"
            :show-file-list="false"
            :on-progress="handleUploadProgress"
            style="margin-bottom: 10px;"
        >
          <i class="el-icon-upload" style="font-size: 28px; color: #409EFF;"></i>
          <div class="el-upload__text">点击或拖拽上传文件</div>

          <el-progress
              v-if="isUploading"
              :percentage="uploadProgress"
              status="success"
              :text-inside="true"
              :stroke-width="18"
              style="width: 100%; margin-top: 10px;"
          />
          <p v-if="showProcessing">上传成功,后台处理中请耐心等待...</p>
        </el-upload>
        <!-- 超过上传限制的提示 -->
        <el-alert
            v-else
            title="已达到上传限制"
            type="warning"
            description="最多上传 5 个文件，如需上传新文件，请先删除部分旧文件。"
            show-icon
            center
            style="margin-bottom: 10px;"
        />

        <div class="scroll-panel">
          <el-row :gutter="12">
            <el-col
                v-for="file in myFiles"
                :key="file.GUID"
                :span="24"
            >
              <el-card shadow="hover" class="file-card compact-card">
                <div class="card-title">{{ file.ORIGINALFILENAME }}</div>
                <div class="card-meta">上传时间：{{ file.CREATE_TIME }}</div>
                <div class="card-meta">下载次数：{{ file.DOWNNUM }}次</div>
                <el-button
                    type="success"
                    :icon="Download"
                    size="small"
                    @click="downloadFile(file)"
                >
                  下载
                </el-button>
                <el-button
                    type="primary"
                    :icon="CopyDocument"
                    size="small"
                    @click="copyFileUrl(file)"
                >
                  复制下载连接
                </el-button>
                <el-button
                    type="danger"
                    :icon="Close"
                    size="small"
                    @click="deleteFile(file)"
                >
                  删除
                </el-button>
              </el-card>
            </el-col>
          </el-row>
        </div>
      </div>
      <div
          v-if="!userStore.userBean.code"
          style="flex: 3; display: flex; justify-content: center; align-items: center;"
      >
        <el-empty
            description="登录后可上传文件"
            :image-size="120"
        >
        </el-empty>
      </div>
    </el-main>
  </el-container>
</template>

<script setup>
import {ref, computed, onMounted} from 'vue'
import {useRoute} from "vue-router";
import {ElMessage} from 'element-plus'
import {User, Clock, Close, CopyDocument, Download, Search} from '@element-plus/icons-vue'
import {
  downloadFileByUrl,
  ele_confirm,
  getGuid,
  sendAxiosRequest,
  pubFormatDate,
  uploadFileWithProgress
} from "@/utils/common.js";
import {useUserStore} from "@/stores/main/user.js";
import Announcement from "@/components/detail/Announcement.vue";
import {getAnnouncementByRouterName} from "@/utils/blogUtil.js";

const route = useRoute();
const userStore = useUserStore();

const searchKeyword = ref('')
const otherFiles = ref([])
const myFiles = ref([])

const uploadProgress = ref(0)
const isUploading = ref(false)

onMounted(() => {
  setFileData()
})

const setFileData = async () => {
  let result = await sendAxiosRequest("/blog-api/resource/getAllFile");
  if (result && !result.isError) {
    otherFiles.value = result.result.filter(item => item["USERCODE"] != userStore.userBean.code);
    myFiles.value = result.result.filter(item => item["USERCODE"] == userStore.userBean.code);
    setFileDataByRouterPms();
  }
}

const setFileDataByRouterPms = () => {
  let fieldGuid = route.query.g;
  if (fieldGuid) {
    otherFiles.value = otherFiles.value.filter(item => item.GUID == fieldGuid);
    myFiles.value = myFiles.value.filter(item => item.GUID == fieldGuid);
  }
}

const beforeUploadCheck = (file) => {
  let isSuccess = true;
  const maxSizeMB = 50;
  const isLtMaxSize = file.size / 1024 / 1024 < maxSizeMB;
  if (!isLtMaxSize) {
    ElMessage.error(`上传文件不能超过 ${maxSizeMB}MB`);
    isSuccess = false;
  }
  if (isSuccess) {
    ElMessage.success(`上传文件需要时间，请耐心等待`);
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

  uploadFileWithProgress({
    url: "/pub-api/upload/uploadFile",
    file,
    fieldName: "file",
    extraData: {spliceUrl: "resourcesFile"},
    onProgress: (percent) => {
      uploadProgress.value = percent;
      onProgress({percent});
      if (percent >= 100) {
        showProcessing.value = true;
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
      showProcessing.value;
      ElMessage.error(err || "上传失败");
      onError(err);
    },
  });
};

const handleUploadSuccess = (res, file) => {
  isUploading.value = false;
  uploadProgress.value = 0;
  ElMessage.success(`上传成功：${file.name}`)
  let fileInfo = {
    GUID: getGuid(),
    ORIGINALFILENAME: res.result.originalFileName,
    SAVEDFILENAME: res.result.savedFileName,
    FILEREALURL: res.result.fileRealUrl,
    FILEVIEWURL: res.result.fileViewUrl,
    USERCODE: userStore.userBean.code,
    USERNAME: userStore.userBean.name,
  }
  sendAxiosRequest("/blog-api/resource/addFileInfo", {fileInfo});
  myFiles.value.unshift({
    GUID: fileInfo.GUID,
    ORIGINALFILENAME: fileInfo.ORIGINALFILENAME,
    CREATE_TIME: `刚刚`,
    FILEVIEWURL: fileInfo.FILEVIEWURL,
    DOWNNUM: '0'
  })
}

const downloadFile = (file) => {
  file.DOWNNUM++;
  ElMessage.success(`下载文件需要时间，请耐心等待!`);
  downloadFileByUrl(file.FILEVIEWURL, file.ORIGINALFILENAME)
  sendAxiosRequest("/blog-api/resource/setFileDownNum", {guid: file.GUID});
}

const deleteFile = (file) => {
  ele_confirm("是否确认删除，删除后不可恢复！", () => {
    myFiles.value = myFiles.value.filter(item => item.GUID != file.GUID);
    ElMessage.success("删除成功！");
    sendAxiosRequest("/blog-api/resource/delFileInfo", {guid: file.GUID, url: file.FILEVIEWURL});
  });
}

const copyFileUrl = (file) => {
  const domain = window.location.origin;
  const input = document.createElement('input');
  input.value = domain + file.FILEVIEWURL;
  document.body.appendChild(input);
  input.select();
  try {
    document.execCommand('copy');
    ElMessage.success("复制成功");
  } catch (err) {
    ElMessage.error("复制失败：" + err);
  }
  document.body.removeChild(input);
}

const filteredOtherFiles = computed(() =>
    searchKeyword.value
        ? otherFiles.value.filter(f => f.ORIGINALFILENAME.toLowerCase().includes(searchKeyword.value.toLowerCase()))
        : otherFiles.value
)
//公告横幅内容
const topAlert = ref([]);
const setTopAlert = async ()=>{
  topAlert.value = await getAnnouncementByRouterName("Resources");
}
setTopAlert();
</script>

<style scoped>
.file-card {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;

  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  padding: 16px 20px;
  margin-bottom: 14px;
  transition: box-shadow 0.2s;
}

.file-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.file-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  word-break: break-all;
}

.file-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 13px;
  color: #606266;
  margin-top: 4px;
  word-break: break-all;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
}

.meta-item .el-icon {
  font-size: 14px;
  color: #909399;
}

.card-actions {
  flex-shrink: 0;
}

.compact-card {
  padding: 12px;
}

.card-title {
  font-weight: 600;
  font-size: 14px;
  color: #303133;
  margin-bottom: 4px;
}

.card-meta {
  font-size: 12px;
  color: #909399;
  margin-bottom: 6px;
}

.scroll-panel {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden; /* ✅ 禁止横向滚动 */
  max-height: calc(100vh - 180px);
  padding-right: 4px;
}

@media (max-width: 768px) {
  /* 主体区域垂直排列 */
  .el-main {
    flex-direction: column !important;
  }

  /* 两个主区域宽度设为 100% */
  .el-main > div {
    width: 100% !important;
    flex: none !important;
  }

  /* 卡片布局调整 */
  .file-card {
    flex-direction: column;
    align-items: flex-start;
    padding: 12px 14px;
  }

  .card-actions {
    margin-top: 10px;
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  .scroll-panel {
    max-height: none;
    height: auto;
    padding-right: 0;
  }

  .el-upload {
    width: 100%;
  }

  .el-card {
    width: 100%;
  }

  .el-input {
    width: 100%;
  }
}

</style>
