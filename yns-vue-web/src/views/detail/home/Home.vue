<template>
  <div class="home">
    <Announcement v-for="al in topAlert" :key="al.GUID" :TEXT="al.TEXT" :URL="al.URL" :URLNAME="al.URLNAME"/>

    <header class="masthead">
      <div class="masthead-inner">
        <div class="masthead-intro">
          <p class="masthead-date">{{ todayText }}</p>
          <h1 class="masthead-title">YnsStudy</h1>
          <p class="masthead-desc">记录编程学习、技术实践与生活思考。</p>
          <nav class="masthead-links" aria-label="快捷入口">
            <a href="#" @click.prevent="goToPublishBlog">写文章</a>
            <a href="#" @click.prevent="goToUpload">上传资源</a>
            <a href="#" @click.prevent="goMe">我的主页</a>
            <a href="#" @click.prevent="goToAdmin">关于站长</a>
          </nav>
        </div>

        <dl v-if="siteStats" class="masthead-stats">
          <div class="stat">
            <dt>文章</dt>
            <dd>{{ formatCount(siteStats.ARTICLENUM) }}</dd>
          </div>
          <div class="stat">
            <dt>阅读</dt>
            <dd>{{ formatCount(siteStats.VIEW_PAGE) }}</dd>
          </div>
          <div class="stat">
            <dt>社区动态</dt>
            <dd>{{ formatCount(siteStats.COMMUNITYNUM) }}</dd>
          </div>
          <div class="stat">
            <dt>用户</dt>
            <dd>{{ formatCount(siteStats.USERNUM) }}</dd>
          </div>
        </dl>
      </div>
    </header>

    <div class="layout">
      <main class="feed">
        <div class="feed-head">
          <h2 class="section-label">
            {{ activeKeyword ? `“${activeKeyword}” 的搜索结果` : '最新文章' }}
          </h2>
          <label class="search">
            <svg class="search-icon" viewBox="0 0 24 24" aria-hidden="true">
              <circle cx="11" cy="11" r="6.5"/>
              <path d="M16 16l4.5 4.5"/>
            </svg>
            <input
                v-model="searchKeyword"
                type="search"
                placeholder="搜索标题或正文"
                aria-label="搜索文章"
                @input="debouncedSearch"
                @keydown.enter="searchNow"
            />
            <button v-if="searchKeyword" type="button" class="search-clear" aria-label="清除搜索" @click="clearSearch">
              ×
            </button>
          </label>
        </div>

        <p v-if="activeKeyword && !loading" class="feed-hint">共找到 {{ total }} 篇</p>

        <ol class="article-list">
          <li
              v-for="article in articles"
              :key="article.GUID"
              class="article"
              :class="{ 'has-cover': article.ILLUSTRATION }"
              @click="openBlog(article)"
          >
            <time class="article-date" :datetime="article.DATE.iso">
              <span class="article-day">{{ article.DATE.monthDay }}</span>
              <span class="article-year">{{ article.DATE.year }}</span>
            </time>

            <div class="article-main">
              <h3 class="article-title">
                <a :href="blogHref(article.GUID)" target="_blank" rel="noopener" @click.stop>{{ article.BLOG_TITLE }}</a>
              </h3>
              <p v-if="article.EXCERPT" class="article-excerpt">{{ article.EXCERPT }}</p>
              <div class="article-meta">
                <el-avatar :src="article.AVATAR" :size="20" class="meta-avatar">
                  {{ article.USERNAME?.charAt(0) }}
                </el-avatar>
                <span class="meta-author">{{ article.USERNAME }}</span>
                <span class="meta-sep"></span>
                <span>{{ formatCount(article.VIEW_PAGE) }} 阅读</span>
              </div>
            </div>

            <img
                v-if="article.ILLUSTRATION"
                :src="article.ILLUSTRATION"
                :alt="`${article.BLOG_TITLE} 的配图`"
                class="article-cover"
                loading="lazy"
                decoding="async"
                @error="article.ILLUSTRATION = ''"
            />
          </li>
        </ol>

        <div v-if="loading && !articles.length" class="article-skeleton" aria-hidden="true">
          <div v-for="n in 4" :key="n" class="skeleton-row">
            <span class="skeleton-date"></span>
            <div class="skeleton-lines">
              <span class="skeleton-line w-60"></span>
              <span class="skeleton-line w-90"></span>
              <span class="skeleton-line w-30"></span>
            </div>
          </div>
        </div>

        <div v-if="!loading && !articles.length" class="feed-empty">
          <p>{{ activeKeyword ? '没有找到相关文章，换个关键词试试。' : '这里还没有文章。' }}</p>
          <button v-if="activeKeyword" type="button" class="text-button" @click="clearSearch">清除搜索</button>
        </div>

        <div ref="sentinelRef" class="feed-foot">
          <span v-if="loading && articles.length" class="feed-status">加载中…</span>
          <button
              v-else-if="!noMore && articles.length"
              type="button"
              class="text-button"
              @click="fetchArticles"
          >加载更多</button>
          <span v-else-if="noMore && articles.length" class="feed-status">已经到底了</span>
        </div>
      </main>

      <aside class="sidebar">
        <section v-if="hotBlogs.length" class="panel">
          <h2 class="section-label">热门文章</h2>
          <ol class="rank-list">
            <li v-for="(blog, index) in hotBlogs" :key="blog.GUID" class="rank-item" @click="openBlog(blog)">
              <span class="rank-no" :class="{ top: index < 3 }">{{ String(index + 1).padStart(2, '0') }}</span>
              <div class="rank-body">
                <a class="rank-title" :href="blogHref(blog.GUID)" :title="blog.BLOG_TITLE" target="_blank" rel="noopener" @click.stop>
                  {{ blog.BLOG_TITLE }}
                </a>
                <span class="rank-meta">
                  {{ blog.USERNAME }} · {{ formatCount(blog.VIEW_PAGE) }} 阅读 · {{ formatCount(blog.COMMENT_COUNT) }} 评论
                </span>
              </div>
            </li>
          </ol>
        </section>

        <section v-if="authors.length" class="panel">
          <h2 class="section-label">活跃作者</h2>
          <ul class="author-list">
            <li v-for="author in authors" :key="author.USERCODE" class="author" @click="openUser(author.USERCODE)">
              <el-avatar :src="author.AVATAR" :size="36" class="author-avatar">
                {{ author.USERNAME?.charAt(0) }}
              </el-avatar>
              <div class="author-body">
                <span class="author-name">{{ author.USERNAME || '未命名' }}</span>
                <span class="author-remark">{{ author.REMARK || '这个人还没有写签名' }}</span>
              </div>
              <span class="author-count">{{ formatCount(author.ARTICLE_COUNT) }}<small>篇</small></span>
            </li>
          </ul>
        </section>

        <section v-if="hotFiles.length" class="panel">
          <h2 class="section-label">资源下载</h2>
          <ul class="file-list">
            <li v-for="file in hotFiles" :key="file.GUID" class="file" @click="openFile(file)">
              <span class="file-ext">{{ fileExt(file.ORIGINALFILENAME) }}</span>
              <span class="file-name" :title="file.ORIGINALFILENAME">{{ file.ORIGINALFILENAME }}</span>
              <span class="file-count">{{ formatCount(file.DOWNNUM) }} 次</span>
            </li>
          </ul>
          <a href="#" class="panel-more" @click.prevent="goToUpload">全部资源</a>
        </section>
      </aside>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import debounce from "lodash/debounce.js";
import { useHomeStore } from "@/stores/detail/home.js";
import { useUserStore } from "@/stores/main/user.js";
import { extractFirstImage, extractPlainTextFromHTML, sendAxiosRequest } from "@/utils/common.js";
import { adminUserCode } from "@/config/vue-config.js";
import { getAnnouncementByRouterName, pubOpenOneBlog, pubOpenUser } from "@/utils/blogUtil.js";
import Announcement from "@/components/detail/Announcement.vue";

const router = useRouter();
const userStore = useUserStore();
const homeStore = useHomeStore();
homeStore.initHomeData();

const hotBlogs = computed(() => homeStore.homeData.hotBlogData || []);
const authors = computed(() => homeStore.homeData.higAuthor || []);
const hotFiles = computed(() => homeStore.homeData.hotFileData || []);

const WEEKDAYS = ['日', '一', '二', '三', '四', '五', '六'];
const todayText = (() => {
  const now = new Date();
  return `${now.getFullYear()}年${now.getMonth() + 1}月${now.getDate()}日 · 星期${WEEKDAYS[now.getDay()]}`;
})();

const formatCount = (value) => {
  const number = Number(value) || 0;
  if (number >= 10000) {
    return `${(number / 10000).toFixed(number >= 100000 ? 0 : 1).replace(/\.0$/, '')}万`;
  }
  return String(number);
};

// 后端返回 “YYYY-MM-DD HH:mm:ss”，直接取日期部分，避免不同浏览器按时区解析出错。
const parseArticleDate = (value) => {
  const match = String(value || '').match(/(\d{4})-(\d{2})-(\d{2})/);
  if (!match) {
    return { year: '', monthDay: '', iso: '' };
  }
  const [, year, month, day] = match;
  return { year, monthDay: `${month}.${day}`, iso: `${year}-${month}-${day}` };
};

const buildExcerpt = (html) => extractPlainTextFromHTML(html || '').replace(/\s+/g, ' ').trim().slice(0, 160);

const fileExt = (name) => {
  const ext = String(name || '').split('.').pop();
  return ext && ext !== name && ext.length <= 4 ? ext.toUpperCase() : 'FILE';
};

// ================= 文章列表 =================
const pageSize = 8;
const articles = ref([]);
const page = ref(1);
const total = ref(0);
const loading = ref(false);
const noMore = ref(false);
const searchKeyword = ref('');
const activeKeyword = ref('');
// 每次重新搜索都会递增，丢弃之前尚未返回的请求，避免旧结果混进新列表。
let requestId = 0;

const fetchArticles = async () => {
  if (loading.value || noMore.value) return;
  const currentRequest = requestId;
  loading.value = true;
  try {
    const res = await sendAxiosRequest('/blog-api/blog/getAllBlog', {
      page: page.value,
      pageSize,
      keyword: activeKeyword.value
    });
    if (currentRequest !== requestId) return;
    if (!res || res.isError) {
      throw new Error(res?.errMsg || '获取文章失败');
    }
    const rows = res.result?.data || [];
    total.value = Number(res.result?.total) || 0;
    articles.value.push(...rows.map(article => ({
      ...article,
      DATE: parseArticleDate(article.CREATE_TIME),
      EXCERPT: buildExcerpt(article.MAINTEXT),
      ILLUSTRATION: extractFirstImage(article.MAINTEXT),
      MAINTEXT: undefined
    })));
    page.value++;
    noMore.value = rows.length < pageSize || articles.value.length >= total.value;
  } catch (e) {
    if (currentRequest === requestId) {
      console.error('获取文章失败', e);
      ElMessage.error('文章加载失败，请稍后重试');
    }
  } finally {
    if (currentRequest === requestId) {
      loading.value = false;
      nextTick(observeSentinel);
    }
  }
};

const resetAndLoad = () => {
  requestId++;
  activeKeyword.value = searchKeyword.value.trim();
  page.value = 1;
  total.value = 0;
  noMore.value = false;
  loading.value = false;
  articles.value = [];
  fetchArticles();
};

const debouncedSearch = debounce(() => {
  if (searchKeyword.value.trim() !== activeKeyword.value) {
    resetAndLoad();
  }
}, 400);

const searchNow = () => {
  debouncedSearch.cancel();
  resetAndLoad();
};

const clearSearch = () => {
  searchKeyword.value = '';
  searchNow();
};

// 滚动到列表底部时自动加载下一页；“加载更多”按钮作为兜底。
const sentinelRef = ref(null);
let observer = null;

function observeSentinel() {
  if (!observer || !sentinelRef.value) return;
  // 重新观察会立即触发一次回调，内容不足一屏时可以继续加载。
  observer.unobserve(sentinelRef.value);
  observer.observe(sentinelRef.value);
}

onMounted(() => {
  fetchArticles();
  if ('IntersectionObserver' in window) {
    observer = new IntersectionObserver((entries) => {
      if (entries.some(entry => entry.isIntersecting) && articles.value.length) {
        fetchArticles();
      }
    }, { rootMargin: '0px 0px 240px 0px' });
    observeSentinel();
  }
});

onBeforeUnmount(() => {
  debouncedSearch.cancel();
  observer?.disconnect();
});

// ================= 站点数据 =================
const siteStats = ref(null);
const loadSiteStats = async () => {
  try {
    const res = await sendAxiosRequest('/blog-api/home/getWebsiteStatistics');
    if (res && !res.isError) {
      siteStats.value = res.result;
    }
  } catch (e) {
    console.error('获取站点统计失败', e);
  }
};
loadSiteStats();

const topAlert = ref([]);
const setTopAlert = async () => {
  topAlert.value = await getAnnouncementByRouterName("Home");
};
setTopAlert();

// ================= 跳转 =================
const blogHref = (guid) => router.resolve({ name: 'oneBlog', params: { g: guid } }).href;

function openBlog(blog) {
  // 用户在拖选文字时不触发跳转
  if (window.getSelection()?.toString()) return;
  pubOpenOneBlog(router, blog.GUID);
}

function openUser(userCode) {
  pubOpenUser(router, userCode);
}

function openFile(file) {
  router.push({ name: "Resources", query: { g: file.GUID } });
}

function goToAdmin() {
  openUser(adminUserCode);
}

function goToPublishBlog() {
  router.push({ name: "MyBlog" });
}

function goToUpload() {
  router.push({ name: "Resources" });
}

function goMe() {
  const userCode = userStore?.userBean?.code;
  if (!userCode) {
    ElMessage.info("登录后可以拥有自己的个人主页");
    return;
  }
  openUser(userCode);
}
</script>

<style scoped>
.home {
  --ink: #1f2328;
  --ink-soft: #4b5563;
  --muted: #8a8f98;
  --line: #e7e5e0;
  --paper: #faf9f6;
  --surface: #ffffff;
  --accent: #0b6fa4;
  --serif: "Noto Serif SC", "Source Han Serif SC", "Songti SC", "STSong", serif;

  min-height: 100%;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
  background: var(--paper);
  color: var(--ink);
  padding-bottom: 64px;
}

/* ============ 页头 ============ */
.masthead {
  border-bottom: 1px solid var(--line);
  background: var(--surface);
}

.masthead-inner {
  max-width: 1120px;
  margin: 0 auto;
  padding: 40px 24px 32px;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 32px;
}

.masthead-date {
  margin: 0 0 10px;
  font-size: 13px;
  color: var(--muted);
  letter-spacing: 0.04em;
}

.masthead-title {
  margin: 0;
  font-family: var(--serif);
  font-size: 44px;
  font-weight: 700;
  line-height: 1.1;
  letter-spacing: -0.01em;
}

.masthead-desc {
  margin: 12px 0 0;
  font-size: 15px;
  color: var(--ink-soft);
}

.masthead-links {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 20px;
  margin-top: 20px;
}

.masthead-links a {
  font-size: 14px;
  color: var(--ink);
  text-decoration: none;
  border-bottom: 1px solid var(--line);
  padding-bottom: 2px;
  transition: border-color 0.15s, color 0.15s;
}

.masthead-links a:hover {
  color: var(--accent);
  border-color: var(--accent);
}

.masthead-stats {
  display: flex;
  margin: 0;
  flex-shrink: 0;
}

.stat {
  padding: 0 24px;
  border-left: 1px solid var(--line);
}

.stat:first-child {
  border-left: none;
  padding-left: 0;
}

.stat:last-child {
  padding-right: 0;
}

.stat dt {
  font-size: 12px;
  color: var(--muted);
  margin-bottom: 6px;
}

.stat dd {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

/* ============ 布局 ============ */
.layout {
  max-width: 1120px;
  margin: 0 auto;
  padding: 32px 24px 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 56px;
  align-items: start;
}

.section-label {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--ink);
  letter-spacing: 0.06em;
}

/* ============ 文章列表 ============ */
.feed-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 12px;
  border-bottom: 2px solid var(--ink);
}

.search {
  position: relative;
  display: flex;
  align-items: center;
  width: 240px;
  border-bottom: 1px solid var(--line);
  transition: border-color 0.15s;
}

.search:focus-within {
  border-color: var(--ink);
}

.search-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  fill: none;
  stroke: var(--muted);
  stroke-width: 1.8;
  stroke-linecap: round;
}

.search input {
  flex: 1;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  padding: 6px 8px;
  font-size: 14px;
  color: var(--ink);
}

.search input::placeholder {
  color: var(--muted);
}

.search input::-webkit-search-cancel-button {
  display: none;
}

.search-clear {
  border: none;
  background: none;
  padding: 0 2px;
  font-size: 18px;
  line-height: 1;
  color: var(--muted);
  cursor: pointer;
}

.search-clear:hover {
  color: var(--ink);
}

.feed-hint {
  margin: 12px 0 0;
  font-size: 13px;
  color: var(--muted);
}

.article-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.article {
  display: grid;
  grid-template-columns: 64px minmax(0, 1fr);
  gap: 20px;
  padding: 24px 0;
  border-bottom: 1px solid var(--line);
  cursor: pointer;
}

.article.has-cover {
  grid-template-columns: 64px minmax(0, 1fr) 148px;
}

.article-date {
  display: flex;
  flex-direction: column;
  padding-top: 3px;
  font-variant-numeric: tabular-nums;
}

.article-day {
  font-size: 18px;
  font-weight: 600;
  color: var(--ink);
}

.article-year {
  margin-top: 2px;
  font-size: 12px;
  color: var(--muted);
}

.article-title {
  margin: 0;
  font-size: 19px;
  font-weight: 600;
  line-height: 1.45;
}

.article-title a {
  color: var(--ink);
  text-decoration: none;
  background-image: linear-gradient(var(--accent), var(--accent));
  background-size: 0 1px;
  background-repeat: no-repeat;
  background-position: 0 100%;
  transition: background-size 0.25s, color 0.15s;
}

.article:hover .article-title a {
  color: var(--accent);
  background-size: 100% 1px;
}

.article-excerpt {
  margin: 8px 0 0;
  font-size: 14px;
  line-height: 1.75;
  color: var(--ink-soft);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.article-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  font-size: 12px;
  color: var(--muted);
}

.meta-avatar {
  flex-shrink: 0;
  font-size: 11px;
  background: #e9e6df;
  color: var(--ink-soft);
}

.meta-author {
  color: var(--ink-soft);
}

.meta-sep {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: currentColor;
}

.article-cover {
  width: 148px;
  height: 100px;
  object-fit: cover;
  border-radius: 4px;
  background: #efede8;
}

/* 加载占位 */
.skeleton-row {
  display: grid;
  grid-template-columns: 64px 1fr;
  gap: 20px;
  padding: 24px 0;
  border-bottom: 1px solid var(--line);
}

.skeleton-date,
.skeleton-line {
  display: block;
  border-radius: 3px;
  background: linear-gradient(90deg, #efede8 25%, #f6f4f0 50%, #efede8 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s ease-in-out infinite;
}

.skeleton-date {
  width: 44px;
  height: 32px;
}

.skeleton-lines {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.skeleton-line {
  height: 12px;
}

.skeleton-line.w-60 { width: 60%; height: 18px; }
.skeleton-line.w-90 { width: 90%; }
.skeleton-line.w-30 { width: 30%; }

@keyframes shimmer {
  from { background-position: 100% 0; }
  to { background-position: -100% 0; }
}

.feed-empty {
  padding: 56px 0;
  text-align: center;
  color: var(--muted);
  font-size: 14px;
}

.feed-empty p {
  margin: 0 0 12px;
}

.feed-foot {
  padding: 28px 0 0;
  text-align: center;
  min-height: 24px;
}

.feed-status {
  font-size: 13px;
  color: var(--muted);
}

.text-button {
  border: 1px solid var(--line);
  background: var(--surface);
  color: var(--ink);
  padding: 8px 22px;
  border-radius: 999px;
  font-size: 13px;
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s;
}

.text-button:hover {
  border-color: var(--ink);
}

/* ============ 侧栏 ============ */
.sidebar {
  display: flex;
  flex-direction: column;
  gap: 40px;
}

.panel .section-label {
  padding-bottom: 12px;
  border-bottom: 2px solid var(--ink);
}

.rank-list,
.author-list,
.file-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.rank-item {
  display: flex;
  gap: 14px;
  padding: 14px 0;
  border-bottom: 1px solid var(--line);
  cursor: pointer;
}

.rank-no {
  flex-shrink: 0;
  width: 22px;
  padding-top: 1px;
  font-family: var(--serif);
  font-size: 15px;
  font-weight: 700;
  color: #c4c1ba;
  font-variant-numeric: tabular-nums;
}

.rank-no.top {
  color: var(--accent);
}

.rank-body {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.rank-title {
  font-size: 14px;
  line-height: 1.55;
  color: var(--ink);
  text-decoration: none;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.rank-item:hover .rank-title {
  color: var(--accent);
}

.rank-meta {
  font-size: 12px;
  color: var(--muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.author {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--line);
  cursor: pointer;
}

.author-avatar {
  flex-shrink: 0;
  background: #e9e6df;
  color: var(--ink-soft);
}

.author-body {
  display: flex;
  flex-direction: column;
  gap: 3px;
  flex: 1;
  min-width: 0;
}

.author-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink);
}

.author:hover .author-name {
  color: var(--accent);
}

.author-remark {
  font-size: 12px;
  color: var(--muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.author-count {
  flex-shrink: 0;
  font-size: 15px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.author-count small {
  margin-left: 2px;
  font-size: 11px;
  font-weight: 400;
  color: var(--muted);
}

.file {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 0;
  border-bottom: 1px solid var(--line);
  font-size: 13px;
  cursor: pointer;
}

.file-ext {
  flex-shrink: 0;
  min-width: 40px;
  padding: 2px 0;
  border: 1px solid var(--line);
  border-radius: 3px;
  text-align: center;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: var(--ink-soft);
}

.file-name {
  flex: 1;
  min-width: 0;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file:hover .file-name {
  color: var(--accent);
}

.file-count {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--muted);
  font-variant-numeric: tabular-nums;
}

.panel-more {
  display: inline-block;
  margin-top: 14px;
  font-size: 13px;
  color: var(--ink-soft);
  text-decoration: none;
}

.panel-more::after {
  content: " →";
}

.panel-more:hover {
  color: var(--accent);
}

/* ============ 响应式 ============ */
@media (max-width: 960px) {
  .masthead-inner {
    flex-direction: column;
    align-items: flex-start;
  }

  .layout {
    grid-template-columns: minmax(0, 1fr);
    gap: 48px;
  }
}

@media (max-width: 640px) {
  .masthead-inner {
    padding: 28px 16px 24px;
    gap: 24px;
  }

  .masthead-title {
    font-size: 34px;
  }

  .masthead-stats {
    width: 100%;
  }

  .stat {
    flex: 1;
    padding: 0 12px;
  }

  .stat dd {
    font-size: 20px;
  }

  .layout {
    padding: 24px 16px 0;
  }

  .feed-head {
    flex-direction: column;
    align-items: stretch;
  }

  .search {
    width: 100%;
  }

  .article,
  .article.has-cover {
    grid-template-columns: minmax(0, 1fr) 96px;
    gap: 14px;
    padding: 20px 0;
  }

  .article:not(.has-cover) {
    grid-template-columns: minmax(0, 1fr);
  }

  .article-date {
    display: none;
  }

  .article-title {
    font-size: 17px;
  }

  .article-cover {
    width: 96px;
    height: 72px;
  }

  .skeleton-row {
    grid-template-columns: 1fr;
  }

  .skeleton-date {
    display: none;
  }
}
</style>
