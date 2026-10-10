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
        <span class="j-tape feed-tape" :class="['j-tape--green', 'j-tape--pink', '', 'j-tape--blue'][index % 4]"></span>
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
        <div class="feed-body" v-if="item.isExpanded" v-html="sanitizeHtml(item.TEXT)"></div>
        <div class="feed-body" v-else>
          <!--不展开时,只显示两行内容-->
          <div v-html="sanitizeHtml(item.TEXT.split('\n').slice(0, 2).join('\n'))"></div>
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
    <el-card class="post-box" shadow="never">
      <span class="j-tape j-tape--white j-tape--top"></span>
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
      <div v-if="showPreview" class="preview-box" v-html="sanitizeHtml(renderedHtml)"></div>
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
  loadScript, sendNotifications, getCurrentUserAdminObject, getUserAdminObjectByUserCode, sanitizeHtml, sendAxiosRequestChecked
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
    TEXT: sanitizeHtml(marked(newPost.value)),
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

async function setTopCommunity(community) {
  let isTop = "1";
  let tip = "已置顶,刷新页面显示最新效果";
  if (community.ISTOP == "1") {
    isTop = "0";
    tip = "已取消置顶,刷新页面显示最新效果";
  }
  const result = await sendAxiosRequestChecked("/blog-api/community/setTopCommunity", {communityGuid: community.GUID, isTop}, "置顶操作失败");
  if (!result) return;
  community.ISTOP = isTop;
  ElMessage.success(tip);
}

// 删除帖子
function deleteCommunity(community) {
  ele_confirm("是否确认删除该内容!", async () => {
    const result = await sendAxiosRequestChecked("/blog-api/community/deleteCommunity", {communityGuid: community.GUID}, "删除失败");
    if (!result) return;
    allPosts.value = allPosts.value.filter(item => item.GUID != community.GUID);
    ElMessage.success("删除成功");
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
.community-page {
  max-width: 860px;
  margin: 0 auto;
  padding: 36px 20px 96px;
  min-height: 100vh;
  color: var(--j-ink);
}

.section {
  margin-top: 36px;
}

.section:first-child {
  margin-top: 0;
}

h3 {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 0 22px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 24px;
  color: var(--j-ink);
}

/* ==========================================
   帖子：一张张贴着胶带的纸
   ========================================== */
.feed-card,
.badge-card {
  position: relative;
  overflow: visible;
  border: 1px solid var(--j-rule) !important;
  border-radius: 2px !important;
  background: var(--j-paper) !important;
  box-shadow: var(--j-shadow) !important;
}

.feed-card {
  margin-bottom: 30px;
  transition: transform 0.2s ease;
}

.feed-card:nth-of-type(odd) {
  transform: rotate(-0.3deg);
}

.feed-card:nth-of-type(even) {
  transform: rotate(0.3deg);
}

.feed-card:hover {
  transform: rotate(0deg);
}

.feed-card :deep(.el-card__body) {
  padding: 26px 26px 18px;
}

.feed-tape {
  top: -10px;
  left: 28px;
  transform: rotate(-4deg);
}

.feed-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}

.author-avatar {
  flex-shrink: 0;
  border: 2px solid #fff;
  box-shadow: 0 0 0 1px var(--j-rule);
  background: #ece4d3;
  color: var(--j-ink-soft);
  cursor: pointer;
}

.author-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.author-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--j-ink);
}

.author-meta {
  font-family: var(--j-hand);
  font-size: 14px;
  color: var(--j-muted);
}

.feed-body {
  margin: 12px 0;
  font-size: 15px;
  line-height: 1.85;
  color: var(--j-ink);
  word-break: break-word;
}

.feed-body :deep(a) {
  color: var(--j-pen);
}

.feed-body :deep(img) {
  max-width: 100%;
  padding: 5px;
  background: #fff;
  box-shadow: 0 1px 3px rgba(60, 50, 30, 0.2);
  box-sizing: border-box;
}

.expand-btn-wrapper {
  position: relative;
  display: flex;
  justify-content: center;
  margin-top: -10px;
  padding-top: 14px;
}

.expand-btn-wrapper::before {
  content: '';
  position: absolute;
  top: -30px;
  left: 0;
  right: 0;
  height: 30px;
  background: linear-gradient(to bottom, rgba(255, 253, 248, 0), var(--j-paper));
  pointer-events: none;
}

.expand-btn {
  padding: 6px 22px !important;
  border: 1px dashed var(--j-rule-strong) !important;
  border-radius: 0 !important;
  background: transparent !important;
  font-family: var(--j-hand);
  font-size: 15px !important;
  color: var(--j-ink-soft) !important;
}

.expand-btn:hover {
  border-color: var(--j-pen) !important;
  color: var(--j-pen) !important;
}

.feed-actions {
  display: flex;
  gap: 4px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px dashed var(--j-rule);
}

.feed-actions .el-button {
  height: auto !important;
  padding: 6px 10px !important;
  color: var(--j-ink-soft) !important;
}

.feed-actions .el-button:hover {
  background-color: rgba(250, 216, 96, 0.25) !important;
  color: var(--j-ink) !important;
}

.feed-actions .el-button:last-child:hover {
  background-color: rgba(194, 72, 62, 0.1) !important;
  color: var(--j-stamp) !important;
}

/* ==========================================
   评论
   ========================================== */
.comment-section {
  margin-top: 16px;
}

.comment-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px dashed var(--j-rule);
}

.comment-item:last-child {
  border-bottom: none;
}

.author-avatar-comment {
  flex-shrink: 0;
  width: 34px !important;
  height: 34px !important;
  background: #ece4d3;
  color: var(--j-ink-soft);
  cursor: pointer;
}

.comment-content {
  flex: 1;
  min-width: 0;
  padding: 10px 14px;
  background: var(--j-paper-warm);
  border: 1px solid var(--j-rule);
}

.comment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.comment-author {
  font-size: 13px;
  font-weight: 600;
  color: var(--j-pen);
}

.comment-text {
  font-size: 14px;
  line-height: 1.65;
  color: var(--j-ink);
}

.children {
  margin-top: 10px;
  padding-left: 12px;
  border-left: 2px solid #ead9a0;
}

.reply-item {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  margin-bottom: 8px;
  font-size: 13px;
  line-height: 1.6;
}

.reply-item:last-child {
  margin-bottom: 0;
}

.reply-item .author-avatar-comment {
  width: 24px !important;
  height: 24px !important;
  font-size: 11px;
}

.reply-author {
  flex-shrink: 0;
  font-weight: 600;
  color: var(--j-pen);
}

.comment-input,
.reply-box {
  margin-top: 14px;
}

.show-more-comments {
  text-align: center;
}

/* ==========================================
   快速发帖：黄色便签
   ========================================== */
.post-box {
  position: relative;
  overflow: visible;
  margin-top: 40px;
  border: none !important;
  border-radius: 2px !important;
  background: var(--j-note) !important;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.1), 0 14px 22px -14px rgba(60, 50, 30, 0.45) !important;
  transform: rotate(-0.6deg);
}

.post-box :deep(.el-card__body) {
  padding: 26px 24px 20px;
}

.post-box h3 {
  margin-bottom: 14px;
}

.post-box :deep(.el-textarea__inner) {
  border: none;
  box-shadow: none;
  background-color: transparent;
  background-image: repeating-linear-gradient(transparent 0 27px, rgba(160, 135, 60, 0.28) 27px 28px);
  line-height: 28px;
  padding: 0 4px;
  font-size: 15px;
  color: var(--j-ink);
}

.post-box :deep(.el-textarea__inner::placeholder) {
  font-family: var(--j-hand);
  color: #a08a52;
}

.post-actions {
  display: flex;
  gap: 8px;
  margin-top: 14px;
}

.preview-box {
  padding: 14px 4px;
  line-height: 1.8;
  color: var(--j-ink);
}

/* ==========================================
   勋章：小圆章
   ========================================== */
.badge-card :deep(.el-card__body) {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.badge {
  height: auto;
  padding: 6px 14px;
  border: 1.5px solid var(--j-stamp) !important;
  border-radius: 20px;
  background: transparent !important;
  font-family: var(--j-hand);
  font-size: 14px;
  color: var(--j-stamp) !important;
  transform: rotate(-3deg);
}

.badge:nth-child(even) {
  border-color: var(--j-pen) !important;
  color: var(--j-pen) !important;
  transform: rotate(2deg);
}

/* ==========================================
   悬浮按钮：两块便签
   ========================================== */
.search-float-btn,
.chat-float-btn,
.ai-float-btn {
  position: fixed;
  right: 32px;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 54px;
  height: 54px;
  border-radius: 2px;
  font-size: 22px;
  cursor: pointer;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.12), 0 10px 16px -10px rgba(60, 50, 30, 0.5);
  transition: transform 0.2s ease;
}

.chat-float-btn {
  bottom: 40px;
  background: #f3cccc;
  transform: rotate(3deg);
}

.search-float-btn {
  bottom: 110px;
  background: #c7e2cf;
  transform: rotate(-3deg);
}

.ai-float-btn {
  bottom: 180px;
  background: var(--j-note);
}

.search-float-btn:hover,
.chat-float-btn:hover,
.ai-float-btn:hover {
  transform: rotate(0deg) translateY(-3px);
}

/* ==========================================
   置顶：红色印章
   ========================================== */
.ribbon-wrapper {
  position: absolute;
  top: 18px;
  right: 22px;
  z-index: 10;
  pointer-events: none;
}

.ribbon {
  padding: 4px 10px;
  border: 2px solid rgba(194, 72, 62, 0.8);
  border-radius: 4px;
  font-family: var(--j-hand);
  font-size: 15px;
  letter-spacing: 0.15em;
  color: rgba(194, 72, 62, 0.85);
  transform: rotate(8deg);
}

.loading-text,
.end-text {
  margin: 36px 0;
  text-align: center;
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-muted);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

@media (max-width: 640px) {
  .community-page {
    padding: 24px 14px 96px;
  }

  .feed-card,
  .post-box {
    transform: none !important;
  }

  .feed-card :deep(.el-card__body) {
    padding: 22px 16px 14px;
  }

  .search-float-btn,
  .chat-float-btn {
    right: 16px;
  }
}
</style>
