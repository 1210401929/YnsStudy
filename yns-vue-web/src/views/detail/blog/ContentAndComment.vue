<template>
  <el-row :gutter="20" class="article-view-row" justify="space-between">
    <!-- 正文 -->
    <el-col :xs="24" :sm="24"
            :md="showComment ? 15 : 23"
            :lg="showComment ? 15 : 23">
      <el-card shadow="hover" class="article-card">
        <div class="author-info">
          <el-avatar
              :src="blogContent.AVATAR"
              size="large"
              class="author-avatar"
              alt="用户头像"
              @click="avatarClick(blogContent)"
              title="查看发布者信息"
          >
            {{ blogContent.USERNAME?.charAt(0) }}
          </el-avatar>
          <div class="author-text" @click="avatarClick(blogContent)" title="查看发布者信息">
            <div class="author-name">{{ blogContent.USERNAME || '匿名用户' }}</div>
            <div class="author-tagline">发布时间: {{ pubFormatDate(blogContent.CREATE_TIME) }}</div>
          </div>
        </div>
        <div class="article-header">
          <h2>{{ blogContent.BLOG_TITLE }}</h2>
          <div style="display: flex; gap: 1px;"
          >
            <el-button size="small" type="primary" plain @click="openOneBlog"
                       v-if="route.name!=='oneBlog'">
              专注模式
            </el-button>
            <!-- 显示逻辑 1.自己发布的内容  2.非超级管理员的发帖,允许普通管理员编辑 3.当前登录用户是超级管理员 -->
            <el-button size="small" type="warning" plain @click="editorVisible = true"
                       v-if="(userStore?.userBean?.code && blogContent.USERCODE===userStore.userBean.code)
                            ||(getCurrentUserAdminObject().isAdmin && blogContent.USERCODE!==adminUserCode)
                            ||getCurrentUserAdminObject().adminLevel==='superAdmin'">
              编辑文章
            </el-button>
            <!-- 显示逻辑 1.自己发布的内容  2.非超级管理员的发帖,允许普通管理员删除 3.当前登录用户是超级管理员 -->
            <el-button size="small" type="danger" plain @click="deleteArticle"
                       v-if="(userStore?.userBean?.code && blogContent.USERCODE===userStore.userBean.code)
                            ||(getCurrentUserAdminObject().isAdmin && blogContent.USERCODE!==adminUserCode)
                            ||getCurrentUserAdminObject().adminLevel==='superAdmin'">
              删除文章
            </el-button>
          </div>
        </div>
        <ArticleEditor :isReadOnly="true" :content="blogContent.MAINTEXT"/>
      </el-card>
    </el-col>
    <!-- 评论或悬浮按钮 -->
    <el-col :xs="24" :sm="24" :md="showComment ? 9 : 1" :lg="showComment ? 9 : 1">
      <!-- 评论区 -->
      <el-card v-if="showComment" shadow="hover" class="comment-card">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;">
          <h3>评论</h3>
          <el-button type="info" link :icon="Right" @click="showComment = false">
            缩回评论
          </el-button>
        </div>

        <div v-for="(comment, i) in visibleComments" :key="comment.GUID || i" class="comment-item">
          <div class="comment-main-row">
            <el-tooltip
                :content="'评论于: '+pubFormatDate(comment.CREATE_TIME)"
                placement="top"
                effect="light"
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
            </el-tooltip>
            <el-tag type="info" size="small">{{ comment.USERNAME }}</el-tag>
            <span class="comment-text">{{ comment.TEXT }}</span>
            <el-button
                link
                type="primary"
                size="small"
                :class="comment.USERCODE === userStore.userBean.code ? 'reply-btn-red' : 'reply-btn'"
                @click="toggleReplyInput(comment.GUID, comment.USERCODE)"
            >
              {{ comment.USERCODE === userStore.userBean.code ? '删除' : '回复' }}
            </el-button>
          </div>
          <div v-if="replyInputVisible[comment.GUID]" class="reply-input">
            <el-input
                v-model="replyInputs[comment.GUID]"
                placeholder="写下你的回复..."
                size="small"
                @keyup.enter="submitReply(comment.GUID)"
                clearable
            />
            <el-button
                type="primary"
                size="small"
                @click="submitReply(comment.GUID)"
                style="margin-top: 6px; width: 100%"
            >
              发送
            </el-button>
          </div>

          <div class="children-comments" v-if="comment.children?.length">
            <el-button
                link
                type="primary"
                size="small"
                class="toggle-children-btn"
                @click="toggleChildren(comment.GUID)"
            >
              {{ isChildrenVisible[comment.GUID] ? '收起回复' : `查看回复 (${comment.children.length})` }}
            </el-button>
            <div v-show="isChildrenVisible[comment.GUID]" class="children-list">
              <div v-for="(child, idx) in comment.children" :key="child.GUID || idx" class="comment-child">
                <el-tooltip
                    :content="'评论于: ' + pubFormatDate(child.CREATE_TIME)"
                    placement="top"
                    effect="light"
                >
                  <el-avatar
                      :src="child.AVATAR"
                      size="large"
                      class="author-avatar-comment"
                      @click="commentAvatarClick(child)"
                      alt="评论用户头像"
                  >
                    {{ child.USERNAME?.charAt(0) }}
                  </el-avatar>
                </el-tooltip>
                <el-tag type="success" size="small">{{ child.USERNAME }}</el-tag>
                <span class="comment-text">{{ child.TEXT }}</span>
              </div>
            </div>
          </div>
        </div>

        <div v-if="blogComment.length > 2" style="margin-bottom: 10px;">
          <el-button text type="primary" @click="toggleComments">
            {{ showAllComments ? '收起评论' : '展开全部评论' }}
          </el-button>
        </div>

        <!-- 评论输入框 -->
        <div class="comment-input">
          <el-input
              v-model="newComment"
              placeholder="写下你的评论..."
              size="small"
              @keyup.enter="submitComment"
              clearable
          />
          <el-button type="primary" size="small" @click="submitComment" style="margin-top: 10px; width: 100%">
            发表评论
          </el-button>
        </div>
      </el-card>

      <!-- 悬浮操作按钮（评论关闭时显示） -->
      <div v-else class="floating-buttons-top">
        <div class="floating-buttons">
          <el-popover
              placement="left"
              :width="280"
              trigger="click"
              v-model:visible="isTocVisible"
              popper-class="toc-popper"
          >
            <template #reference>
                <el-button
                    circle
                    class="toc-main-btn"
                    :icon="List"
                    title="查看目录"
                />
            </template>

            <div class="toc-container">
              <div class="toc-title">文章目录</div>
              <el-scrollbar max-height="400px">
                <div v-if="tocList.length === 0" class="toc-empty">暂无目录或提取中...</div>
                <div
                    v-for="item in tocList"
                    :key="item.id"
                    class="toc-item"
                    :class="'toc-level-' + item.level"
                    @click="scrollToAnchor(item.id)"
                >
                  {{ item.text }}
                </div>
              </el-scrollbar>
            </div>
          </el-popover>
          <el-tooltip :content="'已点赞: ' + blogLikeNum" placement="left" effect="light">
            <el-button circle :class="blogContent.$userIsLike?'comment-btn-success':'comment-btn'" @click="handleLike">
              👍
            </el-button>
          </el-tooltip>
          <el-tooltip :content="'已收藏: ' + blogCollectNum" placement="left" effect="light">
            <el-button circle :class="blogContent.$userIsCollect?'comment-btn-success':'comment-btn'"
                       class="comment-btn" :icon="Star" @click="handleCollect"/>
          </el-tooltip>
          <el-tooltip content="评论" placement="left" effect="light">
            <!-- 评论按钮 -->
            <el-button circle class="comment-btn" :icon="Comment" @click="showCommentFun"
                       style="position: relative;">
              <!-- 使用el-badge显示评论数量，并通过绝对定位调整其位置 -->
              <el-badge :value="blogComment.length" type="primary" class="comment-badge"
                        style="position: absolute; top: -7px; right: -7px;"/>
            </el-button>
          </el-tooltip>
        </div>
      </div>
    </el-col>
  </el-row>

  <!-- 编辑器弹窗 -->
  <el-dialog
      v-model="editorVisible"
      title="文章编辑"
      width="1000px"
      top="2vh"
      :close-on-click-modal="false"
      destroy-on-close
  >
    <ArticleEditor
        :title="blogContent.BLOG_TITLE"
        :content="blogContent.MAINTEXT"
        :save-type="'edit'"
        :isPublic="blogContent.BLOG_TYPE === 'public' ? true : false"
        @submit="handleEditorSubmit"
        @cancel="editorVisible = false"
    />
  </el-dialog>
</template>


<script setup>
import {ref, computed, watch, nextTick} from "vue";
import {useRoute, useRouter} from "vue-router";
import {useUserStore} from "@/stores/main/user.js";
import {useBlogContentStore} from "@/stores/detail/blog.js";
import ArticleEditor from "@/components/detail/ArticleEditor.vue";
import {Star, Comment, Right, List} from '@element-plus/icons-vue'
import {ElMessage} from "element-plus";
import debounce from 'lodash/debounce'
// 1. 引入需要的图标

import {
  buildChildrenData,
  ele_confirm,
  getGuid,
  sendAxiosRequest,
  pubFormatDate,
  sendNotifications, getCurrentUserAdminObject
} from "@/utils/common.js";
import {adminUserCode} from "@/config/vue-config.js";
import {pubOpenOneBlog, pubOpenUser} from "@/utils/blogUtil.js";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();
const blogContentStore = useBlogContentStore();
const props = defineProps({
  blogId: String
})

const contentGuid = ref(route.query.g);
//当不作为路由组件使用时,如果父页面传递了blogId,直接使用blogId
if (props.blogId) {
  contentGuid.value = props.blogId;
}
//该文章点赞数
const blogLikeNum = ref(0);
//该文章收藏数
const blogCollectNum = ref(0);
const blogContent = ref({});
const blogComment = ref([]);
const editorVisible = ref(false);

const showAllComments = ref(false);
const newComment = ref("");

const showComment = ref(false);

const emit = defineEmits(['loaded'])

const showCommentFun = () => {
  showComment.value = true;
  nextTick(() => {
    const commentSection = document.querySelector('.comment-card'); // 找到评论区域
    if (commentSection) {
      const rect = commentSection.getBoundingClientRect();
      //判断该元素,是否已经完全显示在屏幕内
      const isVisible = (
          rect.top >= 0 &&
          rect.bottom <= (window.innerHeight || document.documentElement.clientHeight)
      );

      if (!isVisible) {
        commentSection.scrollIntoView({
          behavior: 'smooth',
          block: 'start'
        });
      }
    }
  });
}

// 控制回复输入框显示，key:评论id，value:bool
const replyInputVisible = ref({});
// 每个回复输入框的文本，key:评论id，value:string
const replyInputs = ref({});
// 子评论是否展开
const isChildrenVisible = ref({});


const handleEditorSubmit = ({blog_type, title, content}) => {

  let result = sendAxiosRequest("/blog-api/blog/updateBlog", {
    guid: contentGuid.value,
    title,
    blog_type,
    content,
  });
  if (result && !result.isError) {
    blogContent.value.BLOG_TITLE = title;
    blogContent.value.MAINTEXT = content;
    blogContent.value.BLOG_TYPE = blog_type;
    blogContentStore.blogContents.forEach(item => {
      if (item["GUID"] === contentGuid.value) {
        item["BLOG_TITLE"] = title;
        item["BLOG_TYPE"] = blog_type;
      }
    });
    ElMessage.success("已修改");
    editorVisible.value = false;
  } else {
    ElMessage.error("修改失败");
  }
};

const loadContentAndComments = async (guid) => {
  let result = await sendAxiosRequest("/blog-api/blog/getBlog", {blogId: guid});
  if (result && !result.isError && result?.result?.[0]) {
    blogContent.value = result.result[0];
  } else {
    ElMessage.error("该文章为私密或已删除");
    return false;
  }
  //获取评论
  result = await sendAxiosRequest("/blog-api/blog/getComment", {blogId: guid});
  if (result && !result.isError) {
    blogComment.value = buildChildrenData(result.result);
  }
  //获取点赞收藏
  result = await sendAxiosRequest("/blog-api/blog/getLikeAndCollectByBlogId", {blogId: guid});
  if (result && !result.isError) {
    let userBean = userStore.userBean;
    let likeNum = 0, collectNum = 0;
    //处理该用户是否已经点赞该文章
    result.result.forEach(item => {
      //计算点赞,收藏数
      if (item["TYPE"] === "like") likeNum++;
      else if (item["TYPE"] === "collect") collectNum++;
      //判断是否是当前登录用户的点赞收藏
      if (item["USERCODE"] === userBean.code) {
        //该用户已经点赞
        if (item["TYPE"] === "like") {
          blogContent.value.$userIsLike = true;
          //该用户已经收藏
        } else if (item["TYPE"] === "collect") {
          blogContent.value.$userIsCollect = true;
        }
      }
    });
    blogLikeNum.value = likeNum;
    blogCollectNum.value = collectNum;
  }
  //加载数据后方法
  emit('loaded', {blogContent: blogContent.value});
  // 初始化控制状态
  replyInputVisible.value = {};
  replyInputs.value = {};
  isChildrenVisible.value = {};
};
//当博客内容有值时,提取正文内容标题 作为目录
watch(() => blogContent.value.MAINTEXT, () => {
  generateToc();
});

//目录
//目录数据
const tocList = ref([]);
//默认是否显示目录
const isTocVisible = ref(true);

// 3. 提取文章内的标题 (h1-4)
const generateToc = () => {
  nextTick(() => {
    //渲染的内容在 .editor-container 内部
    const contentEl = document.querySelector('.editor-container');
    if (!contentEl) return;

    // 获取所有标题
    const headings = contentEl.querySelectorAll('h1, h2,h3,h4');
    const tempToc = [];
    headings.forEach((el, index) => {
      const titleId = `toc-anchor-${index}`;
      el.id = titleId; // 为正文标题添加ID用于跳转
      tempToc.push({
        id: titleId,
        text: el.innerText,
        level: parseInt(el.tagName.replace('H', ''))
      });
    });
    tocList.value = tempToc;
  });
};

// 4. 监听文章内容变化时生成目录
watch(() => blogContent.value.MAINTEXT, () => {
  generateToc();
}, {deep: true});

// 5. 滚动跳转方法
const scrollToAnchor = (anchorId) => {
  const target = document.getElementById(anchorId);
  if (target) {
    target.scrollIntoView({behavior: 'smooth', block: 'start'});
    isTocVisible.value = false; // 点击后关闭弹窗
  }
};

//点赞
function handleLike() {
  let userBean = userStore.userBean;
  if (!userBean || !userBean.code) {
    ElMessage.error("请先登录!");
    return false;
  }
  //如果该用户已经点赞
  if (blogContent.value.$userIsLike) {
    ElMessage.success("已取消点赞");
    blogLikeNum.value--;
    blogContent.value.$userIsLike = false;
    sendAxiosRequest("/blog-api/blog/noGiveLikeBlog", {blogId: contentGuid.value});
  } else {
    ElMessage.success("点赞成功");
    blogLikeNum.value++;
    blogContent.value.$userIsLike = true;
    sendAxiosRequest("/blog-api/blog/giveLikeBlog", {blogId: contentGuid.value});
    //发送消息
    sendNotifications(userBean.code, blogContent.value.USERCODE, "giveLike", null, `${userBean.name}点赞了你的作品《${blogContent.value.BLOG_TITLE}》`)
  }
}

//收藏
function handleCollect() {
  let userBean = userStore.userBean;
  if (!userBean || !userBean.code) {
    ElMessage.error("请先登录!");
    return false;
  }
  //如果该用户已经收藏
  if (blogContent.value.$userIsCollect) {
    ElMessage.success("已取消收藏");
    blogCollectNum.value--;
    blogContent.value.$userIsCollect = false;
    sendAxiosRequest("/blog-api/blog/noCollectBlog", {blogId: contentGuid.value});
  } else {
    ElMessage.success("收藏成功");
    blogCollectNum.value++;
    blogContent.value.$userIsCollect = true;
    sendAxiosRequest("/blog-api/blog/collectBlog", {blogId: contentGuid.value});
    //发送消息
    sendNotifications(userBean.code, blogContent.value.USERCODE, "collect", null, `${userBean.name}收藏了你的作品《${blogContent.value.BLOG_TITLE}》`)
  }
}


//防抖
const loadContentAndCommentsDebounced = debounce((guid) => {
  loadContentAndComments(guid);
}, 100);

//默认执行一次加载数据
if (contentGuid.value) {
  loadContentAndCommentsDebounced(contentGuid.value);
}

const visibleComments = computed(() => {
  return showAllComments.value ? blogComment.value : blogComment.value.slice(0, 2);
});

watch(
    () => props.blogId,
    (newGuid) => {
      if (newGuid) {
        contentGuid.value = newGuid;
        loadContentAndCommentsDebounced(newGuid);
      }
    }
);

watch(
    () => route.query.g,
    (newGuid) => {
      if (newGuid) {
        contentGuid.value = newGuid;
        loadContentAndCommentsDebounced(newGuid);
      }
    }
);


//发表用户点击用户头像
function avatarClick(blogContent) {
  pubOpenUser(router, blogContent.USERCODE);
}

//评论用户点击用户头像
function commentAvatarClick(comment) {
  pubOpenUser(router, comment.USERCODE);
}

function toggleComments() {
  showAllComments.value = !showAllComments.value;
}

function submitComment() {

  let userBean = userStore.userBean;
  if (!userBean || !userBean.code) {
    ElMessage.error("请先登录!");
    return false;
  }
  const value = newComment.value.trim();
  if (value) {
    const oneComment = {
      GUID: getGuid(),
      BLOGID: contentGuid.value,
      USERNAME: userBean.name || "游客",
      USERCODE: userBean.code || "user",
      TEXT: value,
      CREATE_TIME: "刚刚",
      AVATAR: userBean.avatar,
      children: [],
    }
    blogComment.value.unshift(oneComment);
    //克隆
    let comment = {...oneComment};
    //没有children字段,只是前台需要,所有传递到后台之前删除该字段
    delete comment.children;
    //删除更新日期,后台自动生成
    delete comment.CREATE_TIME;
    //清除头像,评论表没有该字段
    delete comment.AVATAR;
    let result = sendAxiosRequest("/blog-api/blog/addComment", {blogComment: comment})
    newComment.value = "";
    //发送消息
    sendNotifications(userBean.code, blogContent.value.USERCODE, "comment", null, `${userBean.name}评论了你的作品《${blogContent.value.BLOG_TITLE}》`)
  }
}

function toggleReplyInput(commentId, commentUserCode) {
  //如果该评论是用户本身的,则处理删除逻辑
  if (commentUserCode == userStore.userBean.code) {
    ele_confirm("是否删除评论?", () => {
      blogComment.value = blogComment.value.filter(item => item["GUID"] != commentId);
      //调用后台删除评论接口
      let result = sendAxiosRequest("/blog-api/blog/deleteComment", {blogGuid: commentId});
    })
    //如果评论是其他用户的,则控制回复的显示逻辑
  } else {
    replyInputVisible.value[commentId] = !replyInputVisible.value[commentId];
    if (!replyInputVisible.value[commentId]) {
      replyInputs.value[commentId] = "";
    }
  }
}

function submitReply(commentId) {
  let userBean = userStore.userBean;
  if (!userBean || !userBean.code) {
    ElMessage.error("请先登录!");
    return false;
  }
  const value = (replyInputs.value[commentId] || "").trim();
  if (!value) return;
  const parentComment = blogComment.value.find((c) => c.GUID === commentId);
  if (parentComment) {
    if (!parentComment.children) parentComment.children = [];
    const oneComment = {
      GUID: getGuid(),
      BLOGID: contentGuid.value,
      SUPERGUID: parentComment.GUID,
      USERNAME: userBean.name || "游客",
      USERCODE: userBean.code || "user",
      CREATE_TIME: "刚刚",
      AVATAR: userBean.avatar,
      TEXT: value,
    }
    parentComment.children.push(oneComment);
    //克隆
    let comment = {...oneComment};
    //删除更新日期,后台自动生成
    delete comment.CREATE_TIME;
    //清除头像,评论表没有该字段
    delete comment.AVATAR;
    let result = sendAxiosRequest("/blog-api/blog/addComment", {blogComment: comment})
    replyInputs.value[commentId] = "";
    replyInputVisible.value[commentId] = false;
    isChildrenVisible.value[commentId] = true;
    //发送通知
    sendNotifications(oneComment.USERCODE, parentComment.USERCODE, "comment", null, `${oneComment.USERNAME}回复了你的评论《${parentComment.TEXT}》`);
  }
}


function toggleChildren(commentId) {
  isChildrenVisible.value[commentId] = !isChildrenVisible.value[commentId];
}

function openOneBlog() {
  pubOpenOneBlog(router, blogContent.value.GUID)
}


function deleteArticle() {
  ele_confirm("确定要删除这篇文章吗？此操作不可撤销。", () => {
    blogContentStore.blogContents = blogContentStore.blogContents.filter(
        item => item["GUID"] !== contentGuid.value
    );
    sendAxiosRequest("/blog-api/blog/deleteBlog", {guid: contentGuid.value});
    blogComment.value = [];
    ElMessage.success("文章已删除");
    //跳转路由到第一篇文章
    router.push({
      name: "BlogContent",
      query: {g: blogContentStore?.blogContents?.[0]?.GUID || ""},
    });
  })
}

</script>
<style scoped>

.author-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding: 10px 0;
  border-bottom: 1px solid #e0e0e0;

}

.author-avatar-comment {
  width: 20px !important;
  height: 20px !important;
  cursor: pointer;
}

.author-avatar {
  width: 48px !important;
  height: 48px !important;
  font-size: 20px;
  background-color: #f2f2f2;
  cursor: pointer;
}

.author-text {
  display: flex;
  flex-direction: column;
  justify-content: center;
  cursor: pointer;
}

.author-name {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.author-tagline {
  font-size: 12px;
  color: #999;
}

.article-view-row {
  margin: 0;
  padding: 0;
  height: 100%;
}

.article-card {
  /*height: 80vh;  限制高度 */
  min-height: 80vh;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.article-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.comment-card {
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  max-height: 70vh; /* 设置最大高度 */
  overflow-y: auto; /* 超出部分滚动显示 */
}

.comment-item {
  margin-bottom: 10px;
  border-bottom: 1px solid #eee;
  padding-bottom: 8px;
}

.comment-main-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.comment-text {
  font-size: 13px;
  color: #555;
  word-wrap: break-word;
  white-space: normal;
  flex: 1 1 auto;
  min-width: 0;
}

.reply-btn {
  margin-left: auto;
  color: #409eff;
  cursor: pointer;
}

.reply-btn-red {
  margin-left: auto;
  color: #ef4916;
  cursor: pointer;
}

.reply-input {
  margin-top: 6px;
  padding-left: 24px;
}

.children-comments {
  margin-left: 24px;
  margin-top: 6px;
}

.toggle-children-btn {
  margin-bottom: 4px;
  padding: 0;
  font-size: 12px;
  color: #409eff;
  cursor: pointer;
}

.children-list {
  border-left: 2px solid #409eff;
  padding-left: 12px;
}

.comment-child {
  margin-bottom: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.comment-input {
  margin-top: 15px;
}

.floating-buttons-top {
  position: fixed; /* 固定定位 */
  top: 50%; /* 垂直居中 */
  right: 30px; /* 距离右侧 20px，可以根据需要调节 */
  transform: translateY(-50%); /* 让元素垂直中心点对齐屏幕中线 */
  z-index: 1000; /* 保证按钮显示在最上层 */
  display: flex;
  justify-content: center;
  align-items: center;
  height: auto; /* 根据内容自适应高度 */
  width: auto; /* 根据内容自适应宽度 */
}

.floating-buttons {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.floating-buttons .el-button + .el-button {
  margin-left: 0 !important;
}

.floating-buttons .el-button {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.2);
}

.comment-btn {
  width: 40px !important;
  height: 40px !important;
  font-size: 20px
}

.comment-btn:hover {
  background-color: #b7daee;
}

.comment-btn-success {
  width: 40px !important;
  height: 40px !important;
  font-size: 20px;
  background-color: #b7daee;
}

.comment-btn-success:hover {
  background-color: #b7daee;
}

/* 目录主按钮：与点赞区分，使用品牌蓝或深色调 */
.toc-main-btn {
  width: 46px !important; /* 稍微大一点点 */
  height: 46px !important;
  font-size: 22px !important;
  background-color: #409eff !important; /* 明显的蓝色 */
  color: white !important;
  border: 2px solid #fff !important;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.4) !important;
  margin-bottom: 8px; /* 与下方按钮拉开间距 */
  transition: all 0.3s;
}

.toc-main-btn:hover {
  transform: scale(1.1);
  background-color: #66b1ff !important;
}

/* 目录弹出卡片样式 */
.toc-title {
  font-weight: bold;
  font-size: 16px;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #eee;
  color: #333;
}

.toc-item {
  padding: 8px 12px;
  cursor: pointer;
  font-size: 14px;
  color: #606266;
  border-radius: 4px;
  transition: background 0.2s;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.toc-item:hover {
  background-color: #f5f7fa;
  color: #409eff;
}

/* 根据 H 标签级别设置缩进 */
.toc-level-2 {
  padding-left: 12px;
  font-weight: 500;
}

.toc-level-3 {
  padding-left: 24px;
  font-size: 13px;
  color: #909399;
}

.toc-level-4 {
  padding-left: 36px;
  font-size: 12px;
  color: #a8abb2;
}

.toc-empty {
  text-align: center;
  color: #999;
  padding: 20px 0;
}

/* 让正文跳转时不要被顶部的导航栏遮挡（根据你项目是否有顶栏调整） */
:deep(h1), :deep(h2), :deep(h3), :deep(h4) {
  scroll-margin-top: 80px;
}
</style>
