<template>
  <div class="desk">
    <Announcement v-for="al in topAlert" :key="al.GUID" :TEXT="al.TEXT" :URL="al.URL" :URLNAME="al.URLNAME"/>

    <div class="desk-inner">
      <!-- 封面标签 -->
      <header class="cover">
        <div class="cover-label">
          <span class="tape tape-left"></span>
          <span class="tape tape-right"></span>
          <h1 class="cover-title">YnsStudy 学习手账</h1>
          <p class="cover-desc">记录编程学习、技术实践与生活思考。</p>
          <p v-if="siteStats" class="cover-stats">
            写了 <b>{{ formatCount(siteStats.ARTICLENUM) }}</b> 篇，
            被翻阅 <b>{{ formatCount(siteStats.VIEW_PAGE) }}</b> 次，
            <b>{{ formatCount(siteStats.USERNUM) }}</b> 位朋友来过
          </p>
        </div>
        <div class="date-stamp" aria-hidden="true">
          <span class="stamp-month">{{ today.month }}月</span>
          <span class="stamp-day">{{ today.day }}</span>
          <span class="stamp-week">周{{ today.weekday }}</span>
        </div>
      </header>

      <div class="layout">
        <main class="notebook">
          <!-- 本子顶部的索引标签 -->
          <nav class="index-tabs" aria-label="快捷入口">
            <a href="#" class="index-tab tab-yellow" @click.prevent="goToPublishBlog">写文章</a>
            <a href="#" class="index-tab tab-green" @click.prevent="goToUpload">上传资源</a>
            <a href="#" class="index-tab tab-pink" @click.prevent="goMe">我的主页</a>
            <a href="#" class="index-tab tab-blue" @click.prevent="goToAdmin">关于站长</a>
          </nav>

          <div class="page">
            <div class="page-head">
              <h2 class="hand page-title">
                {{ activeKeyword ? `找到的「${activeKeyword}」` : '最近写下的' }}
                <small v-if="activeKeyword && !loading">共 {{ total }} 篇</small>
              </h2>
              <label class="search">
                <span class="hand search-label">找一找</span>
                <input
                    v-model="searchKeyword"
                    type="search"
                    placeholder="标题或正文里的字"
                    aria-label="搜索文章"
                    @input="debouncedSearch"
                    @keydown.enter="searchNow"
                />
                <button v-if="searchKeyword" type="button" class="search-clear" aria-label="清除搜索" @click="clearSearch">×</button>
              </label>
            </div>

            <ol class="entries">
              <li
                  v-for="(article, index) in articles"
                  :key="article.GUID"
                  class="entry"
                  @click="openBlog(article)"
              >
                <time class="entry-date hand" :datetime="article.DATE.iso">
                  <span class="entry-md">{{ article.DATE.month }}月{{ article.DATE.day }}日</span>
                  <span class="entry-week">{{ article.DATE.year }} · 周{{ article.DATE.weekday }}</span>
                </time>

                <div class="entry-body">
                  <h3 class="entry-title">
                    <a :href="blogHref(article.GUID)" target="_blank" rel="noopener" @click.stop>{{ article.BLOG_TITLE }}</a>
                  </h3>
                  <p v-if="article.EXCERPT" class="entry-excerpt">{{ article.EXCERPT }}</p>
                  <div class="entry-meta">
                    <el-avatar :src="article.AVATAR" :size="20" class="meta-avatar">
                      {{ article.USERNAME?.charAt(0) }}
                    </el-avatar>
                    <span>{{ article.USERNAME }}</span>
                    <span class="meta-dot">·</span>
                    <span>{{ formatCount(article.VIEW_PAGE) }} 次阅读</span>
                  </div>
                </div>

                <figure v-if="article.ILLUSTRATION" class="polaroid" :class="index % 2 ? 'tilt-left' : 'tilt-right'">
                  <span class="tape tape-photo"></span>
                  <img
                      :src="article.ILLUSTRATION"
                      :alt="`${article.BLOG_TITLE} 的配图`"
                      loading="lazy"
                      decoding="async"
                      @error="article.ILLUSTRATION = ''"
                  />
                </figure>
              </li>
            </ol>

            <div v-if="loading && !articles.length" class="entries-loading" aria-hidden="true">
              <div v-for="n in 3" :key="n" class="ghost-entry">
                <span class="ghost ghost-date"></span>
                <div class="ghost-lines">
                  <span class="ghost ghost-title"></span>
                  <span class="ghost"></span>
                  <span class="ghost ghost-short"></span>
                </div>
              </div>
            </div>

            <div v-if="!loading && !articles.length" class="page-empty hand">
              <p>{{ activeKeyword ? '翻遍了也没找到，换个词试试？' : '这一页还是空白的。' }}</p>
              <button v-if="activeKeyword" type="button" class="paper-button" @click="clearSearch">不找了</button>
            </div>

            <div ref="sentinelRef" class="page-foot">
              <span v-if="loading && articles.length" class="hand foot-text">正在翻页…</span>
              <button
                  v-else-if="!noMore && articles.length"
                  type="button"
                  class="paper-button"
                  @click="fetchArticles"
              >翻下一页</button>
              <span v-else-if="noMore && articles.length" class="hand foot-text">— 写到这里就没有了 —</span>
            </div>
          </div>
        </main>

        <aside class="sidebar">
          <section v-if="hotBlogs.length" class="sticky-note">
            <span class="tape tape-note"></span>
            <h2 class="hand note-title">最近大家在看</h2>
            <ol class="hot-list">
              <li v-for="(blog, index) in hotBlogs" :key="blog.GUID" class="hot-item" @click="openBlog(blog)">
                <span class="hot-no hand">{{ index + 1 }}.</span>
                <div class="hot-body">
                  <a class="hot-title" :href="blogHref(blog.GUID)" :title="blog.BLOG_TITLE" target="_blank" rel="noopener" @click.stop>
                    {{ blog.BLOG_TITLE }}
                  </a>
                  <span class="hot-meta">{{ blog.USERNAME }} · {{ formatCount(blog.VIEW_PAGE) }} 阅读 · {{ formatCount(blog.COMMENT_COUNT) }} 评论</span>
                </div>
              </li>
            </ol>
          </section>

          <section v-if="authors.length" class="card">
            <h2 class="hand card-title">常来写字的人</h2>
            <ul class="author-list">
              <li v-for="author in authors" :key="author.USERCODE" class="author" @click="openUser(author.USERCODE)">
                <el-avatar :src="author.AVATAR" :size="38" class="author-avatar">
                  {{ author.USERNAME?.charAt(0) }}
                </el-avatar>
                <div class="author-body">
                  <span class="author-name">{{ author.USERNAME || '未命名' }}</span>
                  <span class="author-remark">{{ author.REMARK || '还没有写签名' }}</span>
                </div>
                <span class="author-count hand">{{ formatCount(author.ARTICLE_COUNT) }} 篇</span>
              </li>
            </ul>
          </section>

          <section v-if="hotFiles.length" class="card envelope">
            <h2 class="hand card-title">资料袋</h2>
            <ul class="file-list">
              <li v-for="file in hotFiles" :key="file.GUID" class="file" @click="openFile(file)">
                <span class="file-ext">{{ fileExt(file.ORIGINALFILENAME) }}</span>
                <span class="file-name" :title="file.ORIGINALFILENAME">{{ file.ORIGINALFILENAME }}</span>
                <span class="file-count">{{ formatCount(file.DOWNNUM) }} 次</span>
              </li>
            </ul>
            <a href="#" class="hand envelope-more" @click.prevent="goToUpload">去资源页看看 →</a>
          </section>
        </aside>
      </div>
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
const today = (() => {
  const now = new Date();
  return { month: now.getMonth() + 1, day: now.getDate(), weekday: WEEKDAYS[now.getDay()] };
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
    return { year: '', month: '', day: '', weekday: '', iso: '' };
  }
  const [, year, month, day] = match;
  const weekday = WEEKDAYS[new Date(Number(year), Number(month) - 1, Number(day)).getDay()];
  return { year, month: Number(month), day: Number(day), weekday, iso: `${year}-${month}-${day}` };
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

// 滚动到列表底部时自动加载下一页；“翻下一页”按钮作为兜底。
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
.desk {
  --ink: #2b2a27;
  --ink-soft: #57534c;
  --muted: #918b80;
  --paper: #fffdf8;
  --desk: #efe8da;
  --rule: #e6dfd1;
  --margin-red: #e8a59b;
  --pen: #2f5d8a;
  --tape-yellow: rgba(246, 214, 120, 0.78);
  --tape-green: rgba(160, 205, 180, 0.78);
  --tape-pink: rgba(240, 175, 175, 0.75);
  --hand: "Kaiti SC", "STKaiti", "KaiTi", "楷体", "AR PL UKai CN", serif;

  min-height: 100%;
  padding: 0 0 72px;
  color: var(--ink);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
  background-color: var(--desk);
  /* 点阵底纹 */
  background-image: radial-gradient(rgba(120, 104, 80, 0.18) 1px, transparent 1px);
  background-size: 22px 22px;
}

.desk-inner {
  max-width: 1120px;
  margin: 0 auto;
  padding: 36px 24px 0;
}

.hand {
  font-family: var(--hand);
  font-weight: normal;
}

/* 半透明的和纸胶带 */
.tape {
  position: absolute;
  width: 84px;
  height: 22px;
  background: var(--tape-yellow);
  box-shadow: 0 1px 1px rgba(0, 0, 0, 0.04);
  /* 胶带两端的锯齿撕口 */
  -webkit-mask: linear-gradient(90deg, transparent 0 2px, #000 2px calc(100% - 2px), transparent calc(100% - 2px)),
  repeating-linear-gradient(0deg, #000 0 3px, transparent 3px 5px);
  mask: linear-gradient(90deg, transparent 0 2px, #000 2px calc(100% - 2px), transparent calc(100% - 2px)),
  repeating-linear-gradient(0deg, #000 0 3px, transparent 3px 5px);
  pointer-events: none;
}

/* ============ 封面 ============ */
.cover {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 40px;
}

.cover-label {
  position: relative;
  max-width: 560px;
  padding: 28px 36px 24px;
  background: var(--paper);
  border: 1px solid var(--rule);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.06), 0 8px 20px -12px rgba(60, 50, 30, 0.25);
  transform: rotate(-0.8deg);
}

.tape-left {
  top: -10px;
  left: -22px;
  transform: rotate(-32deg);
}

.tape-right {
  top: -9px;
  right: -20px;
  background: var(--tape-green);
  transform: rotate(28deg);
}

.cover-title {
  margin: 0;
  font-family: var(--hand);
  font-size: 34px;
  font-weight: normal;
  letter-spacing: 0.02em;
}

.cover-desc {
  margin: 8px 0 0;
  font-size: 14px;
  color: var(--ink-soft);
}

.cover-stats {
  margin: 14px 0 0;
  padding-top: 12px;
  border-top: 1px dashed var(--rule);
  font-family: var(--hand);
  font-size: 15px;
  color: var(--ink-soft);
}

.cover-stats b {
  font-family: Georgia, "Times New Roman", serif;
  font-weight: normal;
  font-size: 18px;
  color: var(--pen);
  padding: 0 2px;
}

/* 日期印章 */
.date-stamp {
  flex-shrink: 0;
  width: 108px;
  height: 108px;
  margin-right: 12px;
  border: 2px solid rgba(194, 72, 62, 0.75);
  border-radius: 50%;
  outline: 1px solid rgba(194, 72, 62, 0.45);
  outline-offset: 3px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: rgba(194, 72, 62, 0.85);
  font-family: var(--hand);
  transform: rotate(-10deg);
  opacity: 0.9;
}

.stamp-month,
.stamp-week {
  font-size: 13px;
  letter-spacing: 0.1em;
}

.stamp-day {
  font-family: Georgia, "Times New Roman", serif;
  font-size: 38px;
  line-height: 1.05;
}

/* ============ 布局 ============ */
.layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 40px;
  align-items: start;
}

/* ============ 本子 ============ */
.notebook {
  position: relative;
  padding-top: 30px;
}

.index-tabs {
  position: absolute;
  top: 0;
  right: 24px;
  display: flex;
  gap: 6px;
}

.index-tab {
  display: block;
  padding: 6px 14px 12px;
  border-radius: 6px 6px 0 0;
  font-family: var(--hand);
  font-size: 14px;
  color: var(--ink);
  text-decoration: none;
  transform: translateY(4px);
  transition: transform 0.15s ease;
}

.index-tab:hover {
  transform: translateY(0);
}

.tab-yellow { background: #f6e3a1; }
.tab-green { background: #c7e2cf; }
.tab-pink { background: #f3cccc; }
.tab-blue { background: #c9dbeb; }

.page {
  position: relative;
  z-index: 1;
  padding: 28px 32px 32px 64px;
  background: var(--paper);
  border: 1px solid var(--rule);
  border-radius: 2px 6px 6px 2px;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.06), 0 12px 28px -18px rgba(60, 50, 30, 0.35);
}

/* 左侧装订孔 */
.page::before {
  content: "";
  position: absolute;
  top: 18px;
  bottom: 18px;
  left: 16px;
  width: 14px;
  background-image: radial-gradient(circle at 7px 7px, var(--desk) 5px, rgba(120, 104, 80, 0.25) 5.5px, transparent 6.5px);
  background-size: 14px 44px;
}

/* 页边红线 */
.page::after {
  content: "";
  position: absolute;
  top: 0;
  bottom: 0;
  left: 46px;
  width: 1px;
  background: var(--margin-red);
  opacity: 0.6;
}

.page-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--rule);
}

.page-title {
  margin: 0;
  font-size: 24px;
}

.page-title small {
  margin-left: 8px;
  font-size: 14px;
  color: var(--muted);
}

.search {
  display: flex;
  align-items: baseline;
  gap: 8px;
  width: 250px;
  border-bottom: 1px solid var(--ink-soft);
}

.search-label {
  flex-shrink: 0;
  font-size: 15px;
  color: var(--ink-soft);
}

.search input {
  flex: 1;
  min-width: 0;
  padding: 4px 0;
  border: none;
  outline: none;
  background: transparent;
  font-size: 14px;
  color: var(--pen);
}

.search input::placeholder {
  color: #b9b2a6;
}

.search input::-webkit-search-cancel-button {
  display: none;
}

.search-clear {
  border: none;
  background: none;
  padding: 0;
  font-size: 18px;
  line-height: 1;
  color: var(--muted);
  cursor: pointer;
}

/* ============ 一篇日记 ============ */
.entries {
  list-style: none;
  margin: 0;
  padding: 0;
}

.entry {
  display: grid;
  grid-template-columns: 96px minmax(0, 1fr) auto;
  gap: 20px;
  align-items: start;
  padding: 26px 0;
  border-bottom: 1px dashed var(--rule);
  cursor: pointer;
}

.entry:last-child {
  border-bottom: none;
}

.entry-date {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding-top: 2px;
  color: var(--ink-soft);
}

.entry-md {
  font-size: 17px;
  color: var(--ink);
}

.entry-week {
  font-size: 13px;
  color: var(--muted);
}

.entry-title {
  margin: 0;
  font-size: 19px;
  font-weight: 600;
  line-height: 1.5;
}

/* 荧光笔划过标题 */
.entry-title a {
  color: var(--ink);
  text-decoration: none;
  background-image: linear-gradient(transparent 58%, rgba(250, 216, 96, 0.7) 58%, rgba(250, 216, 96, 0.7) 92%, transparent 92%);
  background-size: 0 100%;
  background-repeat: no-repeat;
  transition: background-size 0.35s ease;
  -webkit-box-decoration-break: clone;
  box-decoration-break: clone;
}

.entry:hover .entry-title a {
  background-size: 100% 100%;
}

.entry-excerpt {
  margin: 8px 0 0;
  font-size: 14px;
  line-height: 1.8;
  color: var(--ink-soft);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.entry-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  font-size: 12px;
  color: var(--muted);
}

.meta-avatar {
  flex-shrink: 0;
  font-size: 11px;
  background: #ece4d3;
  color: var(--ink-soft);
}

.meta-dot {
  color: #c9c1b3;
}

/* 拍立得配图 */
.polaroid {
  position: relative;
  margin: 2px 4px 0 0;
  padding: 6px 6px 18px;
  background: #fff;
  box-shadow: 0 1px 3px rgba(60, 50, 30, 0.18);
  transition: transform 0.2s ease;
}

.polaroid img {
  display: block;
  width: 132px;
  height: 92px;
  object-fit: cover;
  background: #efe9dd;
}

.tilt-right { transform: rotate(2deg); }
.tilt-left { transform: rotate(-2deg); }

.entry:hover .polaroid {
  transform: rotate(0deg);
}

.tape-photo {
  top: -9px;
  left: 50%;
  width: 56px;
  height: 18px;
  margin-left: -28px;
  background: var(--tape-pink);
  transform: rotate(-4deg);
}

.tilt-left .tape-photo {
  background: var(--tape-green);
  transform: rotate(5deg);
}

/* 加载占位 */
.ghost-entry {
  display: grid;
  grid-template-columns: 96px 1fr;
  gap: 20px;
  padding: 26px 0;
  border-bottom: 1px dashed var(--rule);
}

.ghost-lines {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.ghost {
  display: block;
  height: 12px;
  border-radius: 3px;
  background: #f1ebdf;
  animation: fade 1.4s ease-in-out infinite;
}

.ghost-date { width: 60px; height: 16px; }
.ghost-title { width: 55%; height: 18px; }
.ghost-short { width: 30%; }

@keyframes fade {
  50% { opacity: 0.45; }
}

.page-empty {
  padding: 56px 0 24px;
  text-align: center;
  font-size: 17px;
  color: var(--muted);
}

.page-empty p {
  margin: 0 0 14px;
}

.page-foot {
  padding-top: 24px;
  text-align: center;
  min-height: 24px;
}

.foot-text {
  font-size: 15px;
  color: var(--muted);
}

.paper-button {
  position: relative;
  padding: 8px 26px 8px 22px;
  border: 1px solid var(--rule);
  background: #fbf6ea;
  font-family: var(--hand);
  font-size: 15px;
  color: var(--ink);
  cursor: pointer;
  /* 右上角折页 */
  clip-path: polygon(0 0, calc(100% - 10px) 0, 100% 10px, 100% 100%, 0 100%);
  transition: background-color 0.15s;
}

.paper-button::after {
  content: "";
  position: absolute;
  top: 0;
  right: 0;
  width: 10px;
  height: 10px;
  background: linear-gradient(225deg, transparent 50%, #e9e0cc 50%);
}

.paper-button:hover {
  background: #f6edd8;
}

/* ============ 侧栏 ============ */
.sidebar {
  display: flex;
  flex-direction: column;
  gap: 36px;
  padding-top: 30px;
}

.sticky-note {
  position: relative;
  padding: 26px 22px 18px;
  background: #fbf0b4;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.1), 0 10px 18px -12px rgba(60, 50, 30, 0.4);
  transform: rotate(1.2deg);
}

.tape-note {
  top: -11px;
  left: 50%;
  margin-left: -42px;
  background: rgba(255, 255, 255, 0.55);
  transform: rotate(-3deg);
}

.note-title,
.card-title {
  margin: 0 0 10px;
  font-size: 20px;
}

.hot-list,
.author-list,
.file-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.hot-item {
  display: flex;
  gap: 8px;
  padding: 9px 0;
  border-bottom: 1px solid rgba(160, 135, 60, 0.18);
  cursor: pointer;
}

.hot-item:last-child {
  border-bottom: none;
}

.hot-no {
  flex-shrink: 0;
  width: 22px;
  font-size: 17px;
  line-height: 1.35;
  color: #a2802a;
}

.hot-body {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}

.hot-title {
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

.hot-item:hover .hot-title {
  text-decoration: underline;
  text-decoration-color: var(--pen);
  text-underline-offset: 3px;
}

.hot-meta {
  font-size: 12px;
  color: #8b7d55;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card {
  position: relative;
  padding: 20px 22px 14px;
  background: var(--paper);
  border: 1px solid var(--rule);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.06);
}

.author {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px dashed var(--rule);
  cursor: pointer;
}

.author:last-child {
  border-bottom: none;
}

.author-avatar {
  flex-shrink: 0;
  background: #ece4d3;
  color: var(--ink-soft);
  border: 2px solid #fff;
  box-shadow: 0 0 0 1px var(--rule);
}

.author-body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.author-name {
  font-size: 14px;
  font-weight: 600;
}

.author:hover .author-name {
  color: var(--pen);
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
  color: var(--pen);
}

/* 牛皮纸资料袋 */
.envelope {
  background: #e7d6b5;
  border-color: #d8c49d;
}

.envelope::before {
  content: "";
  position: absolute;
  top: 8px;
  left: 10px;
  right: 10px;
  border-top: 1px dashed rgba(110, 85, 40, 0.35);
}

.envelope .card-title {
  margin-top: 6px;
}

.file {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  padding: 8px 10px;
  background: #fffaf0;
  font-size: 13px;
  box-shadow: 0 1px 1px rgba(110, 85, 40, 0.15);
  cursor: pointer;
  transition: transform 0.15s ease;
}

.file:hover {
  transform: translateX(3px);
}

.file-ext {
  flex-shrink: 0;
  min-width: 38px;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: #8a6a32;
}

.file-name {
  flex: 1;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-count {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--muted);
}

.envelope-more {
  display: inline-block;
  margin-top: 4px;
  font-size: 15px;
  color: #6e5528;
  text-decoration: none;
}

.envelope-more:hover {
  text-decoration: underline;
  text-underline-offset: 3px;
}

/* ============ 响应式 ============ */
@media (max-width: 960px) {
  .layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .sidebar {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
    gap: 28px;
    padding-top: 8px;
  }
}

@media (max-width: 640px) {
  .desk-inner {
    padding: 24px 16px 0;
  }

  .cover {
    margin-bottom: 28px;
  }

  .cover-label {
    padding: 22px 20px 18px;
    transform: none;
  }

  .cover-title {
    font-size: 27px;
  }

  .date-stamp {
    display: none;
  }

  .index-tabs {
    right: 8px;
    gap: 4px;
  }

  .index-tab {
    padding: 5px 9px 11px;
    font-size: 13px;
  }

  .page {
    padding: 20px 16px 24px 40px;
  }

  .page::before {
    left: 8px;
  }

  .page::after {
    left: 30px;
  }

  .page-head {
    flex-direction: column;
    align-items: stretch;
  }

  .search {
    width: 100%;
  }

  .entry {
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 6px 14px;
    padding: 20px 0;
  }

  .entry-date {
    grid-column: 1 / -1;
    flex-direction: row;
    align-items: baseline;
    gap: 8px;
  }

  .entry-title {
    font-size: 17px;
  }

  .polaroid {
    padding: 4px 4px 12px;
  }

  .polaroid img {
    width: 84px;
    height: 64px;
  }

  .ghost-entry {
    grid-template-columns: 1fr;
  }

  .sticky-note {
    transform: none;
  }
}
</style>
