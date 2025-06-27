<template>
  <div class="community-page">
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
    <div class="main-layout">
      <!-- 文章列表 -->
      <div class="article-section" ref="scrollContainer" @scroll="handleScroll">
        <h2 class="section-title">📚 社区文章</h2>
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
              @click.stop="showFullArticle(article)"
          >
            阅读全文 →
          </el-button>
          <h3 class="article-title">{{ article.BLOG_TITLE }}</h3>
          <div class="article-meta">
            <span>作者：{{ article.USERNAME }}</span>
            <span>时间：{{ pubFormatDate(article.CREATE_TIME) }}</span>
          </div>
          <p class="article-content" v-html="article.MAINTEXT"></p>
        </el-card>
        <el-empty v-if="!articles.length && !loading" description="暂无内容"/>
        <div v-if="loading" class="loading-text">加载中...</div>
        <div v-if="noMore" class="end-text">没有更多文章了</div>
      </div>
      <!-- 优质作者栏 -->
      <div class="author-section">
        <div class="sticky-box">
          <h3 class="section-title">🌟 优质作者</h3>
          <el-card
              v-for="(author, index) in topAuthors"
              :key="index"
              class="author-card"
              @click="hotAuthorClick(author)"
          >
            <div class="author-box">
              <el-avatar
                  :src="author.AVATAR"
                  size="large"
                  class="author-avatar"
                  alt="用户头像"
              >
                {{ author.USERNAME?.charAt(0) }}
              </el-avatar>
              <div class="author-meta">
                <div class="name">{{ author.USERNAME }}</div>
                <div class="stats-row">
                  <span>📝 文章:{{ author.ARTICLE_COUNT }}</span>
                  <span>⬆️ 上传:{{ author.UPLOAD_COUNT }}</span>
                  <span>💬 评论:{{ author.COMMENT_COUNT }}</span>
                </div>
              </div>
            </div>
          </el-card>
        </div>
      </div>
    </div>
  </div>
  <!-- 当前组件的 template 中加上： -->
  <el-dialog
      v-model="showDialog"
      width="80%"
      destroy-on-close
      top="4vh"
      @close="onDialogClose"
  >
    <ContentAndComment :blogId="selectedBlogId"/>
  </el-dialog>
</template>

<script setup>
import {ref, onMounted} from 'vue'
import debounce from 'lodash/debounce'
import {sendAxiosRequest,encrypt,pubFormatDate} from "@/utils/common.js";
import ContentAndComment from "@/views/detail/blog/ContentAndComment.vue";
import {useRoute, useRouter} from 'vue-router'

const route = useRoute()
const router = useRouter()

const showDialog = ref(false);
const selectedBlogId = ref("");
//搜索主键
const searchKeyword = ref('')

const articles = ref([])
const page = ref(1)
const pageSize = 5
const loading = ref(false)
const noMore = ref(false)

const scrollContainer = ref(null)

//优质作者
const topAuthors = ref([])

async function setTopAuthors() {
  if (window.top["topAuthors"])
    topAuthors.value = window.top["topAuthors"];
  else {
    const result = await sendAxiosRequest("/blog-api/home/getHigAuthor");
    window.top["topAuthors"] = result.result;
    topAuthors.value = result.result;
  }
}

setTopAuthors();

const hotAuthorClick = (author)=>{

  const routeUrl = router.resolve({name: 'personInfomation', query: {c: encrypt(author.USERCODE)}}).href;
  window.open(routeUrl, "showPersonInfomation");
}

const fetchArticles = async () => {
  if (loading.value || noMore.value) return
  loading.value = true
  try {
    const res = await sendAxiosRequest("/blog-api/blog/getAllBlog", {
      page: page.value,
      pageSize,
      keyword: searchKeyword.value
    }, "get");

    const newData = res.result.data;
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

const onDialogClose = () => {
  showDialog.value = false
  selectedBlogId.value = ''
  router.replace({query: {}})
}

const showFullArticle = (article) => {
  router.replace({query: {g: article.GUID}});
  selectedBlogId.value = article.GUID;
  showDialog.value = true;
}
const handleScroll = () => {
  const el = scrollContainer.value
  if (el.scrollTop + el.clientHeight >= el.scrollHeight - 10) {
    fetchArticles()
  }
}

const resetAndLoad = () => {
  page.value = 1
  noMore.value = false
  articles.value = []
  fetchArticles()
}

// 添加防抖搜索
const debouncedSearch = debounce(() => {
  resetAndLoad()
}, 500)

onMounted(() => {
  fetchArticles()
  //查看路由参数,如果存在则自动弹出文章
  const blogId = route.query.g
  if (blogId) {
    selectedBlogId.value = blogId
    showDialog.value = true
  }
})
</script>

<style scoped>
.community-page {
  max-width: 1500px;
  margin: 20px auto;
  padding: 0 20px;
}

.search-bar {

  margin-bottom: 20px;
}

.main-layout {
  display: flex;
  gap: 20px;
  min-height: 600px;
}

/* 左侧文章区：可滚动 */
.article-section {
  flex: 3;
  max-height: 80vh;
  overflow-y: auto;
  padding-right: 10px;
}

/* 右侧作者栏：固定位置 */
.author-section {
  flex: 1;
}

.sticky-box {
  position: sticky;
  top: 0;
}

/* 样式优化 */
.section-title {
  font-size: 22px;
  margin-bottom: 15px;
  color: #409EFF;
}

.read-more-btn {
  position: absolute;
  top: 10px;
  right: 12px;
  font-size: 14px;
  color: #409EFF;
  z-index: 1;
  background: white; /* 避免覆盖文字时透明 */
}


.article-card {
  position: relative;
  margin-bottom: 20px;
  height: 200px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  overflow: hidden;
}


.article-content {
  font-size: 15px;
  color: #606266;

  /* 多行截断 */
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 3; /* 显示最多 3 行 */
  -webkit-box-orient: vertical;
}

.article-title {
  font-size: 18px;
  margin: 30px 0 8px 0; /* 增加顶部留白 */
  color: #303133;
}

.article-meta {
  font-size: 13px;
  color: #909399;
  margin-bottom: 10px;
  display: flex;
  gap: 20px;
}

.author-card {
  margin-bottom: 15px;
  padding: 12px;
  background-color: #f9f9f9;
  border-radius: 10px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  cursor: pointer;
}

.author-card:hover {
  transform: scale(1.03); /* 稍微放大 */
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15); /* 添加阴影 */
}


.author-meta .name {
  font-weight: bold;
  font-size: 16px;
  margin-bottom: 8px;
  color: #303133;
}

.stats-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #606266;
}



.author-meta .stats div {
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
}

.author-meta .stats {
  font-size: 13px;
  color: #909399;
}

.loading-text,
.end-text {
  text-align: center;
  color: #909399;
  margin: 20px 0;
}

.author-box {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.author-avatar {
  flex-shrink: 0;
}

.author-meta .name {
  font-weight: bold;
  font-size: 16px;
}

.stats-row span {
  margin-right: 10px;
  color: #666;
  font-size: 14px;
}

</style>
