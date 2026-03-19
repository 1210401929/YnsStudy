<template>
  <div class="blog-detail-page" :style="currentBgStyle">
    <BackgroundAndMusic
        ref="bgMusicComponentRef"
        :is-self="false"
        :user-name="authorInfo.name"
        :init-bg-image="serverBgImage"
        :init-bg-audio="serverBgAudio"
        @update-bg-style="handleBgStyleUpdate"
    />

    <div class="blog-top-bar">
      <div class="blog-title">{{ articleTitle || '博客详情' }}</div>
    </div>

    <div class="main-body">

      <UserInfo
          v-if="targetUserCode"
          :user="authorInfo"
          :target-user-code="targetUserCode"
          @open-chat=""
      />

      <div class="content-side">
        <ContentAndComment :blogId="blogId" @loaded="contentAndCommentIsLoad"/>
      </div>

    </div>

  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  sendAxiosRequest,
  extractPlainTextFromHTML,
  getUserInfoByCode
} from "@/utils/common";
import { useUserStore } from "@/stores/main/user.js";

import ContentAndComment from "@/views/detail/blog/ContentAndComment.vue";
import BackgroundAndMusic from "@/components/detail/personInformation/BackgroundAndMusic.vue";
// 引入左侧用户组件
import UserInfo from "@/components/main/UserInfo.vue";

const userStore = useUserStore();
userStore.initFromLocal();

const route = useRoute();
const router = useRouter();
const blogId = route.params.g;

// 改为响应式变量，这样 UserInfo 组件能够监听到变化并加载数据
const targetUserCode = ref('');

// 作者信息
const authorInfo = ref({});


// ==== 背景与音乐 ====
const bgMusicComponentRef = ref(null);
const serverBgImage = ref('');
const serverBgAudio = ref('');
const currentBgStyle = ref({});

const handleBgStyleUpdate = (style) => {
  currentBgStyle.value = style;
}

function getUserInfo2Data() {
  // 获取账号信息
  const getAuthorInfo = async () => {
    let result = await getUserInfoByCode(targetUserCode.value);
    if (result && !result.isError) {
      authorInfo.value = result.result;
    }
  }

  // 获取用户设置信息（背景图片等）
  const setPersonInfo = async () => {
    let result = await sendAxiosRequest("/blog-api/userInformation/getPersonInfo", { userCode: targetUserCode.value });
    if (result && !result.isError) {
      result = result.result[0] || {};
      serverBgImage.value = result.BGIMAGEURL || "";
    }
  }

  getAuthorInfo();
  setPersonInfo();
}

// 顶部栏信息
const articleTitle = ref("");



const contentAndCommentIsLoad = ({blogContent}) => {
  if (blogContent.USERCODE) {
    // 赋值后，UserInfo 监听到 targetUserCode 变化会自动加载粉丝/点赞数据
    targetUserCode.value = blogContent.USERCODE;
    articleTitle.value = blogContent.BLOG_TITLE;

    // 搜索用户信息
    getUserInfo2Data();

    // 修改浏览器title和meta,有助于搜索排名
    document.title = blogContent.BLOG_TITLE;
    const metaDesc = document.querySelector('meta[name="description"]');
    const blogSummary = extractPlainTextFromHTML(blogContent.BLOG_TITLE + blogContent.MAINTEXT).slice(0, 200) + '...';
    if (metaDesc) {
      metaDesc.setAttribute("content", blogSummary);
    } else {
      const desc = document.createElement('meta')
      desc.name = "description"
      desc.content = blogSummary
      document.head.appendChild(desc)
    }
  }
}

onMounted(() => {})
</script>

<style scoped>
.blog-detail-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-size: cover;
  background-position: center;
  background-attachment: fixed;
  transition: background-image .3s ease;
}

.blog-detail-page.bg-fade {
  animation: bgfade .25s ease;
}

@keyframes bgfade {
  from { opacity: .6; }
  to { opacity: 1; }
}

/* 顶部信息栏 */
.blog-top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: rgba(255, 255, 255, 0.15);
  padding: 20px 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.04);
}

.blog-title {
  font-size: 18px;
  font-weight: bold;
  flex: 1;
  text-align: center;
  color: #333;
}

/* 主体两栏 */
.main-body {
  display: flex;
  align-items: flex-start; /* 🌟 加上这一行，确保左右顶部完美对齐 */
  gap: 24px;
  padding: 20px;
  flex: 1;
  box-sizing: border-box;
}

.content-side {
  flex: 1;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  padding: 24px;
  min-height: 82vh;
  overflow: hidden;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .main-body {
    flex-direction: column;
    gap: 20px;
  }
  .content-side {
    padding: 16px;
  }
}
</style>