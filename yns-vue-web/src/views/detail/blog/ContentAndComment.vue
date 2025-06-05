<template>
  <el-row :gutter="20" class="article-view-row" justify="space-between">
    <!-- 正文 -->
    <el-col :xs="24" :sm="24" :md="18" :lg="19">
      <el-card shadow="hover" class="article-card">
        <div class="article-header">
          <h2>{{ blogContent.BLOG_TITLE }}</h2>
          <div style="display: flex; gap: 1px;" v-if="blogContent.USERCODE==userStore.userBean.code">
            <el-button size="small" type="warning" plain @click="editorVisible = true">
              编辑文章
            </el-button>
            <el-button size="small" type="danger" plain @click="deleteArticle">
              删除文章
            </el-button>
          </div>
        </div>
        <ArticleEditor :isReadOnly="true" :content="blogContent.MAINTEXT" />
      </el-card>
    </el-col>

    <!-- 评论 -->
    <el-col :xs="24" :sm="24" :md="6" :lg="5">
      <el-card shadow="hover" class="comment-card">
        <h3 style="margin-bottom: 10px;">评论</h3>
        <div v-for="(comment, i) in visibleComments" :key="comment.GUID || i" class="comment-item">
          <div class="comment-main-row">
            <el-tag type="info" size="small">{{ comment.USERNAME }}</el-tag>
            <span class="comment-text">{{ comment.TEXT }}</span>
            <el-button
                type="text"
                size="small"
                :class="comment.USERCODE == userStore.userBean.code ? 'reply-btn-red' : 'reply-btn'"
                @click="toggleReplyInput(comment.GUID, comment.USERCODE)"
            >
              {{ comment.USERCODE == userStore.userBean.code ? '删除' : '回复' }}
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
                type="text"
                size="small"
                class="toggle-children-btn"
                @click="toggleChildren(comment.GUID)"
            >
              {{ isChildrenVisible[comment.GUID] ? '收起回复' : `查看回复 (${comment.children.length})` }}
            </el-button>
            <div v-show="isChildrenVisible[comment.GUID]" class="children-list">
              <div v-for="(child, idx) in comment.children" :key="child.GUID || idx" class="comment-child">
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
      </el-card>

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
    </el-col>
  </el-row>

  <!-- 编辑器弹窗 -->
  <el-dialog
      v-model="editorVisible"
      title="文章编辑"
      width="900px"
      top="2vh"
      :close-on-click-modal="false"
  >
    <ArticleEditor
        :title="blogContent.BLOG_TITLE"
        :content="blogContent.MAINTEXT"
        @submit="handleEditorSubmit"
        @cancel="editorVisible = false"
    />
  </el-dialog>
</template>

<script setup>
import {ref, computed, watch} from "vue";
import {useRoute, useRouter} from "vue-router";
import {useUserStore} from "@/stores/main/user.js";
import {useBlogContentStore} from "@/stores/detail/blog.js";
import ArticleEditor from "@/components/detail/ArticleEditor.vue";
import {ElMessageBox, ElMessage} from "element-plus";
import {buildChildrenData, ele_confirm, getGuid, sendAxiosRequest} from "@/utils/common.js";

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

const blogContent = ref({});
const blogComment = ref([]);
const editorVisible = ref(false);

const showAllComments = ref(false);
const newComment = ref("");

// 控制回复输入框显示，key:评论id，value:bool
const replyInputVisible = ref({});
// 每个回复输入框的文本，key:评论id，value:string
const replyInputs = ref({});
// 子评论是否展开
const isChildrenVisible = ref({});

const handleEditorSubmit = ({title, content}) => {
  let result = sendAxiosRequest("/blog-api/blog/updateBlog", {
    guid: contentGuid.value,
    title,
    content,
  });
  if (result && !result.isError) {
    blogContent.value.BLOG_TITLE = title;
    blogContent.value.MAINTEXT = content;
    ElMessage.success("已修改");
    editorVisible.value = false;
  } else {
    ElMessage.error("修改失败");
  }
};

const loadContentAndComments = async (guid) => {

  let result = await sendAxiosRequest("/blog-api/blog/getBlog", {blogId: guid});
  if (result && !result.isError) {
    blogContent.value = result?.result?.[0] || {};
  }
  //获取评论
  result = await sendAxiosRequest("/blog-api/blog/getComment", {blogId: guid});
  if (result && !result.isError) {
    blogComment.value = buildChildrenData(result.result);
  }
  // 初始化控制状态
  replyInputVisible.value = {};
  replyInputs.value = {};
  isChildrenVisible.value = {};
};

loadContentAndComments(contentGuid.value);

const visibleComments = computed(() => {
  return showAllComments.value ? blogComment.value : blogComment.value.slice(0, 2);
});

watch(
    () => props.blogId,
    (newGuid) => {
      contentGuid.value = newGuid;
      loadContentAndComments(newGuid);
    }
);
watch(
    () => route.query.g,
    (newGuid) => {
      contentGuid.value = newGuid;
      loadContentAndComments(newGuid);
    }
);

function toggleComments() {
  showAllComments.value = !showAllComments.value;
}

function submitComment() {
  debugger;
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
      children: [],
    }
    blogComment.value.push(oneComment);
    //没有children字段,只是前台需要,所有传递到后台之前删除该字段
    delete oneComment.children;
    let result = sendAxiosRequest("/blog-api/blog/addComment", {blogComment: oneComment})
    newComment.value = "";
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
  debugger;
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
      TEXT: value,
    }
    parentComment.children.push(oneComment);
    let result = sendAxiosRequest("/blog-api/blog/addComment", {blogComment: oneComment})
    replyInputs.value[commentId] = "";
    replyInputVisible.value[commentId] = false;
    isChildrenVisible.value[commentId] = true;
  }
}


function toggleChildren(commentId) {
  isChildrenVisible.value[commentId] = !isChildrenVisible.value[commentId];
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
.article-view-row {
  margin: 0;
  padding: 0;
  height: 100%;
}

.article-card {
  height: 80vh; /* 限制高度 */
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
</style>
