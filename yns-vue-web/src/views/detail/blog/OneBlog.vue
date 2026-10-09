<template>
  <div class="blog-detail-page j-desk" :style="currentBgStyle">
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
        <router-link to="/ynsStudy/Home">查看最新文章</router-link>
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
const seoImage = ref(`${window.location.origin}/og-image.png`);
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
      url: window.location.origin,
      logo: `${window.location.origin}/icon-512.png`
    },
    image: seoImage.value ? [seoImage.value] : undefined
  };
  return {
    title: seoTitle.value,
    link: [
      // 文章不存在时只输出 noindex，不再声明 canonical，避免两个信号互相矛盾
      ...(articleNotFound.value ? [] : [{rel: 'canonical', href: canonicalUrl.value}]),
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
    seoImage.value = firstImage ? new URL(firstImage, window.location.origin).href : `${window.location.origin}/og-image.png`;

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
.blog-detail-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-size: cover;
  background-position: center;
  background-attachment: fixed;
  transition: background-image .3s ease;
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

/* 顶部：一条纸边 */
.blog-top-bar {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px 20px;
  background: var(--j-paper);
  border-bottom: 1px solid var(--j-rule);
  box-shadow: 0 6px 14px -12px rgba(60, 50, 30, 0.35);
}

.blog-top-bar::after {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  bottom: 3px;
  border-bottom: 1px dashed var(--j-rule);
}

.blog-title {
  flex: 1;
  max-width: 900px;
  overflow: hidden;
  text-align: center;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-family: var(--j-hand);
  font-size: 19px;
  color: var(--j-ink);
}

/* 主体两栏 */
.main-body {
  display: flex;
  align-items: flex-start;
  gap: 28px;
  width: 100%;
  max-width: 1600px;
  margin: 0 auto;
  padding: 32px 24px;
  flex: 1;
  box-sizing: border-box;
}

.content-side {
  flex: 1;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
  min-height: 82vh;
}

/* 文章不存在：一张便签 */
.article-not-found {
  position: relative;
  width: min(560px, calc(100% - 32px));
  margin: 14vh auto;
  padding: 40px 28px 32px;
  box-sizing: border-box;
  text-align: center;
  background: var(--j-note);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.1), 0 16px 28px -16px rgba(60, 50, 30, 0.45);
  transform: rotate(-1deg);
}

.article-not-found h1 {
  margin-top: 0;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 28px;
}

.article-not-found p {
  color: var(--j-ink-soft);
}

.article-not-found a {
  color: var(--j-pen);
}

@media (max-width: 768px) {
  .main-body {
    flex-direction: column;
    gap: 20px;
    padding: 16px 12px;
  }

  .content-side {
    width: 100%;
  }

  :deep(.user-info-container) {
    width: 100%;
    max-width: 100%;
    box-sizing: border-box;
  }
}
</style>
