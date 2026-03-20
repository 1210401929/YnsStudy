<template>
  <div class="person-info" :style="currentBgStyle">
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

    <!-- 顶部渐变装饰条 -->
    <div class="top-bar-deco"></div>

    <div class="page-shell">

      <!-- ===== 左侧用户信息栏 ===== -->
      <aside class="sidebar-left">
        <div class="sidebar-left-inner">
          <UserInfo
              :user="user"
              :target-user-code="targetUserCode"
              @blog-click="blogMainClick"
          />
        </div>
      </aside>

      <!-- ===== 主体内容 ===== -->
      <main class="main-col">

        <!-- 快捷跳转胶囊 -->
        <div class="pill-toolbar">
          <button class="pill-btn" @click="scrollToFiles">
            <span class="pill-icon">📁</span>跳转文件列表
          </button>
        </div>

        <!-- 发表内容卡片 -->
        <section class="bento-card publish-card">
          <div class="card-head">
            <div class="card-title-row">
              <span class="card-emoji">📚</span>
              <span class="card-title">发表内容</span>
              <span class="badge-count" v-if="showBlogs.length">{{ showBlogs.length }}+</span>
            </div>
            <label class="chip-toggle">
              <input type="checkbox" v-model="onlyArticle"/>
              <span class="chip-inner" :class="{ 'chip-on': onlyArticle }">仅文章</span>
            </label>
          </div>

          <div v-if="showBlogs.length > 0" class="post-grid">
            <div
                v-for="(blog, index) in showBlogs"
                :key="blog.GUID"
                class="post-card"
                :style="{ animationDelay: index * 0.07 + 's' }"
                @click="blogMainClick(blog)"
            >
              <div class="post-card-top">
                <span class="post-type-tag" :class="blog.TYPE === 'blog' ? 'tag-art' : 'tag-comm'">
                  {{ blog.TYPE === 'blog' ? '✦ 文章' : '◎ 社区' }}
                </span>
                <span class="post-date">{{ formatDate(blog.CREATE_TIME) }}</span>
              </div>
              <h4 class="post-title">{{ blog.BLOG_TITLE }}</h4>
              <p class="post-excerpt" v-html="stripImages(blog.MAINTEXT)"></p>
              <div class="post-footer">
                <span class="arrow-link">阅读全文 <span class="arrow">→</span></span>
              </div>
            </div>
          </div>

          <div v-if="!showBlogs.length && !loading" class="empty-state">
            <div class="empty-blob">✦</div>
            <p>暂无内容</p>
          </div>

          <button
              v-if="!noMore && !loading"
              class="load-more"
              @click="fetchArticles(null)"
          >
            <span>加载更多</span>
            <span class="bounce-arrow">↓</span>
          </button>

          <div v-if="loading" class="dots-loading">
            <span></span><span></span><span></span>
          </div>

          <div v-if="noMore && showBlogs.length" class="end-tag">
            <span class="end-pill">— 已经到底啦 —</span>
          </div>
        </section>

        <!-- 文件列表卡片 -->
        <section class="bento-card file-card" ref="fileSection">
          <div class="card-head">
            <div class="card-title-row">
              <span class="card-emoji">📁</span>
              <span class="card-title">上传的文件</span>
              <span class="badge-count" v-if="files.length">{{ files.length }}</span>
            </div>
          </div>

          <div v-if="files.length === 0" class="empty-state">
            <div class="empty-blob">◇</div>
            <p>暂无文件</p>
          </div>

          <div v-else class="file-grid">
            <div
                v-for="(file, index) in files"
                :key="index"
                class="file-row"
                :style="{ animationDelay: index * 0.05 + 's' }"
            >
              <div class="file-icon-box">{{ getFileIcon(file.ORIGINALFILENAME) }}</div>
              <div class="file-info">
                <span class="file-name">{{ file.ORIGINALFILENAME }}</span>
                <span class="file-sub">{{ formatDate(file.CREATE_TIME) }} · 下载 {{ file.DOWNNUM }} 次</span>
              </div>
              <button class="dl-btn" @click="downloadFile(file)" title="下载">↓</button>
            </div>
          </div>
        </section>

      </main>

      <!-- ===== 右侧边栏 ===== -->
      <aside class="sidebar-right">
        <div class="sidebar-spacer"></div>

        <!-- 最近文章 -->
        <div class="bento-mini hot-card">
          <div class="mini-head">
            <span>🔥</span>
            <span class="mini-title">最近文章</span>
          </div>
          <ul class="hot-list">
            <li
                v-for="article in recentArticles"
                :key="article.GUID"
                class="hot-item"
                :title="article.BLOG_TITLE"
                @click="blogMainClick(article)"
            >
              <span class="hot-dot"></span>
              <span class="hot-text">{{ article.BLOG_TITLE }}</span>
            </li>
          </ul>
        </div>

        <!-- 归档 -->
        <div class="sticky-wrap">
          <BlogSidebar
              :view_title="'归档'"
              :user-code="targetUserCode"
              @click-blog="blogMainClick"
              v-model:selected-index="selectedCategory"
          />
        </div>
      </aside>
    </div>

    <el-dialog v-model="showDialog" width="80%" destroy-on-close top="4vh" @close="onDialogClose">
      <ContentAndComment :blogId="selectedBlogId"/>
    </el-dialog>
  </div>
</template>

<script setup>
import {ref, onMounted, computed, defineAsyncComponent} from 'vue'
import {ElMessage} from 'element-plus'
import {useRoute} from 'vue-router'
import {
  pubFormatDate,
  decrypt,
  sendAxiosRequest,
  stripImages,
  downloadFileByUrl,
  getUserInfoByCode
} from '@/utils/common.js'

import BackgroundAndMusic from "@/components/detail/personInformation/BackgroundAndMusic.vue";
import WelcomeOverlay from "@/components/detail/personInformation/WelcomeOverlay.vue";
import BlogSidebar from "@/components/detail/myblog/BlogSidebar.vue";
import UserInfo from "@/components/main/UserInfo.vue";

const ContentAndComment = defineAsyncComponent(() => import('@/views/detail/blog/ContentAndComment.vue'))
import {useUserStore} from '@/stores/main/user.js'

const userStore = useUserStore()
userStore.initFromLocal()

const recentArticles = ref([])
const route = useRoute()
const user = ref({})
const selectedCategory = ref('')

const targetUserCode = ref('');
try {
  const raw = route.params.u;
  targetUserCode.value = raw ? decrypt(raw) : '';
} catch (e) {
  targetUserCode.value = '';
}

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

const fetchArticles = async (userCode) => {
  if (loading.value || noMore.value) return;
  if (!userCode) userCode = targetUserCode.value;
  loading.value = true
  try {
    const res = await sendAxiosRequest('/blog-api/userInformation/getBlogAndCommunityByUserCode', {
      userCode, page: page.value, pageSize, keyword: ""
    })
    const newData = res.result.data
    if (newData.length < pageSize) noMore.value = true
    blogs.value.push(...newData)
    page.value++
    if (recentArticles.value.length < 5) {
      recentArticles.value = [...blogs.value].filter(item => item.TYPE === 'blog').splice(0, 5);
    }
  } catch (e) {
    console.error('获取内容失败', e)
  } finally {
    loading.value = false
  }
}

async function getUserInfo2Data() {
  const userCode = targetUserCode.value;
  if (!userCode) return;
  const fetchBasicInfo = async () => {
    const result = await getUserInfoByCode(userCode);
    if (result && !result.isError) {
      user.value = result.result;
      document.title = user.value.name + "的主页";
      const metaDesc = document.querySelector('meta[name="description"]');
      if (metaDesc) {
        metaDesc.setAttribute("content", document.title);
      } else {
        const desc = document.createElement('meta');
        desc.name = "description";
        desc.content = document.title;
        document.head.appendChild(desc);
      }
    }
  };
  const fetchFilesList = async () => {
    const result = await sendAxiosRequest('/blog-api/userInformation/getResourceByUserCode', {userCode});
    if (result && !result.isError) files.value = result.result;
  };
  const setPersonInfo = async () => {
    let result = await sendAxiosRequest("/blog-api/userInformation/getPersonInfo", {userCode});
    if (result && !result.isError) {
      result = result.result[0] || {};
      serverBgAudio.value = result.BGMUSICURL || "";
      serverBgImage.value = result.BGIMAGEURL || "";
    }
  }
  fetchBasicInfo();
  fetchArticles(userCode);
  fetchFilesList();
  setPersonInfo();
}

const blogs = ref([])
const files = ref([])
const onlyArticle = ref(false)
const showDialog = ref(false)
const selectedBlogId = ref('')

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

const fileSection = ref(null)
const scrollToFiles = () => {
  if (fileSection.value?.$el) {
    fileSection.value.$el.scrollIntoView({behavior: 'smooth'})
  } else if (fileSection.value) {
    fileSection.value.scrollIntoView({behavior: 'smooth'})
  }
}

const showWelcome = ref(true);
const enterHomepage = async () => {
  showWelcome.value = false;
  if (bgMusicComponentRef.value) {
    bgMusicComponentRef.value.playMusicForce();
  }
};

onMounted(() => {
  getUserInfo2Data();
});
</script>

<style scoped>
/* ============================================================
   设计语言：白色圆角 Bento 系 · 活力年轻化
   主调：纯白卡片 + 柔粉/天蓝/薄荷绿渐变点缀
   字体：Noto Sans SC + 系统等宽
   ============================================================ */

@import url('https://fonts.googleapis.com/css2?family=Noto+Sans+SC:wght@400;500;600;700;900&display=swap');

/* ── CSS 变量 ── */
:root {
  --white: #ffffff;
  --surface: #f5f6fa;
  --card-bg: rgba(255, 255, 255, 0.96);
  --card-radius: 24px;
  --mini-radius: 16px;
  --shadow-sm: 0 2px 12px rgba(0,0,0,.06);
  --shadow-md: 0 6px 28px rgba(0,0,0,.09);
  --shadow-lg: 0 12px 48px rgba(0,0,0,.11);
  --accent-blue: #4b8dff;
  --accent-pink: #ff6b9d;
  --accent-mint: #2ed8a3;
  --accent-amber: #ffb94a;
  --text-dark: #18191f;
  --text-mid: #5a5f72;
  --text-soft: #9da3b4;
  --border: rgba(0,0,0,.06);
}

/* ── 全局根 ── */
.person-info {
  font-family: 'Noto Sans SC', -apple-system, sans-serif;
  width: 100%;
  height: 100%;
  overflow-y: auto;
  max-height: 100%;
  box-sizing: border-box;
  position: relative;
  background: url('/picture/user/default.webp') no-repeat center center fixed;
  background-size: cover;
  background-attachment: fixed;
}

/* 顶部彩色装饰条 */
.top-bar-deco {
  position: sticky;
  top: 0;
  z-index: 10;
  height: 4px;
  background: linear-gradient(90deg,
  #4b8dff 0%,
  #a78bfa 30%,
  #ff6b9d 55%,
  #2ed8a3 80%,
  #ffb94a 100%
  );
}

/* ── 大布局 ── */
.page-shell {
  display: flex;
  gap: 28px;
  max-width: 1440px;
  margin: 0 auto;
  padding: 28px 24px 48px;
  box-sizing: border-box;
  align-items: flex-start;
}

/* ── 左侧栏 ── */
.sidebar-left {
  width: 264px;
  flex-shrink: 0;
  position: sticky;
  top: 24px;
}

.sidebar-left-inner {
  background: var(--card-bg);
  border-radius: var(--card-radius);
  box-shadow: var(--shadow-md);
  overflow: hidden;
}

/* ── 主列 ── */
.main-col {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* ── 右侧栏 ── */
.sidebar-right {
  width: 256px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.sidebar-spacer { height: 0; }

.sticky-wrap {
  position: sticky;
  top: 24px;
}

/* ── 胶囊工具栏 ── */
.pill-toolbar {
  display: flex;
  justify-content: flex-end;
}

.pill-btn {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 8px 20px;
  background: var(--card-bg);
  border: 1.5px solid var(--border);
  border-radius: 100px;
  font-size: 13px;
  font-weight: 600;
  color: var(--accent-blue);
  cursor: pointer;
  box-shadow: var(--shadow-sm);
  transition: all .22s cubic-bezier(.34,1.56,.64,1);
  letter-spacing: .2px;
}

.pill-btn:hover {
  background: #4b8dff;
  color: #fff;
  border-color: #4b8dff;
  transform: translateY(-2px) scale(1.03);
  box-shadow: 0 8px 24px rgba(75,141,255,.3);
}

.pill-icon { font-size: 15px; }

/* ── 通用 Bento 卡片 ── */
.bento-card {
  background: var(--card-bg);
  border-radius: var(--card-radius);
  box-shadow: var(--shadow-md);
  padding: 28px 28px 24px;
  border: 1px solid rgba(255,255,255,.9);
  transition: box-shadow .25s;
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
}

.bento-card:hover {
  box-shadow: var(--shadow-lg);
}

/* ── 卡片头部 ── */
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 22px;
  padding-bottom: 18px;
  border-bottom: 1.5px solid var(--border);
}

.card-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-emoji { font-size: 22px; line-height: 1; }

.card-title {
  font-size: 17px;
  font-weight: 800;
  color: var(--text-dark);
  letter-spacing: -.1px;
}

.badge-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 26px;
  height: 22px;
  padding: 0 8px;
  background: linear-gradient(135deg, #4b8dff22, #a78bfa22);
  color: var(--accent-blue);
  border-radius: 100px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: .3px;
}

/* ── Chip 开关 ── */
.chip-toggle {
  cursor: pointer;
  user-select: none;
}

.chip-toggle input { display: none; }

.chip-inner {
  display: inline-block;
  padding: 6px 16px;
  border-radius: 100px;
  font-size: 12.5px;
  font-weight: 600;
  background: #f0f1f5;
  color: var(--text-soft);
  border: 1.5px solid transparent;
  transition: all .2s;
}

.chip-inner.chip-on {
  background: linear-gradient(135deg, #4b8dff, #7c6cff);
  color: #fff;
  box-shadow: 0 4px 14px rgba(75,141,255,.35);
}

/* ── 帖子网格 ── */
.post-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.post-card {
  background: #fff;
  border-radius: 18px;
  padding: 18px 18px 14px;
  cursor: pointer;
  border: 1.5px solid #f0f1f5;
  box-shadow: 0 2px 10px rgba(0,0,0,.04);
  transition: all .24s cubic-bezier(.34,1.2,.64,1);
  animation: cardIn .45s ease both;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.post-card:hover {
  transform: translateY(-4px) scale(1.01);
  box-shadow: 0 12px 36px rgba(75,141,255,.12);
  border-color: #c8d9ff;
}

@keyframes cardIn {
  from { opacity: 0; transform: translateY(18px) scale(.97); }
  to   { opacity: 1; transform: translateY(0)   scale(1); }
}

.post-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.post-type-tag {
  display: inline-block;
  padding: 3px 11px;
  border-radius: 100px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: .5px;
}

.tag-art {
  background: linear-gradient(135deg, #e8f0ff, #d4e3ff);
  color: #3a6fd8;
}

.tag-comm {
  background: linear-gradient(135deg, #fff0f5, #ffd6e7);
  color: #d63880;
}

.post-date {
  font-size: 11.5px;
  color: var(--text-soft);
  font-variant-numeric: tabular-nums;
}

.post-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-dark);
  line-height: 1.45;
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  transition: color .2s;
}

.post-card:hover .post-title { color: var(--accent-blue); }

.post-excerpt {
  font-size: 13px;
  line-height: 1.75;
  color: var(--text-mid);
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  flex: 1;
}

.post-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 2px;
}

.arrow-link {
  font-size: 12px;
  font-weight: 600;
  color: var(--accent-blue);
  opacity: 0;
  transform: translateX(-6px);
  transition: all .22s;
}

.post-card:hover .arrow-link {
  opacity: 1;
  transform: translateX(0);
}

.arrow { display: inline-block; transition: transform .2s; }
.post-card:hover .arrow { transform: translateX(3px); }

/* ── 加载更多 ── */
.load-more {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  margin-top: 20px;
  padding: 13px 0;
  background: linear-gradient(135deg, #f0f5ff, #f5f0ff);
  border: 1.5px dashed #b8caff;
  border-radius: 100px;
  font-size: 14px;
  font-weight: 700;
  color: var(--accent-blue);
  cursor: pointer;
  transition: all .22s cubic-bezier(.34,1.4,.64,1);
  letter-spacing: .2px;
}

.load-more:hover {
  background: linear-gradient(135deg, #4b8dff, #7c6cff);
  border-style: solid;
  border-color: transparent;
  color: #fff;
  transform: scale(1.02);
  box-shadow: 0 8px 28px rgba(75,141,255,.3);
}

.bounce-arrow {
  font-size: 16px;
  animation: arrowBounce 1.4s ease infinite;
}

@keyframes arrowBounce {
  0%, 100% { transform: translateY(0); }
  50%       { transform: translateY(5px); }
}

/* ── Loading 点 ── */
.dots-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  padding: 24px 0;
}

.dots-loading span {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: linear-gradient(135deg, #4b8dff, #a78bfa);
  animation: dotPop 1.1s ease infinite;
}

.dots-loading span:nth-child(2) { animation-delay: .18s; }
.dots-loading span:nth-child(3) { animation-delay: .36s; }

@keyframes dotPop {
  0%, 100% { transform: scale(.65); opacity: .4; }
  50%       { transform: scale(1.1); opacity: 1; }
}

/* ── 末尾标记 ── */
.end-tag {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 22px;
}

.end-pill {
  display: inline-block;
  padding: 5px 18px;
  background: #f4f5f9;
  border-radius: 100px;
  font-size: 12px;
  color: var(--text-soft);
  letter-spacing: .8px;
  font-weight: 500;
}

/* ── 空状态 ── */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 20px;
  gap: 10px;
  color: var(--text-soft);
}

.empty-blob {
  font-size: 36px;
  opacity: .4;
  animation: floatBlob 3s ease-in-out infinite;
}

@keyframes floatBlob {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50%       { transform: translateY(-6px) rotate(10deg); }
}

.empty-state p { font-size: 14px; margin: 0; }

/* ── 文件网格 ── */
.file-grid {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.file-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 14px;
  border-radius: var(--mini-radius);
  background: #fafafa;
  border: 1px solid transparent;
  transition: all .2s;
  animation: cardIn .4s ease both;
}

.file-row:hover {
  background: #fff;
  border-color: #dde8ff;
  box-shadow: 0 4px 18px rgba(75,141,255,.08);
}

.file-icon-box {
  width: 38px;
  height: 38px;
  border-radius: 12px;
  background: linear-gradient(135deg, #eef3ff, #f3efff);
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
  gap: 2px;
}

.file-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-dark);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-sub {
  font-size: 12px;
  color: var(--text-soft);
}

.dl-btn {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  border: 1.5px solid #d0dcff;
  background: linear-gradient(135deg, #f0f5ff, #f5f0ff);
  color: var(--accent-blue);
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all .22s cubic-bezier(.34,1.56,.64,1);
  flex-shrink: 0;
}

.dl-btn:hover {
  background: linear-gradient(135deg, #4b8dff, #7c6cff);
  border-color: transparent;
  color: #fff;
  transform: translateY(2px) scale(1.08);
  box-shadow: 0 6px 18px rgba(75,141,255,.35);
}

/* ── 右侧 Mini 卡片 ── */
.bento-mini {
  background: var(--card-bg);
  border-radius: 20px;
  box-shadow: var(--shadow-sm);
  padding: 18px 18px 14px;
  border: 1px solid rgba(255,255,255,.9);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

.mini-head {
  display: flex;
  align-items: center;
  gap: 7px;
  margin-bottom: 14px;
  padding-bottom: 12px;
  border-bottom: 1.5px solid var(--border);
}

.mini-title {
  font-size: 15px;
  font-weight: 800;
  color: var(--text-dark);
}

/* ── 最近文章列表 ── */
.hot-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.hot-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 8px;
  border-radius: 10px;
  cursor: pointer;
  transition: background .18s;
}

.hot-item:hover {
  background: linear-gradient(135deg, #f0f5ff, #f8f0ff);
}

.hot-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: linear-gradient(135deg, #4b8dff, #a78bfa);
  flex-shrink: 0;
}

.hot-text {
  font-size: 13px;
  color: var(--text-mid);
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color .18s;
}

.hot-item:hover .hot-text { color: var(--accent-blue); }

/* ── 响应式 ── */
@media (max-width: 1100px) {
  .page-shell {
    flex-direction: column;
    gap: 20px;
  }

  .sidebar-left,
  .sidebar-right {
    width: 100%;
    position: static;
  }

  .sidebar-left-inner {
    max-width: 100%;
  }

  .sticky-wrap { position: static; }

  .sidebar-spacer { display: none; }
}

@media (max-width: 640px) {
  .page-shell { padding: 16px 14px 36px; }

  .bento-card { padding: 18px 16px 16px; }

  .post-grid { grid-template-columns: 1fr; }
}
</style>