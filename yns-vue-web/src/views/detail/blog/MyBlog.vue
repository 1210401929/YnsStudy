<template>
  <!-- 公告横幅 -->
  <Announcement v-for="al in topAlert" :key="al.GUID" :TEXT="al.TEXT" :URL="al.URL" :URLNAME="al.URLNAME"/>

  <el-container class="blog-container" :style="currentBgStyle">
    <!-- 背景图片/音乐组件-->
    <BackgroundAndMusic
        ref="bgMusicComponentRef"
        :is-self="false"
        :user-name="userStore.userBean.name"
        :init-bg-image="serverBgImage"
        :init-bg-audio="serverBgAudio"
        @update-bg-style="handleBgStyleUpdate"
    />
    <!-- 左侧目录 -->
    <el-aside width="220px" class="sidebar" v-if="blogContentStore.blogContents.length>0">
      <el-menu style="background-color: rgba(14,165,233,0)" :default-active="selectedIndex.toString()"
               @select="handleSelect">
        <el-menu-item
            v-for="(blogContent, index) in blogContentStore.blogContents"
            :key="blogContent.GUID"
            :index="index.toString()"
            class="el-menu_"
        >
          <el-tag :type="blogContent.BLOG_TYPE === 'public'?'success':'danger'" effect="plain" size="small">{{
              blogContent.BLOG_TYPE === "public" ? "公开" : "私密"
            }}
          </el-tag>
          {{ blogContent.BLOG_TITLE }}
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 正文和评论部分 -->
    <el-main class="content" v-if="blogContentStore.blogContents.length>0">
      <RouterView/>
    </el-main>
  </el-container>

  <!-- 引导提示（仅在没有内容时显示） -->
  <div v-if="isInitialized && blogContentStore.blogContents.length === 0" class="guide-tip">
    <div class="guide-box">
      <span>暂无文章，点击右下角按钮发布你的第一篇文章！</span>
      <div class="arrow-down"></div>
    </div>
  </div>

  <!-- 悬浮按钮 -->
  <el-button
      type="primary"
      :icon="Edit"
      @click="subButtonClick"
      class="fab-button"
      :class="{ blinking: blogContentStore.blogContents.length === 0 }"
  >
    发布文章
  </el-button>

  <!-- 编辑器弹窗 -->
  <el-dialog
      v-model="editorVisible"
      title="文章编辑"
      width="900px"
      top="2vh"
      :close-on-click-modal="false"
  >
    <ArticleEditor @submit="handleEditorSubmit" :is-public="true" :save-type="'add'" @cancel="editorVisible = false"/>
  </el-dialog>
</template>

<script setup>
import {ref, onMounted} from 'vue'
import {useRouter} from 'vue-router'
import {useBlogContentStore} from '@/stores/detail/blog.js'
import {useUserStore} from "@/stores/main/user.js";
import {Edit} from '@element-plus/icons-vue'
import {getGuid} from '@/utils/common.js'
import ArticleEditor from '@/components/detail/ArticleEditor.vue'
import {sendAxiosRequest} from "@/utils/common.js";
import {ElMessage} from 'element-plus';
import Announcement from "@/components/detail/Announcement.vue";
import {getAnnouncementByRouterName} from "@/utils/blogUtil.js";
import BackgroundAndMusic from "@/components/detail/personInformation/BackgroundAndMusic.vue";

const router = useRouter()
const userStore = useUserStore();
const blogContentStore = useBlogContentStore()
blogContentStore.blogContents = [];

const selectedIndex = ref(0)
const editorVisible = ref(false)
const isInitialized = ref(false)

function subButtonClick() {
  let userBean = userStore.userBean;
  if (!userBean || !userBean.code) {
    ElMessage.error("请先登录!");
    return false;
  }
  editorVisible.value = true
}

function handleSelect(index) {
  const item = blogContentStore.blogContents[index]
  if (item) {
    router.push({name: 'BlogContent', query: {g: item.GUID}})
    selectedIndex.value = parseInt(index)
  }
}

router.push({name: 'BlogContent', query: {g: "YouDontNeedToPayAttention"}});


async function handleEditorSubmit({blog_type, title, content}) {
  let userBean = userStore.userBean;
  let blogContent = {
    GUID: getGuid(),
    BLOG_TITLE: title,
    MAINTEXT: content,
    BLOG_TYPE: blog_type,
    USERCODE: userBean.code,
    USERNAME: userBean.name
  };

  let result = await sendAxiosRequest("/blog-api/blog/addBlog", {blogContent});
  blogContentStore.blogContents.push(blogContent);
  selectedIndex.value = blogContentStore.blogContents.length - 1
  handleSelect(selectedIndex.value)
  editorVisible.value = false
}

// ==== 背景与音乐：父组件状态对接 ====
const bgMusicComponentRef = ref(null);
const serverBgImage = ref('');
const serverBgAudio = ref('');
const currentBgStyle = ref({});

const handleBgStyleUpdate = (style) => {
  currentBgStyle.value = style;
}

function getUserInfo2Data() {
  //获取用户设置信息 ，例如：背景图片，音乐
  const setPersonInfo = async () => {

    let result = await sendAxiosRequest("/blog-api/userInformation/getPersonInfo", {userCode: userStore.userBean.code});
    if (result && !result.isError) {
      result = result.result[0] || {};
      // 直接赋值给响应式变量，子组件的 watch 会接收到
      //博客阅读界面不适用背景音乐功能，暂时注释
      // serverBgAudio.value = result.BGMUSICURL || "";
      serverBgImage.value = result.BGIMAGEURL || "";
    }
  }
  setPersonInfo();
}

onMounted(async () => {

  await blogContentStore.initBlogContent();
  handleSelect(selectedIndex.value);
  getUserInfo2Data();
  isInitialized.value = true;
})

//公告横幅内容
const topAlert = ref([]);
const setTopAlert = async () => {
  topAlert.value = await getAnnouncementByRouterName("MyBlog");
}
setTopAlert();
</script>

<style scoped>
.blog-container {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 15px 10px;
  gap: 16px;
  min-height: 100vh;
  box-sizing: border-box;
  position: relative;
  /* 新增：确保背景图片能完美铺满并固定 */
  background-size: cover;
  background-position: center;
  background-attachment: fixed;
  transition: background-image .3s ease; /* 柔和的切换动画 */
}

.sidebar {
  background-color: rgba(255, 255, 255, 0.9);
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  padding: 20px 10px;
  height: fit-content;
  position: sticky;
  top: 15px;
  will-change: transform;
  transition: top 0.1s ease;
}

.content {
  flex: 1;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  padding: 24px;
  min-height: 82vh;
}

/* 发布按钮 */
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
  0%, 100% {
    box-shadow: 0 0 20px rgba(64, 158, 255, 0.8);
    transform: scale(1);
  }
  50% {
    box-shadow: 0 0 40px rgba(64, 158, 255, 1);
    transform: scale(1.08);
  }
}

/* 引导提示框 */
.guide-tip {
  position: fixed;
  bottom: 100px; /* 距离页面底部的距离 */
  right: 60px; /* 微调靠近按钮 */
  z-index: 1001;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  pointer-events: none; /* 防止遮挡点击 */
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
  pointer-events: auto; /* 引导框内部可交互 */
}

.arrow-down {
  position: absolute;
  bottom: -10px;
  right: 20px; /* 精准指向按钮的右边中心点 */
  width: 0;
  height: 0;
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-top: 10px solid #fff;
  transform: translateX(50%); /* 让箭头尖端指向按钮中心 */
}

.el-menu_ {
  /* 🌟 新增：文字过长省略 */
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  max-width: 100%;
  display: block;
}

/* 移动端优化 */
@media screen and (max-width: 768px) {
  .blog-container {
    flex-direction: column;
    padding: 10px;
  }

  .sidebar {
    width: 100%;
    margin-top: 0;
    position: static;
    border-radius: 8px;
  }

  .content {
    width: 100%;
    margin-top: 16px;
    border-radius: 8px;
  }

  .guide-tip {
    bottom: 120px;
    right: 20px;
    font-size: 12px;
  }
}
</style>
