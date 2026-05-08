<template>
  <div class="elegant-glass-home" :style="currentBgStyle">
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
                    <span class="pill-text">{{ article.BLOG_TITLE }}</span>
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
                      <h3 class="article-title">{{ blog.BLOG_TITLE }}</h3>
                      <p class="article-desc" v-html="stripImages(blog.MAINTEXT)"></p>
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
import { useHead } from '@vueuse/head'; // 1. 引入 useHead
import { pubFormatDate, sendAxiosRequest, stripImages, downloadFileByUrl } from '@/utils/common.js'

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

// 2. 定义响应式的 SEO 数据源，给定默认值
const seoTitle = ref('ynsStudy的个人博客');
const seoDescription = ref('ynsStudy的个人博客');

useHead({
  title: seoTitle,
  meta: [
    {
      name: 'description',
      content: seoDescription
    }
  ]
});

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
      seoTitle.value = (user.value.name || '用户') + "的个人博客";
      seoDescription.value = seoTitle.value;

      isPageReady.value = true;
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
/* =========================================================
   V4: 扁平画框与展厅 (Flat Gallery Framework)
   核心理念：不干预通用组件，用布局和通透感包容一切
========================================================= */
:root {
  --base-text: #2c3e50;
  --sub-text: #5f6c7b;
  --light-text: #909399;
  --accent-color: #0ea5e9;
  --glass-bg: rgba(255, 255, 255, 0.55); /* 适中的透光率，不刺眼 */
  --glass-border: rgba(255, 255, 255, 0.7);
  --glass-shadow: 0 4px 24px rgba(0, 0, 0, 0.04);
}

.elegant-glass-home {
  width: 100vw;
  height: 100vh;
  overflow-y: overlay;
  background-size: cover;
  background-position: center;
  background-attachment: fixed;
  font-family: system-ui, -apple-system, sans-serif;
  color: var(--base-text);
}

/* 优雅的细体滚动条 */
.elegant-glass-home::-webkit-scrollbar { width: 6px; }
.elegant-glass-home::-webkit-scrollbar-thumb { background: rgba(0,0,0,0.15); border-radius: 10px; }
.elegant-glass-home::-webkit-scrollbar-track { background: transparent; }

.home-container {
  max-width: 1300px;
  margin: 0 auto;
  padding: 20px 24px 60px;
  min-height: 100%;
}

/* --- 顶部通透导航栏 --- */
.glass-navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 24px;
  background: rgba(255, 255, 255, 0.4);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: 100px;
  margin-bottom: 32px;
  position: sticky;
  top: 20px;
  z-index: 100;
  box-shadow: 0 4px 12px rgba(0,0,0,0.02);
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: 10px;
}
.brand-dot { width: 8px; height: 8px; background: var(--accent-color); border-radius: 50%; box-shadow: 0 0 8px var(--accent-color); }
.brand-text { font-weight: 700; font-size: 16px; letter-spacing: 0.5px; }

.nav-actions { display: flex; gap: 12px; }
.nav-btn {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 16px;
  border-radius: 30px;
  font-size: 13px; font-weight: 600;
  cursor: pointer;
  background: rgba(255,255,255,0.6);
  border: 1px solid rgba(255,255,255,0.8);
  color: var(--base-text);
  transition: all 0.25s ease;
}
.nav-btn:hover { background: #fff; transform: translateY(-1px); box-shadow: 0 4px 10px rgba(0,0,0,0.05); }
.primary-btn { color: var(--accent-color); }

/* --- 核心两栏排版 (Layout Grid) --- */
.layout-grid {
  display: flex;
  align-items: flex-start;
  gap: 40px; /* 拉大间距，防止不同风格的组件挤在一起产生视觉冲突 */
}

/* --- 左侧控制塔 (包容通用组件) --- */
.left-control-tower {
  width: 300px;
  flex-shrink: 0;
}
.sticky-tower-inner {
  position: sticky;
  top: 90px; /* 避开顶部的 navbar */
  display: flex;
  flex-direction: column;
  gap: 24px; /* 通用组件之间的安全距离 */
}
/* 给通用组件一个柔和的承载外壳，不修改内部，只加一点呼吸感过渡 */
.component-slot {
  border-radius: 12px; /* 假设通用组件圆角不大，我们也不做夸张包裹 */
  transition: transform 0.3s ease;
}
.component-slot:hover { transform: translateY(-2px); }

/* --- 右侧展厅主体 --- */
.right-main-stage {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 28px;
}

/* 玻璃面板通用类 */
.glass-panel {
  background: var(--glass-bg);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  border: 1px solid var(--glass-border);
  border-radius: 20px;
  padding: 32px;
  box-shadow: var(--glass-shadow);
}

/* 面板标题 */
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 28px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(0,0,0,0.04);
}
.header-left { display: flex; align-items: center; gap: 12px; }
.panel-title { margin: 0; font-size: 20px; font-weight: 700; color: #000; }
.counter-badge {
  background: rgba(0,0,0,0.05); color: var(--sub-text);
  padding: 2px 10px; border-radius: 20px; font-size: 13px; font-weight: bold;
}

/* 近期速览 - 胶囊药丸横排 */
.recent-panel { padding: 20px 24px; }
.panel-header-mini { font-size: 14px; font-weight: 700; color: var(--sub-text); margin-bottom: 16px; display: flex; align-items: center; gap: 6px; }
.pill-list { display: flex; flex-wrap: wrap; gap: 10px; }
.article-pill {
  display: flex; align-items: center; gap: 8px;
  background: rgba(255,255,255,0.7);
  border: 1px solid rgba(255,255,255,0.9);
  padding: 6px 14px;
  border-radius: 30px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.article-pill:hover { background: #fff; transform: scale(1.02); color: var(--accent-color); box-shadow: 0 2px 8px rgba(0,0,0,0.05); }
.pill-idx { font-family: monospace; font-weight: bold; color: #a1a1aa; }
.pill-text { font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 180px; }

/* 自定义扁平开关 */
.flat-switch { display: flex; align-items: center; gap: 8px; cursor: pointer; user-select: none; }
.flat-switch input { display: none; }
.switch-box {
  width: 36px; height: 20px; background: rgba(0,0,0,0.1); border-radius: 12px; position: relative; transition: 0.3s;
}
.switch-handle {
  position: absolute; top: 2px; left: 2px; width: 16px; height: 16px; background: #fff; border-radius: 50%; transition: 0.3s;
}
.flat-switch input:checked + .switch-box { background: var(--base-text); }
.flat-switch input:checked + .switch-box .switch-handle { transform: translateX(16px); }
.switch-label { font-size: 13px; font-weight: 600; color: var(--sub-text); }

/* 极简无连线时间轴 */
.minimal-timeline { display: flex; flex-direction: column; gap: 20px; }
.timeline-card {
  background: rgba(255, 255, 255, 0.6);
  border: 1px solid rgba(255,255,255,0.8);
  border-radius: 16px;
  padding: 20px 24px;
  transition: all 0.3s ease;
  animation: slideFadeIn 0.5s ease both;
}
.timeline-card:hover {
  background: rgba(255,255,255,0.9);
  transform: translateX(4px);
  border-color: #fff;
  box-shadow: 0 4px 16px rgba(0,0,0,0.03);
}

.card-meta { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.meta-type { font-size: 11px; font-weight: 800; padding: 3px 8px; border-radius: 6px; text-transform: uppercase; letter-spacing: 0.5px; }
.meta-type.blog { background: rgba(14, 165, 233, 0.1); color: var(--accent-color); }
.meta-type.community { background: rgba(0,0,0,0.05); color: var(--sub-text); }
.meta-date { font-size: 12px; font-family: monospace; color: var(--light-text); }

.card-body { cursor: pointer; }
.article-title { margin: 0 0 8px 0; font-size: 17px; font-weight: 700; color: #000; line-height: 1.4; transition: 0.2s; }
.timeline-card:hover .article-title { color: var(--accent-color); }
.article-desc { margin: 0; font-size: 14px; color: var(--sub-text); line-height: 1.6; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

/* 现代文件网格 */
.file-grid-modern {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}
.file-box {
  display: flex; align-items: center; gap: 16px;
  padding: 16px;
  background: rgba(255,255,255,0.6);
  border: 1px solid rgba(255,255,255,0.8);
  border-radius: 16px;
  transition: all 0.25s;
}
.file-box:hover {
  background: #fff; transform: translateY(-2px); box-shadow: 0 6px 16px rgba(0,0,0,0.04);
}
.file-icon { font-size: 26px; line-height: 1; }
.file-info { flex: 1; min-width: 0; }
.f-name { font-size: 14px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin-bottom: 4px; }
.f-sub { font-size: 12px; color: var(--light-text); }

.dl-circle-btn {
  width: 32px; height: 32px; border-radius: 50%;
  background: rgba(0,0,0,0.04); border: none; color: var(--sub-text);
  display: flex; align-items: center; justify-content: center;
  cursor: pointer; transition: 0.2s; flex-shrink: 0;
}
.file-box:hover .dl-circle-btn { background: var(--accent-color); color: #fff; transform: scale(1.1); }

/* 加载与状态 */
.empty-view { text-align: center; padding: 40px 0; color: var(--light-text); }
.empty-emoji { font-size: 32px; display: block; margin-bottom: 12px; opacity: 0.8; }

.load-action-area { margin-top: 24px; text-align: center; }
.flat-ghost-btn {
  background: transparent; border: 1.5px dashed rgba(0,0,0,0.15);
  color: var(--sub-text); padding: 12px 32px; border-radius: 30px; font-weight: 600; font-size: 14px;
  cursor: pointer; transition: 0.3s;
}
.flat-ghost-btn:hover { border-style: solid; border-color: var(--base-text); color: var(--base-text); }

.loading-wave { display: inline-flex; gap: 6px; padding: 20px; }
.loading-wave span { width: 6px; height: 6px; background: var(--light-text); border-radius: 50%; animation: wave 1.2s infinite ease-in-out both; }
.loading-wave span:nth-child(2) { animation-delay: 0.15s; }
.loading-wave span:nth-child(3) { animation-delay: 0.3s; }

.end-line { display: flex; align-items: center; gap: 16px; margin-top: 20px; opacity: 0.6; }
.end-line .line { flex: 1; height: 1px; background: rgba(0,0,0,0.1); }
.end-line span { font-size: 12px; font-family: monospace; color: var(--light-text); }

/* 页面加载动画 */
.page-loading-mask { position: fixed; inset: 0; background: #fafafa; display: flex; flex-direction: column; justify-content: center; align-items: center; z-index: 9999; }
.loader-ripple { display: inline-block; position: relative; width: 60px; height: 60px; }
.loader-ripple div { position: absolute; border: 3px solid #cbd5e1; opacity: 1; border-radius: 50%; animation: ripple 1.5s cubic-bezier(0, 0.2, 0.8, 1) infinite; }
.loader-ripple div:nth-child(2) { animation-delay: -0.5s; }
.loading-text { margin-top: 16px; color: #94a3b8; font-size: 13px; letter-spacing: 2px; }

/* 动画定义 */
@keyframes slideFadeIn { 0% { opacity: 0; transform: translateY(16px); } 100% { opacity: 1; transform: translateY(0); } }
@keyframes wave { 0%, 80%, 100% { transform: scale(0.6); opacity: 0.4; } 40% { transform: scale(1); opacity: 1; } }
@keyframes ripple { 0% { top: 28px; left: 28px; width: 0; height: 0; opacity: 0; } 5% { top: 28px; left: 28px; width: 0; height: 0; opacity: 1; } 100% { top: -1px; left: -1px; width: 58px; height: 58px; opacity: 0; } }

/* 响应式适配 */
@media (max-width: 1024px) {
  .layout-grid { flex-direction: column; }
  .left-control-tower { width: 100%; }
  .sticky-tower-inner { position: static; display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
  .file-grid-modern { grid-template-columns: 1fr; }
}

@media (max-width: 768px) {
  .home-container { padding: 12px 16px; }
  .sticky-tower-inner { grid-template-columns: 1fr; }
  .glass-panel { padding: 20px; }
  .timeline-card { padding: 16px; }
  .nav-actions .btn-icon { display: none; } /* 移动端隐藏文字旁的图标节省空间 */
}
</style>