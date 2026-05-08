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
import {nextTick, onMounted, ref} from "vue";
import { useRoute, useRouter } from "vue-router";
import { useHead } from '@vueuse/head'; // 1. 引入 useHead
import {
  sendAxiosRequest,
  extractPlainTextFromHTML,
  getUserInfoByCode
} from "@/utils/common";
import { useUserStore } from "@/stores/main/user.js";

import ContentAndComment from "@/views/detail/blog/ContentAndComment.vue";
import BackgroundAndMusic from "@/components/detail/personInformation/BackgroundAndMusic.vue";
import UserInfo from "@/components/main/UserInfo.vue";

const userStore = useUserStore();
userStore.initFromLocal();

const route = useRoute();
const router = useRouter();
const blogId = route.params.g;

const targetUserCode = ref('');
const authorInfo = ref({});
const articleTitle = ref("");

// ==== 背景与音乐 ====
const bgMusicComponentRef = ref(null);
const serverBgImage = ref('');
const serverBgAudio = ref('');
const currentBgStyle = ref({});

const handleBgStyleUpdate = (style) => {
  currentBgStyle.value = style;
}

// 2. 定义响应式的 SEO 数据源，给定默认值
const seoTitle = ref('博客详情');
const seoDescription = ref('专注于技术分享的内容社区');

// 3. 在 setup 顶层调用 useHead，直接传入 ref 变量
// 当 seoTitle 或 seoDescription 改变时，<head> 标签会自动更新
useHead({
  title: seoTitle,
  meta: [
    {
      name: 'description',
      content: seoDescription
    }
  ]
});

function getUserInfo2Data() {
  const getAuthorInfo = async () => {
    let result = await getUserInfoByCode(targetUserCode.value);
    if (result && !result.isError) {
      authorInfo.value = result.result;
    }
  }

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

const contentAndCommentIsLoad = ({blogContent}) => {
  if (blogContent.USERCODE) {
    targetUserCode.value = blogContent.USERCODE;
    articleTitle.value = blogContent.BLOG_TITLE;

    getUserInfo2Data();
    debugger;
    // 4. 彻底删除 document 原生操作，只需更新 ref 变量
    seoTitle.value = blogContent.BLOG_TITLE;
    seoDescription.value = extractPlainTextFromHTML(blogContent.BLOG_TITLE + blogContent.MAINTEXT).slice(0, 200) + '...';
    //用于Prerender的seo检索   防止只爬虫到外壳没有实际数据
    nextTick(() => {
      window.prerenderReady = true;
    });
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