<template>
  <div class="person-info" :style="currentBgStyle">
    <div v-if="!isPageReady" class="page-loading-mask">
      <div class="loader-content">
        <div class="bouncing-dots">
          <div class="dot"></div>
          <div class="dot"></div>
          <div class="dot"></div>
        </div>
        <p class="loading-text">正在准备用户主页...</p>
      </div>
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

      <div class="inner-container">
        <div class="content-columns">

          <div class="person-left-wrapper">
            <UserInfo
                :user="user"
                :target-user-code="targetUserCode"
                @blog-click="blogMainClick"
            />
          </div>

          <div class="main-content">
            <div class="toolbar">
              <button class="toolbar-btn link-btn" @click="handleFriendLinkClick">
                <span class="toolbar-btn-icon">🤝</span>
                <span>友链</span>
              </button>

              <button class="toolbar-btn jump-btn" @click="scrollToFiles">
                <span class="toolbar-btn-icon">📁</span>
                <span>跳转到文件列表</span>
              </button>
            </div>

            <div class="section-card publish-card">
              <div class="section-header">
                <div class="section-title-group">
                  <span class="section-icon">📚</span>
                  <h3 class="section-title">发表内容</h3>
                  <span class="section-count" v-if="showBlogs.length">{{ showBlogs.length }}+</span>
                </div>
                <label class="toggle-switch">
                  <input type="checkbox" v-model="onlyArticle"/>
                  <span class="toggle-track">
                    <span class="toggle-thumb"></span>
                  </span>
                  <span class="toggle-label">仅文章</span>
                </label>
              </div>

              <div v-if="showBlogs.length > 0" class="blog-timeline">
                <div
                    v-for="(blog, index) in showBlogs"
                    :key="blog.GUID"
                    class="timeline-item"
                    :style="{ animationDelay: index * 0.06 + 's' }"
                >
                  <div class="timeline-connector">
                    <div class="timeline-dot" :class="blog.TYPE === 'blog' ? 'dot-article' : 'dot-community'"></div>
                    <div class="timeline-line" v-if="index < showBlogs.length - 1"></div>
                  </div>
                  <div class="timeline-content">
                    <div class="timeline-meta">
                      <span class="meta-tag" :class="blog.TYPE === 'blog' ? 'tag-article' : 'tag-community'">
                        {{ blog.TYPE === 'blog' ? '文章' : '社区' }}
                      </span>
                      <span class="meta-time">{{ formatDate(blog.CREATE_TIME) }}</span>
                    </div>
                    <div class="blog-card" @click="blogMainClick(blog)">
                      <h4 class="blog-title">{{ blog.BLOG_TITLE }}</h4>
                      <p class="blog-summary" v-html="stripImages(blog.MAINTEXT)"></p>
                      <div class="blog-card-footer">
                        <span class="read-more">阅读全文 →</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div v-if="!showBlogs.length && !loading" class="empty-state">
                <div class="empty-icon">✦</div>
                <p class="empty-text">暂无内容</p>
              </div>

              <button
                  v-if="!noMore && !loading"
                  class="load-more-btn"
                  @click="fetchArticles(null)"
              >
                <span>加载更多</span>
                <span class="load-more-arrow">↓</span>
              </button>
              <div v-if="loading" class="loading-row">
                <span class="loading-dot"></span>
                <span class="loading-dot"></span>
                <span class="loading-dot"></span>
              </div>
              <div v-if="noMore && showBlogs.length" class="end-line">
                <span class="end-dash"></span>
                <span class="end-text">已到末尾</span>
                <span class="end-dash"></span>
              </div>
            </div>

            <div class="section-card file-card" ref="fileSection">
              <div class="section-header">
                <div class="section-title-group">
                  <span class="section-icon">📁</span>
                  <h3 class="section-title">上传的文件</h3>
                  <span class="section-count" v-if="files.length">{{ files.length }}</span>
                </div>
              </div>

              <div v-if="files.length === 0" class="empty-state">
                <div class="empty-icon">◇</div>
                <p class="empty-text">暂无上传文件</p>
              </div>

              <div v-else class="file-list">
                <div
                    v-for="(file, index) in files"
                    :key="index"
                    class="file-item"
                    :style="{ animationDelay: index * 0.05 + 's' }"
                >
                  <div class="file-icon-wrap">
                    <span class="file-icon">{{ getFileIcon(file.ORIGINALFILENAME) }}</span>
                  </div>
                  <div class="file-info">
                    <span class="file-name">{{ file.ORIGINALFILENAME }}</span>
                    <span class="file-meta">{{ formatDate(file.CREATE_TIME) }} · 下载 {{ file.DOWNNUM }} 次</span>
                  </div>
                  <button class="download-btn" @click="downloadFile(file)">
                    <span>↓</span>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div class="right-sidebar">
            <div class="align-spacer"></div>

            <el-card class="sidebar-card recent-articles">
              <template #header>
                <div class="sidebar-card-title">
                  <span style="display: flex; align-items: center; gap: 6px;">
                    🔥 <span>最近文章</span>
                  </span>
                </div>
              </template>
              <ul class="recent-articles-list">
                <li
                    v-for="article in recentArticles"
                    :key="article.GUID"
                    class="recent-article-item"
                    :title="article.BLOG_TITLE"
                    @click="blogMainClick(article)"
                >
                  {{ article.BLOG_TITLE }}
                </li>
              </ul>
            </el-card>

            <div class="sticky-category">
              <BlogSidebar
                  :view_title="'归档'"
                  :user-code="targetUserCode"
                  @click-blog="blogMainClick"
                  v-model:selected-index="selectedCategory"
              />
            </div>
          </div>

        </div>
      </div>
    </template>

    <el-dialog v-model="showDialog" width="80%" destroy-on-close top="4vh" @close="onDialogClose">
      <ContentAndComment :blogId="selectedBlogId"/>
    </el-dialog>
  </div>
</template>

<script setup>
import {ref, onMounted, computed, defineAsyncComponent, watch} from 'vue'
import {ElMessage} from 'element-plus'
import {useRoute} from 'vue-router'
import {useRouter} from "vue-router";
import {
  pubFormatDate,
  sendAxiosRequest,
  stripImages,
  downloadFileByUrl
} from '@/utils/common.js'

import BackgroundAndMusic from "@/components/detail/personInformation/BackgroundAndMusic.vue";
import WelcomeOverlay from "@/components/detail/personInformation/WelcomeOverlay.vue";
import BlogSidebar from "@/components/detail/myblog/BlogSidebar.vue";
import UserInfo from "@/components/main/UserInfo.vue";

//正文与评论组件
const ContentAndComment = defineAsyncComponent(() => import('@/views/detail/blog/ContentAndComment.vue'))

import {useUserStore} from '@/stores/main/user.js'

const userStore = useUserStore()
userStore.initFromLocal()

const recentArticles = ref([])
const route = useRoute();
const router = useRouter();
const user = ref({})
const selectedCategory = ref('')
const targetUserCode = ref('');

// 新增：核心页面渲染锁
const isPageReady = ref(false);

const isSelf = computed(() =>
    !!userStore.userBean.code &&
    !!targetUserCode.value &&
    targetUserCode.value === userStore.userBean.code
);

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
const showWelcome = ref(true);

// ----------------------------------------------------
// 核心优化区
// ----------------------------------------------------
const initPageData = async () => {
  const userNum = route.params.u;
  if (!userNum) return;
  isPageReady.value = false; // 初始化时上锁，隐藏子组件
  loading.value = true;
  try {
    debugger;
    // 1. 先用数字拿到 CODE
    const res = await sendAxiosRequest("/pub-api/login/getUserInfoByNum", { userNum });
    if (res && res.result && res.result.code) {
      const parsedCode = res.result.code;
      targetUserCode.value = parsedCode;
      user.value = res.result;
      // 设置页面标题
      document.title = (user.value.name || '用户') + "的个人博客";
      const metaDesc = document.querySelector('meta[name="description"]');
      if (metaDesc) {
        metaDesc.setAttribute("content", document.title);
      } else {
        const desc = document.createElement('meta');
        desc.name = "description";
        desc.content = document.title;
        document.head.appendChild(desc);
      }
      // 【核心开关】用户信息和 CODE 都有了，允许子组件渲染！
      isPageReady.value = true;
      // 2. 然后再去查列表数据（这个时候组件已经挂载完毕了）
      await Promise.all([
        fetchArticles(parsedCode),
        fetchFilesList(parsedCode),
        setPersonInfo(parsedCode)
      ]);
    } else {
      ElMessage.error("未找到对应用户信息");
    }
  } catch (e) {
    console.error("加载主页数据失败", e);
  } finally {
    loading.value = false;
  }
}

const fetchArticles = async (userCode) => {
  // 如果没传，就用当前的
  const codeToUse = userCode || targetUserCode.value;
  if (noMore.value || !codeToUse) return;

  try {
    const res = await sendAxiosRequest('/blog-api/userInformation/getBlogAndCommunityByUserCode', {
      userCode: codeToUse,
      page: page.value,
      pageSize,
      keyword: ""
    })
    const newData = res.result.data || []
    if (newData.length < pageSize) noMore.value = true
    blogs.value.push(...newData)
    page.value++
    //给最近文章区域赋值
    if (recentArticles.value.length < 5) {
      recentArticles.value = blogs.value.filter(item=>item.TYPE==='blog').slice(0, 5);
    }
  } catch (e) {
    console.error('获取内容失败', e)
  }
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

// ----------------------------------------------------
// 其他方法保持原样
// ----------------------------------------------------

const showBlogs = computed(() => {
  if (!onlyArticle.value) return blogs.value;
  return blogs.value.filter(item => item.TYPE === "blog");
})

function blogMainClick(blog) {
  if (blog.TYPE === "community") {
    ElMessage.success('社区发起内容不用查看详情')
    return false
  }
  selectedBlogId.value = blog.GUID
  showDialog.value = true
}

const onDialogClose = () => {
  selectedBlogId.value = ''
  showDialog.value = false
}

const formatDate = (dateStr) => pubFormatDate(dateStr)

const downloadFile = (file) => {
  ElMessage.success(`准备下载：${file.ORIGINALFILENAME}`)
  downloadFileByUrl(file.FILEVIEWURL, file.ORIGINALFILENAME)
}

const getFileIcon = (filename) => {
  if (!filename) return '📄'
  const ext = filename.split('.').pop().toLowerCase()
  const map = {
    pdf: '📕', doc: '📘', docx: '📘', xls: '📗', xlsx: '📗',
    ppt: '📙', pptx: '📙', zip: '🗜️', rar: '🗜️', jpg: '🖼️',
    jpeg: '🖼️', png: '🖼️', gif: '🖼️', mp4: '🎬', mp3: '🎵',
    txt: '📄', md: '📝', js: '📜', ts: '📜', py: '📜'
  }
  return map[ext] || '📄'
}

const scrollToFiles = () => {
  if (fileSection.value?.$el) {
    fileSection.value.$el.scrollIntoView({behavior: 'smooth'})
  } else if (fileSection.value) {
    fileSection.value.scrollIntoView({behavior: 'smooth'})
  }
}

// 友链点击事件
const handleFriendLinkClick = () => {
  const routeUrl = router.resolve({name: "FriendLink"}).href;
  window.open(routeUrl,"FriendLink");
};

const enterHomepage = async () => {
  showWelcome.value = false;
  if (bgMusicComponentRef.value) {
    bgMusicComponentRef.value.playMusicForce();
  }
};

onMounted(() => {
  initPageData();
});

watch(() => route.params.u, () => {
  page.value = 1;
  blogs.value = [];
  files.value = [];
  noMore.value = false;
  initPageData();
});
</script>
<style scoped>
/* ===== CSS 变量 ===== */
:root {
  --card-bg: rgba(255, 255, 255, 0.82);
  --card-border: rgba(255, 255, 255, 0.6);
  --card-shadow: 0 8px 32px rgba(0, 0, 0, 0.10);
  --accent: #4f7cff;
  --accent-warm: #ff7043;
  --text-primary: #1a1f2e;
  --text-secondary: #6b7280;
  --text-muted: #adb5bd;
  --radius-lg: 18px;
  --radius-md: 12px;
  --radius-sm: 8px;
}

/* ===== 页面根 ===== */
.person-info {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  max-height: 100%;
  padding: 24px 0;
  box-sizing: border-box;
  position: relative;
  background-size: cover;
  background-attachment: fixed;
  transition: background-image 0.001s;
}

/* ===== 布局 ===== */
.inner-container {
  max-width: 1440px;
  margin: 0 auto;
  padding: 0 24px;
  box-sizing: border-box;
}

.content-columns {
  display: flex;
  gap: 40px;
  align-items: flex-start;
}

.person-left-wrapper {
  position: sticky;
  top: 20px;
  width: 280px;
  flex-shrink: 0;
  margin-top: 36px;
  z-index: 10;
}

.main-content {
  flex: 1;
  min-width: 0;
}

.right-sidebar {
  width: 280px;
  flex-shrink: 0;
}

/* 最近文章列表样式 */
.recent-articles-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.recent-article-item {
  padding: 10px 0;
  font-size: 14px;
  color: #333; /* 可以根据你的主题颜色自行调整 */
  cursor: pointer;
  border-bottom: 1px dashed #ebeef5; /* 分割线 */
  transition: color 0.3s ease;

  /* 核心逻辑：单行超出显示省略号 */
  white-space: nowrap; /* 强制不换行 */
  overflow: hidden; /* 超出部分隐藏 */
  text-overflow: ellipsis; /* 显示省略号 */
}

/* 最后一项去掉底部边框 */
.recent-article-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

/* 鼠标悬浮时的交互颜色 */
.recent-article-item:hover {
  color: #409EFF;
}

/* ===== 工具栏 ===== */
.toolbar {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 12px;
  height: 32px;
  margin-bottom: 14px;
}

.toolbar-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 0 16px;
  height: 32px;
  background: rgba(255, 255, 255, 0.75);
  border: 1px solid rgba(255, 255, 255, 0.7);
  border-radius: 20px;
  font-size: 13px;
  cursor: pointer;
  font-weight: 500;
  letter-spacing: 0.3px;
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  transition: all 0.2s ease;
}

.toolbar-btn-icon {
  font-size: 15px;
}

.jump-btn {
  color: #4f7cff;
  box-shadow: 0 2px 12px rgba(79, 124, 255, 0.10);
}

.jump-btn:hover {
  background: rgba(79, 124, 255, 0.12);
  border-color: rgba(79, 124, 255, 0.3);
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(79, 124, 255, 0.18);
}

.link-btn {
  color: #ff7043;
  box-shadow: 0 2px 12px rgba(255, 112, 67, 0.10);
}

.link-btn:hover {
  background: rgba(255, 112, 67, 0.12);
  border-color: rgba(255, 112, 67, 0.3);
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(255, 112, 67, 0.18);
}

/* ===== 通用卡片 ===== */
.section-card {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--card-shadow);
  margin-bottom: 28px;
  padding: 24px 28px;
  /*backdrop-filter: blur(12px);/*中间发表内容区域模糊效果*/
  -webkit-backdrop-filter: blur(12px);
  transition: box-shadow 0.25s ease;
}

.section-card:hover {
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.13);
}

/* ===== 区块标题 ===== */
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 22px;
  padding-bottom: 16px;
  border-bottom: 1.5px solid rgba(0, 0, 0, 0.055);
}

.section-title-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.section-icon {
  font-size: 20px;
  line-height: 1;
}

.section-title {
  font-size: 17px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  letter-spacing: 0.2px;
}

.section-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 24px;
  height: 20px;
  padding: 0 7px;
  background: rgba(79, 124, 255, 0.12);
  color: #4f7cff;
  border-radius: 20px;
  font-size: 11px;
  font-weight: 600;
}

/* ===== 自定义开关 ===== */
.toggle-switch {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.toggle-switch input {
  display: none;
}

.toggle-track {
  position: relative;
  width: 36px;
  height: 20px;
  background: #d1d5db;
  border-radius: 10px;
  transition: background 0.2s;
}

.toggle-switch input:checked + .toggle-track {
  background: #4f7cff;
}

.toggle-thumb {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 14px;
  height: 14px;
  background: #fff;
  border-radius: 50%;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.18);
  transition: transform 0.2s;
}

.toggle-switch input:checked + .toggle-track .toggle-thumb {
  transform: translateX(16px);
}

.toggle-label {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 500;
}

/* ===== 时间线 ===== */
.blog-timeline {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.timeline-item {
  display: flex;
  gap: 16px;
  animation: fadeSlideUp 0.45s ease both;
}

@keyframes fadeSlideUp {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 连接线 */
.timeline-connector {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
  width: 18px;
  padding-top: 14px;
}

.timeline-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
  border: 2.5px solid #fff;
  box-shadow: 0 0 0 2px currentColor;
  transition: transform 0.2s;
}

.dot-article {
  background: #4f7cff;
  color: #4f7cff;
}

.dot-community {
  background: #ff7043;
  color: #ff7043;
}

.timeline-line {
  flex: 1;
  width: 2px;
  background: linear-gradient(to bottom, rgba(79, 124, 255, 0.2), rgba(79, 124, 255, 0.06));
  margin: 6px 0;
  min-height: 24px;
}

/* 时间线内容 */
.timeline-content {
  flex: 1;
  min-width: 0;
  padding-bottom: 20px;
}

.timeline-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.meta-tag {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 20px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.tag-article {
  background: rgba(79, 124, 255, 0.12);
  color: #4f7cff;
}

.tag-community {
  background: rgba(255, 112, 67, 0.12);
  color: #ff7043;
}

.meta-time {
  font-size: 12px;
  color: var(--text-muted);
}

/* ===== 博客卡片 ===== */
.blog-card {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.8);
  border-radius: var(--radius-md);
  padding: 16px 18px;
  cursor: pointer;
  transition: all 0.22s ease;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.04);
}

.blog-card:hover {
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 6px 24px rgba(79, 124, 255, 0.10);
  transform: translateY(-2px);
  border-color: rgba(79, 124, 255, 0.25);
}

.blog-card:hover .timeline-dot {
  transform: scale(1.2);
}

.blog-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 8px 0;
  line-height: 1.45;
  transition: color 0.2s;
}

.blog-card:hover .blog-title {
  color: #4f7cff;
}

.blog-summary {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-secondary);
  margin: 0 0 10px 0;
}

.blog-card-footer {
  display: flex;
  justify-content: flex-end;
}

.read-more {
  font-size: 12px;
  color: #4f7cff;
  font-weight: 500;
  opacity: 0;
  transform: translateX(-6px);
  transition: all 0.2s;
}

.blog-card:hover .read-more {
  opacity: 1;
  transform: translateX(0);
}

/* ===== 加载更多 ===== */
.load-more-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  margin-top: 16px;
  padding: 11px 0;
  background: transparent;
  border: 1.5px dashed rgba(79, 124, 255, 0.3);
  border-radius: var(--radius-md);
  font-size: 14px;
  color: #000000;
  background-color: rgba(14, 165, 233, 0.22);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.load-more-btn:hover {
  background: rgba(79, 124, 255, 0.06);
  border-style: solid;
  border-color: rgba(79, 124, 255, 0.45);
}

.load-more-arrow {
  font-size: 16px;
  animation: bounce 1.4s ease infinite;
}

@keyframes bounce {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(4px);
  }
}

/* ===== 加载动画 ===== */
.loading-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 20px 0;
}

.loading-dot {
  width: 8px;
  height: 8px;
  background: #4f7cff;
  border-radius: 50%;
  animation: loadPulse 1.2s ease infinite;
  opacity: 0.7;
}

.loading-dot:nth-child(2) {
  animation-delay: 0.2s;
}

.loading-dot:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes loadPulse {
  0%, 100% {
    transform: scale(0.7);
    opacity: 0.4;
  }
  50% {
    transform: scale(1);
    opacity: 1;
  }
}

/* ===== 结尾 ===== */
.end-line {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 20px;
  padding: 8px 0;
}

.end-dash {
  flex: 1;
  height: 1px;
  background: linear-gradient(to right, transparent, rgba(0, 0, 0, 0.1), transparent);
}

.end-text {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
  letter-spacing: 0.5px;
}

/* ===== 空状态 ===== */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 20px;
  gap: 12px;
}

.empty-icon {
  font-size: 32px;
  color: #c9d0da;
  line-height: 1;
}

.empty-text {
  font-size: 14px;
  color: var(--text-muted);
  margin: 0;
}

/* ===== 文件列表 ===== */
.file-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 14px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.45);
  border: 1px solid transparent;
  transition: all 0.2s;
  animation: fadeSlideUp 0.4s ease both;
}

.file-item:hover {
  background: rgba(255, 255, 255, 0.85);
  border-color: rgba(79, 124, 255, 0.18);
  box-shadow: 0 3px 14px rgba(79, 124, 255, 0.07);
}

.file-icon-wrap {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  background: rgba(79, 124, 255, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}

.file-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.file-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-meta {
  font-size: 12px;
  color: var(--text-muted);
}

.download-btn {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 1.5px solid rgba(79, 124, 255, 0.25);
  background: rgba(79, 124, 255, 0.07);
  color: #4f7cff;
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  flex-shrink: 0;
}

.download-btn:hover {
  background: #4f7cff;
  border-color: #4f7cff;
  color: #fff;
  transform: translateY(1px);
  box-shadow: 0 4px 12px rgba(79, 124, 255, 0.3);
}

/* ===== 右侧栏 ===== */
.sidebar-card {
  background-color: rgba(255, 255, 255, 0.88) !important;
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: 12px;

  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.07);
  margin-bottom: 22px;
}

.sidebar-card-title {
  font-size: 15px;
  font-weight: bold;
  color: var(--text-primary);
}

.sticky-category {
  position: sticky;
  top: 20px;
}

.align-spacer {
  height: 36px;
}
.page-loading-mask {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100vh;
  background-color: #f5f7fa; /* 非常浅的高级灰底色 */
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 9999; /* 确保它盖住所有东西 */
}

.loader-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.bouncing-dots {
  display: flex;
  gap: 8px;
}

.bouncing-dots .dot {
  width: 12px;
  height: 12px;
  background-color: #409EFF; /* Element Plus 的主题蓝 */
  border-radius: 50%;
  animation: bounce 1.4s infinite ease-in-out both;
}

.bouncing-dots .dot:nth-child(1) { animation-delay: -0.32s; }
.bouncing-dots .dot:nth-child(2) { animation-delay: -0.16s; }

.loading-text {
  color: #606266;
  font-size: 1.1rem;
  letter-spacing: 2px;
  margin: 0;
  animation: pulse-text 2s infinite ease-in-out;
}

@keyframes bounce {
  0%, 80%, 100% { transform: scale(0); }
  40% { transform: scale(1); }
}

@keyframes pulse-text {
  0%, 100% { opacity: 0.6; }
  50% { opacity: 1; }
}
/* ================= 响应式 ================= */
@media (max-width: 1024px) {
  .content-columns {
    flex-direction: column;
    gap: 24px;
    align-items: stretch; /* 关键修复：让内部的栏目自动撑满屏幕宽度 */
  }

  /* 关键修复：把 main-content 也加进来，强制 100% 宽度 */
  .person-left-wrapper,
  .main-content,
  .right-sidebar {
    width: 100%;
    position: static;
    margin-top: 0;
  }

  .align-spacer {
    display: none;
  }
}

@media (max-width: 768px) {
  .inner-container {
    padding: 0 12px; /* 缩小整体页面的左右留白 */
  }

  .section-card {
    padding: 16px 12px; /* 缩小卡片内部的边距，把空间还给内容 */
  }

  .section-header {
    flex-wrap: wrap; /* 防止“仅文章”开关和标题挤在一起重叠 */
    gap: 12px;
  }

  .timeline-connector {
    display: none;
  }

  .timeline-item {
    gap: 0;
  }

  .timeline-content {
    padding-bottom: 16px;
  }

  /* 优化文件列表在手机端的显示，防止文件名过长撑破卡片 */
  .file-item {
    padding: 12px 10px;
    gap: 10px;
  }

  .file-name {
    font-size: 13px;
  }
}
</style>