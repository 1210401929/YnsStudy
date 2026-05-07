<template>
  <!-- 公告横幅 -->
  <Announcement v-for="al in topAlert" :key="al.GUID" :TEXT="al.TEXT" :URL="al.URL" :URLNAME="al.URLNAME" />
  <div class="home-page-wrapper">
    <!-- 主体三栏布局 -->
    <div class="three-column-layout">
      <!-- 左侧：热门文章 -->
      <div class="left-column">
        <h3 class="section-title">🔥 热门文章推荐</h3>
        <ul class="blog-list">
          <li
              v-for="blog in homeStore.homeData.hotBlogData"
              :key="blog.id"
              class="blog-item"
              @click="hotBlogClick(blog)"
          >
            <div class="blog-card">
              <a class="blog-title">{{ blog.BLOG_TITLE }}</a>
              <div class="blog-author">by {{ blog.USERNAME }}</div>
              <div class="blog-meta">
                <span>👁 {{ blog.VIEW_PAGE }}</span>
                <span>👍 {{ blog.LIKE_COUNT }}</span>
                <span>💬 {{ blog.COMMENT_COUNT }}</span>
                <span><el-icon><Star/></el-icon> {{ blog.COLLECT_COUNT }}</span>
              </div>
            </div>
          </li>
        </ul>
      </div>

      <!-- 中间按钮栏 -->
      <div class="middle-column">
        <!-- 滚动容器添加样式 -->
        <div class="article-scroll">
          <span class="top-action-bar">
            <el-button class="action-button" @click="goToAdmin">👨‍💼 站长主页</el-button>
            <el-button class="action-button" @click="goToPublishBlog">📝 发布内容</el-button>
            <el-button class="action-button" @click="goToUpload">📤 上传资源</el-button>
            <el-button class="action-button" @click="goMe"> 我的主页</el-button>
          </span>
          <!-- 搜索栏 -->
          <div class="search-bar">
            <el-input
                v-model="searchKeyword"
                placeholder="搜索文章标题或内容"
                clearable
                prefix-icon="el-icon-search"
                @clear="resetAndLoad"
                @input="debouncedSearch"
            />
          </div>
          <el-card
              v-for="(article, index) in articles"
              :key="index"
              class="article-card"
              shadow="hover"
          >
            <el-button
                link
                type="primary"
                class="read-more-btn"
                @click.stop="hotBlogClick(article)"
            >
              阅读全文 →
            </el-button>
            <h3 class="article-title">
              <el-avatar
                  :src="article.AVATAR"
                  size="large"
                  class="author-avatar title-avatar"
                  alt="用户头像"
              >
                {{ article.USERNAME?.charAt(0) }}
              </el-avatar>
              {{ article.BLOG_TITLE }}
            </h3>
            <div class="article-meta">
              <span>作者：{{ article.USERNAME }}</span>
              <span>时间：{{ pubFormatDate(article.CREATE_TIME) }}</span>
            </div>

            <!-- 文章内容和图片放在同一容器中 -->
            <div class="article-content-wrapper">
              <p class="article-content" v-html="stripImages(article.MAINTEXT)"></p>

              <!-- 图片显示在右侧 -->
              <div v-if="article.ILLUSTRATION" class="article-image">
                <img :src="article.ILLUSTRATION" alt="文章插图" class="article-thumbnail"/>
              </div>
            </div>
          </el-card>

          <el-empty v-if="!articles.length && !loading" description="暂无内容"/>
          <el-button
              v-if="!noMore && !loading"
              type="primary"
              link
              @click="fetchArticles"
              style="margin: 20px auto; display: block;"
          >
            加载更多
          </el-button>
          <div v-if="loading" class="loading-text">加载中...</div>
          <div v-if="noMore" class="end-text">没有更多文章了</div>
        </div>
      </div>

      <!-- 右侧：热门下载 + 优质作者 -->
      <div class="right-column">

        <h3 class="section-title">🌟 优质作者</h3>
        <el-card
            v-for="(author, index) in homeStore.homeData.higAuthor"
            :key="index"
            class="author-card"
            @click="hotAuthorClick(author)"
        >
          <div class="author-box">
            <el-avatar
                :src="author.AVATAR"
                size="large" author-card
                class="author-avatar"
                alt="用户头像"
            >
              {{ author.USERNAME?.charAt(0) }}
            </el-avatar>
            <div class="author-meta">
              <div class="name">{{ author.USERNAME }}</div>
              <div class="stats-row">
                <span>👩🏻‍🤝‍🧑🏽粉丝:{{ author.FOLLOWER_COUNT }}</span>&nbsp;&nbsp;&nbsp;&nbsp;
                <span>📄文章:{{ author.ARTICLE_COUNT }}</span>
              </div>
            </div>
          </div>
        </el-card>
        <h3 class="section-title">📥 热门下载内容</h3>
        <ul class="download-list">
          <li
              v-for="file in homeStore.homeData.hotFileData"
              :key="file.GUID"
              class="download-item"
              @click="downloadClick(file)"
          >
            <div class="download-card">
              <div class="download-title">{{ file.ORIGINALFILENAME }}</div>
              <div class="download-meta">
                <span>🧑 {{ file.USERNAME }}</span> |
                <span style="color:#04c279">下载次数: {{ file.DOWNNUM }}</span>
              </div>
            </div>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup>
import {onMounted, onBeforeUnmount, ref} from "vue";
import {useHomeStore} from "@/stores/detail/home.js";
import {useRouter} from "vue-router";
import {useUserStore} from "@/stores/main/user.js";
import {extractFirstImage, pubFormatDate, sendAxiosRequest, stripImages} from "@/utils/common.js";
import {Star} from "@element-plus/icons-vue";
import {adminUserCode} from "@/config/vue-config.js";
import debounce from "lodash/debounce.js";
import {ElMessage} from "element-plus";
import {getAnnouncementByRouterName, pubOpenOneBlog, pubOpenUser} from "@/utils/blogUtil.js";
import Announcement from "@/components/detail/Announcement.vue";

const router = useRouter();
const userStore = useUserStore();
const homeStore = useHomeStore();
homeStore.initHomeData();

// 文章列表数据
const articles = ref([])
const page = ref(1)
const pageSize = 5
const loading = ref(false)
const noMore = ref(false)
// 滚动容器引用
const scrollContainer = ref(null)

// 搜索文章
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

// 加载文章
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

    // 提取每篇文章的第一张图片
    newData.forEach(article => {
      const firstImage = extractFirstImage(article.MAINTEXT);
      article.ILLUSTRATION = firstImage; // 保存图片 URL
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

const handleScroll = () => {
  const el = scrollContainer.value
  if (el.scrollTop + el.clientHeight >= el.scrollHeight - 10) {
    fetchArticles()
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
  // 获取页面滚动相关数据
  const scrollTop = window.pageYOffset || document.documentElement.scrollTop || document.body.scrollTop;
  const windowHeight = window.innerHeight;
  const docHeight = document.documentElement.scrollHeight;

  // 判断滚动是否接近底部（这里预留10px偏差）
  if (scrollTop + windowHeight >= docHeight - 10) {
    fetchArticles();
  }
};

onMounted(() => {
  fetchArticles();
  //window.addEventListener('scroll', handleWindowScroll);
});

onBeforeUnmount(() => {
  window.removeEventListener('scroll', handleWindowScroll);
});

//公告横幅内容
const topAlert = ref([]);
const setTopAlert = async ()=>{
  topAlert.value = await getAnnouncementByRouterName("Home");
}
setTopAlert();
</script>

<style scoped>
.search-bar {
  margin-bottom: 20px;
  width: 100%;
}

.article-card {
  position: relative;
  margin-bottom: 20px;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  overflow: hidden;
  height: auto;
  min-height: 150px;
  /* 关键：强制长串内容在容器内换行 */
  word-break: break-all;        /* 遇到超长单词/字符也强制断开 */
  overflow-wrap: anywhere;      /* CSS3 推荐写法，任何地方都能换行 */
  white-space: normal;          /* 确保不是 nowrap */
}



.article-card:hover {
  transform: translateY(-2px);
}

.article-title {
  font-size: 18px;
  font-weight: 600;
  margin: 10px 0;
  color: #333;
}

.article-meta {
  font-size: 13px;
  color: #888;
  display: flex;
  gap: 20px;
  margin-bottom: 8px;
}

.article-content-wrapper {
  display: flex;
  align-items: flex-start; /* 确保文本和图片对齐 */
  gap: 16px; /* 控制文本与图片之间的间距 */
  margin-top: 12px;
}

.article-content {
  flex-grow: 1; /* 使得文章内容占据剩余空间 */
  font-size: 14px;
  color: #555;
  line-height: 1.6;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 3; /* 限制显示三行 */
  -webkit-box-orient: vertical;
}

.article-thumbnail {
  max-width: 150px; /* 设置图片最大宽度 */
  max-height: 150px;
  height: auto;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.article-image {
  flex-shrink: 0; /* 防止图片缩放 */
}

.read-more-btn {
  float: right;
  font-size: 13px;
  margin-top: -6px;
  margin-bottom: 6px;
}

.loading-text,
.end-text {
  text-align: center;
  color: #999;
  margin: 16px 0;
}

.home-page-wrapper {
  width: 100%;
  box-sizing: border-box;
  background-color: #f4f6f9;
  padding: 10px 0;
  margin: 10px auto; /* ← 加这一行：上下左右都有外边距 */
}

.top-action-bar {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin: 13px auto;
  flex-wrap: wrap;
}

.action-button {
  border: none;
  background-color: #ffffff;
  color: #409eff;
  font-weight: 500;
  border-radius: 20px;
  padding: 8px 18px;
  font-size: 14px;
  box-shadow: 0 2px 6px rgba(64, 158, 255, 0.1);
  transition: all 0.25s ease;
}

.action-button:hover {
  background-color: #aed7fa;
  color: white;
  transform: translateY(-2px);
}

.three-column-layout {
  display: flex;
  gap: 20px;
  width: 100%;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
}

.left-column,
.right-column {
  width: 280px;
  background: #fafafa;
  border-radius: 10px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  box-sizing: border-box;
}

.middle-column {
  flex: 1;
  border-radius: 12px;
  background: #fff;
  padding: 20px;
  box-sizing: border-box;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  /* 去掉 overflow:hidden 或者 设置为 visible */
  overflow: hidden;
}

.section-title {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 16px;
  color: #333;
  border-left: 4px solid #409EFF;
  padding-left: 8px;
}

.blog-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.blog-item {
  margin-bottom: 12px;
  cursor: pointer;
  transition: transform 0.2s ease;
}

.blog-item:hover {
  transform: translateY(-2px);
}

.blog-card {
  background: #fff;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.03);
}

.blog-title {
  display: block;
  font-weight: 600;
  font-size: 15px;
  color: #333;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.blog-title:hover {
  color: #409EFF;
}


.blog-author {
  font-size: 12px;
  color: #999;
  margin: 4px 0;
}

.blog-meta {
  font-size: 12px;
  color: #666;
  display: flex;
  gap: 12px;
}

.download-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.download-item {
  margin-bottom: 12px;
  cursor: pointer;
  transition: transform 0.2s ease;
}

.download-item:hover {
  transform: translateY(-2px);
}

.download-card {
  background: #fff;
  padding: 12px 14px;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.03);
}

.download-title {
  font-weight: 600;
  font-size: 15px;
  color: #333;
  margin-bottom: 6px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.download-meta {
  font-size: 12px;
  color: #666;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* 优质作者 */
.author-card {
  margin-bottom: 12px; /* 原来是15px */
  background-color: #f9f9f9;
  border-radius: 10px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  cursor: pointer;
}

.author-card:hover {
  transform: scale(1.03);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.author-box {
  display: flex;
  align-items: center;
  gap: 10px; /* 原来是12px */
  margin-bottom: 0; /* ← 去掉多余的 margin-bottom */
}

.author-avatar {
  flex-shrink: 0;
}

.author-meta .name {
  font-weight: bold;
  font-size: 16px;
}

.stats-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #606266;
}

.title-avatar {
  margin-right: 10px;
  vertical-align: middle;
}

@media (max-width: 768px) {
  .three-column-layout {
    flex-direction: column;
  }

  .left-column,
  .middle-column,
  .right-column {
    width: 100%;
  }

  .top-action-bar {
    justify-content: center;
    gap: 10px;
  }
}
</style>
