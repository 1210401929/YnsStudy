<template>
  <el-row ref="layoutRowRef" :gutter="20" class="article-view-row" justify="space-between" align="top">
    <!-- 左侧文章区域 -->
    <el-col :xs="24" :sm="24"
            :md="17"
            :lg="17"
            class="smooth-col">
      <el-card class="article-card">
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
          <div style="display: flex; gap: 8px;">
            <el-button size="small" type="primary" plain @click="openOneBlog"
                       v-if="route.name!=='oneBlog'">
              专注模式
            </el-button>
            <el-button size="small" type="warning" plain @click="editorVisible = true"
                       v-if="canEditOrDelete">
              编辑文章
            </el-button>
            <el-button size="small" type="danger" plain @click="deleteArticle"
                       v-if="canEditOrDelete">
              删除文章
            </el-button>
          </div>
        </div>

        <ArticleEditor :isReadOnly="true" :content="blogContent.MAINTEXT"/>

        <div class="article-bottom-actions">
          <el-divider>文章完</el-divider>
          <div class="bottom-action-buttons">
            <el-button :type="blogContent.$userIsLike ? 'primary' : 'default'"
                       :plain="!blogContent.$userIsLike"
                       round
                       size="large"
                       @click="handleLike">
              👍 {{ blogContent.$userIsLike ? '已赞' : '点赞' }} {{ blogLikeNum > 0 ? `(${blogLikeNum})` : '' }}
            </el-button>

            <el-button :type="blogContent.$userIsCollect ? 'warning' : 'default'"
                       :plain="!blogContent.$userIsCollect"
                       round
                       size="large"
                       :icon="Star"
                       @click="handleCollect">
              {{ blogContent.$userIsCollect ? '已收藏' : '收藏' }} {{ blogCollectNum > 0 ? `(${blogCollectNum})` : '' }}
            </el-button>

            <el-button type="success"
                       plain
                       round
                       size="large"
                       :icon="Comment"
                       @click="showCommentFun"
                       v-if="!showComment">
              参与讨论 {{ blogComment.length > 0 ? `(${blogComment.length})` : '' }}
            </el-button>
          </div>
        </div>
      </el-card>
    </el-col>

    <!-- 右侧区域：评论区 or 目录区 -->
    <el-col :xs="24" :sm="24" :md="7" :lg="7" class="smooth-col">
      <!-- 核心：独立悬浮的包裹层，纯 JS 操控其 translateY -->
      <div ref="rightSidebarRef" style="will-change: transform;">
        <!-- 评论区 -->
        <el-card v-if="showComment" shadow="hover" class="comment-card">
          <div class="comment-header">
            <h3>互动评论 <span class="comment-count-badge">{{ blogComment.length }}</span></h3>
            <el-button type="info" plain size="small" :icon="Right" @click="showComment = false">
              收起评论
            </el-button>
          </div>

          <div v-for="(comment, i) in visibleComments" :key="comment.GUID || i" class="comment-item">
            <div class="comment-main-row" @click="toggleAction(comment.GUID)">
              <div class="avatar-container">
                <el-tooltip :content="'评论于: '+pubFormatDate(comment.CREATE_TIME)" placement="top" effect="light">
                  <el-avatar :src="comment.AVATAR" class="author-avatar-comment"
                             @click.stop="commentAvatarClick(comment)">
                    {{ comment.USERNAME?.charAt(0) }}
                  </el-avatar>
                </el-tooltip>
              </div>

              <div class="comment-content-block">
                <span class="comment-user-name">{{ comment.USERNAME }}:</span>
                <span class="comment-text">{{ comment.TEXT }}</span>
              </div>

              <div class="comment-actions" v-show="activeCommentId === comment.GUID">
                <el-button link type="primary" size="small" class="reply-btn"
                           @click.stop="replyComment(comment.GUID, comment)">
                  回复
                </el-button>
                <el-button link type="danger" size="small" class="reply-btn-red"
                           @click.stop="deleteComment(comment.GUID)"
                           v-if="(userStore?.userBean?.code && comment.USERCODE===userStore.userBean.code) ||(getCurrentUserAdminObject().isAdmin && comment.USERCODE!==adminUserCode) ||getCurrentUserAdminObject().adminLevel==='superAdmin'">
                  删除
                </el-button>
              </div>
            </div>

            <div v-if="replyInputVisible[comment.GUID]" class="reply-input-wrapper">
              <el-row :gutter="10" v-if="!userStore?.userBean?.code" class="guest-form-row">
                <el-col :span="8">
                  <el-input v-model="guestInfo.nickname" :prefix-icon="User" placeholder="昵称 (选填)" size="small" clearable/>
                </el-col>
                <el-col :span="8">
                  <el-input v-model="guestInfo.email" :prefix-icon="Message" placeholder="邮箱 (选填)" size="small" clearable/>
                </el-col>
                <el-col :span="8">
                  <el-input v-model="guestInfo.website" :prefix-icon="Link" placeholder="网址 (选填)" size="small" clearable/>
                </el-col>
              </el-row>
              <div class="reply-action-group">
                <el-input
                    v-model="replyInputs[comment.GUID]"
                    :placeholder="replyTargets[comment.GUID] ? `回复 @${replyTargets[comment.GUID].USERNAME}：` : '写下你的回复...'"
                    size="small"
                    @keyup.enter="submitReply(comment.GUID)"
                    clearable
                />
                <el-button type="primary" size="small" @click="submitReply(comment.GUID)">
                  发送
                </el-button>
              </div>
            </div>

            <div class="children-comments" v-if="comment.children?.length">
              <el-button link type="primary" size="small" class="toggle-children-btn"
                         @click="toggleChildren(comment.GUID)">
                {{ isChildrenVisible[comment.GUID] ? '收起回复' : `查看回复 (${comment.children.length})` }}
              </el-button>

              <div v-show="isChildrenVisible[comment.GUID]" class="children-list">
                <div v-for="(child, idx) in comment.children" :key="child.GUID || idx" class="comment-child"
                     @click="toggleAction(child.GUID)">
                  <div class="avatar-container">
                    <el-tooltip :content="'评论于: ' + pubFormatDate(child.CREATE_TIME)" placement="top" effect="light">
                      <el-avatar :src="child.AVATAR" class="author-avatar-comment child-avatar"
                                 @click.stop="commentAvatarClick(child)">
                        {{ child.USERNAME?.charAt(0) }}
                      </el-avatar>
                    </el-tooltip>
                  </div>

                  <div class="comment-content-block">
                    <span class="comment-user-name child-name">{{ child.USERNAME }}</span>
                    <span v-if="child.RECEIVE_USERNAME && child.RECEIVE_USERCODE !== comment.USERCODE"
                          class="reply-to-text">
                      回复 <span class="reply-to-name">@{{ child.RECEIVE_USERNAME }}</span>：
                    </span>
                    <span class="comment-text">{{ child.TEXT }}</span>
                  </div>

                  <div class="comment-actions" v-show="activeCommentId === child.GUID">
                    <el-button link type="primary" size="small" class="reply-btn"
                               @click.stop="replyComment(comment.GUID, child)">
                      回复
                    </el-button>
                    <el-button link type="danger" size="small" class="reply-btn-red"
                               @click.stop="deleteComment(child.GUID, comment.GUID)"
                               v-if="(userStore?.userBean?.code && child.USERCODE===userStore.userBean.code) ||(getCurrentUserAdminObject().isAdmin && child.USERCODE!==adminUserCode) ||getCurrentUserAdminObject().adminLevel==='superAdmin'">
                      删除
                    </el-button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-if="blogComment.length > 5" style="margin-top: 10px; text-align: center;">
            <el-button text type="primary" @click="toggleComments">
              {{ showAllComments ? '收起部分评论' : '展开全部评论' }}
            </el-button>
          </div>

          <div class="comment-input-area">
            <el-divider border-style="dashed">发表评论</el-divider>
            <el-row :gutter="10" v-if="showMainGuestForm" class="guest-form-row">
              <el-col :xs="24" :sm="8" :md="8" :lg="8">
                <el-input v-model="guestInfo.nickname" :prefix-icon="User" placeholder="昵称 (选填)" size="default" clearable class="guest-input-item"/>
              </el-col>
              <el-col :xs="24" :sm="8" :md="8" :lg="8">
                <el-input v-model="guestInfo.email" :prefix-icon="Message" placeholder="邮箱 (选填)" size="default" clearable class="guest-input-item"/>
              </el-col>
              <el-col :xs="24" :sm="8" :md="8" :lg="8">
                <el-input v-model="guestInfo.website" :prefix-icon="Link" placeholder="网址 (选填)" size="default" clearable class="guest-input-item"/>
              </el-col>
            </el-row>

            <el-input
                v-model="newComment"
                type="textarea"
                :rows="3"
                placeholder="写下你的优质评论..."
                @keyup.enter.ctrl="submitComment"
                clearable
                resize="none"
            />
            <div class="comment-submit-row">
              <span class="hint-text">Ctrl + Enter 快捷发送</span>
              <el-button type="primary" size="default" :icon="Edit" @click="submitComment">
                发表评论
              </el-button>
            </div>
          </div>
        </el-card>

        <!-- 目录区：固定显示 -->
        <el-card v-else shadow="hover" class="toc-fixed-card">
          <div class="toc-title">文章目录</div>
          <el-scrollbar max-height="600px">
            <div v-if="tocList.length === 0" class="toc-empty">暂无目录或提取中...</div>
            <div v-for="item in tocList" :key="item.id"
                 class="toc-item"
                 :class="['toc-level-' + item.level, { 'is-active': activeTocId === item.id }]"
                 @click="scrollToAnchor(item.id)">
              {{ item.text }}
            </div>
          </el-scrollbar>
        </el-card>
      </div>
    </el-col>
  </el-row>

  <!-- 悬浮按钮：始终显示 -->
  <div class="floating-wrapper">
    <div class="floating-buttons glass-effect">
      <el-tooltip :content="blogContent.$userIsLike ? '取消点赞' : '点赞'" placement="left" effect="dark">
        <div class="btn-wrap">
          <el-button circle :class="['floating-btn', blogContent.$userIsLike ? 'is-active-btn' : '']" @click="handleLike">
            👍
          </el-button>
        </div>
      </el-tooltip>

      <el-tooltip :content="blogContent.$userIsCollect ? '取消收藏' : '收藏'" placement="left" effect="dark">
        <div class="btn-wrap">
          <el-button circle :class="['floating-btn', blogContent.$userIsCollect ? 'is-active-btn' : '']" :icon="Star" @click="handleCollect"/>
        </div>
      </el-tooltip>

      <el-tooltip :content="showComment ? '关闭评论' : '打开评论'" placement="left" effect="dark">
        <div class="btn-wrap">
          <el-button circle class="floating-btn" :icon="Comment" @click="showCommentFun" style="position: relative;">
            <el-badge v-if="blogComment.length > 0" :value="blogComment.length" type="danger" class="floating-badge"/>
          </el-button>
        </div>
      </el-tooltip>
    </div>
  </div>

  <!-- 编辑弹窗 -->
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
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useUserStore } from "@/stores/main/user.js";
import { useBlogContentStore } from "@/stores/detail/blog.js";
import ArticleEditor from "@/components/detail/ArticleEditor.vue";
import { Star, Comment, Right, User, Message, Link, Edit } from '@element-plus/icons-vue'
import { ElMessage } from "element-plus";
import debounce from 'lodash/debounce'

import {
  buildChildrenData,
  ele_confirm,
  getGuid,
  sendAxiosRequest,
  pubFormatDate,
  sendNotifications, getCurrentUserAdminObject
} from "@/utils/common.js";

import { adminUserCode } from "@/config/vue-config.js";
import { pubOpenOneBlog, pubOpenUser } from "@/utils/blogUtil.js";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();
const blogContentStore = useBlogContentStore();
const props = defineProps({
  blogId: String
})

const contentGuid = ref(route.query.g);
if (props.blogId) {
  contentGuid.value = props.blogId;
}

const blogLikeNum = ref(0);
const blogCollectNum = ref(0);
const blogContent = ref({});
const blogComment = ref([]);
const editorVisible = ref(false);

const showAllComments = ref(false);
const newComment = ref("");
const showComment = ref(false);

const getLocalGuestData = () => {
  try {
    const cached = localStorage.getItem('blog_guest_info');
    if (cached) return JSON.parse(cached);
  } catch (e) {}
  return { nickname: '', email: '', website: '' };
};

const guestInfo = ref(getLocalGuestData());
const replyInputVisible = ref({});
const replyInputs = ref({});
const isChildrenVisible = ref({});

const showMainGuestForm = computed(() => {
  if (userStore.userBean?.code) return false;
  return !Object.values(replyInputVisible.value).some(visible => visible === true);
});

const emit = defineEmits(['loaded'])

const showCommentFun = () => {
  showComment.value = !showComment.value;
  nextTick(() => {
    const commentSection = document.querySelector('.comment-card');
    if (commentSection && showComment.value) {
      commentSection.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }
  });
}

const activeCommentId = ref("");

function toggleAction(id) {
  activeCommentId.value = activeCommentId.value === id ? "" : id;
}

const handleEditorSubmit = ({blog_type, title, content}) => {
  let result = sendAxiosRequest("/blog-api/blog/updateBlog", {
    guid: contentGuid.value, title, blog_type, content,
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

  result = await sendAxiosRequest("/blog-api/blog/getComment", {blogId: guid});
  if (result && !result.isError) {
    blogComment.value = buildChildrenData(result.result);
  }

  result = await sendAxiosRequest("/blog-api/blog/getLikeAndCollectByBlogId", {blogId: guid});
  if (result && !result.isError) {
    let userBean = userStore.userBean;
    let likeNum = 0, collectNum = 0;
    result.result.forEach(item => {
      if (item["TYPE"] === "like") likeNum++;
      else if (item["TYPE"] === "collect") collectNum++;
      if (userBean && item["USERCODE"] === userBean.code) {
        if (item["TYPE"] === "like") blogContent.value.$userIsLike = true;
        else if (item["TYPE"] === "collect") blogContent.value.$userIsCollect = true;
      }
    });
    blogLikeNum.value = likeNum;
    blogCollectNum.value = collectNum;
  }
  emit('loaded', {blogContent: blogContent.value});
  replyInputVisible.value = {};
  replyInputs.value = {};
  isChildrenVisible.value = {};
};

watch(() => blogContent.value.MAINTEXT, () => {
  generateToc();
});

const canEditOrDelete = computed(() => {
  if (!blogContent.value.USERCODE || !userStore.userBean?.code) return false;
  const currentUser = userStore.userBean;
  const adminObj = getCurrentUserAdminObject();
  return (blogContent.value.USERCODE === currentUser.code) ||
      (adminObj.adminLevel === 'superAdmin') ||
      (adminObj.isAdmin && blogContent.value.USERCODE !== adminUserCode);
});

const tocList = ref([]);
const activeTocId = ref('');
let observer = null;

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
    if (tempToc.length > 0) initObserver();
  });
};

watch(() => blogContent.value.MAINTEXT, () => {
  generateToc();
}, {deep: true});

const initObserver = () => {
  if (observer) observer.disconnect();
  observer = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) activeTocId.value = entry.target.id;
    });
  }, { rootMargin: '-80px 0px -70% 0px' });

  tocList.value.forEach((item) => {
    const el = document.getElementById(item.id);
    if (el) observer.observe(el);
  });
};

const scrollToAnchor = (anchorId) => {
  const target = document.getElementById(anchorId);
  if (target) {
    activeTocId.value = anchorId;
    target.scrollIntoView({behavior: 'smooth', block: 'start'});
  }
};


// ==================== JS 完全独立悬浮核心逻辑 ====================
const layoutRowRef = ref(null);
const rightSidebarRef = ref(null);
let scrollContainer = null;
let ticking = false;

const doScrollUpdate = () => {
  if (!layoutRowRef.value || !rightSidebarRef.value) return;

  const rowEl = layoutRowRef.value.$el || layoutRowRef.value;
  const sidebarEl = rightSidebarRef.value;

  // 移动端排版自动取消悬浮
  if (window.innerWidth < 992) {
    sidebarEl.style.transform = `translateY(0px)`;
    return;
  }

  // 完全基于视口物理位置计算
  const rowRect = rowEl.getBoundingClientRect();
  const sidebarHeight = sidebarEl.offsetHeight;

  const offsetTop = 20; // 悬浮时距离视口顶部的间距

  // 如果父容器顶部已经超出了视口距离（开始往下滚了）
  if (rowRect.top < offsetTop) {
    let translateY = Math.abs(rowRect.top) + offsetTop;

    // 触底限制
    const maxTranslateY = rowRect.height - sidebarHeight;

    if (maxTranslateY <= 0) {
      sidebarEl.style.transform = `translateY(0px)`;
      return;
    }

    if (translateY > maxTranslateY) {
      translateY = maxTranslateY;
    }

    sidebarEl.style.transform = `translateY(${translateY}px)`;
  } else {
    sidebarEl.style.transform = `translateY(0px)`;
  }
};

const handleScroll = () => {
  if (!ticking) {
    window.requestAnimationFrame(() => {
      doScrollUpdate();
      ticking = false;
    });
    ticking = true;
  }
};

// 自动向外寻找真正能滚动的 DOM 节点
const getScrollContainer = (el) => {
  let parent = el.parentElement;
  while (parent) {
    const style = window.getComputedStyle(parent);
    if (/(auto|scroll|overlay)/.test(style.overflow + style.overflowY)) {
      return parent;
    }
    parent = parent.parentElement;
  }
  return window;
};

onMounted(() => {
  requestAnimationFrame(() => {
    const rowEl = layoutRowRef.value.$el || layoutRowRef.value;
    if (rowEl) {
      scrollContainer = getScrollContainer(rowEl);
      scrollContainer.addEventListener('scroll', handleScroll, { passive: true });
      window.addEventListener('scroll', handleScroll, { passive: true });
    }
  });
});
// =================================================================

onUnmounted(() => {
  if (observer) observer.disconnect();

  // 销毁监听事件
  if (scrollContainer) {
    scrollContainer.removeEventListener('scroll', handleScroll);
  }
  window.removeEventListener('scroll', handleScroll);
});


function handleLike() {
  let userBean = userStore.userBean;
  if (!userBean || !userBean.code) {
    ElMessage.warning("登录后体验更多互动功能！");
    return false;
  }
  if (blogContent.value.$userIsLike) {
    blogLikeNum.value--;
    blogContent.value.$userIsLike = false;
    sendAxiosRequest("/blog-api/blog/noGiveLikeBlog", {blogId: contentGuid.value});
  } else {
    blogLikeNum.value++;
    blogContent.value.$userIsLike = true;
    sendAxiosRequest("/blog-api/blog/giveLikeBlog", {blogId: contentGuid.value});
    const routeUrl = router.resolve({name: 'oneBlog', params: {g: contentGuid.value}}).href;
    sendNotifications(userBean.code, blogContent.value.USERCODE, "giveLike", routeUrl, `${userBean.name}点赞了你的作品《${blogContent.value.BLOG_TITLE}》`)
  }
}

function handleCollect() {
  let userBean = userStore.userBean;
  if (!userBean || !userBean.code) {
    ElMessage.warning("请先登录再收藏哦！");
    return false;
  }
  if (blogContent.value.$userIsCollect) {
    blogCollectNum.value--;
    blogContent.value.$userIsCollect = false;
    sendAxiosRequest("/blog-api/blog/noCollectBlog", {blogId: contentGuid.value});
  } else {
    ElMessage.success("收藏成功，可在个人中心查看");
    blogCollectNum.value++;
    blogContent.value.$userIsCollect = true;
    sendAxiosRequest("/blog-api/blog/collectBlog", {blogId: contentGuid.value});
    const routeUrl = router.resolve({name: 'oneBlog', params: {g: contentGuid.value}}).href;
    sendNotifications(userBean.code, blogContent.value.USERCODE, "collect", routeUrl, `${userBean.name}收藏了你的作品《${blogContent.value.BLOG_TITLE}》`)
  }
}

const loadContentAndCommentsDebounced = debounce((guid) => {
  loadContentAndComments(guid);
}, 100);

if (contentGuid.value) {
  loadContentAndCommentsDebounced(contentGuid.value);
}

const visibleComments = computed(() => {
  return showAllComments.value ? blogComment.value : blogComment.value.slice(0, 5);
});

watch(() => props.blogId, (newGuid) => {
  if (newGuid) {
    contentGuid.value = newGuid;
    loadContentAndCommentsDebounced(newGuid);
  }
});

watch(() => route.query.g, (newGuid) => {
  if (newGuid) {
    contentGuid.value = newGuid;
    loadContentAndCommentsDebounced(newGuid);
  }
});

function avatarClick(blogContent) {
  pubOpenUser(router, blogContent.USERCODE);
}

function commentAvatarClick(comment) {
  if(comment.USERWEBSITE) {
    window.open(comment.USERWEBSITE)
  } else {
    pubOpenUser(router, comment.USERCODE);
  }
}

function toggleComments() {
  showAllComments.value = !showAllComments.value;
}

const replyTargets = ref({});

function getCommenterPayload() {
  let userBean = userStore.userBean;
  if (userBean && userBean.code) {
    return {
      name: userBean.name || "匿名用户",
      code: userBean.code,
      avatar: userBean.avatar || "",
      email: "", website: ""
    }
  } else {
    const finalGuest = {
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
    return finalGuest;
  }
}

async function submitComment() {
  const value = newComment.value.trim();
  if (!value) {
    ElMessage.warning("评论内容不能为空哦");
    return;
  }
  const userPayload = getCommenterPayload();
  const oneComment = {
    GUID: getGuid(), BLOGID: contentGuid.value,
    USERNAME: userPayload.name, USERCODE: userPayload.code,
    RECEIVE_USERCODE: blogContent.value.USERCODE, RECEIVE_USERNAME: blogContent.value.USERNAME,
    TEXT: value, CREATE_TIME: "刚刚", AVATAR: userPayload.avatar,
    USEREMAIL: userPayload.email, USERWEBSITE: userPayload.website, children: [],
  }

  let comment = {...oneComment};
  delete comment.children; delete comment.CREATE_TIME; delete comment.AVATAR;
  let result = await sendAxiosRequest("/blog-api/blog/addComment", {blogComment: comment})
  if(result && !result.isError){
    blogComment.value.unshift(oneComment);
    ElMessage.success("评论发表成功！");
  }else{
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
  const userPayload = getCommenterPayload();
  const parentComment = blogComment.value.find((c) => c.GUID === parentGuid);
  const targetUser = replyTargets.value[parentGuid] || parentComment;

  if (parentComment) {
    if (!parentComment.children) parentComment.children = [];
    const oneComment = {
      GUID: getGuid(), BLOGID: contentGuid.value, SUPERGUID: parentComment.GUID,
      USERNAME: userPayload.name, USERCODE: userPayload.code,
      RECEIVE_USERCODE: targetUser.USERCODE, RECEIVE_USERNAME: targetUser.USERNAME,
      CREATE_TIME: "刚刚", AVATAR: userPayload.avatar,
      USEREMAIL: userPayload.email, USERWEBSITE: userPayload.website, TEXT: value,
    }

    let comment = {...oneComment};
    delete comment.CREATE_TIME; delete comment.AVATAR;
    let result = await sendAxiosRequest("/blog-api/blog/addComment", {blogComment: comment})
    if(result && !result.isError){
      parentComment.children.push(oneComment);
      ElMessage.success("回复成功！");
    }else{
      ElMessage.error(result?.errMsg || "回复失败");
    }
    replyInputs.value[parentGuid] = "";
    replyInputVisible.value[parentGuid] = false;
    replyTargets.value[parentGuid] = null;
    isChildrenVisible.value[parentGuid] = true;
  }
}

function deleteComment(commentId, parentId = null) {
  ele_confirm("确定要删除这条评论吗?", () => {
    if (parentId) {
      let parent = blogComment.value.find(c => c.GUID === parentId);
      if (parent) parent.children = parent.children.filter(item => item["GUID"] !== commentId);
    } else {
      blogComment.value = blogComment.value.filter(item => item["GUID"] !== commentId);
    }
    sendAxiosRequest("/blog-api/blog/deleteComment", {blogGuid: commentId});
    ElMessage.success("评论已删除");
  })
}

function toggleChildren(commentId) {
  isChildrenVisible.value[commentId] = !isChildrenVisible.value[commentId];
}

function openOneBlog() { pubOpenOneBlog(router, blogContent.value.GUID) }

function deleteArticle() {
  ele_confirm("确定要删除这篇文章吗？此操作不可撤销。", () => {
    blogContentStore.blogContents = blogContentStore.blogContents.filter(item => item["GUID"] !== contentGuid.value);
    sendAxiosRequest("/blog-api/blog/deleteBlog", {guid: contentGuid.value});
    blogComment.value = [];
    ElMessage.success("文章已删除");
    if(route.name==="BlogContent"){
      router.push({ name: "BlogContent", query: {g: blogContentStore?.blogContents?.[0]?.GUID || ""} });
    }else{
      location.reload();
    }
  })
}
</script>

<style scoped>
.smooth-col {
  transition: all 0.4s cubic-bezier(0.25, 0.8, 0.25, 1);
}

.article-view-row {
  margin: 0;
  padding: 0;
  height: 100%;
  align-items: flex-start;
}

/* ====== 文章主体区域 ====== */
.article-card {
  min-height: 80vh;
  margin-top: 0;
  display: flex;
  flex-direction: column;
  background-color: #ffffff;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.04);
}

.author-info {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.author-avatar {
  width: 50px !important;
  height: 50px !important;
  font-size: 20px;
  background-color: #e6f1fc;
  color: #409eff;
  cursor: pointer;
  transition: transform 0.2s;
}

.author-avatar:hover {
  transform: scale(1.05);
}

.author-text {
  display: flex;
  flex-direction: column;
  cursor: pointer;
}

.author-name {
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
}

.author-tagline {
  font-size: 13px;
  color: #909399;
  margin-top: 2px;
}

.article-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.article-header h2 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

/* 文章底部互动区 */
.article-bottom-actions {
  margin-top: 40px;
  padding-top: 20px;
}

.bottom-action-buttons {
  display: flex;
  justify-content: center;
  gap: 20px;
  margin-top: 20px;
  margin-bottom: 20px;
}

/* ====== 评论区 ====== */
.comment-card {
  display: flex;
  flex-direction: column;
  margin-top: 0;
  border-radius: 12px;
  background-color: #fafbfc;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.04);
  max-height: calc(100vh - 40px);
  /* 清除了原来的 position: sticky */
  overflow-y: auto;
}

.comment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #ebeef5;
}

.comment-header h3 {
  margin: 0;
  font-size: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.comment-count-badge {
  background: #f56c6c;
  color: white;
  border-radius: 10px;
  padding: 0 8px;
  font-size: 12px;
  font-weight: normal;
}

.comment-item {
  margin-bottom: 16px;
  border-bottom: 1px dashed #ebeef5;
  padding-bottom: 12px;
}

.comment-main-row, .comment-child {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 8px 6px;
  border-radius: 8px;
  transition: background-color 0.2s ease;
  cursor: pointer;
}

.comment-main-row:hover, .comment-child:hover {
  background-color: #f0f4f8;
}

.avatar-container {
  flex-shrink: 0;
  margin-top: 2px;
}

.author-avatar-comment {
  width: 36px !important;
  height: 36px !important;
}

.child-avatar {
  width: 28px !important;
  height: 28px !important;
}

.comment-content-block {
  flex: 1;
  min-width: 0;
  line-height: 1.6;
}

.comment-user-name {
  font-size: 14px;
  font-weight: 600;
  color: #409eff;
  margin-right: 6px;
}

.child-name {
  color: #67c23a;
}

.comment-text {
  font-size: 14px;
  color: #333;
  word-wrap: break-word;
  word-break: break-all;
}

.reply-to-text {
  font-size: 13px;
  color: #909399;
  margin: 0 4px;
}

.reply-to-name {
  color: #409eff;
  font-weight: 600;
  cursor: pointer;
  transition: color 0.2s ease;
}

.reply-to-name:hover {
  text-decoration: underline;
  color: #66b1ff;
}

.comment-actions {
  flex-shrink: 0;
  display: flex;
  gap: 8px;
  opacity: 0.8;
  margin-left: 8px;
}

.comment-actions .el-button {
  margin: 0 !important;
}

.reply-input-wrapper {
  margin-top: 8px;
  padding-left: 48px;
  background: transparent;
  padding-bottom: 10px;
  border-radius: 6px;
}

.reply-action-group {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.children-comments {
  margin-left: 48px;
  margin-top: 4px;
}

.children-list {
  border-left: 2px solid #e4e7ed;
  padding-left: 10px;
  margin-top: 8px;
}

.comment-input-area {
  margin-top: 20px;
}

.guest-form-row {
  margin-bottom: 12px;
}

.guest-input-item {
  margin-bottom: 8px;
}

.comment-submit-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
}

.hint-text {
  font-size: 12px;
  color: #b1b3b8;
}

/* ====== 固定目录卡片 ====== */
.toc-fixed-card {
  border-radius: 12px;
  background: #ffffff;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.04);
  /* 清除了原来的 position: sticky */
}

.toc-title {
  font-weight: 600;
  font-size: 16px;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid #f0f2f5;
  color: #303133;
}

.toc-empty {
  font-size: 13px;
  color: #999;
  padding: 10px;
  text-align: center;
}

.toc-item {
  padding: 8px 12px;
  cursor: pointer;
  font-size: 14px;
  color: #606266;
  border-radius: 6px;
  transition: all 0.2s ease;
  border-left: 3px solid transparent;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.toc-item:hover {
  background-color: #f5f7fa;
  color: #409eff;
}

.toc-item.is-active {
  color: #409eff;
  background-color: #ecf5ff;
  font-weight: 600;
  border-left-color: #409eff;
}

.toc-level-2 { padding-left: 12px; }
.toc-level-3 { padding-left: 24px; font-size: 13px; }
.toc-level-4 { padding-left: 36px; font-size: 12px; }

:deep(h1), :deep(h2), :deep(h3), :deep(h4) {
  scroll-margin-top: 80px;
}

/* ====== 悬浮工具栏 ====== */
.floating-wrapper {
  position: fixed;
  top: 50%;
  right: 30px;
  transform: translateY(-50%);
  z-index: 1000;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.glass-effect {
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255,255,255,0.5);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
  border-radius: 40px;
  padding: 12px 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.glass-effect .btn-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 !important;
  padding: 0 !important;
}

.glass-effect .el-button {
  margin: 0 !important;
}

.floating-btn {
  width: 44px !important;
  height: 44px !important;
  font-size: 20px !important;
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1) !important;
  border: none;
  background: transparent;
  box-shadow: none;
  color: #606266;
  display: flex;
  justify-content: center;
  align-items: center;
}

.floating-btn:hover {
  background: rgba(64, 158, 255, 0.1);
  color: #409eff;
  transform: translateY(-2px);
}

.is-active-btn {
  color: #409eff !important;
  background: rgba(64, 158, 255, 0.15) !important;
}

.floating-badge {
  position: absolute;
  top: -2px;
  right: 0;
}

/* =========================================================================
   移动端适配
   ========================================================================= */
@media (max-width: 991px) {
  .comment-card {
    max-height: none !important;
    overflow-y: visible !important;
    margin-top: 15px;
  }

  .toc-fixed-card {
    margin-top: 15px;
  }

  .floating-wrapper {
    right: 15px;
    transform: translateY(-20%);
  }
}
</style>