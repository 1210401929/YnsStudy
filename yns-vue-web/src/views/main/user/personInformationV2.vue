<template>
  <div class="elegant-glass-home j-desk" :style="currentBgStyle">
    <div v-if="!isPageReady" class="page-loading-mask">
      <div class="loader-ripple">
        <div></div><div></div>
      </div>
      <p class="loading-text">构建个人展厅中...</p>
    </div>

    <template v-else>
      <WelcomeOverlay
          :visible="showWelcome"
          :is-self="isSelf"
          :user="user"
          @enter="enterHomepage"
      />

      <BackgroundAndMusic
          ref="bgMusicComponentRef"
          :is-self="isSelf"
          :user-name="user.name"
          :init-bg-image="serverBgImage"
          :init-bg-audio="serverBgAudio"
          @update-bg-style="handleBgStyleUpdate"
      />

      <div class="home-container">
        <header class="glass-navbar">
          <div class="nav-brand">
            <span class="brand-dot"></span>
            <span class="brand-text">导航栏</span>
          </div>
          <div class="nav-actions">
            <button class="nav-btn" @click="handleFriendLinkClick">
              <span class="btn-icon">{{ !showFriendLink ? "🤝" : "⬅" }}</span>
              {{ !showFriendLink ? "友链" : "返回主页" }}
            </button>
            <button class="nav-btn primary-btn" @click="scrollToFiles">
              <span class="btn-icon">📁</span>
              云端文件
            </button>
          </div>
        </header>

        <transition name="fade" mode="out-in">
          <div v-if="showFriendLink" class="friend-link-view">
            <FriendLink :is-embed="true" />
          </div>

          <div v-else class="layout-grid">

            <aside class="left-control-tower">
              <div class="sticky-tower-inner">
                <div class="component-slot">
                  <UserInfo
                      :user="user"
                      :target-user-code="targetUserCode"
                      @blog-click="blogMainClick"
                  />
                </div>

                <div class="component-slot">
                  <BlogSidebar
                      :view_title="'内容归档'"
                      :user-code="targetUserCode"
                      @click-blog="blogMainClick"
                      v-model:selected-index="selectedCategory"
                  />
                </div>
              </div>
            </aside>

            <main class="right-main-stage">

              <div class="glass-panel recent-panel" v-if="recentArticles.length > 0">
                <div class="panel-header-mini">
                  <span class="mini-icon">⚡</span> 最新动态
                </div>
                <div class="pill-list">
                  <div
                      v-for="(article, index) in recentArticles"
                      :key="article.GUID"
                      class="article-pill"
                      @click="blogMainClick(article)"
                  >
                    <span class="pill-idx">{{ index + 1 }}</span>
                    <a :href="'/oneBlog/' + article.GUID" @click.prevent class="pill-text seo-article-link">{{ article.BLOG_TITLE }}</a>
                  </div>
                </div>
              </div>

              <div class="glass-panel content-panel">
                <div class="panel-header">
                  <div class="header-left">
                    <h2 class="panel-title">时间线</h2>
                    <span class="counter-badge" v-if="showBlogs.length">{{ showBlogs.length }}</span>
                  </div>

                  <label class="flat-switch">
                    <input type="checkbox" v-model="onlyArticle"/>
                    <span class="switch-box">
                      <span class="switch-handle"></span>
                    </span>
                    <span class="switch-label">剔除社区留言</span>
                  </label>
                </div>

                <div class="minimal-timeline" v-if="showBlogs.length > 0">
                  <div
                      v-for="(blog, index) in showBlogs"
                      :key="blog.GUID"
                      class="timeline-card"
                      :style="{ animationDelay: index * 0.05 + 's' }"
                  >
                    <div class="card-meta">
                      <span class="meta-type" :class="blog.TYPE">{{ blog.TYPE === 'blog' ? '文章' : '社区' }}</span>
                      <span class="meta-date">{{ formatDate(blog.CREATE_TIME) }}</span>
                    </div>
                    <div class="card-body" @click="blogMainClick(blog)">
                      <a
                          v-if="blog.TYPE === 'blog'"
                          :href="'/oneBlog/' + blog.GUID"
                          @click.prevent
                          class="seo-article-link"
                      ><h3 class="article-title">{{ blog.BLOG_TITLE }}</h3></a>
                      <h3 v-else class="article-title">{{ blog.BLOG_TITLE }}</h3>
                      <p class="article-desc">{{ listExcerpt(blog) }}</p>
                    </div>
                  </div>
                </div>

                <div v-if="!showBlogs.length && !loading" class="empty-view">
                  <span class="empty-emoji">🍃</span>
                  <p>风很轻，这里还是一片空白</p>
                </div>

                <div class="load-action-area">
                  <button v-if="!noMore && !loading" class="flat-ghost-btn" @click="fetchArticles(null)">
                    向下探索更多
                  </button>
                  <div v-if="loading" class="loading-wave">
                    <span></span><span></span><span></span>
                  </div>
                  <div v-if="noMore && showBlogs.length" class="end-line">
                    <div class="line"></div><span>触底了</span><div class="line"></div>
                  </div>
                </div>
              </div>

              <div class="glass-panel file-panel" ref="fileSection">
                <div class="panel-header">
                  <div class="header-left">
                    <h2 class="panel-title">资源库</h2>
                    <span class="counter-badge" v-if="files.length">{{ files.length }}</span>
                  </div>
                </div>

                <div v-if="files.length === 0" class="empty-view">
                  <span class="empty-emoji">🗂️</span>
                  <p>仓库里还没有货物</p>
                </div>

                <div v-else class="file-grid-modern">
                  <div
                      v-for="(file, index) in files"
                      :key="index"
                      class="file-box"
                  >
                    <div class="file-icon">{{ getFileIcon(file.ORIGINALFILENAME) }}</div>
                    <div class="file-info">
                      <div class="f-name" :title="file.ORIGINALFILENAME">{{ file.ORIGINALFILENAME }}</div>
                      <div class="f-sub">{{ formatDate(file.CREATE_TIME) }} · {{ file.DOWNNUM }} 次读取</div>
                    </div>
                    <button class="dl-circle-btn" @click="downloadFile(file)" title="下载">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                    </button>
                  </div>
                </div>
              </div>

            </main>
          </div>
        </transition>
      </div>
    </template>

    <el-dialog v-model="showDialog" width="85%" destroy-on-close top="4vh" @close="onDialogClose" class="custom-flat-dialog">
      <ContentAndComment :blogId="selectedBlogId"/>
    </el-dialog>
  </div>
</template>

<script setup>
// ==========================================
// 逻辑层保持原封不动，确保功能 100% 正常
// ==========================================
import {ref, onMounted, computed, defineAsyncComponent, watch} from 'vue'
import {ElMessage} from 'element-plus'
import {useRoute, useRouter} from 'vue-router'
import {useSeo} from '@/utils/seo.js';
import { pubFormatDate, sendAxiosRequest, listExcerpt, downloadFileByUrl } from '@/utils/common.js'

import BackgroundAndMusic from "@/components/detail/personInformation/BackgroundAndMusic.vue";
import WelcomeOverlay from "@/components/detail/personInformation/WelcomeOverlay.vue";
import BlogSidebar from "@/components/detail/myblog/BlogSidebar.vue";
import UserInfo from "@/components/main/UserInfo.vue";

const ContentAndComment = defineAsyncComponent(() => import('@/views/detail/blog/ContentAndComment.vue'))
const FriendLink = defineAsyncComponent(() => import('@/views/detail/friendLink/FriendLink.vue'))

import {useUserStore} from '@/stores/main/user.js'

const userStore = useUserStore()
userStore.initFromLocal()

const recentArticles = ref([])
const route = useRoute();
const router = useRouter();
const user = ref({})
const selectedCategory = ref('')
const targetUserCode = ref('');
const isPageReady = ref(false);

const isSelf = computed(() => !!userStore.userBean.code && !!targetUserCode.value && targetUserCode.value === userStore.userBean.code);

const bgMusicComponentRef = ref(null);
const serverBgImage = ref('');
const serverBgAudio = ref('');
const currentBgStyle = ref({});
const handleBgStyleUpdate = (style) => currentBgStyle.value = style;

const page = ref(1)
const pageSize = 5
const loading = ref(false)
const noMore = ref(false)
const blogs = ref([])
const files = ref([])
const onlyArticle = ref(false)
const showDialog = ref(false)
const selectedBlogId = ref('')
const fileSection = ref(null)
const showWelcome = ref(false);
const showFriendLink = ref(false);

// SEO：主页统一以 /user/:u 为收录地址；带文章 ID 打开时，文章本身以 /oneBlog/:id 为准
const seoTitle = ref('个人主页 - YnsStudy');
const seoDescription = ref('');
const userNotFound = ref(false);

useSeo(() => ({
  title: seoTitle.value,
  description: seoDescription.value,
  type: 'profile',
  path: route.params.blogId
      ? `/oneBlog/${encodeURIComponent(route.params.blogId)}`
      : `/user/${encodeURIComponent(route.params.u || '')}`,
  noindex: !route.params.u || userNotFound.value
}));

const initPageData = async () => {
  const userNum = route.params.u;
  if (!userNum) return;
  isPageReady.value = false;
  loading.value = true;
  try {
    const res = await sendAxiosRequest("/pub-api/login/getUserInfoByNum", { userNum });
    if (res && res.result && res.result.code) {
      const parsedCode = res.result.code;
      targetUserCode.value = parsedCode;
      user.value = res.result;
      // 设置页面标题
      seoTitle.value = (user.value.name || '用户') + "的个人主页 - YnsStudy";
      seoDescription.value = user.value.remark || `${user.value.name || '用户'} 在 YnsStudy 发表的文章、动态与分享的资源。`;

      isPageReady.value = true;
      await Promise.all([
        fetchArticles(parsedCode),
        fetchFilesList(parsedCode),
        setPersonInfo(parsedCode)
      ]);
    } else {
      userNotFound.value = true;
      ElMessage.error("未找到对应用户信息");
    }
  } catch (e) {
    console.error("加载主页数据失败", e);
  } finally {
    loading.value = false;
  }
}

const fetchArticles = async (userCode) => {
  const codeToUse = userCode || targetUserCode.value;
  if (noMore.value || !codeToUse) return;
  try {
    const res = await sendAxiosRequest('/blog-api/userInformation/getBlogAndCommunityByUserCode', {
      userCode: codeToUse, page: page.value, pageSize, keyword: ""
    })
    const newData = res.result.data || []
    if (newData.length < pageSize) noMore.value = true
    blogs.value.push(...newData)
    page.value++
    if (recentArticles.value.length < 5) {
      recentArticles.value = blogs.value.filter(item=>item.TYPE==='blog').slice(0, 5);
    }
  } catch (e) {}
}

const fetchFilesList = async (userCode) => {
  const result = await sendAxiosRequest('/blog-api/userInformation/getResourceByUserCode', {userCode});
  if (result && !result.isError) files.value = result.result;
};

const setPersonInfo = async (userCode) => {
  let result = await sendAxiosRequest("/blog-api/userInformation/getPersonInfo", {userCode});
  if (result && !result.isError) {
    const data = result.result[0] || {};
    serverBgAudio.value = data.BGMUSICURL || "";
    serverBgImage.value = data.BGIMAGEURL || "";
  }
}

const showBlogs = computed(() => {
  if (!onlyArticle.value) return blogs.value;
  return blogs.value.filter(item => item.TYPE === "blog");
})

function blogMainClick(blog) {
  if (blog.TYPE === "community") return ElMessage.info('该内容为社区留言，非独立文章');
  router.push({name: 'userV2', params: {u: route.params.u, blogId: blog.GUID}});
}

const onDialogClose = () => router.push({name: 'userV2', params: { u: route.params.u }});
const formatDate = (dateStr) => pubFormatDate(dateStr)
const downloadFile = (file) => downloadFileByUrl(file.FILEVIEWURL, file.ORIGINALFILENAME)

const getFileIcon = (filename) => {
  if (!filename) return '📄'
  const ext = filename.split('.').pop().toLowerCase()
  const map = {
    pdf: '📕', doc: '📘', docx: '📘', xls: '📗', xlsx: '📗', ppt: '📙', pptx: '📙',
    zip: '📦', rar: '📦', jpg: '🖼️', jpeg: '🖼️', png: '🖼️', gif: '🖼️',
    mp4: '🎬', mp3: '🎵', txt: '📄', md: '📝', js: '⚡', ts: '⚡', py: '🐍'
  }
  return map[ext] || '📄'
}

const scrollToFiles = () => {
  if (fileSection.value?.$el) fileSection.value.$el.scrollIntoView({behavior: 'smooth'})
  else if (fileSection.value) fileSection.value.scrollIntoView({behavior: 'smooth'})
}

const handleFriendLinkClick = () => showFriendLink.value = !showFriendLink.value;
const enterHomepage = async () => {
  showWelcome.value = false;
  if (bgMusicComponentRef.value) bgMusicComponentRef.value.playMusicForce();
};

onMounted(() => initPageData());
watch(() => route.params.userId, () => {
  page.value = 1; blogs.value = []; files.value = []; noMore.value = false; initPageData();
});
watch(() => route.params.blogId, (newBlogId) => {
  if (newBlogId) { selectedBlogId.value = newBlogId; showDialog.value = true; }
  else { selectedBlogId.value = ''; showDialog.value = false; }
}, { immediate: true });
</script>

<style scoped>
/* 手账风格的个人展厅 */
.elegant-glass-home {
  width: 100vw;
  height: 100vh;
  overflow-y: auto;
  background-size: cover;
  background-position: center;
  background-attachment: fixed;
  color: var(--j-ink);
}

.elegant-glass-home::-webkit-scrollbar { width: 6px; }
.elegant-glass-home::-webkit-scrollbar-thumb { background: rgba(120, 104, 80, 0.25); border-radius: 10px; }
.elegant-glass-home::-webkit-scrollbar-track { background: transparent; }

.home-container {
  max-width: 1300px;
  min-height: 100%;
  margin: 0 auto;
  padding: 24px 24px 64px;
}

/* --- 顶部：一条纸边 --- */
.glass-navbar {
  position: sticky;
  top: 16px;
  z-index: 100;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
  padding: 10px 20px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: 0 6px 14px -10px rgba(60, 50, 30, 0.4);
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: 10px;
}

.brand-dot {
  width: 10px;
  height: 10px;
  border: 2px solid var(--j-stamp);
  border-radius: 50%;
}

.brand-text {
  font-family: var(--j-hand);
  font-size: 18px;
}

.nav-actions {
  display: flex;
  gap: 10px;
}

.nav-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: 1px solid var(--j-rule-strong);
  background: var(--j-paper-warm);
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-ink);
  cursor: pointer;
  transition: background-color 0.2s ease, transform 0.2s ease;
}

.nav-btn:hover {
  background: var(--j-note);
  transform: translateY(-1px);
}

.primary-btn {
  background: #d4ead9;
}

/* --- 两栏 --- */
.layout-grid {
  display: flex;
  align-items: flex-start;
  gap: 36px;
}

.left-control-tower {
  flex-shrink: 0;
  width: 300px;
}

.sticky-tower-inner {
  position: sticky;
  top: 90px;
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.component-slot {
  transition: transform 0.3s ease;
}

.right-main-stage {
  display: flex;
  flex-direction: column;
  flex: 1;
  gap: 30px;
  min-width: 0;
}

/* 纸张面板 */
.glass-panel {
  position: relative;
  padding: 30px 32px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 12px;
  border-bottom: 1px dashed var(--j-rule-strong);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.panel-title {
  margin: 0;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 24px;
  color: var(--j-ink);
}

.counter-badge {
  padding: 0 8px;
  border-radius: 10px;
  background: #efe8da;
  font-size: 12px;
  line-height: 20px;
  color: var(--j-muted);
}

/* 最新动态：黄色便签 */
.recent-panel {
  padding: 22px 24px;
  border: none;
  background: var(--j-note);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.1), 0 12px 20px -14px rgba(60, 50, 30, 0.45);
  transform: rotate(-0.4deg);
}

.panel-header-mini {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 14px;
  font-family: var(--j-hand);
  font-size: 18px;
  color: var(--j-ink);
}

.pill-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.article-pill {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  background: rgba(255, 255, 255, 0.6);
  font-size: 13px;
  color: var(--j-ink);
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.article-pill:hover {
  background: #fff;
}

.pill-idx {
  font-family: var(--j-hand);
  color: #a2802a;
}

.pill-text {
  max-width: 180px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* 开关 */
.flat-switch {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.flat-switch input {
  display: none;
}

.switch-box {
  position: relative;
  width: 36px;
  height: 20px;
  border: 1px solid var(--j-rule-strong);
  border-radius: 10px;
  background: #f3eee3;
  transition: background-color 0.2s ease;
}

.switch-handle {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.3);
  transition: transform 0.2s ease;
}

.flat-switch input:checked + .switch-box {
  background: var(--j-note);
  border-color: #e3cf7a;
}

.flat-switch input:checked + .switch-box .switch-handle {
  transform: translateX(16px);
}

.switch-label {
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-ink-soft);
}

/* 时间线 */
.minimal-timeline {
  display: flex;
  flex-direction: column;
}

.timeline-card {
  padding: 18px 0;
  border-bottom: 1px dashed var(--j-rule);
  opacity: 0;
  animation: slideFadeIn 0.4s ease forwards;
}

.card-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.meta-type {
  padding: 0 8px;
  border: 1px solid currentColor;
  border-radius: 3px;
  font-family: var(--j-hand);
  font-size: 13px;
  line-height: 20px;
}

.meta-type.blog {
  color: var(--j-pen);
}

.meta-type.community {
  color: #4f8a5b;
}

.meta-date {
  font-family: var(--j-hand);
  font-size: 14px;
  color: var(--j-muted);
}

.card-body {
  cursor: pointer;
}

.article-title {
  display: inline;
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  line-height: 1.6;
  color: var(--j-ink);
  background-image: linear-gradient(transparent 58%, var(--j-highlight) 58%, var(--j-highlight) 92%, transparent 92%);
  background-size: 0 100%;
  background-repeat: no-repeat;
  transition: background-size 0.3s ease;
  -webkit-box-decoration-break: clone;
  box-decoration-break: clone;
}

.timeline-card:hover .article-title {
  background-size: 100% 100%;
}

.article-desc {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin: 8px 0 0;
  font-size: 14px;
  line-height: 1.7;
  color: var(--j-ink-soft);
}

.article-desc :deep(*) {
  margin: 0;
  font-size: inherit !important;
  font-weight: normal !important;
  color: inherit !important;
}

/* 资源库：牛皮纸 */
.file-panel {
  background: var(--j-kraft);
  border-color: #d8c49d;
}

.file-panel .panel-header {
  border-bottom-color: rgba(110, 85, 40, 0.3);
}

.file-grid-modern {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.file-box {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 14px;
  background: #fffaf0;
  box-shadow: 0 1px 1px rgba(110, 85, 40, 0.18);
  transition: transform 0.2s ease;
}

.file-box:hover {
  transform: translateY(-2px);
}

.file-icon {
  font-size: 24px;
  line-height: 1;
}

.file-info {
  flex: 1;
  min-width: 0;
}

.f-name {
  margin-bottom: 4px;
  overflow: hidden;
  font-size: 14px;
  font-weight: 600;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.f-sub {
  font-size: 12px;
  color: #7a6440;
}

.dl-circle-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 30px;
  height: 30px;
  border: 1px solid #d8c49d;
  border-radius: 50%;
  background: transparent;
  color: #6e5528;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.file-box:hover .dl-circle-btn {
  background: var(--j-note);
}

.empty-view {
  padding: 40px 0;
  text-align: center;
  font-family: var(--j-hand);
  font-size: 16px;
  color: var(--j-muted);
}

.empty-emoji {
  display: block;
  margin-bottom: 12px;
  font-size: 30px;
  opacity: 0.8;
}

.load-action-area {
  margin-top: 24px;
  text-align: center;
}

.flat-ghost-btn {
  padding: 8px 26px;
  border: 1px dashed var(--j-rule-strong);
  background: transparent;
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-ink-soft);
  cursor: pointer;
  transition: border-color 0.2s ease, color 0.2s ease;
}

.flat-ghost-btn:hover {
  border-color: var(--j-pen);
  color: var(--j-pen);
}

.loading-wave {
  display: inline-flex;
  gap: 6px;
  padding: 20px;
}

.loading-wave span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--j-rule-strong);
  animation: wave 1.2s infinite ease-in-out both;
}

.loading-wave span:nth-child(2) { animation-delay: 0.15s; }
.loading-wave span:nth-child(3) { animation-delay: 0.3s; }

.end-line {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-top: 20px;
}

.end-line .line {
  flex: 1;
  border-top: 1px dashed var(--j-rule-strong);
}

.end-line span {
  font-family: var(--j-hand);
  font-size: 14px;
  color: var(--j-muted);
}

/* 加载 */
.page-loading-mask {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background: var(--j-desk);
}

.loader-ripple {
  position: relative;
  display: inline-block;
  width: 60px;
  height: 60px;
}

.loader-ripple div {
  position: absolute;
  border: 3px solid var(--j-rule-strong);
  border-radius: 50%;
  opacity: 1;
  animation: ripple 1.5s cubic-bezier(0, 0.2, 0.8, 1) infinite;
}

.loader-ripple div:nth-child(2) {
  animation-delay: -0.5s;
}

.loading-text {
  margin-top: 16px;
  font-family: var(--j-hand);
  font-size: 16px;
  letter-spacing: 2px;
  color: var(--j-ink-soft);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@keyframes slideFadeIn { 0% { opacity: 0; transform: translateY(16px); } 100% { opacity: 1; transform: translateY(0); } }
@keyframes wave { 0%, 80%, 100% { transform: scale(0.6); opacity: 0.4; } 40% { transform: scale(1); opacity: 1; } }
@keyframes ripple { 0% { top: 28px; left: 28px; width: 0; height: 0; opacity: 0; } 5% { top: 28px; left: 28px; width: 0; height: 0; opacity: 1; } 100% { top: -1px; left: -1px; width: 58px; height: 58px; opacity: 0; } }

@media (max-width: 1024px) {
  .layout-grid { flex-direction: column; }
  .left-control-tower { width: 100%; }
  .sticky-tower-inner { position: static; display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
  .file-grid-modern { grid-template-columns: 1fr; }
}

@media (max-width: 768px) {
  .home-container { padding: 12px 14px 48px; }
  .sticky-tower-inner { grid-template-columns: 1fr; }
  .glass-panel { padding: 22px 18px; }
  .recent-panel { transform: none; }
  .nav-actions .btn-icon { display: none; }
}
</style>
