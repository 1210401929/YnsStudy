<!-- 评论区：文章和便签墙的随记共用（楼中楼、匿名评论、作者章、横线信纸） -->
<template>
  <section class="comment-card">
    <div class="comment-header">
      <h3>{{ title }} <span class="comment-count-badge">{{ commentTotal }}</span></h3>
    </div>

    <p v-if="loaded && comments.length === 0" class="comment-empty">
      <svg class="pencil-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 20l1.2-4.6L15.8 4.8a2 2 0 0 1 2.8 0l.6.6a2 2 0 0 1 0 2.8L8.6 18.8z M13.5 7l3.5 3.5 M4 20l4.6-1.2"/></svg>
      还没有人留言，来写下第一条吧
    </p>

    <div v-for="(comment, i) in visibleComments" :key="comment.GUID || i" class="comment-item">
      <div class="comment-main-row" @click="toggleAction(comment.GUID)">
        <div class="avatar-container">
          <el-tooltip :content="'评论于: '+pubFormatDate(comment.CREATE_TIME)" placement="top" effect="light">
            <el-avatar :src="comment.AVATAR" class="author-avatar-comment"
                       @click.stop="commentAvatarClick(comment)"
                       :style="!comment.AVATAR ? getAvatarStyle(comment.USERNAME) : {}">
              {{ comment.USERNAME?.charAt(0) }}
            </el-avatar>
          </el-tooltip>
        </div>

        <div class="comment-content-block">
          <span class="comment-user-name">{{ comment.USERNAME }}<span v-if="isAuthor(comment)" class="j-stamp author-stamp">作者</span>:</span>
          <span class="comment-text">{{ comment.TEXT }}</span>
        </div>

        <div class="comment-actions" v-show="activeCommentId === comment.GUID">
          <button type="button" class="tool-link" @click.stop="replyComment(comment.GUID, comment)">回复</button>
          <button type="button" class="tool-link danger"
                  @click.stop="deleteComment(comment.GUID)"
                  v-if="canDelete(comment)">
            删除
          </button>
        </div>
      </div>

      <!-- 用真正的表单包起来，浏览器才会识别并记住昵称、邮箱、网址，下次自动填充 -->
      <form v-if="replyInputVisible[comment.GUID]" class="reply-input-wrapper" @submit.prevent="submitReply(comment.GUID)">
        <el-row :gutter="[10, 10]" v-if="!userStore?.userBean?.code" class="guest-form-row">
          <el-col :xs="24" :sm="12">
            <el-input v-model="guestInfo.nickname" name="author" autocomplete="name" :prefix-icon="User" placeholder="昵称 (选填)" size="small" clearable/>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-input v-model="guestInfo.email" name="mail" type="email" autocomplete="email" :prefix-icon="Message" placeholder="邮箱 (选填)" size="small" clearable/>
          </el-col>
          <el-col :span="24">
            <el-input v-model="guestInfo.website" v-url-hint="guestInfo.website" name="url" type="url" autocomplete="on" :prefix-icon="Link" placeholder="http://" size="small" clearable/>
          </el-col>
        </el-row>
        <div class="reply-action-group">
          <el-input
              v-model="replyInputs[comment.GUID]"
              :placeholder="replyTargets[comment.GUID] ? `回复 @${replyTargets[comment.GUID].USERNAME}：` : '写下你的回复...'"
              size="small"
              clearable
          />
          <el-button type="primary" size="small" native-type="submit">
            发送
          </el-button>
        </div>
      </form>

      <div class="children-comments" v-if="comment.children?.length">
        <button type="button" class="tool-link toggle-children-btn" @click="toggleChildren(comment.GUID)">
          {{ repliesOpen(comment.GUID) ? '收起回复' : `查看回复 (${comment.children.length})` }}
        </button>

        <div v-show="repliesOpen(comment.GUID)" class="children-list">
          <div v-for="(child, idx) in comment.children" :key="child.GUID || idx" class="comment-child"
               @click="toggleAction(child.GUID)">
            <div class="avatar-container">
              <el-tooltip :content="'评论于: ' + pubFormatDate(child.CREATE_TIME)" placement="top" effect="light">
                <el-avatar :src="child.AVATAR" class="author-avatar-comment child-avatar"
                           @click.stop="commentAvatarClick(child)"
                           :style="!child.AVATAR ? getAvatarStyle(child.USERNAME) : {}">
                  {{ child.USERNAME?.charAt(0) }}
                </el-avatar>
              </el-tooltip>
            </div>

            <div class="comment-content-block">
              <span class="comment-user-name child-name">{{ child.USERNAME }}<span v-if="isAuthor(child)" class="j-stamp author-stamp">作者</span></span>
              <span v-if="child.RECEIVE_USERNAME && child.RECEIVE_USERCODE !== comment.USERCODE"
                    class="reply-to-text">
                回复 <span class="reply-to-name">@{{ child.RECEIVE_USERNAME }}</span>：
              </span>
              <span class="comment-text">{{ child.TEXT }}</span>
            </div>

            <div class="comment-actions" v-show="activeCommentId === child.GUID">
              <button type="button" class="tool-link" @click.stop="replyComment(comment.GUID, child)">回复</button>
              <button type="button" class="tool-link danger"
                      @click.stop="deleteComment(child.GUID, comment.GUID)"
                      v-if="canDelete(child)">
                删除
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="comments.length > 5" class="comment-more">
      <button type="button" class="tool-link" @click="toggleComments">
        {{ showAllComments ? '收起部分评论' : '展开全部评论' }}
      </button>
    </div>

    <form class="comment-input-area" @submit.prevent="submitComment">
      <p class="input-title">发表评论</p>
      <el-row :gutter="[10, 10]" v-if="showMainGuestForm" class="guest-form-row">
        <el-col :xs="24" :sm="12">
          <el-input v-model="guestInfo.nickname" name="author" autocomplete="name" :prefix-icon="User" placeholder="昵称 (选填)" size="default" clearable />
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-input v-model="guestInfo.email" name="mail" type="email" autocomplete="email" :prefix-icon="Message" placeholder="邮箱 (选填)" size="default" clearable />
        </el-col>
        <el-col :span="24">
          <el-input v-model="guestInfo.website" v-url-hint="guestInfo.website" name="url" type="url" autocomplete="on" :prefix-icon="Link" placeholder="http://" size="default" clearable />
        </el-col>
      </el-row>

      <el-input
          v-model="newComment"
          type="textarea"
          :rows="3"
          placeholder="写下你的优质评论..."
          class="lined-paper"
          @keyup.enter.ctrl="submitComment"
          clearable
          resize="none"
      />
      <div class="comment-submit-row">
        <span class="hint-text">Ctrl + Enter 快捷发送</span>
        <el-button type="primary" size="default" :icon="Edit" native-type="submit">
          发表评论
        </el-button>
      </div>
    </form>
  </section>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { User, Message, Link, Edit } from '@element-plus/icons-vue'
import { useUserStore } from "@/stores/main/user.js";
import { adminUserCode } from "@/config/vue-config.js";
import { pubOpenUser } from "@/utils/blogUtil.js";
import { getAvatarStyle } from "@/utils/avatarStyle.js";
import {
  buildChildrenData,
  ele_confirm,
  getGuid,
  sendAxiosRequest,
  sendAxiosRequestChecked,
  pubFormatDate,
  getCurrentUserAdminObject
} from "@/utils/common.js";

const props = defineProps({
  // blog：文章评论；note：便签墙随记的评论
  kind: { type: String, default: 'blog' },
  // 文章或随记的编号
  targetId: { type: [String, Number], required: true },
  // 文章或随记作者，用于“作者”章和默认的被回复人
  authorCode: { type: String, default: '' },
  authorName: { type: String, default: '' },
  title: { type: String, default: '互动评论' }
})
const emit = defineEmits(['count'])

// 两种评论的接口地址和字段名不同，其余逻辑完全一样
const API = {
  blog: {
    targetField: 'BLOGID',
    list: (id) => ['/blog-api/blog/getComment', { blogId: id }],
    add: (comment) => ['/blog-api/blog/addComment', { blogComment: comment }],
    remove: (id) => ['/blog-api/blog/deleteComment', { blogGuid: id }],
  },
  note: {
    targetField: 'COMMUNITYID',
    list: (id) => ['/blog-api/community/getComment', { communityId: id }],
    add: (comment) => ['/blog-api/community/addComment', { comment }],
    remove: (id) => ['/blog-api/community/deleteComment', { commentGuid: id }],
  }
}
const api = computed(() => API[props.kind] || API.blog)

const router = useRouter();
const userStore = useUserStore();

// 网址输入框是 type="url"，浏览器默认只提示“请输入网址”，看不出缺的是 http:// 或 https://。
// 格式不对时换成明确的提示；内容一变（包括点清除按钮）就清掉，交给浏览器重新校验，
// 否则自定义错误会一直保留，清空后也无法提交。
const URL_HINT = '网址需要以 http:// 或 https:// 开头，例如 https://example.com';
const urlInputOf = (el) => (el.tagName === 'INPUT' ? el : el.querySelector('input'));
const vUrlHint = {
  mounted(el) {
    const input = urlInputOf(el);
    if (!input) return;
    input.addEventListener('invalid', () => {
      if (input.validity.typeMismatch) input.setCustomValidity(URL_HINT);
    });
  },
  updated(el, binding) {
    if (binding.value !== binding.oldValue) urlInputOf(el)?.setCustomValidity('');
  }
};

const comments = ref([]);
const loaded = ref(false);
const showAllComments = ref(false);
const newComment = ref("");
const replyInputVisible = ref({});
const replyInputs = ref({});
const replyTargets = ref({});
const isChildrenVisible = ref({});
const activeCommentId = ref("");

// 访客上次填写的昵称、邮箱、网址，文章和随记共用
const getLocalGuestData = () => {
  try {
    const cached = localStorage.getItem('blog_guest_info');
    if (cached) return { nickname: '', email: '', website: '', ...JSON.parse(cached) };
  } catch (e) {}
  return { nickname: '', email: '', website: '' };
};
const guestInfo = ref(getLocalGuestData());

const showMainGuestForm = computed(() => {
  if (userStore.userBean?.code) return false;
  return !Object.values(replyInputVisible.value).some(visible => visible === true);
});

// 评论总数包含楼中楼
const commentTotal = computed(() => comments.value.reduce((sum, c) => sum + 1 + (c.children?.length || 0), 0));
watch(commentTotal, (total) => emit('count', total), { immediate: true });

const visibleComments = computed(() => showAllComments.value ? comments.value : comments.value.slice(0, 5));

const load = async (id) => {
  loaded.value = false;
  comments.value = [];
  replyInputVisible.value = {};
  replyInputs.value = {};
  replyTargets.value = {};
  isChildrenVisible.value = {};
  const [url, body] = api.value.list(id);
  const result = await sendAxiosRequest(url, body);
  if (result && !result.isError) {
    comments.value = buildChildrenData(result.result || []);
  }
  loaded.value = true;
};
watch(() => props.targetId, (id) => id && load(id), { immediate: true });

// 评论者是文章或随记作者时盖一个“作者”章
const isAuthor = (c) => !!c.USERCODE && c.USERCODE === props.authorCode;

// 删除权限与后台一致：自己的、管理员可删普通用户的、超级管理员都能删
const canDelete = (c) => {
  const admin = getCurrentUserAdminObject();
  return (userStore?.userBean?.code && c.USERCODE === userStore.userBean.code) ||
      (admin.isAdmin && c.USERCODE !== adminUserCode) ||
      admin.adminLevel === 'superAdmin';
};

function toggleAction(id) {
  activeCommentId.value = activeCommentId.value === id ? "" : id;
}

function toggleComments() {
  showAllComments.value = !showAllComments.value;
}

function commentAvatarClick(comment) {
  // 访客填写的网址只打开 http/https 链接，防止 javascript: 之类的链接执行脚本
  const website = String(comment.USERWEBSITE || '').trim();
  if (/^https?:\/\//i.test(website)) {
    window.open(website, '_blank', 'noopener');
  } else if (comment.USERCODE) {
    pubOpenUser(router, comment.USERCODE);
  }
}

function getCommenterPayload() {
  const userBean = userStore.userBean;
  if (userBean && userBean.code) {
    return { name: userBean.name || "匿名用户", code: userBean.code, avatar: userBean.avatar || "", email: "", website: "" };
  }
  const guest = {
    name: guestInfo.value.nickname.trim() || "访客",
    code: "guest_" + getGuid().substring(0, 8),
    avatar: "",
    email: guestInfo.value.email.trim(),
    website: guestInfo.value.website.trim()
  };
  localStorage.setItem('blog_guest_info', JSON.stringify({
    nickname: guestInfo.value.nickname,
    email: guestInfo.value.email,
    website: guestInfo.value.website
  }));
  return guest;
}

// 发送到后台的评论：去掉只用于页面显示的字段
const postComment = async (oneComment, extra = {}) => {
  const comment = { ...oneComment, ...extra };
  delete comment.children; delete comment.CREATE_TIME; delete comment.AVATAR;
  const [url, body] = api.value.add(comment);
  return sendAxiosRequest(url, body);
};

async function submitComment() {
  const value = newComment.value.trim();
  if (!value) {
    ElMessage.warning("评论内容不能为空哦");
    return;
  }
  const userPayload = getCommenterPayload();
  const oneComment = {
    GUID: getGuid(), [api.value.targetField]: props.targetId,
    USERNAME: userPayload.name, USERCODE: userPayload.code,
    RECEIVE_USERCODE: props.authorCode, RECEIVE_USERNAME: props.authorName,
    TEXT: value, CREATE_TIME: "刚刚", AVATAR: userPayload.avatar,
    USEREMAIL: userPayload.email, USERWEBSITE: userPayload.website, children: [],
  };
  const result = await postComment(oneComment);
  if (result && !result.isError) {
    comments.value.unshift(oneComment);
    ElMessage.success("评论发表成功！");
    // 给作者的消息由后台在保存评论时发送，匿名评论也能通知到
  } else {
    ElMessage.error(result?.errMsg || "发表评论出错");
  }
  newComment.value = "";
}

function replyComment(parentGuid, targetComment) {
  if (replyInputVisible.value[parentGuid] && replyTargets.value[parentGuid]?.GUID === targetComment.GUID) {
    replyInputVisible.value[parentGuid] = false;
    replyInputs.value[parentGuid] = "";
    replyTargets.value[parentGuid] = null;
  } else {
    Object.keys(replyInputVisible.value).forEach(key => replyInputVisible.value[key] = false);
    replyInputVisible.value[parentGuid] = true;
    replyTargets.value[parentGuid] = targetComment;
    isChildrenVisible.value[parentGuid] = true;
  }
}

async function submitReply(parentGuid) {
  const value = (replyInputs.value[parentGuid] || "").trim();
  if (!value) return;
  const parentComment = comments.value.find((c) => c.GUID === parentGuid);
  if (!parentComment) return;
  const userPayload = getCommenterPayload();
  const targetUser = replyTargets.value[parentGuid] || parentComment;
  if (!parentComment.children) parentComment.children = [];
  const oneComment = {
    GUID: getGuid(), [api.value.targetField]: props.targetId, SUPERGUID: parentComment.GUID,
    USERNAME: userPayload.name, USERCODE: userPayload.code,
    RECEIVE_USERCODE: targetUser.USERCODE, RECEIVE_USERNAME: targetUser.USERNAME,
    CREATE_TIME: "刚刚", AVATAR: userPayload.avatar,
    USEREMAIL: userPayload.email, USERWEBSITE: userPayload.website, TEXT: value,
  };
  // 告诉后台回复的是哪条评论，用于给对方发站内消息或邮件（匿名访客靠留下的邮箱）
  const result = await postComment(oneComment, { REPLY_TO: targetUser.GUID });
  if (result && !result.isError) {
    parentComment.children.push(oneComment);
    ElMessage.success("回复成功！");
  } else {
    ElMessage.error(result?.errMsg || "回复失败");
  }
  replyInputs.value[parentGuid] = "";
  replyInputVisible.value[parentGuid] = false;
  replyTargets.value[parentGuid] = null;
  isChildrenVisible.value[parentGuid] = true;
}

function deleteComment(commentId, parentId = null) {
  ele_confirm("确定要删除这条评论吗?", async () => {
    const [url, body] = api.value.remove(commentId);
    const result = await sendAxiosRequestChecked(url, body, "删除评论失败");
    if (!result) return;
    if (parentId) {
      const parent = comments.value.find(c => c.GUID === parentId);
      if (parent) parent.children = parent.children.filter(item => item["GUID"] !== commentId);
    } else {
      comments.value = comments.value.filter(item => item["GUID"] !== commentId);
    }
    ElMessage.success("评论已删除");
  })
}

// 楼中楼默认展开；手动点过“收起回复/查看回复”的按手动结果
function repliesOpen(commentId) {
  if (commentId in isChildrenVisible.value) return isChildrenVisible.value[commentId];
  return true;
}

function toggleChildren(commentId) {
  isChildrenVisible.value[commentId] = !repliesOpen(commentId);
}
</script>

<style scoped>
/* 文字按钮 */
.tool-link {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 2px 6px;
  border: none;
  background: none;
  font-size: 13px;
  color: var(--j-pen);
  cursor: pointer;
  text-decoration: underline;
  text-decoration-color: transparent;
  text-underline-offset: 3px;
  transition: text-decoration-color 0.15s;
}

.tool-link:hover {
  text-decoration-color: currentColor;
}

.tool-link.danger {
  color: var(--j-stamp);
}

/* 暂无评论：楷体一句话 + 一支铅笔 */
.comment-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin: 0;
  padding: 18px 0 8px;
  font-family: var(--j-hand);
  font-size: 16px;
  color: var(--j-muted);
}

.pencil-icon {
  width: 20px;
  height: 20px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.5;
  stroke-linecap: round;
  stroke-linejoin: round;
}

/* ====== 评论区 ====== */
.comment-card {
  position: relative;
  display: flex;
  flex-direction: column;
  padding: 20px 18px 18px;
  background: var(--j-paper-warm);
  border: 1px solid var(--j-rule);
  margin-top: 28px;
  box-sizing: border-box;
  scroll-margin-top: 20px;
}

.comment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 1px dashed var(--j-rule-strong);
}

.comment-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 19px;
}

.comment-count-badge {
  min-width: 20px;
  padding: 0 6px;
  border-radius: 10px;
  background: var(--j-stamp);
  color: #fff;
  font-family: var(--j-sans);
  font-size: 12px;
  line-height: 18px;
  text-align: center;
}

.comment-item {
  padding: 6px 0 10px;
  border-bottom: 1px dashed var(--j-rule);
}

.comment-main-row,
.comment-child {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 8px 6px;
  border-radius: 3px;
  transition: background-color 0.2s ease;
  cursor: pointer;
}

.comment-main-row:hover,
.comment-child:hover {
  background-color: rgba(250, 216, 96, 0.18);
}

.avatar-container {
  flex-shrink: 0;
  margin-top: 2px;
}

.author-avatar-comment {
  width: 34px !important;
  height: 34px !important;
}

.child-avatar {
  width: 26px !important;
  height: 26px !important;
  font-size: 12px;
}

.comment-content-block {
  flex: 1;
  min-width: 0;
  line-height: 1.65;
}

.comment-user-name {
  margin-right: 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--j-pen);
}

.child-name {
  color: #4f8a5b;
}

.author-stamp {
  margin: 0 4px 0 6px;
  font-size: 11px;
  line-height: 15px;
  vertical-align: 1px;
}

.comment-text {
  font-size: 14px;
  color: var(--j-ink);
  word-wrap: break-word;
  word-break: break-all;
}

.reply-to-text {
  margin: 0 4px;
  font-size: 13px;
  color: var(--j-muted);
}

.reply-to-name {
  color: var(--j-pen);
  font-weight: 600;
}

.comment-actions {
  flex-shrink: 0;
  display: flex;
  gap: 2px;
}

.reply-input-wrapper {
  margin-top: 6px;
  padding: 0 0 8px 44px;
}

.reply-action-group {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.children-comments {
  margin: 2px 0 0 44px;
}

.toggle-children-btn {
  padding-left: 0;
}

/* 楼中楼左边一道铅笔线 */
.children-list {
  margin-top: 6px;
  padding-left: 10px;
  border-left: 1.5px solid rgba(110, 100, 85, 0.35);
}

.comment-more {
  margin-top: 10px;
  text-align: center;
}

.comment-input-area {
  margin-top: 18px;
}

.input-title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 0 12px;
  font-family: var(--j-hand);
  font-size: 16px;
  color: var(--j-ink-soft);
}

.input-title::after {
  content: "";
  flex: 1;
  border-top: 1px dashed var(--j-rule-strong);
}

.guest-form-row {
  margin-bottom: 10px;
}

/* 评论输入框：横线信纸，横线跟着文字一起滚动 */
.lined-paper :deep(.el-textarea__inner) {
  padding-top: 4px;
  line-height: 26px;
  background-color: var(--j-paper);
  background-image: repeating-linear-gradient(transparent 0 25px, var(--j-rule) 25px 26px);
  background-position: 0 4px;
  background-attachment: local;
}

.comment-submit-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
}

.hint-text {
  font-size: 12px;
  color: var(--j-muted);
}

@media (max-width: 991px) {
  .comment-card {
    margin-top: 18px;
  }
}
</style>
