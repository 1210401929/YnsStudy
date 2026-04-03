<template>
  <Announcement v-for="al in topAlert" :key="al.GUID" :TEXT="al.TEXT" :URL="al.URL" :URLNAME="al.URLNAME"/>

  <el-container class="blog-container" :style="currentBgStyle">
    <BackgroundAndMusic
        ref="bgMusicComponentRef"
        :is-self="false"
        :user-name="userStore.userBean.name"
        :init-bg-image="serverBgImage"
        :init-bg-audio="serverBgAudio"
        @update-bg-style="handleBgStyleUpdate"
    />

    <div class="layout-wrapper">

      <div class="sidebar-wrapper" v-if="blogContentStore.blogContents.length > 0">
        <BlogSidebar
            :user-code="userStore.userBean.code"
            :is-router="true"
            v-model:selectedIndex="selectedIndex"
        />
      </div>

      <el-main class="content" v-if="blogContentStore.blogContents.length > 0">
        <RouterView/>
      </el-main>

    </div>
  </el-container>

  <div v-if="isInitialized && blogContentStore.blogContents.length === 0" class="guide-tip">
    <div class="guide-box">
      <span>暂无文章，点击右下角按钮发布你的第一篇文章！</span>
      <div class="arrow-down"></div>
    </div>
  </div>

  <el-button
      type="primary"
      :icon="Edit"
      @click="subButtonClick"
      class="fab-button"
      :class="{ blinking: blogContentStore.blogContents.length === 0 }"
  >
    发布文章
  </el-button>

  <el-dialog v-model="editorVisible" title="文章编辑" width="900px" top="2vh" :close-on-click-modal="false">
    <ArticleEditor @submit="handleEditorSubmit" :is-public="true" :save-type="'add'" @cancel="editorVisible = false"/>
  </el-dialog>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useBlogContentStore } from '@/stores/detail/blog.js'
import { useUserStore } from "@/stores/main/user.js"
import { Edit } from '@element-plus/icons-vue'
import { getGuid, sendAxiosRequest } from '@/utils/common.js'
import ArticleEditor from '@/components/detail/ArticleEditor.vue'
import { ElMessage } from 'element-plus'
import Announcement from "@/components/detail/Announcement.vue"
import { getAnnouncementByRouterName } from "@/utils/blogUtil.js"
import BackgroundAndMusic from "@/components/detail/personInformation/BackgroundAndMusic.vue"

// 分类归档目录组件
import BlogSidebar from "@/components/detail/myblog/BlogSidebar.vue";

const router = useRouter()
const userStore = useUserStore()
const blogContentStore = useBlogContentStore()

blogContentStore.blogContents = []
const selectedIndex = ref('')
const editorVisible = ref(false)
const isInitialized = ref(false)

// ================= 基础交互 =================
function subButtonClick() {
  let userBean = userStore.userBean;
  if (!userBean || !userBean.code) {
    return ElMessage.error("请先登录!");
  }
  editorVisible.value = true
}

router.push({name: 'BlogContent', query: {g: "YouDontNeedToPayAttention"}});

async function handleEditorSubmit({blog_type, title, content}) {
  let userBean = userStore.userBean;
  let blogContent = {
    GUID: null,//不传递guid,后台构造
    BLOG_TITLE: title,
    MAINTEXT: content,
    BLOG_TYPE: blog_type,
    USERCODE: userBean.code,
    USERNAME: userBean.name,
    CATEGORY_ID: null // 默认新文章暂不分类
  };

  let result = await sendAxiosRequest("/blog-api/blog/addBlog", {blogContent});
  blogContent = result.result[0];
  blogContentStore.blogContents.unshift(blogContent);

  // 更新选中项并跳转
  selectedIndex.value = blogContent.GUID;
  router.push({name: 'BlogContent', query: {g: blogContent.GUID}});
  editorVisible.value = false;
}

// ================= 背景与公告 =================
const bgMusicComponentRef = ref(null);
const serverBgImage = ref('');
const serverBgAudio = ref('');
const currentBgStyle = ref({});

const handleBgStyleUpdate = (style) => currentBgStyle.value = style;

function getUserInfo2Data() {
  const setPersonInfo = async () => {
    let result = await sendAxiosRequest("/blog-api/userInformation/getPersonInfo", {userCode: userStore.userBean.code});
    if (result && !result.isError) {
      serverBgImage.value = (result.result[0] || {}).BGIMAGEURL || "";
    }
  }
  setPersonInfo();
}

const topAlert = ref([]);
const setTopAlert = async () => {
  topAlert.value = await getAnnouncementByRouterName("MyBlog");
}

onMounted( () => {
  blogContentStore.initBlogContent();

  if (blogContentStore.blogContents.length > 0) {
    // 默认选中第一篇并跳转
    // selectedIndex.value = blogContentStore.blogContents[0].GUID;
    // router.push({name: 'BlogContent', query: {g: selectedIndex.value}});
  }

  getUserInfo2Data();
  setTopAlert();
  isInitialized.value = true;
})
</script>

<style scoped>
/* 最外层容器，只负责背景 */
.blog-container {
  min-height: 100vh;
  box-sizing: border-box;
  position: relative;
  background-size: cover;
  background-position: center;
  background-attachment: fixed;
  transition: background-image .3s ease;
}

/* 🎯 核心布局层：限制最大宽度并居中 */
.layout-wrapper {
  width: 100%;
  max-width: 1440px; /* 和个人主页保持一致的大气宽屏 */
  margin: 0 auto;
  padding: 24px 16px;
  display: flex;
  align-items: flex-start;
  gap: 32px; /* 左边侧边栏和右侧正文的间距 */
  box-sizing: border-box;
}

/* 🎯 侧边栏约束容器 */
.sidebar-wrapper {
  width: 280px;    /* 死死锁住 280px 宽度 */
  flex-shrink: 0;  /* 绝对不允许被挤压变窄 */
  position: sticky;
  top: 24px;       /* 开启独立吸顶 */
}

/* 右侧正文区域 */
.content {
  flex: 1; /* 占据剩余全部空间 */
  min-width: 0; /* 防止内容过长撑破 Flex 布局 */

  /* 统一卡片美化风格 */
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.5);
  border-radius: 12px;
  box-shadow: 0 4px 18px rgba(0, 0, 0, 0.08);

  padding: 24px;
  min-height: 82vh;
}

/* =============== 悬浮按钮和提示框保持不变 =============== */
.fab-button {
  position: fixed;
  bottom: 30px;
  right: 30px;
  z-index: 1000;
  background-color: #409eff;
  border-radius: 30px;
  box-shadow: 0 0 20px rgba(64, 158, 255, 0.8);
  padding: 12px 20px;
  font-weight: bold;
}

.fab-button:hover {
  background-color: #66b1ff;
}

.blinking {
  animation: pulse 1.2s infinite;
}

@keyframes pulse {
  0%, 100% { box-shadow: 0 0 20px rgba(64, 158, 255, 0.8); transform: scale(1); }
  50% { box-shadow: 0 0 40px rgba(64, 158, 255, 1); transform: scale(1.08); }
}

.guide-tip {
  position: fixed;
  bottom: 100px;
  right: 60px;
  z-index: 1001;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  pointer-events: none;
}

.guide-box {
  background: #fff;
  padding: 10px 16px;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  font-size: 14px;
  color: #333;
  position: relative;
  max-width: 240px;
  pointer-events: auto;
}

.arrow-down {
  position: absolute;
  bottom: -10px;
  right: 20px;
  width: 0;
  height: 0;
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-top: 10px solid #fff;
  transform: translateX(50%);
}

/* 响应式：屏幕较小时变成上下结构 */
@media screen and (max-width: 992px) {
  .layout-wrapper {
    flex-direction: column;
    padding: 16px;
    gap: 20px;
  }
  .sidebar-wrapper {
    width: 100%;
    position: static; /* 手机端取消吸顶 */
  }
  .content {
    width: 100%;
    margin-top: 0;
  }
}
</style>