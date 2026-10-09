<template>
  <div class="rss-page-wrapper j-desk">
    <div class="rss-banner">
      <div class="banner-content">
        <span class="j-tape banner-tape"></span>
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
import {useHead} from '@vueuse/head';

const rssList = ref([]);
const loading = ref(true);
const rssUrl = `${window.location.origin}/rss.xml`;
const rssFetchUrl = import.meta.env.DEV
    ? getSendAxiosUrl('/blog-api/home/rss.xml')
    : '/rss.xml';

// 这是给访客使用的订阅预览页，不应与每篇文章争夺搜索结果。
useHead({
  title: 'RSS 订阅 - YnsStudy',
  meta: [
    {name: 'description', content: '订阅 YnsStudy 的最新博客文章。'},
    {name: 'robots', content: 'noindex,follow'}
  ],
  link: [{rel: 'alternate', type: 'application/rss+xml', title: 'YnsStudy RSS', href: rssUrl}]
});

const fetchRss = async () => {
  loading.value = true;
  try {
    const response = await fetch(rssFetchUrl);
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
.rss-page-wrapper {
  min-height: 100vh;
  padding: 40px 0 56px;
  color: var(--j-ink);
}

/* 顶部：贴胶带的标签 */
.rss-banner {
  display: flex;
  justify-content: center;
  padding: 0 20px;
}

.banner-content {
  position: relative;
  padding: 22px 34px 18px;
  text-align: center;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
  transform: rotate(-0.6deg);
}

.banner-tape {
  top: -11px;
  left: 50%;
  margin-left: -42px;
  transform: rotate(-3deg);
}

.banner-content h1 {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin: 0;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 28px;
}

.banner-content h1 .el-icon {
  color: #e38b2d;
}

.banner-content p {
  margin: 8px 0 0;
  font-size: 14px;
  color: var(--j-ink-soft);
}

.rss-container {
  max-width: 900px;
  margin: 36px auto 0;
  padding: 0 20px;
}

.main-card {
  border: 1px solid var(--j-rule);
  border-radius: 2px;
  background: var(--j-paper);
}

.main-card :deep(.el-card__header) {
  border-bottom: 1px dashed var(--j-rule-strong);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 2px 0;
}

.custom-btn-group {
  display: inline-flex;
  align-items: stretch;
}

.custom-btn-group :deep(.el-button) {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 32px;
  padding: 8px 15px;
}

.custom-btn-group :deep(.el-icon) {
  margin-right: 4px;
}

/* RSS 2.0：橙色小印章 */
.feed-badge {
  display: inline-block;
  margin-right: 12px;
  padding: 2px 8px;
  border: 1.5px solid #e38b2d;
  border-radius: 3px;
  font-size: 12px;
  font-weight: bold;
  color: #c9711a;
  transform: rotate(-4deg);
}

.article-count {
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-muted);
}

/* 时间线 */
.custom-timeline {
  padding: 20px 10px 4px;
}

.custom-timeline :deep(.el-timeline-item__tail) {
  border-left: 2px dashed var(--j-rule-strong);
}

.custom-timeline :deep(.el-timeline-item__node) {
  background: var(--j-paper);
  border-color: var(--j-margin-red);
}

.custom-timeline :deep(.el-timeline-item__timestamp) {
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-ink-soft);
}

.item-card {
  margin-bottom: 5px;
  border: 1px solid var(--j-rule);
  border-radius: 2px;
  background: #fffefb;
  transition: transform 0.2s ease;
}

.item-card:hover {
  transform: translateX(4px);
}

.blog-title {
  margin: 0 0 10px;
  font-size: 18px;
}

.blog-title a {
  color: var(--j-ink);
  text-decoration: none;
  background-image: linear-gradient(transparent 58%, var(--j-highlight) 58%, var(--j-highlight) 92%, transparent 92%);
  background-size: 0 100%;
  background-repeat: no-repeat;
  transition: background-size 0.3s ease;
}

.item-card:hover .blog-title a {
  background-size: 100% 100%;
}

.blog-desc {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin: 0 0 14px;
  font-size: 14px;
  line-height: 1.8;
  color: var(--j-ink-soft);
}

.item-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 10px;
  border-top: 1px dashed var(--j-rule);
}

.time-tag {
  border: none;
  background: #f3eee3;
  color: var(--j-ink-soft);
}

.rss-footer-tips {
  margin-top: 28px;
  text-align: center;
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-muted);
}

@media (max-width: 768px) {
  .rss-page-wrapper { padding-top: 28px; }
  .banner-content { transform: none; padding: 20px 18px 16px; }
  .banner-content h1 { font-size: 22px; }
  .card-header { flex-direction: column; gap: 15px; align-items: flex-start; }
  .header-right { width: 100%; }
  .custom-btn-group { width: 100%; display: flex; }
  .custom-btn-group :deep(.el-button) { flex: 1; }
}
</style>
