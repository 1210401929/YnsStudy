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

    <template v-if="articleNotFound">
      <main class="article-not-found">
        <h1>文章不存在</h1>
        <p>这篇文章可能已删除、设为私密或地址有误。</p>
        <router-link to="/ynsStudy/MyBlog">返回博客列表</router-link>
      </main>
    </template>

    <template v-else>
      <div class="blog-top-bar">
        <div class="blog-title">{{ articleTitle || '博客详情' }}</div>
      </div>

      <main class="main-body">

        <UserInfo
            v-if="targetUserCode"
            :user="authorInfo"
            :target-user-code="targetUserCode"
            @open-chat=""
        />

        <article class="content-side">
          <ContentAndComment
              :blogId="blogId"
              @loaded="contentAndCommentIsLoad"
              @not-found="handleArticleNotFound"
          />
        </article>

      </main>
    </template>

  </div>
</template>

<script setup>
import {computed, nextTick, ref} from "vue";
import { useRoute } from "vue-router";
import { useHead } from '@vueuse/head';
import {
  sendAxiosRequest,
  extractPlainTextFromHTML,
  extractFirstImage,
  getUserInfoByCode
} from "@/utils/common";
import { useUserStore } from "@/stores/main/user.js";

import ContentAndComment from "@/views/detail/blog/ContentAndComment.vue";
import BackgroundAndMusic from "@/components/detail/personInformation/BackgroundAndMusic.vue";
import UserInfo from "@/components/main/UserInfo.vue";

const userStore = useUserStore();
userStore.initFromLocal();

const route = useRoute();
const blogId = route.params.g;

const targetUserCode = ref('');
const authorInfo = ref({});
const articleTitle = ref("");
const articleNotFound = ref(false);

// ==== 背景与音乐 ====
const bgMusicComponentRef = ref(null);
const serverBgImage = ref('');
const serverBgAudio = ref('');
const currentBgStyle = ref({});

const handleBgStyleUpdate = (style) => {
  currentBgStyle.value = style;
}

const canonicalUrl = computed(() => `${window.location.origin}/oneBlog/${encodeURIComponent(blogId)}`);
const seoTitle = ref('博客详情 - YnsStudy');
const seoDescription = ref('YnsStudy 博客文章详情');
const seoImage = ref(`${window.location.origin}/finder.png`);
const seoAuthor = ref('YnsStudy');
const seoPublished = ref('');
const seoModified = ref('');

const toISODate = (value) => {
  if (!value) return '';
  const text = String(value).trim();
  if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(text)) {
    return `${text.replace(' ', 'T')}+08:00`;
  }
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? '' : date.toISOString();
};

// Go 会为首次 HTTP 响应写入同样的数据；这里负责 Vue 接管和站内跳转后的 head 更新。
useHead(() => {
  const structuredData = {
    '@context': 'https://schema.org',
    '@type': 'BlogPosting',
    headline: articleTitle.value || '博客详情',
    description: seoDescription.value,
    url: canonicalUrl.value,
    datePublished: seoPublished.value || undefined,
    dateModified: seoModified.value || seoPublished.value || undefined,
    mainEntityOfPage: {
      '@type': 'WebPage',
      '@id': canonicalUrl.value
    },
    author: {
      '@type': 'Person',
      name: seoAuthor.value
    },
    publisher: {
      '@type': 'Organization',
      name: 'YnsStudy',
      url: window.location.origin
    },
    image: seoImage.value ? [seoImage.value] : undefined
  };
  return {
    title: seoTitle.value,
    link: [
      {rel: 'canonical', href: canonicalUrl.value},
      {rel: 'alternate', type: 'application/rss+xml', title: 'YnsStudy RSS', href: `${window.location.origin}/rss.xml`}
    ],
    meta: [
      {name: 'description', content: seoDescription.value},
      {name: 'robots', content: articleNotFound.value ? 'noindex,follow' : 'index,follow,max-image-preview:large,max-snippet:-1'},
      {property: 'og:type', content: 'article'},
      {property: 'og:site_name', content: 'YnsStudy'},
      {property: 'og:title', content: articleTitle.value || '博客详情'},
      {property: 'og:description', content: seoDescription.value},
      {property: 'og:url', content: canonicalUrl.value},
      {property: 'og:image', content: seoImage.value},
      {property: 'article:published_time', content: seoPublished.value},
      {property: 'article:modified_time', content: seoModified.value || seoPublished.value},
      {name: 'twitter:card', content: 'summary_large_image'},
      {name: 'twitter:title', content: articleTitle.value || '博客详情'},
      {name: 'twitter:description', content: seoDescription.value},
      {name: 'twitter:image', content: seoImage.value}
    ],
    script: [{type: 'application/ld+json', children: JSON.stringify(structuredData)}]
  };
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
    articleNotFound.value = false;
    targetUserCode.value = blogContent.USERCODE;
    articleTitle.value = blogContent.BLOG_TITLE;

    getUserInfo2Data();
    seoTitle.value = `${blogContent.BLOG_TITLE} - YnsStudy`;
    const plainText = extractPlainTextFromHTML(blogContent.MAINTEXT).replace(/\s+/g, ' ').trim();
    seoDescription.value = plainText.length > 180 ? `${plainText.slice(0, 180)}…` : plainText;
    seoAuthor.value = blogContent.USERNAME || 'YnsStudy';
    seoPublished.value = toISODate(blogContent.CREATE_TIME);
    seoModified.value = toISODate(blogContent.UPDATE_TIME) || seoPublished.value;
    const firstImage = extractFirstImage(blogContent.MAINTEXT);
    seoImage.value = firstImage ? new URL(firstImage, window.location.origin).href : `${window.location.origin}/finder.png`;

    nextTick(() => {
      window.prerenderReady = true;
    });
  }
}

const handleArticleNotFound = () => {
  articleNotFound.value = true;
  seoTitle.value = '文章不存在 - YnsStudy';
  seoDescription.value = '这篇文章可能已删除、设为私密或地址有误。';
  window.prerenderReady = true;
};
</script>

<style scoped>
/* 🌟 修复 1：最外层加上防溢出外壳，彻底消灭横向滚动条 */
.blog-detail-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-size: cover;
  background-position: center;
  background-attachment: fixed;
  transition: background-image .3s ease;

  /* 核心限制属性 */
  width: 100%;
  max-width: 100vw;
  overflow-x: hidden;
  box-sizing: border-box;
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
  align-items: flex-start;
  gap: 24px;
  padding: 20px;
  flex: 1;
  box-sizing: border-box;
  /* 🌟 防止内部子元素过大 */
  max-width: 100%;
}

.content-side {
  flex: 1;
  /* 🌟 修复 2：加入 Flex 终极防撑破属性与盒模型约束 */
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;

  background: rgba(255, 255, 255, 0.5);
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  padding: 24px;
  min-height: 82vh;
}

.article-not-found {
  width: min(680px, calc(100% - 32px));
  margin: 12vh auto;
  padding: 42px 28px;
  box-sizing: border-box;
  text-align: center;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 16px 45px rgba(31, 45, 61, 0.12);
}

.article-not-found h1 {
  margin-top: 0;
  font-size: 28px;
}

.article-not-found p {
  color: #68788a;
}

.article-not-found a {
  color: #087cad;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .main-body {
    flex-direction: column;
    gap: 20px;
    padding: 12px; /* 移动端稍微减小外边距 */
  }
  .content-side {
    width: 100%; /* 🌟 确保手机端占比 100% 不越界 */
    padding: 16px; /* 移动端减小内边距 */
  }
  /* 🌟 确保左侧的 UserInfo 在移动端也不会撑破屏幕 */
  :deep(.user-info-container) {
    width: 100%;
    max-width: 100%;
    box-sizing: border-box;
  }
}
</style>
