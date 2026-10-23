<template>
  <el-row :gutter="28" class="article-view-row" :class="{ 'is-focus': isFocusPage }" justify="space-between" align="top">
    <!-- 文章区域：文章详情页独占整行（目录在页面左栏），其他页面右侧留一栏放目录 -->
    <el-col :xs="24" :sm="24" :md="isFocusPage ? 24 : 17" :lg="isFocusPage ? 24 : 17">
      <article ref="articleSheetRef" class="article-sheet">
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
            <!-- 有用户编号时输出真实链接，搜索引擎可以由文章发现作者主页 -->
            <a
                v-if="blogContent.USERNUM"
                class="author-name author-link"
                :href="`/user/${encodeURIComponent(blogContent.USERNUM)}`"
                target="_blank"
                @click.stop
            >{{ blogContent.USERNAME || '匿名用户' }}</a>
            <div v-else class="author-name">{{ blogContent.USERNAME || '匿名用户' }}</div>
            <div class="author-tagline">
              发布时间: {{ pubFormatDate(blogContent.CREATE_TIME) }}
              <span v-if="readingStats" class="reading-stats">· 约 {{ readingStats.words }} 字 · 阅读约 {{ readingStats.minutes }} 分钟</span>
              <!-- 点赞、收藏只展示数量（操作在文章末尾），点评论数跳到评论区 -->
              <span class="meta-stats" @click.stop>
                <span class="meta-stat" title="点赞数">👍 {{ blogLikeNum }}</span>
                <span class="meta-stat" title="收藏数">⭐ {{ blogCollectNum }}</span>
                <button type="button" class="meta-stat meta-comment" title="查看评论" @click="scrollToComments">💬 {{ commentTotal }}</button>
              </span>
            </div>
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

        <!-- 文章标签：公开文章点开是标签页；私密文章不提供链接（标签页只列公开文章） -->
        <ArticleSeriesBox :series="blogContent.SERIES" :current-guid="blogContent.GUID" placement="top"/>

        <div v-if="blogContent.TAGS?.length" class="article-tags">
          <template v-for="tag in blogContent.TAGS" :key="tag">
            <router-link v-if="blogContent.BLOG_TYPE === 'public'" :to="blogTagHref(tag)" class="j-tag" rel="tag">#{{ tag }}</router-link>
            <span v-else class="j-tag">#{{ tag }}</span>
          </template>
        </div>

        <p v-if="!blogContent.GUID" class="sheet-loading">
          <svg class="pencil-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 20l1.2-4.6L15.8 4.8a2 2 0 0 1 2.8 0l.6.6a2 2 0 0 1 0 2.8L8.6 18.8z M13.5 7l3.5 3.5 M4 20l4.6-1.2"/></svg>
          正在翻开这一页…
        </p>
        <ArticleViewer v-if="articleRenderMode === 'html'" :content="blogContent.MAINTEXT"/>
        <ArticleEditor v-else :isReadOnly="true" :content="blogContent.MAINTEXT"/>

        <div class="article-bottom-actions">
          <p class="end-mark">文章完</p>
          <ArticleSeriesBox :series="blogContent.SERIES" :current-guid="blogContent.GUID" placement="bottom"/>
          <div class="bottom-action-buttons">
            <!-- 点过之后在按钮角上盖一个红章；按钮有折角裁切，章放在外层才不会被裁掉 -->
            <span class="stamp-holder">
              <button type="button" class="j-button" :class="{ 'is-on': blogContent.$userIsLike }"
                      :title="blogContent.$userIsLike ? '再点一次取消点赞' : '点赞'" @click="handleLike">
                👍 点赞 {{ blogLikeNum > 0 ? `(${blogLikeNum})` : '' }}
              </button>
              <span v-if="blogContent.$userIsLike" class="j-stamp action-stamp" :class="{ 'j-stamp--press': stampPressed.like }">已赞</span>
            </span>

            <span class="stamp-holder">
              <button type="button" class="j-button" :class="{ 'is-on': blogContent.$userIsCollect }"
                      :title="blogContent.$userIsCollect ? '再点一次取消收藏' : '收藏'" @click="handleCollect">
                <el-icon><Star/></el-icon>
                收藏 {{ blogCollectNum > 0 ? `(${blogCollectNum})` : '' }}
              </button>
              <span v-if="blogContent.$userIsCollect" class="j-stamp action-stamp" :class="{ 'j-stamp--press': stampPressed.collect }">已收藏</span>
            </span>

          </div>
        </div>

        <!-- 评论区：在文章下方（与便签墙共用组件） -->
        <section class="comment-anchor" ref="commentAnchorRef">
          <CommentSection
              v-if="blogContent.GUID"
              kind="blog"
              :target-id="blogContent.GUID"
              :author-code="blogContent.USERCODE"
              :author-name="blogContent.USERNAME"
              @count="commentTotal = $event"
          />
        </section>

        <!-- 上一篇 / 下一篇 / 同作者文章放在评论之后，读完先看讨论，最后再去别的文章 -->
        <ArticleLinks
            v-if="route.name === 'oneBlog' && blogContent.GUID"
            :blog-id="blogContent.GUID"
            :author="blogContent.USERNAME"
        />
      </article>
    </el-col>

    <!-- 其他页面（我的博客、弹窗）：右侧目录，滚动时吸附在顶部 -->
    <el-col v-if="!isFocusPage" :xs="24" :sm="24" :md="7" :lg="7" class="toc-col" :class="{ 'use-drawer': useTocDrawer }">
      <div class="side-float">
        <ArticleToc :items="tocList" :active-id="activeTocId" @select="scrollToAnchor"/>
      </div>
    </el-col>
  </el-row>

  <!-- 文章详情页：目录放进页面左栏、用户信息卡片下方 -->
  <Teleport v-if="isFocusPage && tocList.length" defer to="#yns-toc-dock">
    <ArticleToc :items="tocList" :active-id="activeTocId" @select="scrollToAnchor"/>
  </Teleport>

  <!-- 小屏：目录收进右下角按钮，点开从右侧滑出 -->
  <template v-if="useTocDrawer && tocList.length">
    <button type="button" class="toc-drawer-btn" title="文章目录" @click="tocDrawerVisible = true">
      <el-icon><Memo/></el-icon>
      <span>目录</span>
    </button>
    <el-drawer v-model="tocDrawerVisible" direction="rtl" size="min(300px, 82vw)" :with-header="false">
      <ArticleToc plain :items="tocList" :active-id="activeTocId" max-height="calc(100vh - 80px)" @select="selectFromDrawer"/>
    </el-drawer>
  </template>

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
        ref="articleEditorRef"
        :draft-key="'edit-' + contentGuid"
        :title="blogContent.BLOG_TITLE"
        :content="blogContent.MAINTEXT"
        :tags="blogContent.TAGS || []"
        :series-id="blogContent.SERIES_ID || ''"
        :series-editable="blogContent.USERCODE === userStore.userBean?.code"
        :save-type="'edit'"
        :isPublic="blogContent.BLOG_TYPE === 'public' ? true : false"
        @submit="handleEditorSubmit"
        @cancel="editorVisible = false"
    />
  </el-dialog>
</template>

<script setup>
import { ref, computed, onUnmounted, watch, nextTick, defineAsyncComponent } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useUserStore } from "@/stores/main/user.js";
import { useBlogContentStore } from "@/stores/detail/blog.js";
import ArticleViewer from "@/components/detail/ArticleViewer.vue";
import ArticleLinks from "@/components/detail/myblog/ArticleLinks.vue";
import ArticleToc from "@/components/detail/myblog/ArticleToc.vue";
import CommentSection from "@/components/detail/comment/CommentSection.vue";
import { getAvatarStyle } from "@/utils/avatarStyle.js";
import { Star, Memo } from '@element-plus/icons-vue'
import { ElMessage } from "element-plus";
import debounce from 'lodash/debounce'

import {
  ele_confirm,
  sendAxiosRequest,
  sendAxiosRequestChecked,
  pubFormatDate,
  sendNotifications, getCurrentUserAdminObject, extractPlainTextFromHTML
} from "@/utils/common.js";

import { adminUserCode, articleRenderMode } from "@/config/vue-config.js";
import { pubOpenOneBlog, pubOpenUser } from "@/utils/blogUtil.js";
import { blogTagHref } from "@/utils/blogTags.js";
import ArticleSeriesBox from "@/components/detail/myblog/ArticleSeriesBox.vue";

// 编辑器体积较大，只在打开编辑弹窗（或切回编辑器展示方式）时才下载
const ArticleEditor = defineAsyncComponent(() => import("@/components/detail/ArticleEditor.vue"));

const articleEditorRef = ref(null);

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
const editorVisible = ref(false);

// 文章详情页：目录由页面放在左栏；我的博客页和详情页在小屏时把目录收进抽屉
const isFocusPage = computed(() => route.name === 'oneBlog');
const useTocDrawer = computed(() => ['oneBlog', 'BlogContent'].includes(route.name));
const tocDrawerVisible = ref(false);
const articleSheetRef = ref(null);
const commentAnchorRef = ref(null);

// 只有当次点击才播放盖章动画，刷新页面看到的已赞状态直接显示
const stampPressed = ref({ like: false, collect: false });

// 评论总数（包含楼中楼）由评论组件告知
const commentTotal = ref(0);

const scrollToComments = () => {
  commentAnchorRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' });
};

const emit = defineEmits(['loaded', 'not-found', 'toc-change'])

const handleEditorSubmit = async ({blog_type, title, content, tags, seriesId}) => {
  const payload = {guid: contentGuid.value, title, blog_type, content, tags: tags || []};
  // 不能选系列（管理员修改别人的文章）时不传，后台保留原来的系列
  if (seriesId !== undefined) payload.seriesId = seriesId;
  const result = await sendAxiosRequestChecked("/blog-api/blog/updateBlog", payload, "修改失败");
  if (result) {
    articleEditorRef.value?.onSaved();
    blogContent.value.BLOG_TITLE = title;
    blogContent.value.MAINTEXT = content;
    blogContent.value.BLOG_TYPE = blog_type;
    blogContent.value.TAGS = tags || [];
    // 系列序号、上一篇下一篇要重新计算，重新读取一次
    if (seriesId !== undefined && String(seriesId || '') !== String(blogContent.value.SERIES_ID || '')) {
      reloadSeries();
    }
    blogContentStore.blogContents.forEach(item => {
      if (item["GUID"] === contentGuid.value) {
        item["BLOG_TITLE"] = title;
        item["BLOG_TYPE"] = blog_type;
      }
    });
    ElMessage.success("已修改");
    editorVisible.value = false;
  }
};

// 修改了所属系列后重新读取系列信息（第几篇、上一篇下一篇）
async function reloadSeries() {
  const res = await sendAxiosRequest("/blog-api/blog/getBlog", {blogId: contentGuid.value});
  const row = res && !res.isError ? res.result?.[0] : null;
  if (row) {
    blogContent.value.SERIES_ID = row.SERIES_ID;
    blogContent.value.SERIES = row.SERIES;
  }
}

const loadContentAndComments = async (guid) => {
  // 正文、点赞收藏只依赖文章 ID，两个请求同时发出；评论由评论组件自己加载
  const [blogResult, reactionResult] = await Promise.all([
    sendAxiosRequest("/blog-api/blog/getBlog", {blogId: guid}),
    sendAxiosRequest("/blog-api/blog/getLikeAndCollectByBlogId", {blogId: guid}),
  ]);
  if (blogResult && !blogResult.isError && blogResult?.result?.[0]) {
    blogContent.value = blogResult.result[0];
  } else {
    ElMessage.error("该文章为私密或已删除");
    emit('not-found');
    return false;
  }
  // 作者信息等由父组件加载，拿到正文就通知
  emit('loaded', {blogContent: blogContent.value});

  if (reactionResult && !reactionResult.isError) {
    let userBean = userStore.userBean;
    let likeNum = 0, collectNum = 0;
    reactionResult.result.forEach(item => {
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
};

// 文章字数和预计阅读时长：中文按字、英文按单词计数，阅读速度按每分钟 400 字 / 200 词估算
const readingStats = computed(() => {
  const text = extractPlainTextFromHTML(blogContent.value.MAINTEXT || '');
  const cjk = (text.match(/[\u3400-\u9fff\uf900-\ufaff]/g) || []).length;
  const words = (text.replace(/[\u3400-\u9fff\uf900-\ufaff]/g, ' ').match(/[A-Za-z0-9_]+(?:['’.-][A-Za-z0-9_]+)*/g) || []).length;
  const total = cjk + words;
  if (!total) return null;
  return {words: total, minutes: Math.max(1, Math.round(cjk / 400 + words / 200))};
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
    const contentEl = articleSheetRef.value?.querySelector('.editor-container');
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
    emit('toc-change', tempToc.length);
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


const selectFromDrawer = (anchorId) => {
  tocDrawerVisible.value = false;
  scrollToAnchor(anchorId);
};

onUnmounted(() => {
  if (observer) observer.disconnect();
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
    stampPressed.value.like = true;
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
    stampPressed.value.collect = true;
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

function openOneBlog() { pubOpenOneBlog(router, blogContent.value.GUID) }

function deleteArticle() {
  ele_confirm("确定要删除这篇文章吗？此操作不可撤销。", async () => {
    const result = await sendAxiosRequestChecked("/blog-api/blog/deleteBlog", {guid: contentGuid.value}, "删除文章失败");
    if (!result) return;
    blogContentStore.blogContents = blogContentStore.blogContents.filter(item => item["GUID"] !== contentGuid.value);
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
.article-view-row {
  margin: 0;
  padding: 0;
  height: 100%;
  align-items: flex-start;
}

/* 文章详情页：正文一行控制在 760px 左右，宽屏时也不会拉得太长 */
.article-view-row.is-focus .article-sheet {
  max-width: 864px;
}

/* 右侧目录栏拉满整行高度，目录卡片才能一路吸附到文章结束 */
.toc-col {
  align-self: stretch;
}

.side-float {
  position: sticky;
  top: 20px;
}

/* ====== 文章：一张贴着胶带的纸 ====== */
.article-sheet {
  position: relative;
  min-height: 80vh;
  padding: 34px 40px 28px 64px;
  /* 页边红线画在背景里，两个伪元素留给左侧装订孔 */
  background-color: var(--j-paper);
  background-image: linear-gradient(90deg, transparent 40px, rgba(232, 165, 155, 0.55) 40px, rgba(232, 165, 155, 0.55) 41px, transparent 41px);
  border: 1px solid var(--j-rule);
  border-radius: 2px 6px 6px 2px;
  box-shadow: var(--j-shadow);
  box-sizing: border-box;
}

/* 左侧装订孔（与首页的纸一致）：::before 画孔边的阴影，::after 是孔里露出的东西 */
.article-sheet::before,
.article-sheet::after {
  content: "";
  position: absolute;
  top: 18px;
  bottom: 18px;
  left: 16px;
  width: 14px;
  pointer-events: none;
}

.article-sheet::before {
  background-image: radial-gradient(circle at 7px 7px, transparent 5px, rgba(120, 104, 80, 0.25) 5.5px, transparent 6.5px);
  background-size: 14px 44px;
}

/* 孔里露出桌面；文章详情页设置了背景时露出背景（--page-bg 由页面提供，和页面背景一样固定在视口上，位置能对上） */
.article-sheet::after {
  background-color: var(--j-desk);
  background-image: var(--page-bg, none);
  background-size: cover;
  background-position: center;
  background-attachment: fixed;
  -webkit-mask: radial-gradient(circle at 7px 7px, #000 5px, transparent 5.5px) 0 0 / 14px 44px repeat-y;
  mask: radial-gradient(circle at 7px 7px, #000 5px, transparent 5.5px) 0 0 / 14px 44px repeat-y;
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

.reading-stats {
  margin-left: 4px;
}

.meta-stats {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-left: 10px;
  cursor: default;
}

.meta-stat {
  font-family: var(--j-sans);
  font-size: 13px;
  color: var(--j-muted);
  white-space: nowrap;
}

.meta-comment {
  padding: 0;
  border: none;
  background: none;
  cursor: pointer;
  text-decoration: underline;
  text-decoration-color: transparent;
  text-underline-offset: 3px;
  transition: color 0.15s, text-decoration-color 0.15s;
}

.meta-comment:hover {
  color: var(--j-pen);
  text-decoration-color: currentColor;
}

.article-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 18px;
}

.article-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: -8px 0 18px;
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
.author-link {
  display: block;
  text-decoration: none;
}

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

.stamp-holder {
  position: relative;
  display: inline-flex;
}

.action-stamp {
  position: absolute;
  top: -10px;
  right: -14px;
  z-index: 1;
  background: rgba(255, 253, 248, 0.6);
  pointer-events: none;
}

/* 加载中：楷体一句话 + 一支铅笔 */
.sheet-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin: 0;
  font-family: var(--j-hand);
  color: var(--j-muted);
}

.sheet-loading {
  padding: 80px 0;
  font-size: 18px;
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

.sheet-loading .pencil-icon {
  animation: pencil-write 1.2s ease-in-out infinite alternate;
}

@keyframes pencil-write {
  from { transform: translateX(-3px) rotate(-4deg); }
  to { transform: translateX(3px) rotate(4deg); }
}

:deep(h1), :deep(h2), :deep(h3), :deep(h4) {
  scroll-margin-top: 80px;
}

/* ====== 小屏目录按钮 ====== */
.toc-drawer-btn {
  display: none;
}

@media (max-width: 991px) {
  /* 详情页、我的博客页的目录改到抽屉里 */
  .toc-col.use-drawer {
    display: none;
  }

  .toc-col .side-float {
    position: static;
    margin-top: 18px;
  }

  .toc-drawer-btn {
    position: fixed;
    right: 0;
    bottom: 90px;
    z-index: 1000;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    width: 44px;
    padding: 8px 0 6px 4px;
    border: none;
    border-radius: 6px 0 0 6px;
    background: #c7e2cf;
    color: var(--j-ink);
    font-size: 12px;
    cursor: pointer;
    box-shadow: -1px 1px 3px rgba(60, 50, 30, 0.15);
  }

  .toc-drawer-btn .el-icon {
    font-size: 17px;
  }
}

@media (max-width: 640px) {
  .article-sheet {
    padding: 26px 16px 20px 30px;
  }

  /* 手机上左边距放不下装订孔，只保留红线 */
  .article-sheet {
    background-image: linear-gradient(90deg, transparent 16px, rgba(232, 165, 155, 0.55) 16px, rgba(232, 165, 155, 0.55) 17px, transparent 17px);
  }

  .article-sheet::before,
  .article-sheet::after {
    display: none;
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
