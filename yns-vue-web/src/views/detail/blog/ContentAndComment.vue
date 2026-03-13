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
          <div class="comment-main-row" @click="toggleAction(comment.GUID)">
            <el-tooltip :content="'评论于: '+pubFormatDate(comment.CREATE_TIME)" placement="top" effect="light">
              <el-avatar :src="comment.AVATAR" size="large" class="author-avatar-comment"
                         @click.stop="commentAvatarClick(comment)"> {{ comment.USERNAME?.charAt(0) }}
              </el-avatar>
            </el-tooltip>
            <el-tag type="info" size="small">{{ comment.USERNAME }}</el-tag>
            <span class="comment-text">{{ comment.TEXT }}</span>

            <div class="comment-actions" v-show="activeCommentId === comment.GUID">
              <el-button link type="primary" size="small" class="reply-btn"
                         @click.stop="replyComment(comment.GUID, comment)"> 回复
              </el-button>
              <el-button link type="primary" size="small" class="reply-btn-red"
                         @click.stop="deleteComment(comment.GUID)"
                         v-if="(userStore?.userBean?.code && comment.USERCODE===userStore.userBean.code) ||(getCurrentUserAdminObject().isAdmin && comment.USERCODE!==adminUserCode) ||getCurrentUserAdminObject().adminLevel==='superAdmin'">
                删除
              </el-button>
            </div>
          </div>

          <div v-if="replyInputVisible[comment.GUID]" class="reply-input">
            <el-input
                v-model="replyInputs[comment.GUID]"
                :placeholder="replyTargets[comment.GUID] ? `回复 @${replyTargets[comment.GUID].USERNAME}：` : '写下你的回复...'"
                size="small"
                @keyup.enter="submitReply(comment.GUID)"
                clearable
            />
            <el-button type="primary" size="small" @click="submitReply(comment.GUID)"
                       style="margin-top: 6px; width: 100%">
              发送
            </el-button>
          </div>

          <div class="children-comments" v-if="comment.children?.length">
            <el-button link type="primary" size="small" class="toggle-children-btn"
                       @click="toggleChildren(comment.GUID)">
              {{ isChildrenVisible[comment.GUID] ? '收起回复' : `查看回复 (${comment.children.length})` }}
            </el-button>

            <div v-show="isChildrenVisible[comment.GUID]" class="children-list">
              <div v-for="(child, idx) in comment.children" :key="child.GUID || idx" class="comment-child"
                   @click="toggleAction(child.GUID)">
                <el-tooltip :content="'评论于: ' + pubFormatDate(child.CREATE_TIME)" placement="top" effect="light">
                  <el-avatar :src="child.AVATAR" size="large" class="author-avatar-comment"
                             @click.stop="commentAvatarClick(child)"> {{ child.USERNAME?.charAt(0) }}
                  </el-avatar>
                </el-tooltip>
                <el-tag type="success" size="small">{{ child.USERNAME }}</el-tag>

                <span v-if="child.RECEIVE_USERNAME && child.RECEIVE_USERCODE !== comment.USERCODE"
                      class="reply-to-text">
    回复 <span class="reply-to-name">@{{ child.RECEIVE_USERNAME }}</span>：
  </span>

                <span class="comment-text">{{ child.TEXT }}</span>

                <div class="comment-actions" v-show="activeCommentId === child.GUID">
                  <el-button link type="primary" size="small" class="reply-btn"
                             @click.stop="replyComment(comment.GUID, child)"> 回复
                  </el-button>
                  <el-button link type="primary" size="small" class="reply-btn-red"
                             @click.stop="deleteComment(child.GUID, comment.GUID)"
                             v-if="(userStore?.userBean?.code && child.USERCODE===userStore.userBean.code) ||(getCurrentUserAdminObject().isAdmin && child.USERCODE!==adminUserCode) ||getCurrentUserAdminObject().adminLevel==='superAdmin'">
                    删除
                  </el-button>
                </div>
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
                    :class="[
                      'toc-level-' + item.level,
                      { 'is-active': activeTocId === item.id }
                    ]"
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
import {ref, computed, onUnmounted, watch, nextTick} from "vue";
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
// 记录当前点击(激活)的评论ID，用于控制操作按钮的显示
const activeCommentId = ref("");

// 切换按钮显隐状态
function toggleAction(id) {
  if (activeCommentId.value === id) {
    activeCommentId.value = ""; // 再次点击同一个，收起按钮
  } else {
    activeCommentId.value = id; // 展开当前点击的按钮
  }
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
  debugger;
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

// 3. 修改 generateToc在生成目录后启动观察
const generateToc = () => {
  nextTick(() => {
    const contentEl = document.querySelector('.editor-container');
    if (!contentEl) return;

    const headings = contentEl.querySelectorAll('h1, h2, h3, h4');
    const tempToc = [];
    headings.forEach((el, index) => {
      const titleId = `toc-anchor-${index}`;
      el.id = titleId;
      tempToc.push({
        id: titleId,
        text: el.innerText,
        level: parseInt(el.tagName.replace('H', ''))
      });
    });
    tocList.value = tempToc;

    // 目录生成后，初始化监听
    if (tempToc.length > 0) {
      initObserver();
    }
  });
};

// 4. 监听文章内容变化时生成目录
watch(() => blogContent.value.MAINTEXT, () => {
  generateToc();
}, {deep: true});

const activeTocId = ref(''); // 新增：记录当前激活的目录ID
let observer = null; // 用于存储观察者实例

// 2. 定义滚动监听核心逻辑
const initObserver = () => {
  // 如果之前有观察者，先断开（防止重复绑定）
  if (observer) observer.disconnect();

  observer = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      // 当标题进入视口（这里设定距离顶部 100px 就算激活）
      if (entry.isIntersecting) {
        activeTocId.value = entry.target.id;
      }
    });
  }, {
    // 根元素为 null 表示 viewport，rootMargin 调整触发位置
    // "0px 0px -80% 0px" 表示标题滚动到视口上半部分时触发
    rootMargin: '-80px 0px -70% 0px'
  });

  // 开始观察所有的标题元素
  tocList.value.forEach((item) => {
    const el = document.getElementById(item.id);
    if (el) observer.observe(el);
  });
};

//滚动跳转方法
const scrollToAnchor = (anchorId) => {
  const target = document.getElementById(anchorId);
  if (target) {
    // 1. 设置当前激活的ID
    activeTocId.value = anchorId;

    // 2. 执行平滑滚动
    target.scrollIntoView({behavior: 'smooth', block: 'start'});

    // 3. 延迟关闭目录（给用户一个看到高亮反馈的时间，或者直接关闭）
    // setTimeout(() => {
    //   isTocVisible.value = false;
    // }, 900);
  }
};

/**
 * 监听当前激活的目录项 ID 变化
 * 用于自动滑动右侧目录到可视窗口
 */
watch(activeTocId, (newId) => {
  if (!newId) return;

  // 必须在 nextTick 确保 DOM 已经渲染/更新
  nextTick(() => {
    // 找到当前高亮的那个目录 DOM 元素
    // 注意：这里的选择器要匹配你模板里的 class
    const activeItemEl = document.querySelector(`.toc-item.is-active`);

    if (activeItemEl) {
      // 让目录列表平滑滚动，使高亮项出现在视野内
      activeItemEl.scrollIntoView({
        behavior: 'smooth', // 平滑滚动
        block: 'nearest',   // 如果已经在视野内就不动，不在就滚到最近的边缘
        inline: 'start'
      });
    }
  });
});

// 4. 组件卸载前断开连接
onUnmounted(() => {
  if (observer) observer.disconnect();
});

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

// 记录当前输入框正在回复的目标对象。key: 顶级评论GUID, value: 被回复的评论(可能是顶级也可能是子级)
const replyTargets = ref({});

// 1. 发起顶级评论时，一并把 RECEIVE 字段给到文章作者
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
      RECEIVE_USERCODE: blogContent.value.USERCODE, // 发给文章作者
      RECEIVE_USERNAME: blogContent.value.USERNAME,
      TEXT: value,
      CREATE_TIME: "刚刚",
      AVATAR: userBean.avatar,
      children: [],
    }
    blogComment.value.unshift(oneComment);
    let comment = {...oneComment};
    delete comment.children;
    delete comment.CREATE_TIME;
    delete comment.AVATAR;
    sendAxiosRequest("/blog-api/blog/addComment", {blogComment: comment})
    newComment.value = "";
    sendNotifications(userBean.code, blogContent.value.USERCODE, "comment", null, `${userBean.name}评论了你的作品《${blogContent.value.BLOG_TITLE}》`)
  }
}

// 2. 点击回复按钮：记录目标是谁
function replyComment(parentGuid, targetComment) {
  // 如果再次点击同一个人的回复按钮，且当前输入框开着，则收起
  if (replyInputVisible.value[parentGuid] && replyTargets.value[parentGuid]?.GUID === targetComment.GUID) {
    replyInputVisible.value[parentGuid] = false;
    replyInputs.value[parentGuid] = "";
    replyTargets.value[parentGuid] = null;
  } else {
    // 展开并记录被回复的目标
    replyInputVisible.value[parentGuid] = true;
    replyTargets.value[parentGuid] = targetComment;
    isChildrenVisible.value[parentGuid] = true; // 顺便把子列表展开
  }
}


// 3. 提交回复：带上 RECEIVE 字段
function submitReply(parentGuid) {
  let userBean = userStore.userBean;
  if (!userBean || !userBean.code) {
    ElMessage.error("请先登录!");
    return false;
  }
  const value = (replyInputs.value[parentGuid] || "").trim();
  if (!value) return;

  const parentComment = blogComment.value.find((c) => c.GUID === parentGuid);
  // 获取刚刚记录的回复目标（保底是顶级评论）
  const targetUser = replyTargets.value[parentGuid] || parentComment;

  if (parentComment) {
    if (!parentComment.children) parentComment.children = [];
    const oneComment = {
      GUID: getGuid(),
      BLOGID: contentGuid.value,
      SUPERGUID: parentComment.GUID, // 依然挂在顶级评论下，保持两层嵌套
      USERNAME: userBean.name || "游客",
      USERCODE: userBean.code || "user",
      RECEIVE_USERCODE: targetUser.USERCODE, // 重点：接收方编码
      RECEIVE_USERNAME: targetUser.USERNAME, // 重点：接收方名字
      CREATE_TIME: "刚刚",
      AVATAR: userBean.avatar,
      TEXT: value,
    }
    parentComment.children.push(oneComment);

    let comment = {...oneComment};
    delete comment.CREATE_TIME;
    delete comment.AVATAR;
    sendAxiosRequest("/blog-api/blog/addComment", {blogComment: comment})

    // 清理状态
    replyInputs.value[parentGuid] = "";
    replyInputVisible.value[parentGuid] = false;
    replyTargets.value[parentGuid] = null;
    isChildrenVisible.value[parentGuid] = true;

    // 发送通知：通知目标被回复了
    sendNotifications(oneComment.USERCODE, targetUser.USERCODE, "comment", null, `${oneComment.USERNAME}回复了你的评论《${targetUser.TEXT}》`);
  }
}

// 4. 删除评论：兼容删除子评论的情况
function deleteComment(commentId, parentId = null) {
  ele_confirm("是否删除评论?", () => {
    if (parentId) {
      // 如果传了 parentId，说明删的是子评论
      let parent = blogComment.value.find(c => c.GUID === parentId);
      if (parent) {
        parent.children = parent.children.filter(item => item["GUID"] !== commentId);
      }
    } else {
      // 删顶级评论
      blogComment.value = blogComment.value.filter(item => item["GUID"] !== commentId);
    }
    //调用后台删除评论接口
    sendAxiosRequest("/blog-api/blog/deleteComment", {blogGuid: commentId});
  })
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


/* 调整不同层级的缩进补偿（可选） */
.toc-level-2.is-active {
  padding-left: 9px;
}

.toc-level-3.is-active {
  padding-left: 21px;
}

.toc-level-4.is-active {
  padding-left: 33px;
}

.toc-item {
  /* 基础样式中增加过渡 */
  transition: all 0.2s ease-in-out;
  border-left: 3px solid transparent; /* 预留边框位置防止抖动 */
}

.toc-item.is-active {
  color: #409eff;
  background-color: #ecf5ff;
  font-weight: bold;
  border-left: 3px solid #409eff;
  /* 根据层级微调 padding，确保文字对齐 */
}

.comment-child {
  margin-bottom: 6px;
  display: flex;
  align-items: center;
  flex-wrap: wrap; /* 允许换行 */
  gap: 6px;
}

/* 针对 回复 @xxx 的文字修饰 */
.reply-to-text {
  font-size: 13px;
  color: #909399;
  margin: 0 2px;
}

.reply-to-name {
  color: #409eff;
  cursor: pointer;
}

.reply-to-name:hover {
  text-decoration: underline;
}

/* 让整行鼠标变成小手，并给一点点悬浮底色反馈 */
.comment-main-row, .comment-child {
  cursor: pointer;
  border-radius: 4px;
  padding: 4px;
  transition: background-color 0.2s ease;
}

.comment-main-row:hover, .comment-child:hover {
  background-color: #f7f9fc; /* 非常浅的蓝色/灰色，提升交互感 */
}

/* 控制按钮组靠右对齐 */
.comment-actions {
  margin-left: auto;
  display: flex;
  gap: 8px; /* 按钮之间的间距 */
  align-items: center;
}

/* 覆盖你原本的左边距，因为外层 wrapper 已经做过排版了 */
.reply-btn, .reply-btn-red {
  margin-left: 0 !important;
}
</style>
