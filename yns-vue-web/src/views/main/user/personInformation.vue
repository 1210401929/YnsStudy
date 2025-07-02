<template>
  <div class="person-info">
    <div class="inner-container">
    <!-- 顶部跳转按钮 -->
    <div class="toolbar">
      <el-button type="success" size="small" @click="scrollToFiles">
        📁 跳转到文件列表
      </el-button>
    </div>

    <!-- 用户信息卡片 -->
    <el-card class="profile-card">
      <div class="profile-header">
        <!-- 左侧：头像和基础信息 -->
        <div class="left-info">
          <el-avatar
              :src="user.avatar"
              size="large"
              class="author-avatar"
              alt="用户头像"
          >
            {{ user.name?.charAt(0) }}
          </el-avatar>
          <div class="profile-details">
            <h2 class="username">
              <span class="name">{{ user.name }}</span>
              <span v-if="user.code == adminUserCode" class="admin-badge">管理员</span>
            </h2>
            <p>邮箱: {{ user.email || "未知" }}</p>
            <p class="userip">IP: {{ user.loginaddress || "未知" }}</p>
          </div>
        </div>

        <!-- 右侧：互动信息 -->
        <div class="right-info">
          <div class="stats">
            <span>关注 <strong>0</strong></span>
            <span>粉丝 <strong>0</strong></span>
          </div>
          <el-button
              :type="isFollowing ? 'danger' : 'primary'"
              size="small"
              round
              @click="toggleFollow"
              class="follow-button"
          >
            {{ isFollowing ? '取消关注' : '关注' }}
          </el-button>
          <div class="interaction-buttons">
            <el-button type="primary" link @click="goToLiked">
              👍 点赞数:{{likeNum}}
            </el-button>
            <el-button type="warning" link @click="goToFavorites">
              ⭐ 收藏数:{{collectNum}}
            </el-button>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 博客内容 -->
    <el-card class="section-card">
      <h3 class="section-title">📚 发表内容</h3>
      <el-empty v-if="blogs.length === 0" description="暂无发表内容" />
      <el-timeline v-else>
        <el-timeline-item
            v-for="(blog, index) in displayedBlogs"
            :key="blog.GUID"
            :timestamp="formatDate(blog.CREATE_TIME)"
            placement="top"
            type="primary"
        >
          <el-card shadow="hover" class="blog-card"  @click="blogMainClick(blog)">
            <h4>{{ blog.BLOG_TITLE }}</h4>
            <p class="blog-summary" v-html="blog.MAINTEXT"></p>
          </el-card>
        </el-timeline-item>
      </el-timeline>

      <div v-if="blogs.length > 3" class="collapse-toggle">
        <el-button type="text" @click="showAllBlogs = !showAllBlogs">
          {{ showAllBlogs ? '收起博客内容' : '查看全部博客内容' }}
        </el-button>
      </div>
    </el-card>

    <!-- 上传的文件 -->
    <el-card class="section-card" ref="fileSection">
      <h3 class="section-title">📁 上传的文件</h3>
      <el-empty v-if="files.length === 0" description="暂无上传文件" />
      <el-table
          v-else
          :data="files"
          stripe
          border
          style="width: 100%"
          class="file-table"
      >
        <el-table-column prop="ORIGINALFILENAME" label="文件名" />
        <el-table-column prop="DOWNNUM" label="下载次数" width="100" />
        <el-table-column
            prop="CREATE_TIME"
            label="上传时间"
            :formatter="(row) => formatDate(row.CREATE_TIME)"
        />
        <el-table-column label="操作" width="100">
          <template #default="scope">
            <el-button type="primary" link @click="downloadFile(scope.row)">
              下载
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
  </div>
  <el-dialog
      v-model="showDialog"
      width="80%"
      destroy-on-close
      top="4vh"
      @close="onDialogClose"
  >
    <ContentAndComment  :blogId="selectedBlogId" />
  </el-dialog>
  <!-- 通用卡片结构（点赞/收藏列表都用） -->
  <el-dialog v-model="showLikeDialog" title="👍 点赞列表" width="60%" :modal-class="'fixed-dialog-height'" top="6vh">
    <div class="card-list-scroll">
    <div class="card-list">
      <!-- 卡片内容保持统一结构 -->
      <el-card
          v-for="item in likeList"
          :key="item.GUID"
          class="list-card"
          shadow="hover"
          @click="blogMainClick(item)"
      >
        <div class="card-top">
          <span class="author-name">{{ item.USERNAME || user.name }}</span>
          <span class="card-time">{{ formatDate(item.CREATE_TIME) }}</span>
        </div>
        <div class="card-body">
          <h4 class="card-title">{{ item.BLOG_TITLE }}</h4>
          <p class="card-summary" v-html="item.MAINTEXT"></p>
        </div>
      </el-card>
    </div>
    </div>
  </el-dialog>

  <el-dialog v-model="showCollectDialog" title="⭐ 收藏列表" width="60%" :modal-class="'fixed-dialog-height'" top="6vh">
    <div class="card-list-scroll">
    <div class="card-list">
      <!-- 卡片内容保持统一结构 -->
      <el-card
          v-for="item in collectList"
          :key="item.GUID"
          class="list-card"
          shadow="hover"
          @click="blogMainClick(item)"
      >
        <div class="card-top">
          <span class="author-name">{{ item.USERNAME || user.name }}</span>
          <span class="card-time">{{ formatDate(item.CREATE_TIME) }}</span>
        </div>
        <div class="card-body">
          <h4 class="card-title">{{ item.BLOG_TITLE }}</h4>
          <p class="card-summary" v-html="item.MAINTEXT"></p>
        </div>
      </el-card>

    </div>
    </div>
  </el-dialog>


</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {useRoute,useRouter} from "vue-router";
import {pubFormatDate, decrypt, sendAxiosRequest, downloadFileByUrl} from "@/utils/common.js";
import ContentAndComment from "@/views/detail/blog/ContentAndComment.vue";
import {adminUserCode} from "@/config/vue-config.js";
import {useUserStore} from "@/stores/main/user.js";

const userStore = useUserStore();
userStore.initFromLocal();

const showLikeDialog = ref(false)
const showCollectDialog = ref(false)
//点赞数
const likeNum = ref(0);
//点赞列表
const likeList = ref([]);
//收藏数
const collectNum = ref(0);
//收藏列表
const collectList = ref([]);

const route = useRoute();
const router = useRouter();
const user = ref({

})

async function getUserInfo2Data(){

  let userCode = route.query.c;
  if(userCode){
    userCode = decrypt(userCode);
    //获取账号信息
    let result = await sendAxiosRequest("/pub-api/login/getUserInfoByCode",{userCode});
    if(result && !result.isError){
      user.value = result.result;
    }
    //获取账号发布得内容和文件
    result = await  sendAxiosRequest("/blog-api/userInformation/getBlogAndResourceByUserCode",{userCode});
    if(result && !result.isError){
      blogs.value = result.result.blog;
      files.value = result.result.resource;
    }
    result = await  sendAxiosRequest("/blog-api/blog/getLikeAndCollectByUserCode",{userCode});
    if(result && !result.isError){
      result.result.forEach(item=>{
        if(item["TYPE"]=="like"){
          likeNum.value++;
          likeList.value.push(item);
        }else if (item["TYPE"]=="collect"){
          collectNum.value++;
          collectList.value.push(item);
        }
      })
    }
  }
}

getUserInfo2Data();

const isFollowing = ref(false)

const toggleFollow = () => {
  isFollowing.value = !isFollowing.value
  ElMessage.success(isFollowing.value ? '已关注' : '已取消关注')
}

const blogs = ref([])
const files = ref([])

const showAllBlogs = ref(false)
const displayedBlogs = computed(() => {
  return showAllBlogs.value ? blogs.value : blogs.value.slice(0, 3)
})

const showDialog = ref(false);
const selectedBlogId = ref("");
function blogMainClick(blog){
  selectedBlogId.value = blog.GUID;
  showDialog.value = true;
}

const onDialogClose = () => {
  selectedBlogId.value = ''
  showDialog.value = false
}

const formatDate = (dateStr) => {
  return pubFormatDate(dateStr);
}

const downloadFile = (file) => {

  ElMessage.success(`准备下载：${file.ORIGINALFILENAME}`);
  downloadFileByUrl(file.FILEVIEWURL,file.ORIGINALFILENAME);
}

const fileSection = ref(null)

const scrollToFiles = () => {
  if (fileSection.value?.$el) {
    fileSection.value.$el.scrollIntoView({ behavior: 'smooth' })
  }
}

const goToLiked = () => {
  showLikeDialog.value = true;
}

const goToFavorites = () => {
  showCollectDialog.value = true;
}

onMounted(() => {

})
</script>

<style scoped>
.person-info {
  background-color: #f5f7fa; /* 柔和的浅灰蓝背景 */
  height: 100%;
  overflow-y: auto;
  max-height: 100%;
  padding: 24px;
  box-sizing: border-box;
}

.inner-container {
  max-width: 900px; /* 控制内部内容宽度，可根据需要调整 */
  margin: 0 auto;   /* 居中 */
  padding: 0 20px;  /* 适当左右内边距，防止太贴边 */
}

.toolbar {
  text-align: right;
  margin-bottom: 12px;
}

.profile-card {
  margin-bottom: 24px;
  padding: 24px;
}

.profile-header {
  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 20px;
}

.left-info {
  display: flex;
  align-items: center;
}

.author-avatar {
  width: 100px !important;
  height: 100px !important;
  font-size: 20px;
  background-color: #f2f2f2;
}

.profile-details {
  margin-left: 24px;
}

.right-info {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: flex-end;
  gap: 10px;
}

.interaction-buttons {
  display: flex;
  gap: 10px;
  margin-top: 4px;
}

.username {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin: 0;
  font-size: 24px;
  font-weight: bold;
}

.admin-badge {
  background-color: #ffdf02;
  color: #000;
  font-size: 12px;
  font-weight: bold;
  border-radius: 12px;
  padding: 2px 8px;
  line-height: 1;
}

.admin-badge {
  background-color: #ffdf02;
  color: #000;
  font-size: 12px;
  font-weight: bold;
  border-radius: 12px;
  padding: 2px 8px;
  line-height: 1; /* 避免徽章高度不一致 */
}

.userip {
  color: #888;
  margin: 4px 0;
}

.stats {
  margin-top: 8px;
  font-size: 14px;
  color: #666;
  display: flex;
  gap: 24px;
}

.follow-button {
  margin-top: 12px;
  transition: all 0.3s ease;
}

.section-card {
  margin-bottom: 30px;
  padding: 16px;
}

.section-title {
  font-size: 18px;
  margin-bottom: 16px;
  font-weight: bold;
}

.blog-card {
  transition: 0.3s;
  cursor: pointer;
}

.blog-card:hover {
  transform: scale(1.01);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.blog-summary {
  display: -webkit-box;
  -webkit-line-clamp: 8;   /* 最多显示 8 行 */
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.6;    /* 可调，建议和字体配合设置 */
  color: #666;
  margin-top: 8px;
}

.file-table :deep(.el-button) {
  font-weight: bold;
}

.collapse-toggle {
  margin-top: 12px;
  text-align: center;
}

/* 弹窗固定高度 */
:deep(.fixed-dialog-height .el-dialog) {
  max-height: 500px;
  display: flex;
  flex-direction: column;
}

/* 弹窗 body 内部滚动 */
:deep(.fixed-dialog-height .el-dialog__body) {
  flex: 1;
  overflow: hidden; /* ⚠️ 禁止 body 滚动 */
  padding-top: 0;
}

.card-list-scroll {
  max-height: 560px;
  overflow-y: auto;
  padding: 12px 16px; /* 上下12px，左右16px */
  box-sizing: border-box;
}

/* 卡片列表 */
.card-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.list-card {
  height: 160px;
  padding: 12px 16px;
  background-color: #edf8f4;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  transition: all 0.3s ease;
  cursor: pointer;
}

.list-card:hover {
  transform: scale(1.01);
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
}

.card-top {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: #666;
  margin-bottom: 4px;
}

.card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  margin: 6px 0;
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow: hidden;
  color: #333;
}

.card-summary {
  color: #666;
  font-size: 13px;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

</style>
