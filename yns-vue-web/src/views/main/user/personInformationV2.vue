<template>
  <div class="person-info" :style="currentBgStyle">
    <WelcomeOverlay
        :visible="showWelcome"
        :is-self="isSelf"
        :user="user"
        @enter="enterHomepage"
    />

    <BackgroundAndMusic
        ref="bgMusicComponentRef"
        :is-self="isSelf"
        :user-name="user.name"
        :init-bg-image="serverBgImage"
        :init-bg-audio="serverBgAudio"
        @update-bg-style="handleBgStyleUpdate"
    />

    <div class="inner-container">
      <div class="content-columns">

        <div class="person-left-wrapper">
          <UserInfo
              :user="user"
              :target-user-code="targetUserCode"
              @blog-click="blogMainClick"
          />
        </div>

        <div class="main-content">
          <div class="toolbar-enhanced">
            <div class="stats-badge" v-if="blogs.length > 0">
              <el-icon><Histogram /></el-icon>
              <span>共发表 {{ blogs.length }} 条动态</span>
            </div>
            <el-button type="primary" round class="jump-btn" @click="scrollToFiles">
              <el-icon><FolderOpened /></el-icon> 快速跳转至文件
            </el-button>
          </div>

          <div class="section-container">
            <div class="section-header-fancy">
              <div class="title-group">
                <span class="icon-cube">📚</span>
                <h3 class="section-title">发表内容</h3>
              </div>
              <div class="filter-group">
                <el-switch
                    v-model="onlyArticle"
                    inline-prompt
                    active-text="文章"
                    inactive-text="全部"
                    style="--el-switch-on-color: #67c23a;"
                />
              </div>
            </div>

            <el-timeline v-if="showBlogs.length > 0" class="custom-timeline">
              <transition-group name="list-fade">
                <el-timeline-item
                    v-for="(blog) in showBlogs"
                    :key="blog.GUID"
                    placement="top"
                    :hollow="true"
                >
                  <template #dot>
                    <div class="timeline-dot" :class="blog.TYPE"></div>
                  </template>

                  <div class="timestamp-label">
                    {{ blog.TYPE === 'blog' ? '📝 文章发布于 ' : '👥 社区发起于 ' }}
                    {{ formatDate(blog.CREATE_TIME) }}
                  </div>

                  <el-card shadow="never" class="blog-card-refined" @click="blogMainClick(blog)">
                    <div class="blog-card-header">
                      <h4>{{ blog.BLOG_TITLE }}</h4>
                    </div>
                    <p class="blog-summary" v-html="stripImages(blog.MAINTEXT)"></p>
                    <div class="blog-footer">
                      <span class="read-more">阅读详情 <el-icon><ArrowRight /></el-icon></span>
                    </div>
                  </el-card>
                </el-timeline-item>
              </transition-group>
            </el-timeline>

            <div v-if="!showBlogs.length && !loading" class="empty-wrapper">
              <el-empty description="暂无内容发布" :image-size="120" />
            </div>

            <div class="pagination-trigger">
              <el-button
                  v-if="!noMore && !loading"
                  type="primary"
                  round
                  plain
                  @click="fetchArticles(null)"
              >
                探索更多内容
              </el-button>
              <div v-if="loading" class="loading-state">
                <el-icon class="is-loading"><Loading /></el-icon> 正在努力加载...
              </div>
              <p v-if="noMore && showBlogs.length > 0" class="end-pro-text">—— 已触达知识的边界 ——</p>
            </div>
          </div>

          <el-card class="section-card-refined file-section" ref="fileSection">
            <template #header>
              <div class="card-header-fancy">
                <div class="title-group">
                  <span class="icon-cube">📁</span>
                  <h3 class="section-title">上传的文件</h3>
                </div>
              </div>
            </template>
            <el-table
                v-if="files.length > 0"
                :data="files"
                class="modern-table"
                style="width: 100%"
            >
              <el-table-column prop="ORIGINALFILENAME" label="文件名" min-width="180" />
              <el-table-column prop="DOWNNUM" label="下载" width="80" align="center" />
              <el-table-column label="上传日期" width="120">
                <template #default="scope">{{ formatDate(scope.row.CREATE_TIME).split(' ')[0] }}</template>
              </el-table-column>
              <el-table-column label="操作" width="100" fixed="right">
                <template #default="scope">
                  <el-button type="primary" link @click="downloadFile(scope.row)">下载</el-button>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-else description="暂无共享文件" :image-size="80" />
          </el-card>
        </div>

        <div class="right-sidebar">
          <div class="align-spacer"></div>

          <el-card class="sidebar-card recent-articles">
            <template #header>
              <div class="sidebar-card-title">
                <el-icon><TrendCharts /></el-icon> <span>最近文章</span>
              </div>
            </template>
            <div class="recent-articles-placeholder">
              <el-skeleton :rows="3" animated />
            </div>
          </el-card>

          <div class="sticky-category">
            <BlogSidebar
                :view_title="'内容归档'"
                :user-code="targetUserCode"
                @click-blog="blogMainClick"
                v-model:selected-index="selectedCategory"
            />
          </div>
        </div>

      </div>
    </div>

    <el-dialog v-model="showDialog" width="70%" destroy-on-close top="6vh" custom-class="glass-dialog">
      <ContentAndComment :blogId="selectedBlogId"/>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, defineAsyncComponent } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute } from 'vue-router'
import {
  Histogram, FolderOpened, ArrowRight, Loading, TrendCharts
} from '@element-plus/icons-vue'
import {
  pubFormatDate, decrypt, sendAxiosRequest, stripImages,
  downloadFileByUrl, getUserInfoByCode
} from '@/utils/common.js'

import BackgroundAndMusic from "@/components/detail/personInformation/BackgroundAndMusic.vue";
import WelcomeOverlay from "@/components/detail/personInformation/WelcomeOverlay.vue";
import BlogSidebar from "@/components/detail/myblog/BlogSidebar.vue";
import UserInfo from "@/components/main/UserInfo.vue";

const ContentAndComment = defineAsyncComponent(() => import('@/views/detail/blog/ContentAndComment.vue'))
import { useUserStore } from '@/stores/main/user.js'

const userStore = useUserStore()
userStore.initFromLocal()

const route = useRoute()
const user = ref({})
const selectedCategory = ref('')
const targetUserCode = ref('')

// 初始化逻辑
try {
  const raw = route.params.u;
  targetUserCode.value = raw ? decrypt(raw) : '';
} catch (e) {
  targetUserCode.value = '';
}

const isSelf = computed(() =>
    !!userStore.userBean.code &&
    !!targetUserCode.value &&
    targetUserCode.value === userStore.userBean.code
);

const bgMusicComponentRef = ref(null);
const serverBgImage = ref('');
const serverBgAudio = ref('');
const currentBgStyle = ref({});
const handleBgStyleUpdate = (style) => currentBgStyle.value = style;

const page = ref(1)
const pageSize = 5 // 增加单页展示数量
const loading = ref(false)
const noMore = ref(false)
const blogs = ref([])
const files = ref([])
const onlyArticle = ref(false)
const showDialog = ref(false)
const selectedBlogId = ref('')

const fetchArticles = async (userCode) => {
  if (loading.value || noMore.value) return;
  if (!userCode) userCode = targetUserCode.value;
  loading.value = true
  try {
    const res = await sendAxiosRequest('/blog-api/userInformation/getBlogAndCommunityByUserCode', {
      userCode,
      page: page.value,
      pageSize,
      keyword: ""
    })
    const newData = res.result.data
    if (newData.length < pageSize) noMore.value = true
    blogs.value.push(...newData)
    page.value++
  } catch (e) {
    console.error('获取内容失败', e)
  } finally {
    loading.value = false
  }
}

async function getUserInfo2Data() {
  const userCode = targetUserCode.value;
  if (!userCode) return;

  // 并行请求提高速度
  getUserInfoByCode(userCode).then(result => {
    if (result && !result.isError) {
      user.value = result.result;
      document.title = user.value.name + "的主页";
    }
  });

  sendAxiosRequest('/blog-api/userInformation/getResourceByUserCode', {userCode}).then(res => {
    if (res && !res.isError) files.value = res.result;
  });

  sendAxiosRequest("/blog-api/userInformation/getPersonInfo", {userCode}).then(res => {
    if (res && !res.isError) {
      const data = res.result[0] || {};
      serverBgAudio.value = data.BGMUSICURL || "";
      serverBgImage.value = data.BGIMAGEURL || "";
    }
  });

  fetchArticles(userCode);
}

const showBlogs = computed(() => {
  if (!onlyArticle.value) return blogs.value;
  return blogs.value.filter(item => item.TYPE === "blog");
})

function blogMainClick(blog) {
  if (blog.TYPE === "community") {
    ElMessage.info('社区动态暂不支持详情查看')
    return
  }
  selectedBlogId.value = blog.GUID
  showDialog.value = true
}

const onDialogClose = () => {
  selectedBlogId.value = ''
  showDialog.value = false
}

const formatDate = (dateStr) => pubFormatDate(dateStr)

const downloadFile = (file) => {
  ElMessage.success(`准备下载：${file.ORIGINALFILENAME}`)
  downloadFileByUrl(file.FILEVIEWURL, file.ORIGINALFILENAME)
}

const fileSection = ref(null)
const scrollToFiles = () => {
  if (fileSection.value?.$el) {
    fileSection.value.$el.scrollIntoView({behavior: 'smooth'})
  }
}

const showWelcome = ref(true);
const enterHomepage = () => {
  showWelcome.value = false;
  if (bgMusicComponentRef.value) bgMusicComponentRef.value.playMusicForce();
};

onMounted(() => getUserInfo2Data());
</script>

<style scoped>
/* 基础背景与布局 */
.person-info {
  width: 100%;
  height: 100vh;
  overflow-y: auto;
  padding: 40px 0;
  box-sizing: border-box;
  background: url('/picture/user/default.webp') no-repeat center center fixed;
  background-size: cover;
  transition: background-image 0.5s ease-in-out;
}

.inner-container {
  max-width: 1300px;
  margin: 0 auto;
  padding: 0 20px;
}

.content-columns {
  display: flex;
  gap: 32px;
  align-items: flex-start;
}

/* 左侧栏 */
.person-left-wrapper {
  position: sticky;
  top: 40px;
  width: 280px;
  flex-shrink: 0;
  margin-top: 52px; /* 对应右侧对其高度 */
}

/* ================= 中间区域美化 ================= */
.main-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* 顶部工具条 */
.toolbar-enhanced {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 4px;
}

.stats-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 16px;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 20px;
  color: #fff;
  font-size: 13px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.section-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-header-fancy {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 10px;
  margin-bottom: 8px;
}

.title-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.icon-cube {
  font-size: 20px;
  filter: drop-shadow(0 2px 4px rgba(0,0,0,0.3));
}

.section-title {
  font-size: 20px;
  font-weight: 700;
  color: #fff;
  text-shadow: 0 2px 4px rgba(0,0,0,0.5);
  margin: 0;
}

/* 文章卡片重构 */
.blog-card-refined {
  border: none !important;
  background: rgba(255, 255, 255, 0.75) !important;
  backdrop-filter: blur(12px);
  border-radius: 16px !important;
  border: 1px solid rgba(255, 255, 255, 0.4) !important;
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  cursor: pointer;
  margin-bottom: 4px;
}

.blog-card-refined:hover {
  transform: translateY(-6px) scale(1.01);
  background: rgba(255, 255, 255, 0.95) !important;
  box-shadow: 0 15px 35px rgba(0, 0, 0, 0.15) !important;
}

.blog-card-header h4 {
  margin: 0;
  font-size: 1.25rem;
  color: #1a1a1a;
  letter-spacing: -0.02em;
}

.blog-summary {
  margin: 14px 0;
  font-size: 14px;
  line-height: 1.7;
  color: #4c4c4c;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.blog-footer {
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid rgba(0,0,0,0.05);
  padding-top: 12px;
}

.read-more {
  font-size: 13px;
  font-weight: 600;
  color: #409eff;
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 时间轴 */
.custom-timeline {
  margin-top: 10px;
}

.timestamp-label {
  font-size: 12px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  margin-bottom: 10px;
  text-shadow: 0 1px 2px rgba(0,0,0,0.3);
}

.timeline-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 3px solid #fff;
  box-sizing: border-box;
}
.timeline-dot.blog { background: #409eff; box-shadow: 0 0 10px rgba(64,158,255,0.6); }
.timeline-dot.community { background: #e6a23c; box-shadow: 0 0 10px rgba(230,162,60,0.6); }

/* 加载更多 */
.pagination-trigger {
  margin: 20px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.loading-state {
  color: #fff;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.end-pro-text {
  color: rgba(255,255,255,0.6);
  font-size: 13px;
  letter-spacing: 2px;
}

/* 文件列表卡片 */
.section-card-refined {
  background: rgba(255, 255, 255, 0.8) !important;
  backdrop-filter: blur(12px);
  border-radius: 20px !important;
  border: 1px solid rgba(255, 255, 255, 0.4) !important;
}

.modern-table {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-header-bg-color: rgba(0, 0, 0, 0.04);
  font-size: 14px;
}

/* 侧边栏 */
.sidebar-card {
  background: rgba(255, 255, 255, 0.85) !important;
  backdrop-filter: blur(10px);
  border-radius: 16px !important;
  margin-bottom: 24px;
}

.sidebar-card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: bold;
  color: #333;
}

.align-spacer { height: 52px; }

/* 列表动画 */
.list-fade-enter-active,
.list-fade-leave-active {
  transition: all 0.5s ease;
}
.list-fade-enter-from {
  opacity: 0;
  transform: translateX(30px);
}

/* 响应式 */
@media (max-width: 1100px) {
  .content-columns { flex-direction: column; }
  .person-left-wrapper, .right-sidebar { width: 100%; position: static; margin-top: 0; }
  .align-spacer { display: none; }
}
</style>