<template>
  <el-row ref="layoutRowRef" :gutter="28" class="article-view-row" justify="space-between" align="top">
    <!-- 左侧文章区域 -->
    <el-col :xs="24" :sm="24" :md="17" :lg="17">
      <article class="article-sheet">
        <span class="j-tape j-tape--top"></span>

        <div class="author-info">
          <el-avatar
              :src="blogContent.AVATAR"
              :size="42"
              class="author-avatar"
              alt="用户头像"
              @click="avatarClick(blogContent)"
              :style="!blogContent.AVATAR ? getAvatarStyle(blogContent.USERNAME || '匿名用户') : {}"
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
          <h1 v-if="route.name === 'oneBlog'">{{ blogContent.BLOG_TITLE }}</h1>
          <h2 v-else>{{ blogContent.BLOG_TITLE }}</h2>
          <div class="article-tools">
            <button v-if="route.name!=='oneBlog'" type="button" class="tool-link" @click="openOneBlog">专注模式</button>
            <button v-if="canEditOrDelete" type="button" class="tool-link" @click="editorVisible = true">编辑文章</button>
            <button v-if="canEditOrDelete" type="button" class="tool-link danger" @click="deleteArticle">删除文章</button>
          </div>
        </div>

        <ArticleEditor :isReadOnly="true" :content="blogContent.MAINTEXT"/>

        <div class="article-bottom-actions">
          <p class="end-mark">文章完</p>
          <div class="bottom-action-buttons">
            <button type="button" class="j-button" :class="{ 'is-on': blogContent.$userIsLike }" @click="handleLike">
              👍 {{ blogContent.$userIsLike ? '已赞' : '点赞' }} {{ blogLikeNum > 0 ? `(${blogLikeNum})` : '' }}
            </button>

            <button type="button" class="j-button" :class="{ 'is-on': blogContent.$userIsCollect }" @click="handleCollect">
              <el-icon><Star/></el-icon>
              {{ blogContent.$userIsCollect ? '已收藏' : '收藏' }} {{ blogCollectNum > 0 ? `(${blogCollectNum})` : '' }}
            </button>

            <button v-if="!showComment" type="button" class="j-button" @click="showCommentFun">
              <el-icon><Comment/></el-icon>
              参与讨论 {{ blogComment.length > 0 ? `(${blogComment.length})` : '' }}
            </button>
          </div>
        </div>
      </article>
    </el-col>

    <!-- 右侧区域：评论区 or 目录区 -->
    <el-col :xs="24" :sm="24" :md="7" :lg="7">
      <!-- 独立悬浮的包裹层，纯 JS 操控其 translateY -->
      <div ref="rightSidebarRef" class="side-float">
        <!-- 评论区 -->
        <section v-if="showComment" class="comment-card">
          <div class="comment-header">
            <h3>互动评论 <span class="comment-count-badge">{{ blogComment.length }}</span></h3>
            <button type="button" class="tool-link" @click="showComment = false">
              收起评论 <el-icon><Right/></el-icon>
            </button>
          </div>

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
                <span class="comment-user-name">{{ comment.USERNAME }}:</span>
                <span class="comment-text">{{ comment.TEXT }}</span>
              </div>

              <div class="comment-actions" v-show="activeCommentId === comment.GUID">
                <button type="button" class="tool-link" @click.stop="replyComment(comment.GUID, comment)">回复</button>
                <button type="button" class="tool-link danger"
                        @click.stop="deleteComment(comment.GUID)"
                        v-if="(userStore?.userBean?.code && comment.USERCODE===userStore.userBean.code) ||(getCurrentUserAdminObject().isAdmin && comment.USERCODE!==adminUserCode) ||getCurrentUserAdminObject().adminLevel==='superAdmin'">
                  删除
                </button>
              </div>
            </div>

            <div v-if="replyInputVisible[comment.GUID]" class="reply-input-wrapper">
              <el-row :gutter="[10, 10]" v-if="!userStore?.userBean?.code" class="guest-form-row">
                <el-col :xs="24" :sm="12">
                  <el-input v-model="guestInfo.nickname" :prefix-icon="User" placeholder="昵称 (选填)" size="small" clearable/>
                </el-col>
                <el-col :xs="24" :sm="12">
                  <el-input v-model="guestInfo.email" :prefix-icon="Message" placeholder="邮箱 (选填)" size="small" clearable/>
                </el-col>
                <el-col :span="24">
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
              <button type="button" class="tool-link toggle-children-btn" @click="toggleChildren(comment.GUID)">
                {{ isChildrenVisible[comment.GUID] ? '收起回复' : `查看回复 (${comment.children.length})` }}
              </button>

              <div v-show="isChildrenVisible[comment.GUID]" class="children-list">
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
                    <span class="comment-user-name child-name">{{ child.USERNAME }}</span>
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
                            v-if="(userStore?.userBean?.code && child.USERCODE===userStore.userBean.code) ||(getCurrentUserAdminObject().isAdmin && child.USERCODE!==adminUserCode) ||getCurrentUserAdminObject().adminLevel==='superAdmin'">
                      删除
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-if="blogComment.length > 5" class="comment-more">
            <button type="button" class="tool-link" @click="toggleComments">
              {{ showAllComments ? '收起部分评论' : '展开全部评论' }}
            </button>
          </div>

          <div class="comment-input-area">
            <p class="input-title">发表评论</p>
            <el-row :gutter="[10, 10]" v-if="showMainGuestForm" class="guest-form-row">
              <el-col :xs="24" :sm="12">
                <el-input v-model="guestInfo.nickname" :prefix-icon="User" placeholder="昵称 (选填)" size="default" clearable />
              </el-col>
              <el-col :xs="24" :sm="12">
                <el-input v-model="guestInfo.email" :prefix-icon="Message" placeholder="邮箱 (选填)" size="default" clearable />
              </el-col>
              <el-col :span="24">
                <el-input v-model="guestInfo.website" :prefix-icon="Link" placeholder="网址 (选填)" size="default" clearable />
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
        </section>

        <!-- 目录区：固定显示 -->
        <section v-else class="toc-fixed-card">
          <span class="j-tape j-tape--green toc-tape"></span>
          <div class="toc-title">文章目录</div>
          <el-scrollbar max-height="600px">
            <div v-if="tocList.length === 0" class="toc-empty">暂无目录或提取中...</div>
            <div v-for="item in tocList" :key="item.id"
                 class="toc-item"
                 :class="['toc-level-' + item.level, { 'is-active': activeTocId === item.id }]"
                 @click="scrollToAnchor(item.id)">
              <span>{{ item.text }}</span>
            </div>
          </el-scrollbar>
        </section>
      </div>
    </el-col>
  </el-row>

  <!-- 悬浮按钮：贴在屏幕右边缘的索引标签 -->
  <div class="floating-wrapper">
    <el-tooltip :content="blogContent.$userIsLike ? '取消点赞' : '点赞'" placement="left">
      <button type="button" class="side-tab tab-yellow" :class="{ 'is-on': blogContent.$userIsLike }" @click="handleLike">
        👍
      </button>
    </el-tooltip>

    <el-tooltip :content="blogContent.$userIsCollect ? '取消收藏' : '收藏'" placement="left">
      <button type="button" class="side-tab tab-green" :class="{ 'is-on': blogContent.$userIsCollect }" @click="handleCollect">
        <el-icon><Star/></el-icon>
      </button>
    </el-tooltip>

    <el-tooltip :content="showComment ? '关闭评论' : '打开评论'" placement="left">
      <button type="button" class="side-tab tab-pink" :class="{ 'is-on': showComment }" @click="showCommentFun">
        <el-icon><Comment/></el-icon>
        <span v-if="blogComment.length > 0" class="side-tab-badge">{{ blogComment.length }}</span>
      </button>
    </el-tooltip>
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

const emit = defineEmits(['loaded', 'not-found'])

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
    emit('not-found');
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
    const layoutElement = layoutRowRef.value?.$el || layoutRowRef.value;
    const contentEl = layoutElement?.querySelector?.('.editor-container');
    if (!contentEl) return;

    // 详情页标题已经是唯一 H1；旧文章正文若含 H1，则降为 H2，保持语义层级正确。
    if (route.name === 'oneBlog') {
      contentEl.querySelectorAll('h1').forEach((heading) => {
        const replacement = document.createElement('h2');
        for (const attribute of heading.attributes) {
          replacement.setAttribute(attribute.name, attribute.value);
        }
        while (heading.firstChild) replacement.appendChild(heading.firstChild);
        heading.replaceWith(replacement);
      });
    }

    // 为正文图片补充可理解的 alt，并延迟加载非首图，兼顾图片 SEO 与页面性能。
    contentEl.querySelectorAll('img').forEach((image, index) => {
      const alt = (image.getAttribute('alt') || '').trim();
      if (!alt || /^image(?:\.[a-z0-9]+)?$/i.test(alt) || alt === '图片') {
        image.setAttribute('alt', `${blogContent.value.BLOG_TITLE || '文章'} 配图 ${index + 1}`);
      }
      image.setAttribute('decoding', 'async');
      if (index === 0) {
        image.setAttribute('loading', 'eager');
        image.setAttribute('fetchpriority', 'high');
      } else {
        image.setAttribute('loading', 'lazy');
      }
    });

    const headings = contentEl.querySelectorAll('h2, h3, h4');
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

function getCommentNotificationUrl() {
  return router.resolve({name: 'oneBlog', params: {g: contentGuid.value}}).href;
}

function isLoggedInCommentUser(comment) {
  const userCode = String(comment?.USERCODE || '').trim();
  return userCode !== '' && !userCode.toLowerCase().startsWith('guest_');
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
    sendNotifications(
        userPayload.code,
        blogContent.value.USERCODE,
        "comment",
        getCommentNotificationUrl(),
        `${userPayload.name}评论了你的文章《${blogContent.value.BLOG_TITLE}》`
    );
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
      if (isLoggedInCommentUser(targetUser)) {
        sendNotifications(
            userPayload.code,
            targetUser.USERCODE,
            "comment",
            getCommentNotificationUrl(),
            `${userPayload.name}回复了你在文章《${blogContent.value.BLOG_TITLE}》下的评论`
        );
      }
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

// 根据用户名生成固定的头像底色（取自胶带配色）
const getAvatarStyle = (name) => {
  if (!name) return {};

  const colors = [
    { bg: '#f6e3a1', text: '#7a5c12' },
    { bg: '#c7e2cf', text: '#2f6040' },
    { bg: '#f3cccc', text: '#8a3a34' },
    { bg: '#c9dbeb', text: '#2f5d8a' },
    { bg: '#e2d5c0', text: '#6b5330' }
  ];

  // 简单的字符串哈希算法，确保同一个名字每次计算出的颜色都是固定的
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }

  const index = Math.abs(hash) % colors.length;
  const selectedColor = colors[index];

  return {
    backgroundColor: selectedColor.bg,
    color: selectedColor.text,
    fontWeight: '600'
  };
};
</script>

<style scoped>
.article-view-row {
  margin: 0;
  padding: 0;
  height: 100%;
  align-items: flex-start;
}

/* 给屏幕右边缘的索引标签留出位置 */
@media (min-width: 992px) {
  .article-view-row {
    padding-right: 44px;
  }
}

/* ====== 文章：一张贴着胶带的纸 ====== */
.article-sheet {
  position: relative;
  min-height: 80vh;
  padding: 34px 40px 28px 64px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
  box-sizing: border-box;
}

/* 页边红线 */
.article-sheet::before {
  content: "";
  position: absolute;
  top: 0;
  bottom: 0;
  left: 40px;
  width: 1px;
  background: var(--j-margin-red);
  opacity: 0.55;
}

.author-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 18px;
  padding-bottom: 14px;
  border-bottom: 1px dashed var(--j-rule);
}

.author-avatar {
  flex-shrink: 0;
  font-size: 18px;
  background-color: #ece4d3;
  color: var(--j-ink-soft);
  cursor: pointer;
}

.author-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  cursor: pointer;
}

.author-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--j-ink);
}

.author-text:hover .author-name {
  color: var(--j-pen);
}

.author-tagline {
  font-family: var(--j-hand);
  font-size: 14px;
  color: var(--j-muted);
}

.article-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 18px;
}

.article-header h1,
.article-header h2 {
  margin: 0;
  font-size: 26px;
  line-height: 1.45;
  color: var(--j-ink);
}

.article-tools {
  display: flex;
  flex-shrink: 0;
  gap: 4px;
  padding-top: 6px;
}

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

/* 文章底部互动区 */
.article-bottom-actions {
  margin-top: 40px;
  padding-top: 8px;
}

.end-mark {
  display: flex;
  align-items: center;
  gap: 14px;
  margin: 0;
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-muted);
}

.end-mark::before,
.end-mark::after {
  content: "";
  flex: 1;
  border-top: 1px dashed var(--j-rule-strong);
}

.bottom-action-buttons {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 16px;
  margin: 24px 0 8px;
}

.bottom-action-buttons .j-button.is-on {
  background: var(--j-note);
  border-color: #e3cf7a;
}

/* ====== 评论区 ====== */
.comment-card {
  position: relative;
  display: flex;
  flex-direction: column;
  padding: 20px 18px 18px;
  background: var(--j-paper-warm);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
  max-height: calc(100vh - 40px);
  overflow-y: auto;
  box-sizing: border-box;
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

.children-list {
  margin-top: 6px;
  padding-left: 10px;
  border-left: 2px solid #ead9a0;
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

/* ====== 目录：一张索引卡 ====== */
.toc-fixed-card {
  position: relative;
  padding: 24px 18px 16px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
}

.toc-tape {
  top: -10px;
  left: 22px;
  transform: rotate(-4deg);
}

.toc-title {
  margin-bottom: 12px;
  font-family: var(--j-hand);
  font-size: 19px;
  color: var(--j-ink);
}

.toc-empty {
  padding: 10px 4px;
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-muted);
}

.toc-item {
  padding: 7px 4px;
  border-bottom: 1px dashed var(--j-rule);
  font-size: 14px;
  line-height: 20px;
  color: var(--j-ink-soft);
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.toc-item span {
  background-repeat: no-repeat;
  background-image: linear-gradient(transparent 55%, var(--j-highlight) 55%, var(--j-highlight) 95%, transparent 95%);
  background-size: 0 100%;
  transition: background-size 0.3s ease;
}

.toc-item:hover {
  color: var(--j-ink);
}

.toc-item.is-active {
  color: var(--j-ink);
}

.toc-item.is-active span {
  background-size: 100% 100%;
}

.toc-level-2 { padding-left: 4px; }
.toc-level-3 { padding-left: 18px; font-size: 13px; }
.toc-level-4 { padding-left: 32px; font-size: 12px; }

:deep(h1), :deep(h2), :deep(h3), :deep(h4) {
  scroll-margin-top: 80px;
}

/* ====== 悬浮工具：屏幕右边缘的索引标签 ====== */
.floating-wrapper {
  position: fixed;
  top: 50%;
  right: 0;
  transform: translateY(-50%);
  z-index: 1000;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
}

.side-tab {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 40px;
  padding: 0 0 0 4px;
  border: none;
  border-radius: 6px 0 0 6px;
  font-size: 17px;
  color: var(--j-ink);
  cursor: pointer;
  box-shadow: -1px 1px 3px rgba(60, 50, 30, 0.15);
  transition: width 0.2s ease;
}

.side-tab:hover,
.side-tab.is-on {
  width: 52px;
}

.side-tab.is-on {
  box-shadow: -2px 2px 6px rgba(60, 50, 30, 0.25);
}

.tab-yellow { background: #f6e3a1; }
.tab-green { background: #c7e2cf; }
.tab-pink { background: #f3cccc; }

.side-tab-badge {
  position: absolute;
  top: -6px;
  left: -6px;
  min-width: 18px;
  padding: 0 4px;
  border-radius: 9px;
  background: var(--j-stamp);
  color: #fff;
  font-size: 11px;
  line-height: 18px;
  box-sizing: border-box;
}

/* =========================================================================
   移动端适配
   ========================================================================= */
@media (max-width: 991px) {
  .comment-card {
    max-height: none !important;
    overflow-y: visible !important;
    margin-top: 18px;
  }

  .toc-fixed-card {
    margin-top: 18px;
  }

  .floating-wrapper {
    top: auto;
    bottom: 90px;
    transform: none;
  }
}

@media (max-width: 640px) {
  .article-sheet {
    padding: 26px 16px 20px 30px;
  }

  .article-sheet::before {
    left: 16px;
  }

  .article-header {
    flex-direction: column;
    gap: 8px;
  }

  .article-header h1,
  .article-header h2 {
    font-size: 22px;
  }

  .article-tools {
    padding-top: 0;
  }
}
</style>
