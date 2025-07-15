<template>
  <!-- 公告横幅 -->
  <el-alert
      type="success"
      show-icon
      class="banner-alert"
      closable
  >
    <template #title>
      <div style="text-align: center;">
        🎉 欢迎来到 YNS 社区！新用户请阅读
        <a
            href="http://ynsstudy.cn/oneBlog?g=F4D23BEE65D1432196379195FF8F1EC0&u=7E5B0BB01BBEB7CF4C8A528794A99698&n=%E7%A4%BE%E5%8C%BA%E6%8C%87%E5%8D%97"
            target="_blank"
            style="color: #409EFF; text-decoration: underline; margin-left: 4px;"
        >《社区指南》</a> ~
      </div>
    </template>
  </el-alert>

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
        <div class="feed-body" v-html="item.TEXT"></div>
        <div class="feed-actions">
          <el-button text size="small" :icon="ChatDotSquare" @click="toggleComments(item)">
            {{ item.showComments ? '收起评论' : '评论' }}
          </el-button>
          <el-button v-if="userStore.userBean.code == adminUserCode" text size="small" :icon="Star"
                     @click="setTopCommunity(item)">{{ item.ISTOP == "1" ? "取消置顶" : "置顶" }}
          </el-button>
          <el-button v-if="userStore.userBean.code == item.USERCODE" text size="small" :icon="Delete"
                     @click="deleteCommunity(item)">删除
          </el-button>
        </div>
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
                  @click="commentAvatarClick(child)"
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
                        @click="commentAvatarClick(child)"
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
          style="margin: 20px auto; display: block; border-radius: 8px; padding: 8px 20px; background-color: #409EFF; color: white;"
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
          v-model="newPost"
          type="textarea"
          :rows="3"
          placeholder="分享你的想法、提问、或者一句话"
          class="post-textarea"
      />
      <el-button type="primary" @click="submitPost" class="submit-btn">发布</el-button>
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
    <div class="chat-window" v-if="chatVisible">
      <div class="chat-header">
        社区聊天
        <span class="close-btn" @click="chatVisible = false">×</span>
      </div>
      <div class="chat-messages" ref="chatMessagesRef">
        <div
            v-for="(msg, index) in chatMessages"
            :key="index"
            :class="['chat-message', msg.user === '我' ? 'self' : 'other']"
        >
          <div class="message-bubble">
            <strong v-if="msg.user !== '我'" class="sender">{{ msg.user }}：</strong>
            {{ msg.text }}
          </div>
        </div>
      </div>
      <div class="chat-input-row">
        <el-input
            v-model="chatText"
            size="small"
            placeholder="输入消息，回车或点击发送"
            @keyup.enter="sendChat"
            class="chat-input"
        />
        <el-button type="primary" size="small" @click="sendChat" class="chat-send-btn">发送</el-button>
      </div>
    </div>

    <!-- 搜索用户对话框 -->
    <el-dialog title="搜索用户" v-model="searchUserDialogVisible" width="400px" :before-close="handleClose">
      <el-input
          v-model="searchUserInput"
          placeholder="输入用户名"
          clearable
          prefix-icon="el-icon-search"
          size="small"
          style="margin-bottom: 12px"
      />
      <div v-for="userInfo in searchUserArr" :key="userInfo.CODE">
        <el-card class="user-card" shadow="hover" style="margin-bottom: 10px">
          <div class="user-card-content">
            <el-avatar
                :src="userInfo.AVATAR"
                size="large"
                class="author-avatar-search"
                alt="评论用户头像"
            >
              {{ userInfo.NAME?.charAt(0) }}
            </el-avatar>
            <span class="user-name">{{ userInfo.NAME }}</span>
            <span class="user-remark">{{ userInfo.REMARK }}</span>
          </div>
        </el-card>
      </div>
      <template #footer>
        <el-button @click="searchUserDialogVisible = false" class="btn-cancel">取消</el-button>
        <el-button type="primary" @click="searchUser" class="btn-search">搜索</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import {ref, computed, nextTick, onMounted, onBeforeUnmount} from 'vue';
import {ElMessage} from "element-plus";
import {ChatDotSquare, Delete, Star} from '@element-plus/icons-vue'
import {encrypt, sendAxiosRequest, pubFormatDate, getGuid, buildChildrenData, ele_confirm} from "@/utils/common.js";
import {useUserStore} from "@/stores/main/user.js";
import {adminUserCode} from "@/config/vue-config.js";
import {useRouter} from "vue-router";

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
    TEXT: newPost.value,
    comments: [],
    newComment: '',
    showComments: false,
    showAllComments: false,
    hasLoadedComments: true,
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
    }, 'get');
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

//删除帖子
function deleteCommunity(community) {
  ele_confirm("是否确认删除该内容!", () => {
    sendAxiosRequest("/blog-api/community/deleteCommunity", {communityGuid: community.GUID});
    allPosts.value = allPosts.value.filter(item => item.GUID != community.GUID);
  });
}

//加载评论
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
  postItem.comments.push(comment);
  let sendComment = {
    GUID: comment.GUID,
    USERCODE: comment.USERCODE,
    USERNAME: comment.USERNAME,
    COMMUNITYID: comment.COMMUNITYID,
    TEXT: comment.TEXT,
  }
  sendAxiosRequest("/blog-api/community/addComment", {comment: sendComment});
  postItem.newComment = '';
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

function avatarClick(community) {
  const routeUrl = router.resolve({name: 'personInfomation', query: {c: encrypt(community.USERCODE)}}).href;
  window.open(routeUrl, community.USERCODE);
}

function commentAvatarClick(comment) {
  const routeUrl = router.resolve({name: 'personInfomation', query: {c: encrypt(comment.USERCODE)}}).href;
  window.open(routeUrl, comment.USERCODE);
}

const chatVisible = ref(false);
const chatText = ref('');
const chatMessages = ref([
  {user: '小明', text: '有人在吗？'},
  {user: '我', text: '我在！'},
]);
const chatMessagesRef = ref(null);

function sendChat() {
  if (!chatText.value.trim()) return;
  chatMessages.value.push({user: '我', text: chatText.value});
  chatText.value = '';
  nextTick(() => {
    if (chatMessagesRef.value) {
      chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight;
    }
  });
}

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

const badges = ref(['原始股']);
</script>

<style scoped>
.author-avatar-comment {
  width: 20px !important;
  height: 20px !important;
  cursor: pointer;
}

.community-page {
  padding: 30px;
  background: #f7f8fa;
}

.section {
  margin-top: 40px;
}

.post-box {
  padding: 20px;
  border-radius: 8px;
}

.submit-btn {
  margin-top: 10px;
}

.feed-card {
  position: relative;
  margin-bottom: 15px;
  padding: 15px;
  border-radius: 6px;
}

.feed-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.author-avatar {
  cursor: pointer;
}

.author-info {
  display: flex;
  flex-direction: column;
  font-size: 13px;
  color: #555;
}

.author-name {
  font-weight: bold;
  font-size: 14px;
  color: #333;
}

.author-meta {
  font-size: 12px;
  color: #999;
}

.feed-body {
  font-size: 15px;
  margin: 10px 0;
}

.feed-actions {
  margin-top: 10px;
  display: flex;
  gap: 10px;
}

.comment-section {
  margin-top: 10px;
  background-color: #f9f9f9;
  border-radius: 6px;
  padding: 10px;
}

.comment-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 12px;
  border-bottom: 1px solid #eee;
  padding-bottom: 6px;
}

.comment-avatar {
  width: 26px;
  height: 26px;
}

.comment-content {
  flex: 1;
}

.comment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  color: #333;
  margin-bottom: 4px;
}

.comment-text {
  font-size: 14px;
  color: #555;
}

.reply-box {
  margin-top: 8px;
}

.children {
  margin-top: 8px;
  padding-left: 30px;
  border-left: 2px solid #ddd;
}

.reply-item {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 4px 0;
}

.reply-avatar {
  width: 22px;
  height: 22px;
}

.reply-author {
  font-size: 13px;
  font-weight: bold;
  color: #444;
}

.reply-text {
  font-size: 13px;
  color: #333;
}

.loading-text,
.end-text {
  text-align: center;
  margin: 20px 0;
  color: #888;
}

.search-float-btn,
.chat-float-btn {
  position: fixed;
  right: 20px;
  padding: 10px 14px;
  font-size: 24px;

  color: white;
  border-radius: 50%;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  cursor: pointer;
  z-index: 1000;
}

.search-float-btn {
  background: #239b19;
  bottom: 90px;
}

.chat-float-btn {
  background: #409EFF;
  bottom: 30px;
}

.chat-window {
  position: fixed;
  right: 20px;
  bottom: 90px;
  width: 340px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 25px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  z-index: 1000;
  border: 1px solid #e0e0e0;
}

.chat-header {
  background: #409EFF;
  color: white;
  padding: 12px 16px;
  font-size: 16px;
  font-weight: bold;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.close-btn {
  cursor: pointer;
  font-size: 18px;
}

.chat-messages {
  padding: 10px;
  max-height: 240px;
  overflow-y: auto;
  font-size: 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: #f8f8f8;
}
.chat-message {
  display: flex;
}

.chat-message.self {
  justify-content: flex-end;
}

.chat-message.other {
  justify-content: flex-start;
}

.message-bubble {
  max-width: 70%;
  padding: 8px 12px;
  border-radius: 14px;
  background-color: #f0f0f0;
  word-break: break-word;
  font-size: 14px;
}

.chat-message.self .message-bubble {
  background-color: #409eff;
  color: white;
  border-bottom-right-radius: 0;
}

.chat-message.other .message-bubble {
  background-color: #e4e6eb;
  color: #333;
  border-bottom-left-radius: 0;
}

.chat-input-row {
  display: flex;
  padding: 10px;
  border-top: 1px solid #eee;
  background-color: #fff;
  gap: 6px;
}

.chat-input {
  flex: 1;
}

.chat-send-btn {
  white-space: nowrap;
}

.badge-card {
  padding: 15px;
}

.badge {
  margin: 5px;
}

.ribbon-wrapper {
  width: 75px;
  height: 75px;
  overflow: hidden;
  position: absolute;
  top: 0;
  right: 0;
  z-index: 10;
}

.ribbon {
  font: bold 12px sans-serif;
  color: #1e1d1d;
  text-align: center;
  transform: rotate(45deg);
  position: absolute;
  top: 12px;
  right: -20px;
  width: 100px;
  background: #fadb14; /* 明亮黄色 */
  box-shadow: 0 0 3px rgba(0, 0, 0, 0.3);
  padding: 4px 0;
}

.banner-alert {
  display: flex;
  justify-content: center;
  align-items: center;
  text-align: center;
}

.user-card {
  margin-bottom: 12px;
  padding: 10px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.user-card-content {
  display: flex;
  align-items: center;
  gap: 15px;            /* 增加头像与用户名之间的间距 */
  flex-wrap: wrap;       /* 允许内容换行 */
  width: 100%;
}

.user-name {
  font-size: 16px;
  font-weight: bold;
  color: #333;
  max-width: 5em;        /* 限制最大宽度为5个字符 */
  white-space: nowrap;   /* 禁止换行 */
  overflow: hidden;      /* 隐藏超出部分 */
  text-overflow: ellipsis; /* 显示省略号 */
  flex-shrink: 0;        /* 防止用户名被压缩 */
}

.user-remark {
  font-size: 13px;
  color: #888;
  margin-top: 4px;
  white-space: nowrap;   /* 禁止换行 */
  overflow: hidden;      /* 隐藏超出部分 */
  text-overflow: ellipsis; /* 显示省略号 */
  flex-shrink: 1;        /* 允许个性签名压缩 */
}

.author-avatar-search{
  width: 40px !important;
  height: 40px !important;
  border-radius: 50%;
}
</style>
