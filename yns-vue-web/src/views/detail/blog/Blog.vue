<template>
  <el-container class="blog-container">
    <!-- 左侧目录 -->
    <el-aside width="220px" class="sidebar" v-if="blogContentStore.blogContents.length>0">
      <el-menu :default-active="selectedIndex.toString()" @select="handleSelect">
        <el-menu-item
            v-for="(blogContent, index) in blogContentStore.blogContents"
            :key="blogContent.GUID"
            :index="index.toString()"
            class="el-menu_"
        >
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
    <ArticleEditor @submit="handleEditorSubmit" @cancel="editorVisible = false"/>
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

onMounted(async () => {
  await blogContentStore.initBlogContent();
  handleSelect(selectedIndex.value);
  isInitialized.value = true;
})

async function handleEditorSubmit({title, content}) {
  let userBean = userStore.userBean;
  let blogContent = {
    GUID: getGuid(),
    BLOG_TITLE: title,
    MAINTEXT: content,
    USERCODE: userBean.code,
    USERNAME: userBean.name
  };

  let result = await sendAxiosRequest("/blog-api/blog/addBlog", {blogContent});
  blogContentStore.blogContents.push(blogContent);
  selectedIndex.value = blogContentStore.blogContents.length - 1
  handleSelect(selectedIndex.value)
  editorVisible.value = false
}
</script>

<style scoped>
.blog-container {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 15px 10px;
  background-color: #f2f3f5;
  gap: 16px;
  min-height: 100vh;
  box-sizing: border-box;
  position: relative;
}

.sidebar {
  background: #fff;
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
  background: #fff;
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
