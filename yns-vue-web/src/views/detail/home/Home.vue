<template>
  <div class="page-container">
    <Announcement v-for="al in topAlert" :key="al.GUID" :TEXT="al.TEXT" :URL="al.URL" :URLNAME="al.URLNAME" />

    <div class="home-page-wrapper">
      <div class="three-column-layout">

        <aside class="left-column">
          <div class="side-card">
            <h3 class="section-title">
              <span class="title-icon">🔥</span> 热门榜单
            </h3>
            <ul class="blog-list">
              <li
                  v-for="(blog, index) in homeStore.homeData.hotBlogData"
                  :key="blog.id"
                  class="blog-item"
                  @click="hotBlogClick(blog)"
              >
                <div class="blog-rank" :class="'rank-' + (index + 1)">{{ index + 1 }}</div>

                <div class="blog-content">
                  <a class="blog-title" :title="blog.BLOG_TITLE">{{ blog.BLOG_TITLE }}</a>
                  <div class="blog-meta">
                    <span class="author-name">{{ blog.USERNAME }}</span>
                    <div class="stats">
                      <span><el-icon><View/></el-icon> {{ blog.VIEW_PAGE }}</span>
                      <span><el-icon><ChatDotSquare /></el-icon> {{ blog.COMMENT_COUNT }}</span>
                      <span><el-icon><Star/></el-icon> {{ blog.COLLECT_COUNT }}</span>
                    </div>
                  </div>
                </div>
              </li>
            </ul>
          </div>

          <div class="side-footer">
            <div class="footer-icon">🚀</div>
            <div class="footer-slogan">探索技术无限可能</div>
            <div class="footer-divider"></div>
          </div>
        </aside>

        <main class="middle-column">
          <div class="top-action-bar">
            <div class="action-btn-group">
              <div class="action-item" @click="goToAdmin">
                <div class="icon-wrapper admin-bg">👨‍💼</div>
                <span>站长主页</span>
              </div>
              <div class="action-item" @click="goToPublishBlog">
                <div class="icon-wrapper publish-bg">📝</div>
                <span>发布内容</span>
              </div>
              <div class="action-item" @click="goToUpload">
                <div class="icon-wrapper upload-bg">📤</div>
                <span>上传资源</span>
              </div>
              <div class="action-item" @click="goMe">
                <div class="icon-wrapper me-bg">🏠</div>
                <span>我的主页</span>
              </div>
            </div>
          </div>

          <div class="search-container">
            <el-input
                v-model="searchKeyword"
                class="premium-search"
                placeholder="搜索感兴趣的文章标题或内容..."
                clearable
                :prefix-icon="Search"
                @clear="resetAndLoad"
                @input="debouncedSearch"
            />
          </div>

          <div class="article-scroll">
            <div
                v-for="(article, index) in articles"
                :key="index"
                class="article-card"
                @click="hotBlogClick(article)"
            >
              <div class="article-header">
                <el-avatar
                    :src="article.AVATAR"
                    :size="28"
                    class="author-avatar"
                >
                  {{ article.USERNAME?.charAt(0) }}
                </el-avatar>
                <span class="author-name">{{ article.USERNAME }}</span>
                <span class="divider">·</span>
                <span class="publish-time">{{ pubFormatDate(article.CREATE_TIME) }}</span>
              </div>

              <div class="article-body">
                <div class="article-main">
                  <h3 class="article-title">{{ article.BLOG_TITLE }}</h3>
                  <p class="article-content" v-html="stripImages(article.MAINTEXT)"></p>
                </div>

                <div v-if="article.ILLUSTRATION" class="article-image">
                  <img :src="article.ILLUSTRATION" alt="文章插图" class="article-thumbnail"/>
                </div>
              </div>
            </div>

            <el-empty v-if="!articles.length && !loading" description="暂时没有发现内容哦" />

            <div class="load-more-wrapper">
              <el-button
                  v-if="!noMore && !loading && articles.length > 0"
                  type="primary"
                  round
                  plain
                  class="load-more-btn"
                  @click="fetchArticles"
              >
                浏览更多内容
              </el-button>
              <div v-if="loading" class="status-text loading-text">
                <span class="dot-typing"></span> 正在努力加载...
              </div>
              <el-divider v-if="noMore && articles.length > 0" class="end-text">我是有底线的</el-divider>
            </div>
          </div>
        </main>

        <aside class="right-column">
          <div class="side-card">
            <h3 class="section-title">
              <span class="title-icon">🌟</span> 优质作者
            </h3>
            <div class="author-list">
              <div
                  v-for="(author, index) in homeStore.homeData.higAuthor"
                  :key="index"
                  class="author-card"
                  @click="hotAuthorClick(author)"
              >
                <el-avatar :src="author.AVATAR" :size="46" class="author-avatar">
                  {{ author.USERNAME?.charAt(0) }}
                </el-avatar>
                <div class="author-info">
                  <div class="name">{{ author.USERNAME }}</div>
                  <div class="signature">{{ author.REMARK || '无签名' }}</div>
                  <div class="stats-row">
                    <span>粉丝 {{ author.FOLLOWER_COUNT }}</span>
                    <span class="v-divider"></span>
                    <span>文章 {{ author.ARTICLE_COUNT }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="side-card">
            <h3 class="section-title">
              <span class="title-icon">📥</span> 热门资源
            </h3>
            <ul class="download-list">
              <li
                  v-for="file in homeStore.homeData.hotFileData"
                  :key="file.GUID"
                  class="download-item"
                  @click="downloadClick(file)"
              >
                <div class="download-info">
                  <div class="download-title" :title="file.ORIGINALFILENAME">{{ file.ORIGINALFILENAME }}</div>
                  <div class="download-meta">
                    <span class="uploader">{{ file.USERNAME }}</span>
                    <span class="download-count">↓ {{ file.DOWNNUM }}</span>
                  </div>
                </div>
              </li>
            </ul>
          </div>

          <div class="side-footer">
            <div class="footer-slogan">记录生活 · 分享点滴</div>
            <div class="footer-copyright">© 2026 YnsStudy</div>
          </div>
        </aside>

      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onBeforeUnmount, ref } from "vue";
import { useHomeStore } from "@/stores/detail/home.js";
import { useRouter } from "vue-router";
import { useUserStore } from "@/stores/main/user.js";
import { extractFirstImage, pubFormatDate, sendAxiosRequest, stripImages } from "@/utils/common.js";
import { Star, View, Search,ChatDotSquare } from "@element-plus/icons-vue";
import { adminUserCode } from "@/config/vue-config.js";
import debounce from "lodash/debounce.js";
import { ElMessage } from "element-plus";
import { getAnnouncementByRouterName, pubOpenOneBlog, pubOpenUser } from "@/utils/blogUtil.js";
import Announcement from "@/components/detail/Announcement.vue";

const router = useRouter();
const userStore = useUserStore();
const homeStore = useHomeStore();
homeStore.initHomeData();

const articles = ref([])
const page = ref(1)
const pageSize = 5
const loading = ref(false)
const noMore = ref(false)
const searchKeyword = ref('')

const resetAndLoad = () => {
  page.value = 1
  noMore.value = false
  articles.value = []
  fetchArticles()
}

const debouncedSearch = debounce(() => {
  resetAndLoad()
}, 500)

const fetchArticles = async () => {
  if (loading.value || noMore.value) return
  loading.value = true
  try {
    const res = await sendAxiosRequest('/blog-api/blog/getAllBlog', {
      page: page.value,
      pageSize,
      keyword: searchKeyword.value
    })
    const newData = res.result.data;

    newData.forEach(article => {
      const firstImage = extractFirstImage(article.MAINTEXT);
      article.ILLUSTRATION = firstImage;
    });

    if (newData.length < pageSize) {
      noMore.value = true
    }
    articles.value.push(...newData)
    page.value++
  } catch (e) {
    console.error('获取文章失败', e)
  } finally {
    loading.value = false
  }
}

function hotBlogClick(blog) {
  pubOpenOneBlog(router,blog.GUID)
}

function downloadClick(file) {
  router.push({name: "Resources", query: {g: file.GUID}});
}

function goToAdmin() {
  pubOpenUser(router,adminUserCode);
}

async function goToPublishBlog() {
  router.push({name: "MyBlog"});
}

function goToUpload() {
  router.push({name: "Resources"});
}

function goMe(){
  let userCode = userStore?.userBean?.code || false;
  if(!userCode){
    ElMessage.info("登录后可以拥有自己的个人主页");
  }
  pubOpenUser(router,userCode);
}

const hotAuthorClick = (author) => {
  pubOpenUser(router,author.USERCODE);
};

const handleWindowScroll = () => {
  const scrollTop = window.pageYOffset || document.documentElement.scrollTop || document.body.scrollTop;
  const windowHeight = window.innerHeight;
  const docHeight = document.documentElement.scrollHeight;

  if (scrollTop + windowHeight >= docHeight - 100) {
    fetchArticles();
  }
};

onMounted(() => {
  fetchArticles();
  window.addEventListener('scroll', handleWindowScroll);
});

onBeforeUnmount(() => {
  window.removeEventListener('scroll', handleWindowScroll);
});

const topAlert = ref([]);
const setTopAlert = async ()=>{
  topAlert.value = await getAnnouncementByRouterName("Home");
}
setTopAlert();
</script>

<style scoped>
.page-container {
  background-color: #f2f5f8;
  min-height: 100vh;
  padding-bottom: 40px;
}

.home-page-wrapper {
  max-width: 1440px;
  margin: 0 auto;
  padding: 20px;
  box-sizing: border-box;
}

/* ====================================
   自然流排版：不再强行拉伸高度
   ==================================== */
.three-column-layout {
  display: flex;
  gap: 24px;
  align-items: flex-start; /* 顺其自然的高度 */
}

.left-column { width: 290px; flex-shrink: 0; }
.right-column { width: 300px; flex-shrink: 0; }
.middle-column { flex: 1; min-width: 0; }

.side-card {
  background: #ffffff;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.03);
  transition: box-shadow 0.3s ease;
}
.side-card:hover {
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.06);
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
  margin: 0 0 16px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}
.title-icon {
  font-size: 18px;
}

/* === 左侧：热门榜单 === */
.blog-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
}
.blog-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px 0;
  border-bottom: 1px solid #f2f3f5;
  cursor: pointer;
}
.blog-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}
.blog-rank {
  font-size: 15px;
  font-weight: bold;
  color: #b5b9c2;
  width: 20px;
  text-align: center;
  flex-shrink: 0;
  margin-top: 2px;
}
.rank-1 { color: #f53f3f; font-size: 18px; }
.rank-2 { color: #ff7d00; font-size: 17px; }
.rank-3 { color: #fadc19; font-size: 16px; }

.blog-content {
  flex: 1;
  min-width: 0;
}
.blog-title {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-wrap: break-word;
  word-break: break-all;
  white-space: normal;
  font-size: 15px;
  font-weight: 500;
  color: #1d2129;
  line-height: 1.6;
  transition: color 0.2s ease;
  margin-bottom: 10px;
}
.blog-item:hover .blog-title {
  color: #165dff;
}
.blog-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: #86909c;
}
.blog-meta .stats {
  display: flex;
  gap: 12px;
}
.blog-meta .stats span {
  display: flex;
  align-items: center;
  gap: 3px;
}

/* === 中间区域 === */
.top-action-bar {
  background: #ffffff;
  border-radius: 12px;
  padding: 24px 20px;
  margin-bottom: 20px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.03);
}
.action-btn-group {
  display: flex;
  justify-content: space-around;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}
.action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  transition: transform 0.2s ease;
}
.action-item:hover {
  transform: translateY(-4px);
}
.action-item span {
  font-size: 14px;
  font-weight: 500;
  color: #4e5969;
}
.action-item:hover span {
  color: #165dff;
}
.icon-wrapper {
  width: 48px;
  height: 48px;
  border-radius: 16px;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 22px;
  color: white;
  box-shadow: 0 4px 10px rgba(0,0,0,0.1);
}
.admin-bg { background: linear-gradient(135deg, #7280ed, #4d5af0); }
.publish-bg { background: linear-gradient(135deg, #42d392, #34a873); }
.upload-bg { background: linear-gradient(135deg, #ff9a9e, #fecfef); }
.me-bg { background: linear-gradient(135deg, #fbc2eb, #a6c1ee); }

.search-container {
  margin-bottom: 20px;
}
:deep(.premium-search .el-input__wrapper) {
  border-radius: 20px;
  padding: 4px 16px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.03);
  border: 1px solid transparent;
  transition: all 0.3s ease;
}
:deep(.premium-search .el-input__wrapper:hover),
:deep(.premium-search .el-input__wrapper.is-focus) {
  box-shadow: 0 6px 20px rgba(22, 93, 255, 0.1);
  border-color: #165dff;
}
:deep(.premium-search .el-input__inner) {
  font-size: 15px;
  height: 40px;
}

.article-card {
  background: #ffffff;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.03);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid transparent;
  overflow: hidden;
}
.article-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  border-color: rgba(22, 93, 255, 0.1);
}
.article-header {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  font-size: 13px;
  color: #86909c;
}
.author-avatar {
  margin-right: 8px;
}
.author-name {
  color: #4e5969;
  font-weight: 500;
}
.divider {
  margin: 0 8px;
  color: #c9cdd4;
}
.article-body {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}
.article-main {
  flex: 1;
  min-width: 0;
}
.article-title {
  margin: 0 0 8px 0;
  font-size: 18px;
  font-weight: 600;
  color: #1d2129;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-wrap: break-word;
  word-break: break-all;
  white-space: normal;
}
.article-content {
  margin: 0;
  font-size: 14px;
  color: #86909c;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-wrap: break-word;
  word-break: break-all;
  white-space: normal;
}
.article-image {
  flex-shrink: 0;
}
.article-thumbnail {
  width: 140px;
  height: 90px;
  object-fit: cover;
  border-radius: 6px;
  border: 1px solid #f2f3f5;
}

.load-more-wrapper {
  text-align: center;
  padding: 20px 0;
}
.load-more-btn {
  width: 160px;
  font-weight: 500;
}
.status-text {
  font-size: 14px;
  color: #86909c;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}
:deep(.el-divider__text) {
  background-color: #f2f5f8;
  color: #86909c;
  font-size: 13px;
}

/* === 右侧：优质作者 === */
.author-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.author-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  transition: background-color 0.2s ease;
  cursor: pointer;
}
.author-card:hover {
  background-color: #f7f8fa;
}
.author-info {
  flex: 1;
  min-width: 0;
}
.author-info .name {
  font-weight: 600;
  font-size: 15px;
  color: #1d2129;
  margin-bottom: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.signature {
  font-size: 12px;
  color: #86909c;
  margin-bottom: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.stats-row {
  display: flex;
  align-items: center;
  font-size: 12px;
  color: #86909c;
}
.v-divider {
  width: 1px;
  height: 10px;
  background-color: #e5e6eb;
  margin: 0 8px;
}

.download-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.download-item {
  padding: 12px;
  border-radius: 8px;
  background: #f7f8fa;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}
.download-item:hover {
  background: #ffffff;
  border-color: rgba(22, 93, 255, 0.2);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transform: translateX(4px);
}
.download-title {
  font-weight: 500;
  font-size: 14px;
  color: #1d2129;
  margin-bottom: 6px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-wrap: break-word;
  word-break: break-all;
  white-space: normal;
}
.download-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: #86909c;
}
.download-count {
  color: #00b42a;
  font-weight: 500;
  background: rgba(0, 180, 42, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
}

/* ====================================
   ✨ 侧边栏底部视觉收尾样式
   ==================================== */
.side-footer {
  text-align: center;
  padding: 10px 0 30px;
  color: #b5b9c2;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}
.footer-icon {
  font-size: 20px;
  margin-bottom: 4px;
}
.footer-slogan {
  font-size: 13px;
  letter-spacing: 1px;
}
.footer-copyright {
  font-size: 12px;
  color: #c9cdd4;
}
.footer-divider {
  width: 40px;
  height: 2px;
  background-color: #e5e6eb;
  margin-top: 8px;
  border-radius: 2px;
}

@media (max-width: 1024px) {
  .three-column-layout {
    flex-direction: column;
  }

  .left-column,
  .middle-column,
  .right-column {
    width: 100%;
  }

  .middle-column { order: 1; }
  .left-column { order: 2; }
  .right-column { order: 3; }

  /* 手机端可以隐藏侧边栏收尾，以免太长 */
  .side-footer {
    display: none;
  }
}

@media (max-width: 768px) {
  .article-body {
    flex-direction: column-reverse;
  }
  .article-image {
    width: 100%;
  }
  .article-thumbnail {
    width: 100%;
    height: auto;
    max-height: 180px;
    margin-bottom: 12px;
  }
  .home-page-wrapper {
    padding: 10px;
  }
}
</style>