<template>
  <!-- 公告横幅 -->
  <Announcement v-for="al in topAlert" :key="al.GUID" :TEXT="al.TEXT" :URL="al.URL" :URLNAME="al.URLNAME"/>

  <div class="community-page">
    <!-- 讨论区 -->
    <div class="section">
      <h3>🔨️ 讨论区</h3>
      <el-card
          v-for="(item, index) in displayedPosts"
          :key="index"
          class="feed-card"
          shadow="never"
      >
        <div v-if="item.ISTOP === '1'" class="ribbon-wrapper">
          <div class="ribbon">已置顶</div>
        </div>

        <div class="feed-header">
          <el-avatar
              :src="item.AVATAR"
              size="medium"
              class="author-avatar"
              alt="用户头像"
              @click="avatarClick(item)"
              title="查看发布者信息"
          >
            {{ item.USERNAME?.charAt(0) }}
          </el-avatar>
          <div class="author-info">
            <div class="author-name">{{ item.USERNAME || item.USERCODE }}</div>
            <div class="author-meta">{{ pubFormatDate(item.CREATE_TIME) }}</div>
          </div>
        </div>

        <!-- 内容展示部分 -->
        <div class="feed-body" v-if="item.isExpanded" v-html="item.TEXT"></div>
        <div class="feed-body" v-else>
          <!--不展开时,只显示两行内容-->
          <div v-html="item.TEXT.split('\n').slice(0, 2).join('\n')"></div>
        </div>
        <!-- 展开/收起帖子内容按钮 -->
        <div class="expand-btn-wrapper" v-if="item.TEXT.split('\n').length > 2">
          <el-button
              v-if="!item.isExpanded"
              type="primary"
              plain
              class="expand-btn"
              @click="expandPost(item)"
          >
            展开
          </el-button>
          <el-button
              v-if="item.isExpanded"
              type="primary"
              plain
              class="expand-btn"
              @click="collapsePost(item)"
          >
            收起
          </el-button>
        </div>
        <!-- 展开/收起按钮 -->
        <div class="feed-actions">
          <el-button text size="small" :icon="ChatDotSquare" @click="toggleComments(item)">
            {{ item.showComments ? '收起评论' : '评论' }}
          </el-button>
          <!-- 只有超级管理员有置顶权限 -->
          <el-button v-if="getCurrentUserAdminObject().adminLevel==='superAdmin'" text size="small" :icon="Star"
                     @click="setTopCommunity(item)">{{ item.ISTOP === "1" ? "取消置顶" : "置顶" }}
          </el-button>
          <!-- 允许删除逻辑:1.非置顶 是登录用户自己的,允许删除  2.非置顶 且非超级管理员的发帖,允许普通管理员删除  3.当前登录用户是超级管理员  -->
          <el-button v-if="getCurrentUserAdminObject().adminLevel==='superAdmin' || (item.ISTOP!=='1' && userStore.userBean.code === item.USERCODE)
              || (getCurrentUserAdminObject().isAdmin && item.ISTOP!=='1' && item.USERCODE!==adminUserCode)" text
                     size="small" :icon="Delete"
                     @click="deleteCommunity(item)">删除
          </el-button>
        </div>

        <!-- 评论部分 -->
        <transition name="fade">
          <div v-show="item.showComments" class="comment-section">
            <div
                v-for="(comment, i) in limitedComments(item)"
                :key="i"
                class="comment-item"
            >
              <el-avatar
                  :src="comment.AVATAR"
                  size="large"
                  class="author-avatar-comment"
                  @click="commentAvatarClick(comment)"
                  alt="评论用户头像"
              >
                {{ comment.USERNAME?.charAt(0) }}
              </el-avatar>
              <div class="comment-content">
                <div class="comment-header">
                  <span class="comment-author">{{ comment.USERNAME }}</span>
                  <el-button link size="small" @click="replyTarget = i">回复</el-button>
                </div>
                <div class="comment-text">{{ comment.TEXT }}</div>
                <div v-if="replyTarget === i" class="reply-box">
                  <el-input
                      v-model="comment.replyText"
                      size="small"
                      placeholder="写下你的回复..."
                      @keyup.enter="submitReply(item, comment)"
                  />
                  <el-button type="primary" size="small" @click="submitReply(item, comment)" style="margin-top: 5px">
                    回复
                  </el-button>
                </div>
                <div v-if="comment.children" class="children">
                  <div v-for="(r, j) in comment.children" :key="j" class="reply-item">
                    <el-avatar
                        :src="r.AVATAR"
                        size="large"
                        class="author-avatar-comment"
                        @click="commentAvatarClick(r)"
                        alt="评论用户头像"
                    >
                      {{ r.USERNAME?.charAt(0) }}
                    </el-avatar>
                    <span class="reply-author">{{ r.USERNAME }}：</span>
                    <span class="reply-text">{{ r.TEXT }}</span>
                  </div>
                </div>
              </div>
            </div>
            <div v-if="item.comments && item.comments.length > 3 && !item.showAllComments" class="show-more-comments">
              <el-button text @click="item.showAllComments = true">展示全部评论</el-button>
            </div>
            <div class="comment-input">
              <el-input
                  v-model="item.newComment"
                  placeholder="写下你的评论..."
                  size="small"
                  @keyup.enter="submitComment(item)"
                  clearable
              />
              <el-button type="primary" size="small" @click="submitComment(item)" style="margin-top: 6px; width: 100%">
                发表评论
              </el-button>
            </div>
          </div>
        </transition>
      </el-card>
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
      <div v-if="noMore" class="end-text">没有更多内容了</div>
    </div>

    <!-- 快速发帖 -->
    <el-card class="post-box" shadow="always">
      <h3>💬 快速发帖</h3>
      <el-input
          v-if="!showPreview"
          v-model="newPost"
          type="textarea"
          :rows="3"
          placeholder="说点什么吧? （支持 Markdown 语法）"
          class="post-textarea"
          :autosize="true"
      />
      <!-- Markdown 预览 -->
      <div v-if="showPreview" class="preview-box" v-html="renderedHtml"></div>
      <div class="post-actions">
        <el-button type="primary" @click="submitPost">发布</el-button>
        <el-button @click="togglePreview">{{ showPreview ? '编辑' : '预览' }}</el-button>
      </div>
    </el-card>
    <!-- 荣誉勋章 -->
    <div class="section">
      <h3>🏇 我的勋章</h3>
      <el-card class="badge-card" shadow="hover">
        <el-tag
            v-for="badge in badges"
            :key="badge"
            type="success"
            effect="dark"
            class="badge"
        >
          {{ badge }}
        </el-tag>
      </el-card>
    </div>

    <!-- 悬浮按钮：搜索用户 -->
    <div class="search-float-btn" @click="searchUserDialogVisible = true">
      🔍
    </div>

    <!-- 悬浮按钮：聊天 -->
    <div class="chat-float-btn" @click="chatVisible = !chatVisible">
      💬
    </div>

    <!-- 聊天窗口 -->
    <Chat v-if="chatVisible" title="社区聊天" @closeChat="closeChat"/>

    <!-- 搜索用户对话框 -->
    <el-dialog title="搜索用户" v-model="searchUserDialogVisible" width="60%">
      <user-list></user-list>
    </el-dialog>
  </div>
</template>

<script setup>
import {ref, computed, onMounted} from 'vue';
import Chat from "@/components/detail/Chat.vue";
import {ElMessage} from "element-plus";
import Announcement from "@/components/detail/Announcement.vue";
import {ChatDotSquare, Delete, Star, Close} from '@element-plus/icons-vue'
import {
  encrypt,
  sendAxiosRequest,
  pubFormatDate,
  getGuid,
  buildChildrenData,
  ele_confirm,
  loadScript, sendNotifications, getCurrentUserAdminObject, getUserAdminObjectByUserCode
} from "@/utils/common.js";
import {useUserStore} from "@/stores/main/user.js";
import {adminUserCode} from "@/config/vue-config.js";
import {useRouter} from "vue-router";
import {marked} from 'marked';
import {getAnnouncementByRouterName, pubOpenUser} from "@/utils/blogUtil.js";
import UserList from "@/components/detail/UserList.vue";

const router = useRouter();
const userStore = useUserStore();

const newPost = ref('');
const replyTarget = ref(null);
const allPosts = ref([]);
const page = ref(1)
const pageSize = 5
const loading = ref(false)
const noMore = ref(false)
const searchKeyword = ref('')

// 帖子展示
const displayedPosts = computed(() => allPosts.value);

onMounted(() => {
  fetchArticles(); // 只加载第一页
  //loadAiFun();//加载Ai按钮
});

// 提交帖子
function submitPost() {
  if (!newPost.value.trim()) return;
  if (!userStore.userBean.code) {
    ElMessage.error("请先登录!");
    return false;
  }
  let community = {
    GUID: getGuid(),
    USERCODE: userStore.userBean.code,
    USERNAME: userStore.userBean.name,
    AVATAR: userStore.userBean.avatar || "",
    CREATE_TIME: '刚刚',
    TEXT: marked(newPost.value),
    comments: [],
    newComment: '',
    showComments: false,
    showAllComments: false,
    hasLoadedComments: true,
    isExpanded: false, // 用于控制帖子是否展开
  }
  allPosts.value.unshift(community);
  sendAxiosRequest("/blog-api/community/addCommunity", {
    community: {
      GUID: community.GUID,
      USERCODE: community.USERCODE,
      USERNAME: community.USERNAME,
      TEXT: community.TEXT
    }
  });
  newPost.value = '';
}

// 加载帖子
const fetchArticles = async () => {
  if (loading.value || noMore.value) return
  loading.value = true
  try {
    const res = await sendAxiosRequest('/blog-api/community/getAllCommunity', {
      page: page.value,
      pageSize,
      keyword: searchKeyword.value
    });
    const newData = res.result.data;
    if (newData.length < pageSize) noMore.value = true;
    allPosts.value.push(...newData);
    page.value++;
  } catch (e) {
    console.error('获取文章失败', e)
  } finally {
    loading.value = false
  }
}

function setTopCommunity(community) {
  let isTop = "1";
  let tip = "已置顶,刷新页面显示最新效果";
  if (community.ISTOP == "1") {
    isTop = "0";
    tip = "已取消置顶,刷新页面显示最新效果";
  }
  community.ISTOP = isTop;
  sendAxiosRequest("/blog-api/community/setTopCommunity", {communityGuid: community.GUID, isTop});
  ElMessage.success(tip);
}

// 删除帖子
function deleteCommunity(community) {
  ele_confirm("是否确认删除该内容!", () => {
    sendAxiosRequest("/blog-api/community/deleteCommunity", {communityGuid: community.GUID});
    allPosts.value = allPosts.value.filter(item => item.GUID != community.GUID);
  });
}

// 加载评论
async function toggleComments(postItem) {
  postItem.showComments = !postItem.showComments;
  if (postItem.showComments && !postItem.hasLoadedComments) {
    try {
      const res = await sendAxiosRequest('/blog-api/community/getComment', {
        communityId: postItem.GUID // 或你的帖子唯一标识字段
      });
      postItem.comments = buildChildrenData(res.result || []);
      postItem.hasLoadedComments = true;
    } catch (e) {
      ElMessage.error('加载评论失败');
    }
  }
}

function limitedComments(item) {
  if (item.showAllComments) return item.comments;
  return (item.comments || []).slice(0, 3);
}

function submitComment(postItem) {
  if (!postItem.newComment.trim()) return;
  if (!userStore.userBean.code) {
    ElMessage.error("请先登录!");
    return false;
  }
  let comment = {
    GUID: getGuid(),
    USERCODE: userStore.userBean.code,
    USERNAME: userStore.userBean.name,
    COMMUNITYID: postItem.GUID,
    AVATAR: userStore.userBean.avatar,
    TEXT: postItem.newComment,
    children: [],
    replyText: ''
  }
  postItem.comments.unshift(comment);
  let sendComment = {
    GUID: comment.GUID,
    USERCODE: comment.USERCODE,
    USERNAME: comment.USERNAME,
    COMMUNITYID: comment.COMMUNITYID,
    TEXT: comment.TEXT,
  }
  sendAxiosRequest("/blog-api/community/addComment", {comment: sendComment});
  postItem.newComment = '';
  //发送通知
  sendNotifications(comment.USERCODE, postItem.USERCODE, "comment", null, `${userStore.userBean.name}评论了你在${pubFormatDate(postItem.CREATE_TIME)}发布的社区内容`)
}

function submitReply(postItem, comment) {
  if (!comment.replyText || !comment.replyText.trim()) return;
  if (!userStore.userBean.code) {
    ElMessage.error("请先登录!");
    return false;
  }
  comment.children = comment.children || [];
  let oneComment = {
    USERCODE: userStore.userBean.code,
    USERNAME: userStore.userBean.name,
    AVATAR: userStore.userBean.avatar,
    COMMUNITYID: postItem.GUID,
    SUPERGUID: comment.GUID,
    TEXT: comment.replyText,
  }
  comment.children.push(oneComment);

  let sendComment = {
    USERCODE: oneComment.USERCODE,
    USERNAME: oneComment.USERNAME,
    COMMUNITYID: oneComment.COMMUNITYID,
    SUPERGUID: oneComment.SUPERGUID,
    TEXT: oneComment.TEXT,
  }
  sendAxiosRequest("/blog-api/community/addComment", {comment: sendComment});
  comment.replyText = '';
  replyTarget.value = null;
}

function expandPost(postItem) {
  postItem.isExpanded = true;
}

function collapsePost(postItem) {
  postItem.isExpanded = false;
}

function avatarClick(community) {
  pubOpenUser(router, community.USERCODE);
}

function commentAvatarClick(comment) {
  pubOpenUser(router, comment.USERCODE);
}

function userInfoCLick(userInfo) {
  pubOpenUser(router, userInfo.CODE);
}

const showPreview = ref(false)
const togglePreview = () => {
  showPreview.value = !showPreview.value
}
const renderedHtml = computed(() => {
  return marked.parse(newPost.value || '')
})

const searchUserDialogVisible = ref(false);
const searchUserInput = ref('');
const searchUserArr = ref([]);

async function searchUser() {
  searchUserArr.value = [];
  if (!searchUserInput.value.trim()) return;
  let result = await sendAxiosRequest("/pub-api/login/getUserInfoByName", {userName: searchUserInput.value});
  if (!result || result.isError) {
    ElMessage.error("发生错误!");
    return false;
  }
  if (result.result.length == 0) {
    ElMessage.success("未搜索到用户");
    return false;
  }
  searchUserArr.value = result.result;
}

const chatVisible = ref(false);

const closeChat = () => {
  chatVisible.value = false;
};


const loadAiFun = async () => {
  // 动态加载 SDK 脚本
  const sdkUrl = "https://agi-dev-platform-web.bj.bcebos.com/ai_apaas/embed/output/embedLiteSDK.js?responseExpires=0";
  if (!window.EmbedLiteSDK) {
    await loadScript(sdkUrl);
  }
  const appId = "f85ab2ae-b66c-4b7b-98e0-7241ed296953";
  const code = "embedgbotWcUfzUsuj4CQw9Wj";
  new window.EmbedLiteSDK({
    appId,
    code,
  });
}
//公告横幅内容
const topAlert = ref([]);
const setTopAlert = async () => {
  topAlert.value = await getAnnouncementByRouterName("Community");
}
setTopAlert();
const badges = ref(["原始股"]);
</script>

<style scoped>
/* ==========================================
   ✦ 设计令牌 (Design Tokens) - 核心变量
   ========================================== */
.community-page {
  /* 色彩系统 */
  --brand-primary: #4f46e5;      /* 现代靛蓝，比普通蓝更高级 */
  --brand-primary-hover: #4338ca;
  --bg-page: #f8fafc;            /* 极浅蓝灰背景 */
  --bg-card: #ffffff;
  --bg-comment: #f1f5f9;
  --text-main: #0f172a;          /* 深蓝灰，代替死板的纯黑 */
  --text-regular: #334155;
  --text-muted: #64748b;
  --border-light: #e2e8f0;

  /* 阴影系统 - 多层弥散阴影打造空间感 */
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -1px rgba(0, 0, 0, 0.03);
  --shadow-hover: 0 20px 25px -5px rgba(0, 0, 0, 0.05), 0 10px 10px -5px rgba(0, 0, 0, 0.02);
  --shadow-float: 0 10px 30px -5px rgba(79, 70, 229, 0.3);

  /* 形状与动画 */
  --radius-xl: 16px;
  --radius-lg: 12px;
  --radius-md: 8px;
  --transition-spring: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  --transition-smooth: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);

  /* 页面基础设置 */
  max-width: 860px;
  margin: 0 auto;
  padding: 30px 20px 80px; /* 底部留白防止遮挡悬浮按钮 */
  background-color: var(--bg-page);
  min-height: 100vh;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  color: var(--text-main);
}

/* ==========================================
   ✦ 标题与基础区块
   ========================================== */
.section { margin-top: 40px; }

h3 {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-main);
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 10px;
  letter-spacing: -0.02em;
}

/* ==========================================
   ✦ 现代卡片设计 (发帖/帖子列表)
   ========================================== */
.post-box, .feed-card, .badge-card {
  background: var(--bg-card) !important;
  border-radius: var(--radius-xl) !important;
  border: 1px solid rgba(255,255,255,0.8) !important; /* 配合阴影实现微刻线效果 */
  box-shadow: var(--shadow-md) !important;
  transition: var(--transition-smooth) !important;
  padding: 24px !important;
  position: relative;
  overflow: visible; /* 为了让头像和角标能突破边界一点点 */
}

.feed-card { margin-bottom: 24px; }

.feed-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-hover) !important;
}

/* ==========================================
   ✦ 发帖者头部信息
   ========================================== */
.feed-header {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 16px;
}

.author-avatar {
  border: 2px solid #fff;
  box-shadow: var(--shadow-sm);
  transition: var(--transition-spring);
}
.author-avatar:hover { transform: scale(1.1); }

.author-info { display: flex; flex-direction: column; gap: 2px; }

.author-name {
  font-weight: 700;
  font-size: 1.05rem;
  color: var(--text-main);
  letter-spacing: -0.01em;
}

.author-meta { font-size: 0.85rem; color: var(--text-muted); }

/* ==========================================
   ✦ 帖子正文与展开逻辑
   ========================================== */
.feed-body {
  font-size: 0.95rem;
  line-height: 1.8;
  color: var(--text-regular);
  margin: 16px 0;
  word-break: break-word;
}

.expand-btn-wrapper {
  display: flex;
  justify-content: center;
  margin-top: -10px;
  padding-top: 15px;
  position: relative;
}

.expand-btn-wrapper::before {
  content: '';
  position: absolute;
  top: -30px; left: 0; right: 0; height: 30px;
  background: linear-gradient(to bottom, rgba(255,255,255,0), var(--bg-card));
  pointer-events: none;
}

.expand-btn {
  background: var(--bg-comment) !important;
  border: none !important;
  color: var(--brand-primary) !important;
  border-radius: 20px !important;
  padding: 8px 24px !important;
  font-weight: 600 !important;
  font-size: 0.85rem !important;
  transition: var(--transition-smooth) !important;
}
.expand-btn:hover {
  background: var(--brand-primary) !important;
  color: #fff !important;
  transform: translateY(-1px);
}

/* ==========================================
   ✦ 交互操作区 (点赞/评论/删除按钮)
   ========================================== */
.feed-actions {
  display: flex;
  gap: 8px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--border-light);
}

.feed-actions .el-button {
  color: var(--text-muted) !important;
  font-weight: 600;
  border-radius: var(--radius-md) !important;
  padding: 8px 12px !important;
  height: auto !important;
  transition: var(--transition-smooth);
}
.feed-actions .el-button:hover {
  background-color: var(--bg-comment) !important;
  color: var(--brand-primary) !important;
}
.feed-actions .el-button:last-child:hover {
  background-color: #fee2e2 !important; /* 删除按钮特型：浅红底 */
  color: #ef4444 !important; /* 删除按钮特型：红字 */
}

/* ==========================================
   ✦ 评论区 (极致的嵌套层级视觉)
   ========================================== */
.comment-section {
  margin-top: 20px;
  background-color: #fff;
  border-radius: var(--radius-lg);
  padding: 0; /* 移除外框背景，融入卡片 */
}

.comment-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px 0;
  border-bottom: 1px solid var(--border-light);
}
.comment-item:last-child { border-bottom: none; }

.author-avatar-comment {
  width: 36px !important;
  height: 36px !important;
  box-shadow: var(--shadow-sm);
}

.comment-content {
  flex: 1;
  background: var(--bg-comment);
  padding: 14px 16px;
  border-radius: 0 var(--radius-lg) var(--radius-lg) var(--radius-lg);
}

.comment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.comment-author {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-main);
}

.comment-text {
  font-size: 0.9rem;
  color: var(--text-regular);
  line-height: 1.6;
}

/* 回复引导线 (Thread Line) */
.children {
  margin-top: 12px;
  padding-left: 16px;
  border-left: 2px solid var(--border-light);
  position: relative;
}

.reply-item {
  margin-bottom: 10px;
  font-size: 0.85rem;
  line-height: 1.5;
  background: rgba(255,255,255,0.5);
  padding: 8px 12px;
  border-radius: var(--radius-md);
}
.reply-item:last-child { margin-bottom: 0; }
.reply-author { font-weight: 700; color: var(--brand-primary); }

.comment-input, .reply-box { margin-top: 16px; }

/* ==========================================
   ✦ 输入框与发帖区增强
   ========================================== */
:deep(.el-textarea__inner), :deep(.el-input__inner) {
  border-radius: var(--radius-md);
  background-color: var(--bg-comment);
  border: 1px solid transparent;
  transition: var(--transition-smooth);
  box-shadow: none !important;
}
:deep(.el-textarea__inner:focus), :deep(.el-input__inner:focus) {
  background-color: #fff;
  border-color: var(--brand-primary);
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1) !important; /* 现代Focus环 */
}

/* ==========================================
   ✦ 悬浮操作按钮 (毛玻璃特效)
   ========================================== */
.search-float-btn, .chat-float-btn, .ai-float-btn {
  position: fixed;
  right: 32px;
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: white;
  border-radius: 50%;
  backdrop-filter: blur(8px); /* 毛玻璃 */
  -webkit-backdrop-filter: blur(8px);
  cursor: pointer;
  z-index: 1000;
  transition: var(--transition-spring);
  border: 1px solid rgba(255,255,255,0.2);
}

.search-float-btn:hover, .chat-float-btn:hover, .ai-float-btn:hover {
  transform: translateY(-5px) scale(1.05);
}

.chat-float-btn { background: rgba(79, 70, 229, 0.9); box-shadow: var(--shadow-float); bottom: 40px; }
.search-float-btn { background: rgba(16, 185, 129, 0.9); box-shadow: 0 10px 30px -5px rgba(16, 185, 129, 0.3); bottom: 112px; }
.ai-float-btn { background: rgba(15, 23, 42, 0.9); box-shadow: 0 10px 30px -5px rgba(15, 23, 42, 0.3); bottom: 184px; }

/* ==========================================
   ✦ 精致的置顶角标 (无缝贴合圆角)
   ========================================== */
.ribbon-wrapper {
  width: 85px; height: 85px;
  overflow: hidden;
  position: absolute;
  top: 0; right: 0;
  border-radius: 0 var(--radius-xl) 0 0;
  z-index: 10;
  pointer-events: none;
}

.ribbon {
  font: bold 11px 'Inter', sans-serif;
  color: #78350f;
  text-align: center;
  text-transform: uppercase;
  letter-spacing: 1px;
  transform: rotate(45deg);
  position: absolute;
  top: 18px; right: -24px;
  width: 110px;
  background: linear-gradient(135deg, #fde68a, #fbbf24);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  padding: 6px 0;
}

/* ==========================================
   ✦ 其他组件补全
   ========================================== */
.badge-card { display: flex; flex-wrap: wrap; gap: 8px; }
.badge {
  border-radius: 20px;
  padding: 4px 12px;
  font-weight: 600;
  border: none;
}

.preview-box {
  margin-top: 16px;
  padding: 20px;
  background-color: var(--bg-comment);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-light);
  line-height: 1.7;
}

.loading-text, .end-text {
  text-align: center;
  margin: 40px 0;
  color: var(--text-muted);
  font-size: 0.9rem;
  letter-spacing: 0.05em;
}

/* 淡入淡出动画过渡 */
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s ease, transform 0.3s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; transform: translateY(-10px); }
</style>
