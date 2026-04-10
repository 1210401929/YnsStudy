<template>
  <div class="rss-page-wrapper">
    <div class="rss-banner">
      <div class="banner-content">
        <h1><el-icon><Compass /></el-icon> RSS Feed 订阅预览</h1>
        <p>ynsstudy.cn 的最新动态，支持通过 RSS 阅读器订阅</p>
      </div>
    </div>

    <div class="rss-container">
      <el-card class="main-card" shadow="always">
        <template #header>
          <div class="card-header">
            <div class="header-left">
              <span class="feed-badge">RSS 2.0</span>
              <span class="article-count" v-if="!loading">共 {{ rssList.length }} 篇文章</span>
            </div>

            <div class="header-right">
              <el-button-group class="custom-btn-group">
                <el-button type="primary" @click="copyRssUrl">
                  <el-icon><CopyDocument /></el-icon>
                  <span>复制订阅地址</span>
                </el-button>
                <el-button @click="fetchRss" :loading="loading">
                  <el-icon v-if="!loading"><Refresh /></el-icon>
                  <span>刷新</span>
                </el-button>
              </el-button-group>
            </div>
          </div>
        </template>

        <el-skeleton :loading="loading" animated :count="3">
          <template #template>
            <div style="padding: 20px">
              <el-skeleton-item variant="h3" style="width: 50%" />
              <el-skeleton-item variant="text" style="margin-top: 10px;" />
              <el-skeleton-item variant="text" style="width: 30%; margin-top: 10px;" />
            </div>
          </template>

          <el-timeline v-if="rssList.length > 0" class="custom-timeline">
            <el-timeline-item
                v-for="(item, index) in rssList"
                :key="index"
                :timestamp="item.pubDate"
                placement="top"
                type="primary"
                :hollow="true"
            >
              <el-card class="item-card" shadow="hover">
                <div class="item-content">
                  <h3 class="blog-title">
                    <a :href="item.link" target="_blank">{{ item.title }}</a>
                  </h3>
                  <p class="blog-desc">{{ item.description }}</p>
                  <div class="item-footer">
                    <el-tag size="small" effect="light" class="time-tag">
                      <el-icon><Calendar /></el-icon> {{ item.pubDate }}
                    </el-tag>
                    <el-link :href="item.link" type="primary" :underline="false" target="_blank">
                      继续阅读 <el-icon><ArrowRight /></el-icon>
                    </el-link>
                  </div>
                </div>
              </el-card>
            </el-timeline-item>
          </el-timeline>

          <el-empty v-else :image-size="200" description="订阅源暂时空空如也" />
        </el-skeleton>
      </el-card>

      <div class="rss-footer-tips">
        <p>提示：你可以将订阅地址粘贴至 Feedly, Inoreader 或微信阅读等工具中</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { Compass, ArrowRight, CopyDocument, Refresh, Calendar } from '@element-plus/icons-vue';
import { getSendAxiosUrl } from "@/utils/common.js";

const rssList = ref([]);
const loading = ref(true);
const rssUrl = window.location.origin + getSendAxiosUrl('/blog-api/home/rss.xml');

const fetchRss = async () => {
  loading.value = true;
  try {
    const response = await fetch(getSendAxiosUrl('/blog-api/home/rss.xml'));
    const xmlText = await response.text();
    const parser = new DOMParser();
    const xmlDoc = parser.parseFromString(xmlText, "text/xml");
    const items = xmlDoc.getElementsByTagName("item");

    rssList.value = Array.from(items).map(item => ({
      title: item.getElementsByTagName("title")[0]?.textContent || '无标题',
      link: item.getElementsByTagName("link")[0]?.textContent || '#',
      pubDate: item.getElementsByTagName("pubDate")[0]?.textContent || '',
      description: item.getElementsByTagName("description")[0]?.textContent || '暂无摘要'
    }));
  } catch (error) {
    console.error('RSS Fetch Error:', error);
    ElMessage.error('RSS 数据抓取失败');
  } finally {
    loading.value = false;
  }
};

const copyRssUrl = () => {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(rssUrl);
    ElMessage.success('订阅链接已成功复制');
  } else {
    ElMessage.warning('当前环境不支持自动复制，请手动复制浏览器地址');
  }
};

onMounted(fetchRss);
</script>

<style scoped>
/* 页面整体背景 */
.rss-page-wrapper {
  min-height: 100vh;
  background-color: #f5f7fa;
  padding-bottom: 50px;
}

/* 顶部 Banner */
.rss-banner {
  background: linear-gradient(135deg, #409eff 0%, #3a8ee6 100%);
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  text-align: center;
}
.banner-content h1 {
  font-size: 2.2rem;
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}
.banner-content p {
  opacity: 0.9;
  font-size: 1.1rem;
}

/* 容器布局 */
.rss-container {
  max-width: 900px;
  margin: -40px auto 0;
  padding: 0 20px;
}

.main-card {
  border-radius: 12px;
  border: none;
}

/* 头部对齐优化 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 0;
}

.custom-btn-group {
  display: inline-flex;
  align-items: stretch; /* 核心：让子元素高度拉平 */
}

/* 强制按钮内部布局一致 */
.custom-btn-group :deep(.el-button) {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 32px; /* 固定高度确保一致 */
  padding: 8px 15px;
}

.custom-btn-group :deep(.el-icon) {
  margin-right: 4px;
}

.feed-badge {
  background: #f0f9eb;
  color: #67c23a;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: bold;
  margin-right: 12px;
}

.article-count {
  color: #909399;
  font-size: 14px;
}

/* 时间线样式 */
.custom-timeline {
  padding: 20px 10px;
}

.item-card {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  transition: all 0.3s;
  margin-bottom: 5px;
}

.item-card:hover {
  transform: translateY(-3px);
  border-color: #409eff;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.blog-title {
  margin: 0 0 10px 0;
  font-size: 1.25rem;
}

.blog-title a {
  color: #2c3e50;
  text-decoration: none;
}

.blog-title a:hover {
  color: #409eff;
}

.blog-desc {
  color: #5e6d82;
  font-size: 14px;
  line-height: 1.8;
  background: #f8f9fb;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 15px;
  /* 限制高度并显示省略号（可选） */
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.item-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.rss-footer-tips {
  text-align: center;
  margin-top: 30px;
  color: #909399;
  font-size: 13px;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .banner-content h1 { font-size: 1.6rem; }
  .rss-container { margin-top: -20px; }
  .card-header { flex-direction: column; gap: 15px; align-items: flex-start; }
  .header-right { width: 100%; }
  .custom-btn-group { width: 100%; display: flex; }
  .custom-btn-group :deep(.el-button) { flex: 1; }
}
</style>