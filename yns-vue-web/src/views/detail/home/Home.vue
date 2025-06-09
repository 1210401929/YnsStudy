<template>
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
    <!-- 中间预留区域 -->
    <div class="middle-column">
      <div class="placeholder"> 未开发区域</div>
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
              <span>🧑 {{ file.USERNAME }}</span>
              |
              <span style="color:#04c279">下载次数: {{ file.DOWNNUM }}</span>
            </div>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import {useHomeStore} from "@/stores/detail/home.js";
import {useRouter} from "vue-router";

const router = useRouter();
const homeStore = useHomeStore();
homeStore.initHomeData();

function hotBlogClick(blog) {
  router.push({name: "Community", query: {g: blog.GUID}});
}

function downloadClick(file) {
  router.push({name: "Resources", query: {g: file.GUID}});
}
</script>

<style scoped>
.three-column-layout {
  display: flex;
  gap: 20px;
  width: 100%;
  box-sizing: border-box;
  justify-content: space-between;
  align-items: flex-start; /* 避免等高 */
}

/* 左右列固定宽度，中间自适应 */
.left-column,
.right-column {
  width: 280px;
  background: #fafafa;
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 12px 16px;
  font-size: 14px;
  box-sizing: border-box;
}

.middle-column {
  flex: 1;
  min-height: 300px;
  border: 1px dashed #ccc;
  border-radius: 8px;
  background: #fff;
  padding: 20px;
  box-sizing: border-box;
}

.placeholder {
  color: #999;
  font-size: 16px;
  text-align: center;
  padding: 40px 0;
}

/* 通用标题样式 */
.section-title {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 16px;
  color: #333;
  border-left: 4px solid #409EFF;
  padding-left: 8px;
}

/* 热门文章样式 */
.blog-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.blog-item {
  margin-bottom: 12px;
  cursor: pointer;
}

.blog-card {
  background: #fff;
  padding: 12px;
  border-radius: 6px;
  border: 1px solid #e0e0e0;
}

.blog-title {
  font-weight: 600;
  font-size: 15px;
  color: #333;
  display: block;
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

/* 热门下载样式 */
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
  border-radius: 6px;
  border: 1px solid #e0e0e0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  transition: box-shadow 0.3s ease;
}

.download-item:hover .download-card {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
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

/* 响应式：移动端变为上下布局 */
@media (max-width: 768px) {
  .three-column-layout {
    flex-direction: column;
  }

  .left-column,
  .middle-column,
  .right-column {
    width: 100%;
  }
}
</style>
