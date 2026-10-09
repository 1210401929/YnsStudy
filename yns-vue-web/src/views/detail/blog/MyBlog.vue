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
import {ref, onMounted, watch} from 'vue'
import { useRouter } from 'vue-router'
import { useBlogContentStore } from '@/stores/detail/blog.js'
import { useUserStore } from "@/stores/main/user.js"
import { Edit } from '@element-plus/icons-vue'
import { sendAxiosRequest } from '@/utils/common.js'
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

  setTopAlert();
  isInitialized.value = true;
})

watch(()=>userStore.userBean.code,()=>{
  getUserInfo2Data();
},{ immediate: true } )
</script>

<style scoped>
/* 最外层容器，只负责背景（用户可在个人主页设置背景图） */
.blog-container {
  min-height: 100vh;
  box-sizing: border-box;
  position: relative;
  background-size: cover;
  background-position: center;
  background-attachment: fixed;
  transition: background-image .3s ease;
}

.layout-wrapper {
  width: 100%;
  max-width: 1360px;
  margin: 0 auto;
  padding: 32px 24px;
  display: flex;
  align-items: flex-start;
  gap: 28px;
  box-sizing: border-box;
}

.sidebar-wrapper {
  width: 280px;
  flex-shrink: 0;
  position: sticky;
  top: 24px;
}

/* 右侧正文：纸张样式由 ContentAndComment 自己提供 */
.content {
  flex: 1;
  min-width: 0;
  box-sizing: border-box;
  overflow-x: hidden;
  padding: 10px 4px 24px;
  min-height: 82vh;
}

/* 发布文章：一张黄色便签 */
.fab-button {
  position: fixed;
  bottom: 30px;
  right: 30px;
  z-index: 1000;
  height: auto;
  padding: 12px 20px;
  border: none;
  border-radius: 2px;
  background-color: var(--j-note);
  color: var(--j-ink);
  font-family: var(--j-hand);
  font-size: 16px;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.12), 0 12px 20px -10px rgba(60, 50, 30, 0.5);
  transform: rotate(-3deg);
  transition: transform 0.2s ease, background-color 0.2s ease;
}

.fab-button:hover,
.fab-button:focus {
  background-color: #f8e88f;
  color: var(--j-ink);
  transform: rotate(0deg) translateY(-2px);
}

.blinking {
  animation: wiggle 1.6s ease-in-out infinite;
}

@keyframes wiggle {
  0%, 100% { transform: rotate(-3deg); }
  50% { transform: rotate(2deg) translateY(-3px); }
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
  position: relative;
  max-width: 240px;
  padding: 12px 16px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-ink);
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
  border-top: 10px solid var(--j-paper);
  transform: translateX(50%);
}

/* 屏幕较小时变成上下结构 */
@media screen and (max-width: 992px) {
  .layout-wrapper {
    flex-direction: column;
    padding: 16px 12px;
    gap: 20px;
    overflow-x: hidden;
  }

  .sidebar-wrapper {
    width: 100%;
    position: static;
  }

  .content {
    width: 100%;
    margin-top: 0;
    padding: 4px 0 16px;
  }
}
</style>
