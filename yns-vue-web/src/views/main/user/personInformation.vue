<template>
  <div class="person-info j-desk" :style="currentBgStyle">
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
          :auto-play="true"
          @update-bg-style="handleBgStyleUpdate"
      />

      <div class="inner-container">
        <div class="content-columns">

          <div class="person-left-wrapper">
            <UserInfo
                :user="user"
                :target-user-code="targetUserCode"
                :creation-stats="overview.stats"
                @blog-click="blogMainClick"
            />
          </div>

          <div class="main-content">
            <!-- 内容：全部 / 文章 / 随记 / 文件四个标签页，列表在后台按类型筛选 -->
            <div class="section-card publish-card">
              <span class="j-tape j-tape--top"></span>
              <div class="section-header">
                <div class="section-title-group">
                  <span class="section-icon">📚</span>
                  <h3 class="section-title">{{ isSelf ? '我的内容' : 'TA 的内容' }}</h3>
                </div>
                <nav class="content-tabs" role="tablist" aria-label="内容分类">
                  <button
                      v-for="tab in contentTabs"
                      :key="tab.key"
                      type="button"
                      role="tab"
                      class="content-tab"
                      :class="{ 'is-active': activeTab === tab.key }"
                      :aria-selected="activeTab === tab.key"
                      @click="switchTab(tab.key)"
                  >
                    {{ tab.label }}<span v-if="tab.count !== null" class="tab-count">{{ tab.count }}</span>
                  </button>
                </nav>
              </div>

              <!-- 文件 -->
              <template v-if="activeTab === 'files'">
                <div v-if="files.length === 0" class="empty-state">
                  <div class="empty-icon">◇</div>
                  <p class="empty-text">暂无上传文件</p>
                </div>
                <div v-else class="file-list">
                  <div
                      v-for="(file, index) in files"
                      :key="file.GUID || index"
                      class="file-item"
                      :style="{ animationDelay: index * 0.05 + 's' }"
                  >
                    <div class="file-icon-wrap">
                      <span class="file-icon">{{ getFileIcon(file.ORIGINALFILENAME) }}</span>
                    </div>
                    <div class="file-info">
                      <span class="file-name" :title="file.ORIGINALFILENAME">{{ file.ORIGINALFILENAME }}</span>
                      <span class="file-meta">{{ formatDate(file.CREATE_TIME) }} · 下载 {{ file.DOWNNUM || 0 }} 次</span>
                    </div>
                    <button class="download-btn" :title="`下载 ${file.ORIGINALFILENAME}`" @click="downloadFile(file)">
                      <span>↓</span>
                    </button>
                  </div>
                </div>
              </template>

              <!-- 文章、随记 -->
              <template v-else>
                <div v-if="blogs.length > 0" class="blog-timeline">
                  <div
                      v-for="(blog, index) in blogs"
                      :key="blog.TYPE + blog.GUID"
                      class="timeline-item"
                      :style="{ animationDelay: (index % pageSize) * 0.06 + 's' }"
                  >
                    <div class="timeline-connector">
                      <div class="timeline-dot" :class="blog.TYPE === 'blog' ? 'dot-article' : 'dot-community'"></div>
                      <div class="timeline-line" v-if="index < blogs.length - 1"></div>
                    </div>
                    <div class="timeline-content">
                      <div class="timeline-meta">
                        <span class="meta-tag" :class="blog.TYPE === 'blog' ? 'tag-article' : 'tag-community'">
                          {{ blog.TYPE === 'blog' ? '文章' : '随记' }}
                        </span>
                        <span class="meta-time">{{ formatDate(blog.CREATE_TIME) }}</span>
                      </div>
                      <div class="blog-card" @click="blogMainClick(blog)">
                        <a
                            v-if="blog.TYPE === 'blog'"
                            :href="'/oneBlog/' + blog.GUID"
                            @click.prevent
                            class="seo-link"
                        >
                          <h4 class="blog-title">{{ blog.BLOG_TITLE }}</h4>
                        </a>
                        <a v-else :href="'/note/' + encodeURIComponent(blog.GUID)" @click.prevent class="seo-link">
                          <h4 class="blog-title">{{ blog.BLOG_TITLE || '一张便签' }}</h4>
                        </a>
                        <p class="blog-summary">{{ listExcerpt(blog) }}</p>
                        <div class="blog-card-footer">
                          <span class="read-more">阅读全文 →</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div v-if="!blogs.length && !loading" class="empty-state">
                  <div class="empty-icon">✦</div>
                  <p class="empty-text">{{ emptyText }}</p>
                </div>

                <button
                    v-if="!noMore && !loading && blogs.length"
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

                <div v-if="noMore && blogs.length" class="end-line">
                  <span class="end-dash"></span>
                  <span class="end-text">已到末尾</span>
                  <span class="end-dash"></span>
                </div>
              </template>
            </div>
          </div>

          <div class="right-sidebar">
            <div class="align-spacer"></div>

            <!-- TA 的系列：按专题阅读的入口，没有系列时不显示 -->
            <el-card v-if="overview.series.length" class="sidebar-card series-card" shadow="never">
              <template #header>
                <div class="sidebar-card-title">{{ isSelf ? '我的系列' : 'TA 的系列' }}</div>
              </template>
              <ul class="side-list">
                <li v-for="item in overview.series" :key="item.ID" class="side-item">
                  <router-link :to="seriesHref(item.ID)" class="side-link" :title="item.DESCRIPTION || item.TITLE">
                    <span class="side-title">{{ item.TITLE }}</span>
                    <span class="side-meta">{{ item.ARTICLE_COUNT }} 篇</span>
                  </router-link>
                </li>
              </ul>
            </el-card>

            <!-- 热门文章：按阅读量排，和中间按时间排的列表互补 -->
            <el-card v-if="overview.hotBlogs.length" class="sidebar-card hot-card" shadow="never">
              <template #header>
                <div class="sidebar-card-title">热门文章</div>
              </template>
              <ol class="side-list">
                <li v-for="(article, index) in overview.hotBlogs" :key="article.GUID" class="side-item">
                  <a :href="'/oneBlog/' + article.GUID" class="side-link" :title="article.BLOG_TITLE" @click.prevent="blogMainClick({...article, TYPE: 'blog'})">
                    <span class="side-rank" :class="{ 'is-top': index < 3 }">{{ index + 1 }}</span>
                    <span class="side-title">{{ article.BLOG_TITLE }}</span>
                    <span class="side-meta">{{ formatCount(article.VIEW_PAGE) }} 阅读</span>
                  </a>
                </li>
              </ol>
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
import {ref, onMounted, computed, defineAsyncComponent, watch, nextTick} from 'vue'
import {ElMessage} from 'element-plus'
import {useRoute} from 'vue-router'
import {useRouter} from "vue-router";
import {useSeo, buildProfileSeo} from '@/utils/seo.js';
import {
  pubFormatDate,
  sendAxiosRequest,
  listExcerpt,
  downloadUploadedFile
} from '@/utils/common.js'
import {seriesHref} from '@/utils/blogSeries.js'

import BackgroundAndMusic from "@/components/detail/personInformation/BackgroundAndMusic.vue";
import WelcomeOverlay from "@/components/detail/personInformation/WelcomeOverlay.vue";
import BlogSidebar from "@/components/detail/myblog/BlogSidebar.vue";
import UserInfo from "@/components/main/UserInfo.vue";

//正文与评论组件
const ContentAndComment = defineAsyncComponent(() => import('@/views/detail/blog/ContentAndComment.vue'))

import {useUserStore} from '@/stores/main/user.js'

const userStore = useUserStore()
userStore.initFromLocal()

const route = useRoute();
const router = useRouter();
const user = ref({})
const selectedCategory = ref('');
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
const showDialog = ref(false)
const selectedBlogId = ref('')

// ================= 内容标签页 =================
// all 全部、blog 文章、community 随记（在后台筛选）、files 文件
const activeTab = ref('all')
// 作者概览：创作数据（公开文章数、随记数、总阅读）、热门文章、系列
const emptyOverview = () => ({stats: {articles: null, notes: null, views: null}, hotBlogs: [], series: []})
const overview = ref(emptyOverview())
const filesLoaded = ref(false)

const countOrNull = (value) => (value === null || value === undefined ? null : Number(value) || 0)
const contentTabs = computed(() => {
  const {articles, notes} = overview.value.stats
  const all = articles === null || notes === null ? null : Number(articles) + Number(notes)
  return [
    {key: 'all', label: '全部', count: all},
    {key: 'blog', label: '文章', count: countOrNull(articles)},
    {key: 'community', label: '随记', count: countOrNull(notes)},
    {key: 'files', label: '文件', count: filesLoaded.value ? files.value.length : null}
  ]
})

const emptyText = computed(() => ({blog: '还没有公开文章', community: '还没有随记'}[activeTab.value] || '暂无内容'))

// 切换标签页：文章、随记重新从第一页加载；文件已经一次性加载过
function switchTab(key) {
  if (key === activeTab.value) return
  activeTab.value = key
  if (key === 'files') return
  requestId++
  page.value = 1
  blogs.value = []
  noMore.value = false
  loading.value = false
  fetchArticles(null)
}

// 每次重新加载列表都递增，丢弃之前还没返回的请求，避免快速切换标签时旧结果混进新列表
let requestId = 0

const fetchOverview = async (userCode) => {
  const res = await sendAxiosRequest('/blog-api/userInformation/getAuthorOverview', {userCode}).catch(() => null)
  if (res && !res.isError && res.result) {
    overview.value = {
      stats: res.result.stats || emptyOverview().stats,
      hotBlogs: res.result.hotBlogs || [],
      series: res.result.series || []
    }
  }
}

const formatCount = (value) => {
  const number = Number(value) || 0
  if (number >= 10000) return `${(number / 10000).toFixed(number >= 100000 ? 0 : 1).replace(/\.0$/, '')}万`
  return String(number)
}
//控制是否显示欢迎页
const showWelcome = ref(false);

// SEO：主页统一以 /user/:u 为收录地址；带文章 ID 打开时，文章本身以 /oneBlog/:id 为准
const seoTitle = ref('个人主页 - YnsStudy');
const seoDescription = ref('');
const seoImage = ref(undefined);
const seoJsonLd = ref(undefined);
// 还没有公开文章的作者主页不收录（与 Go 输出的首屏一致）
const seoNoindex = ref(false);
const userNotFound = ref(false);

useSeo(() => ({
  title: seoTitle.value,
  description: seoDescription.value,
  image: seoImage.value,
  jsonLd: route.params.blogId ? undefined : seoJsonLd.value,
  type: 'profile',
  path: route.params.blogId
      ? `/oneBlog/${encodeURIComponent(route.params.blogId)}`
      : `/user/${encodeURIComponent(route.params.u || '')}`,
  noindex: !route.params.u || userNotFound.value || (!route.params.blogId && seoNoindex.value)
}));
// ----------------------------------------------------
// 核心优化区
// ----------------------------------------------------
const initPageData = async () => {
  const userNum = route.params.u;
  if (!userNum) return;
  isPageReady.value = false; // 初始化时上锁，隐藏子组件
  loading.value = true;
  try {

    // 1. 先用数字拿到 CODE
    const res = await sendAxiosRequest("/pub-api/login/getUserInfoByNum", { userNum });
    if (res && res.result && res.result.code) {
      const parsedCode = res.result.code;
      targetUserCode.value = parsedCode;
      user.value = res.result;
      // 设置页面标题
      // 标题、描述和结构化数据与 Go 输出的首屏一致（站长主页会带上别名和外部主页）
      const profileSeo = buildProfileSeo(user.value, userNum);
      seoTitle.value = profileSeo.title;
      seoDescription.value = profileSeo.description;
      seoImage.value = profileSeo.image;
      seoJsonLd.value = profileSeo.jsonLd;
      seoNoindex.value = profileSeo.noindex;

      // 【核心开关】用户信息和 CODE 都有了，允许子组件渲染！
      isPageReady.value = true;
      // 2. 然后再去查列表数据（这个时候组件已经挂载完毕了）
      await Promise.all([
        fetchArticles(parsedCode),
        fetchFilesList(parsedCode),
        fetchOverview(parsedCode),
        setPersonInfo(parsedCode)
      ]);
      //用于Prerender的seo检索   防止只爬虫到外壳没有实际数据
      nextTick(() => {
        window.prerenderReady = true;
      });
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

const openUserBlog = ()=>{
  const blogId = route.params.blogId;
  if(blogId){
    selectedBlogId.value = blogId;
    showDialog.value = true
  }
}

const fetchArticles = async (userCode) => {
  // 如果没传，就用当前的
  const codeToUse = userCode || targetUserCode.value;
  if (noMore.value || !codeToUse) return;
  const current = requestId
  loading.value = true

  try {
    const res = await sendAxiosRequest('/blog-api/userInformation/getBlogAndCommunityByUserCode', {
      userCode: codeToUse,
      page: page.value,
      pageSize,
      keyword: "",
      // 全部时不传类型；文章、随记在后台筛选
      type: activeTab.value === 'all' ? '' : activeTab.value
    })
    if (current !== requestId) return
    if (!res || res.isError) throw new Error(res?.errMsg || '获取内容失败')
    const newData = res.result.data || []
    if (newData.length < pageSize || blogs.value.length + newData.length >= Number(res.result.total)) noMore.value = true
    blogs.value.push(...newData)
    page.value++
  } catch (e) {
    console.error('获取内容失败', e)
  } finally {
    if (current === requestId) loading.value = false
  }
}

const fetchFilesList = async (userCode) => {
  const result = await sendAxiosRequest('/blog-api/userInformation/getResourceByUserCode', {userCode});
  if (result && !result.isError) files.value = result.result || [];
  filesLoaded.value = true;
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


function blogMainClick(blog) {
  // 便签墙的随记打开自己的页面
  if (blog.TYPE === "community") {
    router.push('/note/' + encodeURIComponent(blog.GUID));
    return false
  }
  const userNum = route.params.u;
  //改变路由地址   watch会监听route.params.blogId  自动打开/关闭文章
  router.push({name: 'user', params: {u: userNum,blogId:blog.GUID}});
}

const onDialogClose = () => {
  //改变路由地址   watch会监听route.params.blogId  自动打开/关闭文章
    router.push({
      name: 'user',
      params: { u: route.params.u }
    });
}

const formatDate = (dateStr) => pubFormatDate(dateStr)

// 直接交给浏览器下载并记一次下载，见 common.js 的 downloadUploadedFile
const downloadFile = (file) => downloadUploadedFile(file)

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


const enterHomepage = async () => {
  showWelcome.value = false;
  if (bgMusicComponentRef.value) {
    bgMusicComponentRef.value.playMusicForce();
  }
};

onMounted(() => {
  initPageData();
});

// 路由是 /user/:u，站内从一个作者主页跳到另一个作者主页时组件会复用，需要按新的 u 重新加载
watch(() => route.params.u, () => {
  userNotFound.value = false;
  requestId++;
  page.value = 1;
  blogs.value = [];
  files.value = [];
  filesLoaded.value = false;
  noMore.value = false;
  activeTab.value = 'all';
  overview.value = emptyOverview();
  initPageData();
});

watch(() => route.params.blogId, (newBlogId) => {
      if (newBlogId) {
        selectedBlogId.value = newBlogId;
        showDialog.value = true;
      } else {
        selectedBlogId.value = '';
        showDialog.value = false;
      }
    }, { immediate: true } // immediate 确保刷新页面时也能触发
);
</script>
<style scoped>
/* ===== 页面根 ===== */
.person-info {
  position: relative;
  width: 100%;
  height: 100%;
  max-height: 100%;
  overflow-y: auto;
  padding: 28px 0 56px;
  box-sizing: border-box;
  background-size: cover;
  background-attachment: fixed;
  transition: background-image 0.001s;
  color: var(--j-ink);
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
  gap: 36px;
  align-items: flex-start;
}

.person-left-wrapper {
  position: sticky;
  top: 20px;
  z-index: 10;
  flex-shrink: 0;
  width: 280px;
  margin-top: 36px;
}

.main-content {
  flex: 1;
  min-width: 0;
}

.right-sidebar {
  flex-shrink: 0;
  width: 280px;
}

/* ===== 区块：一张活页纸 ===== */
.section-card {
  position: relative;
  margin-bottom: 30px;
  padding: 28px 30px 24px 56px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
}

/* 页边红线 */
.section-card::before {
  content: "";
  position: absolute;
  top: 0;
  bottom: 0;
  left: 36px;
  width: 1px;
  background: var(--j-margin-red);
  opacity: 0.55;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
  padding-bottom: 12px;
  border-bottom: 1px dashed var(--j-rule-strong);
}

.section-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-icon {
  font-size: 18px;
}

.section-title {
  margin: 0;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 22px;
  color: var(--j-ink);
}

.section-count {
  padding: 0 7px;
  border-radius: 10px;
  background: #efe8da;
  font-size: 12px;
  line-height: 20px;
  color: var(--j-muted);
}

/* ===== 时间线 ===== */
.blog-timeline {
  display: flex;
  flex-direction: column;
}

.timeline-item {
  display: flex;
  gap: 16px;
  opacity: 0;
  animation: fadeSlideUp 0.4s ease forwards;
}

@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.timeline-connector {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
  width: 14px;
  padding-top: 26px;
}

/* 手绘的小圆圈 */
.timeline-dot {
  flex-shrink: 0;
  width: 10px;
  height: 10px;
  border: 2px solid currentColor;
  border-radius: 50% 45% 55% 48%;
  background: var(--j-paper);
}

.dot-article {
  color: var(--j-pen);
}

.dot-community {
  color: #4f8a5b;
}

.timeline-line {
  flex: 1;
  width: 0;
  margin-top: 4px;
  border-left: 1px dashed var(--j-rule-strong);
}

.timeline-content {
  flex: 1;
  min-width: 0;
  padding-bottom: 20px;
}

.timeline-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.meta-tag {
  padding: 0 8px;
  border: 1px solid currentColor;
  border-radius: 3px;
  font-family: var(--j-hand);
  font-size: 13px;
  line-height: 20px;
}

.tag-article {
  color: var(--j-pen);
}

.tag-community {
  color: #4f8a5b;
}

.meta-time {
  font-family: var(--j-hand);
  font-size: 14px;
  color: var(--j-muted);
}

.blog-card {
  padding: 4px 0 14px;
  border-bottom: 1px dashed var(--j-rule);
  cursor: pointer;
}

.blog-title {
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

.blog-card:hover .blog-title {
  background-size: 100% 100%;
}

.blog-summary {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin: 8px 0 0;
  font-size: 14px;
  line-height: 1.75;
  color: var(--j-ink-soft);
  word-break: break-word;
}

.blog-summary :deep(*) {
  margin: 0;
  font-size: inherit !important;
  font-weight: normal !important;
  color: inherit !important;
}

.blog-card-footer {
  margin-top: 8px;
}

.read-more {
  font-family: var(--j-hand);
  font-size: 14px;
  color: var(--j-pen);
  transition: transform 0.2s ease;
  display: inline-block;
}

.blog-card:hover .read-more {
  transform: translateX(4px);
}

.load-more-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin: 18px auto 0;
  padding: 8px 26px;
  border: 1px dashed var(--j-rule-strong);
  background: transparent;
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-ink-soft);
  cursor: pointer;
  transition: border-color 0.2s ease, color 0.2s ease;
}

.load-more-btn:hover {
  border-color: var(--j-pen);
  color: var(--j-pen);
}

.load-more-arrow {
  animation: bounceArrow 1.4s ease-in-out infinite;
}

@keyframes bounceArrow {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(3px); }
}

.loading-row {
  display: flex;
  justify-content: center;
  gap: 6px;
  padding: 16px 0;
}

.loading-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--j-rule-strong);
  animation: loadPulse 1.2s ease-in-out infinite;
}

.loading-dot:nth-child(2) { animation-delay: 0.15s; }
.loading-dot:nth-child(3) { animation-delay: 0.3s; }

@keyframes loadPulse {
  0%, 100% { opacity: 0.3; transform: scale(0.8); }
  50% { opacity: 1; transform: scale(1); }
}

.end-line {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 18px;
}

.end-dash {
  flex: 1;
  border-top: 1px dashed var(--j-rule-strong);
}

.end-text {
  font-family: var(--j-hand);
  font-size: 14px;
  color: var(--j-muted);
}

.empty-state {
  padding: 36px 0;
  text-align: center;
}

.empty-icon {
  font-size: 22px;
  color: var(--j-rule-strong);
}

.empty-text {
  margin: 8px 0 0;
  font-family: var(--j-hand);
  font-size: 16px;
  color: var(--j-muted);
}

/* ===== 文件：一张张小卡片 ===== */
.file-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 10px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  background: #fffaf0;
  box-shadow: 0 1px 1px rgba(110, 85, 40, 0.18);
  opacity: 0;
  animation: fadeSlideUp 0.4s ease forwards;
  transition: transform 0.2s ease;
}

.file-item:hover {
  transform: translateY(-2px);
}

.file-icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 34px;
  height: 34px;
  border: 1px solid var(--j-rule);
  background: var(--j-paper);
}

.file-icon {
  font-size: 18px;
}

.file-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.file-name {
  overflow: hidden;
  font-size: 14px;
  font-weight: 600;
  color: var(--j-ink);
  white-space: nowrap;
  text-overflow: ellipsis;
}

.file-meta {
  font-size: 12px;
  color: #7a6440;
}

.download-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 30px;
  height: 30px;
  border: 1px solid #d8c49d;
  border-radius: 50%;
  background: transparent;
  font-size: 15px;
  color: #6e5528;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.download-btn:hover {
  background: var(--j-note);
}

/* ===== 右侧：便签卡片（系列、热门文章） ===== */
.sidebar-card {
  position: relative;
  overflow: visible;
  margin-bottom: 30px;
  border: none;
  border-radius: 2px;
  background: var(--j-note);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.1), 0 12px 20px -14px rgba(60, 50, 30, 0.45) !important;
  transform: rotate(1deg);
}

.sidebar-card :deep(.el-card__header) {
  padding: 16px 18px 8px;
  border-bottom: 1px dashed rgba(160, 135, 60, 0.35);
}

.sidebar-card :deep(.el-card__body) {
  padding: 8px 18px 14px;
}

.sidebar-card-title {
  font-family: var(--j-hand);
  font-size: 19px;
  color: var(--j-ink);
}

/* 系列卡片换成浅蓝便签，和热门文章的黄色便签区分开，并朝另一边歪 */
.series-card {
  background: #dfe9f2;
  transform: rotate(-1deg);
}

.series-card :deep(.el-card__header) {
  border-bottom-color: rgba(47, 93, 138, 0.25);
}

.side-list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.side-item {
  border-bottom: 1px solid rgba(160, 135, 60, 0.18);
}

.series-card .side-item {
  border-bottom-color: rgba(47, 93, 138, 0.15);
}

.side-item:last-child {
  border-bottom: none;
}

.side-link {
  display: flex;
  align-items: baseline;
  gap: 8px;
  padding: 8px 0;
  font-size: 14px;
  color: var(--j-ink);
  text-decoration: none;
  cursor: pointer;
}

.side-link:hover .side-title {
  text-decoration: underline;
  text-decoration-color: var(--j-pen);
  text-underline-offset: 3px;
}

.side-rank {
  flex-shrink: 0;
  width: 1.2em;
  font-family: var(--j-num);
  color: var(--j-muted);
  text-align: center;
}

.side-rank.is-top {
  color: var(--j-stamp);
  font-weight: 600;
}

.side-title {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.side-meta {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--j-muted);
}

/* ===== 内容标签页：本子侧边的索引贴 ===== */
.content-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.content-tab {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 12px;
  border: 1px solid var(--j-rule-strong);
  border-bottom-width: 2px;
  background: var(--j-paper-warm);
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-ink-soft);
  cursor: pointer;
  transition: background-color 0.15s, color 0.15s, transform 0.15s;
}

.content-tab:hover {
  background: var(--j-note);
  color: var(--j-ink);
}

.content-tab.is-active {
  background: var(--j-note);
  border-color: var(--j-stamp);
  color: var(--j-stamp);
  transform: translateY(-2px);
}

.tab-count {
  font-family: var(--j-num);
  font-size: 12px;
  color: var(--j-muted);
}

.content-tab.is-active .tab-count {
  color: var(--j-stamp);
}

.sticky-category {
  position: sticky;
  top: 20px;
}

.align-spacer {
  height: 36px;
}

/* ===== 加载中 ===== */
.page-loading-mask {
  position: fixed;
  top: 0;
  left: 0;
  z-index: 9999;
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100vh;
  background-color: var(--j-desk);
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
  border-radius: 50%;
  background-color: var(--j-pen);
  animation: bounce 1.4s infinite ease-in-out both;
}

.bouncing-dots .dot:nth-child(2) { background-color: #e8c95b; }
.bouncing-dots .dot:nth-child(3) { background-color: var(--j-stamp); }
.bouncing-dots .dot:nth-child(1) { animation-delay: -0.32s; }
.bouncing-dots .dot:nth-child(2) { animation-delay: -0.16s; }

.loading-text {
  margin: 0;
  font-family: var(--j-hand);
  font-size: 18px;
  letter-spacing: 2px;
  color: var(--j-ink-soft);
  animation: pulse-text 2s infinite ease-in-out;
}

.seo-link {
  display: block;
  color: inherit;
  text-decoration: none;
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
    align-items: stretch;
  }

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

  .sidebar-card {
    transform: none;
  }
}

@media (max-width: 768px) {
  .inner-container {
    padding: 0 12px;
  }

  .section-card {
    padding: 22px 14px 18px 30px;
  }

  .section-card::before {
    left: 16px;
  }

  .section-header {
    flex-wrap: wrap;
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

  .file-item {
    padding: 12px 10px;
    gap: 10px;
  }

  .file-name {
    font-size: 13px;
  }
}
</style>
