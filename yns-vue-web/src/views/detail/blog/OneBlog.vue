<template>
  <div class="blog-detail-page">
    <!-- 顶部信息栏 -->
    <div class="blog-top-bar">
      <div class="back-btn" @click="goBack">← 回到主页</div>
      <div class="blog-title">{{ articleTitle || '博客详情' }}</div>
    </div>
    <!-- 主体内容 -->
    <div class="main-body">
      <!-- 左侧作者信息 -->
      <div class="author-side">
        <el-card class="author-card" shadow="hover">
          <div class="author-avatar-wrapper" @click="avatarClick(authorInfo)" title="查看发布者信息">
            <el-avatar
                :src="authorInfo.avatar"
                class="author-avatar"
                alt="用户头像"
            >
              {{ authorInfo.name?.charAt(0) }}
            </el-avatar>
          </div>
          <div class="author-name">{{ authorInfo.name }}</div>
          <p class="userip">IP: {{ authorInfo.loginaddress || "未知" }}</p>
          <div class="author-desc" :title="authorInfo.remark">{{ authorInfo.remark || '这个人很神秘~' }}</div>
          <div class="stats">
            <div class="stat-item" title="关注">
              <div class="stat-num">{{ followingNum }}</div>
              <div class="stat-label">关注</div>
            </div>
            <div class="stat-item" title="粉丝">
              <div class="stat-num">{{ followersNum }}</div>
              <div class="stat-label">粉丝</div>
            </div>
          </div>
          <div class="author-actions">
            <el-button size="default" :type="isFollowing ? 'danger' : 'primary'" round @click="toggleFollow"
                       style="flex: 1; margin-right: 8px;">{{ isFollowing ? '取消关注' : '关注' }}
            </el-button>
            <el-button size="default" round @click="messageAuthor" style="flex: 1;">私信</el-button>
          </div>
        </el-card>
      </div>
      <!-- 右侧正文内容 -->
      <div class="content-side">
        <ContentAndComment :blogId="blogId" @loaded="contentAndCommentIsLoad"/>
      </div>
    </div>
    <!-- 聊天窗口 -->
    <Chat v-if="chatVisible" :title="'与'+authorInfo.name+'的聊天'" @closeChat="messageAuthor"/>
  </div>
</template>

<script setup>
import {ref, onMounted} from "vue";
import Chat from "@/components/detail/Chat.vue";
import {useRoute, useRouter} from "vue-router";
import ContentAndComment from "@/views/detail/blog/ContentAndComment.vue";
import {sendAxiosRequest, decrypt, encrypt, sendNotifications, extractPlainTextFromHTML} from "@/utils/common";
import {useUserStore} from "@/stores/main/user.js";
import {ElMessage} from "element-plus";
import {pubOpenUser} from "@/utils/blogUtil.js";

const userStore = useUserStore();
userStore.initFromLocal();

const route = useRoute();
const router = useRouter();
const blogId = route.params.g;
let userCode;

// 作者信息
const authorInfo = ref({});

//关注相关
const isFollowing = ref(false);
//粉丝数
const followersNum = ref(0);
//粉丝列表
const followersUser = ref([]);
//关注数
const followingNum = ref(0);
//关注列表
const followingUser = ref([]);

// 聊天相关
const chatVisible = ref(false);


async function getUserInfo2Data() {
  if (userCode) {
    //获取账号信息
    let result = await sendAxiosRequest("/pub-api/login/getUserInfoByCode", {userCode});
    if (result && !result.isError) {
      authorInfo.value = result.result;
    }

    result = await sendAxiosRequest("/blog-api/userInformation/getFollowUser", {userCode});
    if (result && !result.isError) {
      followersNum.value = result.result.followersUser.length;
      followersUser.value = result.result.followersUser;
      followingNum.value = result.result.followingUser.length;
      followingUser.value = result.result.followingUser;
      //粉丝列表包含当前登录用户,表示已经关注
      let arr = result.result.followersUser.filter(item => item["CODE"] == userStore.userBean.code);
      if (arr.length > 0) {
        isFollowing.value = true;
      }
    }
  }
}

const toggleFollow = () => {

  if (!userStore.userBean.code) {
    ElMessage.error("用户过期,请返回主页面重新登录!");
    return false;
  }
  isFollowing.value = !isFollowing.value

  //关注
  if (isFollowing.value) {
    followersNum.value++;
    sendAxiosRequest("/blog-api/userInformation/followUser", {
      followUserCode: authorInfo.value.code,
      followUserName: authorInfo.value.name
    });
    //发送消息
    sendNotifications(userStore.userBean.code, authorInfo.value.code, "followUser", null, `${userStore.userBean.name}关注了你`)
    //取消关注
  } else {
    followersNum.value--;
    sendAxiosRequest("/blog-api/userInformation/noFollowUser", {followUserCode: authorInfo.value.code});
  }
  ElMessage.success(isFollowing.value ? '已关注' : '已取消关注')
}

// 顶部栏信息
const articleTitle = ref("");

function avatarClick(authorInfo) {
  pubOpenUser(router, authorInfo.code);
}

const messageAuthor = () => {
  if (!userStore.userBean.code) {
    ElMessage.error("用户过期,请返回主页面重新登录!");
    return false;
  }
  if (userStore.userBean.code == authorInfo.value.code) {
    ElMessage.error("和自己就别聊了");
    return false;
  }
  chatVisible.value = !chatVisible.value; // 显示聊天窗口
};

const goBack = () => {
  router.push("/");
};

const contentAndCommentIsLoad = ({blogContent}) => {
  if (blogContent.USERCODE) {
    userCode = blogContent.USERCODE;
    articleTitle.value = blogContent.BLOG_TITLE;
    //搜索用户信息
    getUserInfo2Data();
    //修改浏览器title和meta,有助于搜索排名
    document.title = blogContent.BLOG_TITLE;
    const metaDesc = document.querySelector('meta[name="description"]');
    const blogSummary = extractPlainTextFromHTML(blogContent.BLOG_TITLE + blogContent.MAINTEXT).slice(0, 200) + '...';
    if (metaDesc) {
      metaDesc.setAttribute("content", blogSummary);
    } else {
      const desc = document.createElement('meta')
      desc.name = "description"
      desc.content = blogSummary
      document.head.appendChild(desc)
    }
  }
}

</script>

<style scoped>
.blog-detail-page {
  min-height: 100vh;
  background-color: #f5f6f9;
  display: flex;
  flex-direction: column;
}

/* 顶部信息栏 */
.blog-top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: #ffffff;
  padding: 20px 20px;
  border-bottom: 1px solid #eee;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.04);
}

.back-btn {
  color: #409EFF;
  font-size: 14px;
  cursor: pointer;
  min-width: 50px;
}

.blog-title {
  font-size: 18px;
  font-weight: bold;
  flex: 1;
  text-align: center;
  color: #333;
}

.author-avatar {
  width: 48px !important;
  height: 48px !important;
  font-size: 20px;
  background-color: #f2f2f2;
  cursor: pointer;
}

.blog-meta {
  display: flex;
  gap: 14px;
  font-size: 13px;
  color: #999;
}

/* 主体两栏 */
.main-body {
  display: flex;
  gap: 24px;
  padding: 20px;
  flex: 1;
  box-sizing: border-box;
}

/* 左侧作者信息 */
.author-side {
  width: 260px;
  flex-shrink: 0;
}

.author-card {
  padding: 24px 20px;
  text-align: center;
  border-radius: 12px;
  box-shadow: 0 2px 10px rgb(0 0 0 / 0.06);
  background-color: #fff;
}

.author-avatar-wrapper {
  margin: 0 auto;
  width: 72px;
  height: 72px;
  border-radius: 50%;
  overflow: hidden;
  border: 3px solid #409EFF;
  cursor: pointer;
  transition: box-shadow 0.3s ease;
}

.author-avatar-wrapper:hover {
  box-shadow: 0 0 8px #409EFF;
}

.author-avatar {
  width: 72px !important;
  height: 72px !important;
  font-size: 28px;
  background-color: #f0f4ff;
  color: #409EFF;
}

.author-name {
  font-size: 20px;
  font-weight: 700;
  margin-top: 14px;
  color: #2c3e50;
  user-select: none;
}

.userip {
  color: #999;
  margin: 6px 0 10px;
  font-size: 13px;
  user-select: text;
}

.author-desc {
  font-size: 14px;
  color: #666;
  margin-bottom: 20px;
  user-select: text;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2; /* 最多两行 */
  overflow: hidden;
  text-overflow: ellipsis;

  /* 美化：视觉上更柔和 */
  background-color: #f8f8f8;
  padding: 4px 8px;
  border-radius: 6px;
  font-style: italic;

  /* 删除 min-height 以适应内容 */
  min-height: auto;
}


/* 关注粉丝统计区 */
.stats {
  display: flex;
  justify-content: center;
  gap: 40px;
  margin-bottom: 20px;
}

.stat-item {
  cursor: pointer;
  user-select: none;
  text-align: center;
  color: #444;
  transition: color 0.2s ease;
}

.stat-item:hover {
  color: #409EFF;
}

.stat-num {
  font-weight: 700;
  font-size: 22px;
  line-height: 1;
}

.stat-label {
  font-size: 13px;
  color: #888;
  margin-top: 4px;
}

/* 按钮组 */
.author-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.author-actions .el-button {
  box-shadow: 0 2px 8px rgb(64 158 255 / 0.3);
  font-weight: 600;
  user-select: none;
}

.content-side {
  flex: 1;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  padding: 24px;
  min-height: 82vh;
}

/* 移动端适配 */
@media (max-width: 768px) {
  /* 让主体内容的作者信息和正文竖向排列 */
  .main-body {
    flex-direction: column;
    gap: 20px;
  }

  .author-side {
    width: 100%; /* 让作者信息在手机端占据全宽 */
  }

  .author-card {
    padding: 20px 16px;
  }

  .stats {
    gap: 30px;
  }

  .author-name {
    font-size: 18px;
  }

  .author-avatar-wrapper {
    width: 60px;
    height: 60px;
    border-width: 2px;
  }

  .author-avatar {
    width: 60px !important;
    height: 60px !important;
    font-size: 22px;
  }

  .content-side {
    padding: 16px;
  }
}
</style>

