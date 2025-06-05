<template>
  <div class="hot-blogs">
    <h3 class="section-title">🔥 热门文章推荐</h3>
    <ul class="blog-list">
      <li v-for="blog in homeStore.homeData.hotBlogData" :key="blog.id" class="blog-item" @click="hotBlogCLick(blog)">
        <div class="blog-card">
          <a role="link" class="blog-title">{{ blog.BLOG_TITLE }}</a>
          <div class="blog-author">by {{ blog.USERNAME }}</div>
          <div class="blog-meta">
            <span>👁 {{ blog.VIEW_PAGE }}</span>
            <span>💬 {{ blog.COMMENT_COUNT }}</span>
          </div>
        </div>
      </li>
    </ul>
  </div>
</template>

<script setup>
import {useHomeStore} from "@/stores/detail/home.js";
import {ref} from "vue";
import {useRouter} from "vue-router";

const router = useRouter();

const homeStore = useHomeStore();
homeStore.initHomeData();

const showDialog = ref(false);
const selectedBlogId = ref("");

function hotBlogCLick(blog) {
  router.push({name:"Community",query:{blogId:blog.GUID}})
}
</script>

<style scoped>
.hot-blogs {
  width: 280px;
  background: #fafafa;
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 12px 16px;
  font-size: 14px;
  box-sizing: border-box;
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
  padding: 12px 14px;
  border-radius: 6px;
  border: 1px solid #e0e0e0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  transition: box-shadow 0.3s ease;
}

.blog-item:hover .blog-card {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.blog-title {
  display: inline-block;
  color: #333;
  font-weight: 600;
  text-decoration: none;
  line-height: 1.4;
  cursor: pointer;
  font-size: 15px;

  /* 🌟 新增：文字过长省略 */
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  max-width: 100%;
  display: block;
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
  flex-wrap: wrap;
}

/* ✅ 响应式支持移动端 */
@media (max-width: 768px) {
  .hot-blogs {
    width: 100%;
    margin-top: 20px;
    font-size: 15px;
    padding: 16px;
  }

  .section-title {
    font-size: 18px;
  }

  .blog-meta {
    font-size: 13px;
  }

  .blog-author {
    font-size: 13px;
  }
}
</style>
