<template>
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
                <span>💬 {{ blog.COMMENT_COUNT }}</span>
              </div>
            </div>
          </li>
        </ul>
      </div>

      <!-- 中间轮播图 -->
      <div class="middle-column">
        <!-- 精致顶部按钮区域 -->
        <div class="top-action-bar">
          <el-button class="action-button" @click="goToAdmin">👨‍💼 站长主页</el-button>
          <el-button class="action-button" @click="goToPublishBlog">📝 发布内容</el-button>
          <el-button class="action-button" @click="goToUpload">📤 上传资源</el-button>
        </div>

        <el-carousel
            height="300px"
            trigger="click"
            indicator-position="outside"
            :interval="4000"
            arrow="always"
        >
          <el-carousel-item v-for="(item, index) in carouselImages" :key="index">
            <div class="carousel-image-container">
              <img :src="item.url" alt="轮播图" class="carousel-image" @click="imageClick(item)" />
            </div>
          </el-carousel-item>
        </el-carousel>
      </div>

      <!-- 右侧：热门下载 -->
      <div class="right-column">
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
import { ref } from "vue";
import { useHomeStore } from "@/stores/detail/home.js";
import { useRouter } from "vue-router";
import {encrypt} from "@/utils/common.js";

const router = useRouter();
const homeStore = useHomeStore();
homeStore.initHomeData();

function hotBlogClick(blog) {
  router.push({ name: "Community", query: { g: blog.GUID } });
}
function downloadClick(file) {
  router.push({ name: "Resources", query: { g: file.GUID } });
}
function goToAdmin() {
  const routeUrl = router.resolve({name: 'personInfomation', query: {c: encrypt("yulei")}}).href;
  window.open(routeUrl, "showPersonInfomation");
}
function goToPublishBlog() {
  router.push({ name: "Blog" });
}
function goToUpload() {
  router.push({ name: "Resources" });
}

const carouselImages = ref([
  { url: "/picture/blog.png", routeName: "Blog" },
  { url: "/picture/resources.png", routeName: "Resources" },
  { url: "/picture/community.png", routeName: "Community" }
]);

const imageClick = (item) => {
  router.push({ name: item.routeName });
};
</script>

<style scoped>
/* 页面整体包装 */
.home-page-wrapper {
  width: 100%;
  box-sizing: border-box;
}

/* 顶部按钮栏：美观居中 */
.top-action-bar {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin: 20px auto;
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

/* 三栏主区域 */
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
  min-height: 300px;
  border-radius: 10px;
  background: #fff;
  padding: 20px;
  box-sizing: border-box;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

/* 轮播图样式 */
.carousel-image-container {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
  border-radius: 10px;
}

.carousel-image {
  height: 100%;
  object-fit: cover;
  border-radius: 10px;
  cursor: pointer;
  transition: transform 0.3s ease;
}
.carousel-image:hover {
  transform: scale(1.03);
}

/* 标题样式 */
.section-title {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 16px;
  color: #333;
  border-left: 4px solid #409EFF;
  padding-left: 8px;
}

/* 热门文章区域 */
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

/* 下载区域 */
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

/* 响应式优化 */
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
