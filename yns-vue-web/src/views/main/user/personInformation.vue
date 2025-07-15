<template>
  <div class="person-info">
    <div class="inner-container">
      <!-- 顶部跳转按钮 -->
      <div class="toolbar">
        <el-button type="success" size="small" @click="scrollToFiles">
          📁 跳转到文件列表
        </el-button>
      </div>

      <!-- 用户信息卡片 -->
      <el-card class="profile-card">
        <div class="profile-header">
          <!-- 左侧：头像和基础信息 -->
          <div class="left-info">
            <el-avatar
                :src="user.avatar"
                size="large"
                class="author-avatar"
                alt="用户头像"
            >
              {{ user.name?.charAt(0) }}
            </el-avatar>
            <div class="profile-details">
              <h2 class="username">
                <span class="name">{{ user.name }}</span>
                <span v-if="user.code == adminUserCode" class="admin-badge">管理员</span>
              </h2>
              <p>邮箱: {{ user.email || "未知" }}</p>
              <p class="user-remark" :title="user.remark">个性签名: {{ user.remark || "这个人很神秘~" }}</p>
              <p class="userip">IP: {{ user.loginaddress || "未知" }}</p>
            </div>
          </div>

          <!-- 右侧：互动信息 -->
          <div class="right-info">
            <div class="stats">
              <span @click="followingUserClick" style="cursor: pointer;">关注 <strong>{{ followingNum }}</strong></span>
              <span @click="followersUserClick" style="cursor: pointer;">粉丝 <strong>{{ followersNum }}</strong></span>
            </div>
            <div class="author-actions">
              <el-button
                  :type="isFollowing ? 'danger' : 'primary'"
                  size="small"
                  round
                  @click="toggleFollow"
                  class="follow-button"
              >
                {{ isFollowing ? '取消关注' : '关注' }}
              </el-button>
              <el-button size="small" class="follow-button" round @click="messageAuthor" >私信</el-button>
            </div>
            <div class="interaction-buttons">
              <el-button type="primary" link @click="goToLiked">
                👍 点赞数:{{ likeNum }}
              </el-button>
              <el-button type="warning" link @click="goToFavorites">
                ⭐ 收藏数:{{ collectNum }}
              </el-button>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 博客内容 -->
      <el-card class="section-card">
        <h3 class="section-title">📚 发表内容</h3>
        <el-empty v-if="blogs.length === 0" description="暂无发表内容"/>
        <el-timeline v-else>
          <el-timeline-item
              v-for="(blog, index) in displayedBlogs"
              :key="blog.GUID"
              :timestamp=" (!blog.TEXT ? '发表文章：' : '社区发起：') + formatDate(blog.CREATE_TIME)"
              placement="top"
              type="primary"
          >
            <el-card shadow="hover" class="blog-card" @click="blogMainClick(blog)">
              <h4>{{ blog.BLOG_TITLE }}</h4>
              <p class="blog-summary" v-html="blog.MAINTEXT||blog.TEXT"></p>
            </el-card>
          </el-timeline-item>
        </el-timeline>

        <div v-if="blogs.length > 3" class="collapse-toggle">
          <el-button link type="primary" @click="showAllBlogs = !showAllBlogs">
            {{ showAllBlogs ? '收起博客内容' : '查看全部博客内容' }}
          </el-button>
        </div>
      </el-card>

      <!-- 上传的文件 -->
      <el-card class="section-card" ref="fileSection">
        <h3 class="section-title">📁 上传的文件</h3>
        <el-empty v-if="files.length === 0" description="暂无上传文件"/>
        <el-table
            v-else
            :data="files"
            stripe
            border
            style="width: 100%"
            class="file-table"
        >
          <el-table-column prop="ORIGINALFILENAME" label="文件名"/>
          <el-table-column prop="DOWNNUM" label="下载次数" width="100"/>
          <el-table-column
              prop="CREATE_TIME"
              label="上传时间"
              :formatter="(row) => formatDate(row.CREATE_TIME)"
          />
          <el-table-column label="操作" width="100">
            <template #default="scope">
              <el-button type="primary" link @click="downloadFile(scope.row)">
                下载
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
  <el-dialog
      v-model="showDialog"
      width="80%"
      destroy-on-close
      top="4vh"
      @close="onDialogClose"
  >
    <ContentAndComment :blogId="selectedBlogId"/>
  </el-dialog>
  <!-- 通用卡片结构（点赞/收藏列表都用） -->
  <el-dialog v-model="showLikeDialog" title="👍 点赞列表" width="60%" :modal-class="'fixed-dialog-height'" top="6vh">
    <div class="card-list-scroll">
      <div class="card-list">
        <!-- 卡片内容保持统一结构 -->
        <el-card
            v-for="item in likeList"
            :key="item.GUID"
            class="list-card"
            shadow="hover"
            @click="blogMainClick(item)"
        >
          <div class="card-top">
            <span class="author-name" style="color: #0029fc">作者:{{ item.USERNAME }}</span>
            <span class="card-time">{{ formatDate(item.CREATE_TIME) }}</span>
          </div>
          <div class="card-body">
            <h4 class="card-title">{{ item.BLOG_TITLE }}</h4>
            <p class="card-summary" v-html="item.MAINTEXT"></p>
          </div>
        </el-card>
      </div>
    </div>
  </el-dialog>

  <el-dialog v-model="showCollectDialog" title="⭐ 收藏列表" width="60%" :modal-class="'fixed-dialog-height'" top="6vh">
    <div class="card-list-scroll">
      <div class="card-list">
        <!-- 卡片内容保持统一结构 -->
        <el-card
            v-for="item in collectList"
            :key="item.GUID"
            class="list-card"
            shadow="hover"
            @click="blogMainClick(item)"
        >
          <div class="card-top">
            <span class="author-name" style="color: #0029fc">作者:{{ item.USERNAME }}</span>
            <span class="card-time">{{ formatDate(item.CREATE_TIME) }}</span>
          </div>
          <div class="card-body">
            <h4 class="card-title">{{ item.BLOG_TITLE }}</h4>
            <p class="card-summary" v-html="item.MAINTEXT"></p>
          </div>
        </el-card>

      </div>
    </div>
  </el-dialog>
  <!-- 粉丝列表-->
  <el-dialog v-model="showFollowersUser" title="🙋粉丝列表" width="60%" :modal-class="'fixed-dialog-height'" top="6vh">
    <div class="card-list-scroll">
      <div class="card-list">
        <!-- 卡片内容保持统一结构 -->
        <el-card
            v-for="item in followersUser"
            :key="item.CODE"
            class="list-card-follow"
            shadow="hover"
            @click="openFollowersUser(item)"
        >
          <div class="follower-card-content">
            <el-avatar
                :src="item.avatar"
                size="large"
                class="author-avatar-follow"
                alt="用户头像"
            >
              {{ item.NAME?.charAt(0) }}
            </el-avatar>
            <span class="author-name">{{ item.NAME }}</span>
          </div>
        </el-card>
      </div>
    </div>
  </el-dialog>
  <!-- 关注列表-->
  <el-dialog v-model="showFollowingUser" title="👀关注列表" width="60%" :modal-class="'fixed-dialog-height'" top="6vh">
    <div class="card-list-scroll">
      <div class="card-list">
        <!-- 卡片内容保持统一结构 -->
        <el-card
            v-for="item in followingUser"
            :key="item.CODE"
            class="list-card-follow"
            shadow="hover"
            @click="openFollowingUser(item)"
        >
          <div class="follower-card-content">
            <el-avatar
                :src="item.avatar"
                size="large"
                class="author-avatar-follow"
                alt="用户头像"
            >
              {{ item.NAME?.charAt(0) }}
            </el-avatar>
            <span class="author-name">{{ item.NAME }}</span>
          </div>
        </el-card>
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import {ref, onMounted, computed} from 'vue'
import {ElMessage} from 'element-plus'
import {useRoute, useRouter} from "vue-router";
import {pubFormatDate, decrypt, encrypt, sendAxiosRequest, downloadFileByUrl} from "@/utils/common.js";
import ContentAndComment from "@/views/detail/blog/ContentAndComment.vue";
import {adminUserCode} from "@/config/vue-config.js";
import {useUserStore} from "@/stores/main/user.js";

const userStore = useUserStore();
userStore.initFromLocal();

const showLikeDialog = ref(false)
const showCollectDialog = ref(false)
//点赞数
const likeNum = ref(0);
//点赞列表
const likeList = ref([]);
//收藏数
const collectNum = ref(0);
//收藏列表
const collectList = ref([]);

const route = useRoute();
const router = useRouter();
const user = ref({})

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

const showFollowersUser = ref(false);
const showFollowingUser = ref(false);
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
      followUserCode: user.value.code,
      followUserName: user.value.name
    });

    //取消关注
  } else {
    followersNum.value--;
    sendAxiosRequest("/blog-api/userInformation/noFollowUser", {followUserCode: user.value.code});
  }
  ElMessage.success(isFollowing.value ? '已关注' : '已取消关注')
}

const followingUserClick = () => {
  if (followingUser.value.length == 0) {
    ElMessage.success("他很高冷,没有关注任何人");
    return false;
  }
  showFollowingUser.value = true;
}

const followersUserClick = () => {
  if (followersUser.value.length == 0) {
    ElMessage.success("他没有粉丝,仍需努力");
    return false;
  }
  showFollowersUser.value = true;
}

const openFollowersUser = (item) => {
  const routeUrl = router.resolve({name: 'personInfomation', query: {c: encrypt(item.CODE)}}).href;
  window.open(routeUrl, item.CODE);
}

const openFollowingUser = (item) => {
  const routeUrl = router.resolve({name: 'personInfomation', query: {c: encrypt(item.CODE)}}).href;
  window.open(routeUrl, item.CODE);
}

async function getUserInfo2Data() {
  let userCode = route.query.c;
  if (userCode) {
    userCode = decrypt(userCode);
    //获取账号信息
    let result = await sendAxiosRequest("/pub-api/login/getUserInfoByCode", {userCode});
    if (result && !result.isError) {
      user.value = result.result;
    }
    //获取账号发布得内容和文件
    result = await sendAxiosRequest("/blog-api/userInformation/getBlogAndResourceByUserCode", {userCode});
    if (result && !result.isError) {
      let arr = [...result.result.blog, ...result.result.community];
      arr.sort((a, b) => new Date(b.CREATE_TIME) - new Date(a.CREATE_TIME));
      blogs.value = arr;
      files.value = result.result.resource;
    }
    result = await sendAxiosRequest("/blog-api/blog/getLikeAndCollectByUserCode", {userCode});
    if (result && !result.isError) {
      result.result.forEach(item => {
        if (item["TYPE"] == "like") {
          likeNum.value++;
          likeList.value.push(item);
        } else if (item["TYPE"] == "collect") {
          collectNum.value++;
          collectList.value.push(item);
        }
      })
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

getUserInfo2Data();

const blogs = ref([])
const files = ref([])

const showAllBlogs = ref(false)
const displayedBlogs = computed(() => {
  return showAllBlogs.value ? blogs.value : blogs.value.slice(0, 3)
})

const showDialog = ref(false);
const selectedBlogId = ref("");

function blogMainClick(blog) {
  if (blog.TEXT) {
    ElMessage.success("社区发起内容不用查看详情");
    return false;
  }
  selectedBlogId.value = blog.GUID;
  showDialog.value = true;
}

const onDialogClose = () => {
  selectedBlogId.value = ''
  showDialog.value = false
}

const formatDate = (dateStr) => {
  return pubFormatDate(dateStr);
}

const downloadFile = (file) => {

  ElMessage.success(`准备下载：${file.ORIGINALFILENAME}`);
  downloadFileByUrl(file.FILEVIEWURL, file.ORIGINALFILENAME);
}

const fileSection = ref(null)

const scrollToFiles = () => {
  if (fileSection.value?.$el) {
    fileSection.value.$el.scrollIntoView({behavior: 'smooth'})
  }
}

const goToLiked = () => {
  if (likeList.value.length == 0) {
    ElMessage.success("他没有点赞的文章,也没有喜欢的人");
    return false;
  }
  showLikeDialog.value = true;
}

const goToFavorites = () => {
  if (collectList.value.length == 0) {
    ElMessage.success("他没有收藏的文章,也没有喜欢的人");
    return false;
  }
  showCollectDialog.value = true;
}

const messageAuthor = () => {
  ElMessage.success("私信作者:" + user.value.name);
}

onMounted(() => {

})
</script>

<style scoped>
.person-info {
  width: 100%;
  background-color: #f5f7fa;
  height: 100%;
  overflow-y: auto;
  max-height: 100%;
  padding: 24px 0;
  box-sizing: border-box;
}

.inner-container {
  width: 100%;
  max-width: 900px;
  margin: 0 auto;
  padding: 0 12px;
  box-sizing: border-box;
}

.toolbar {
  text-align: right;
  margin-bottom: 12px;
}

.profile-card {
  margin-bottom: 24px;
  padding: 24px;
}

.profile-header {
  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 20px;
}

.left-info {
  display: flex;
  align-items: center;
}

.author-avatar {
  width: 100px !important;
  height: 100px !important;
  font-size: 20px;
  background-color: #f2f2f2;
}

.author-avatar-follow {
  width: 60px !important;
  height: 60px !important;
  font-size: 20px;
  background-color: #f2f2f2;
}

.profile-details {
  margin-left: 24px;
}

.right-info {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: flex-end;
  gap: 10px;
}

.interaction-buttons {
  display: flex;
  gap: 10px;
  margin-top: 4px;
}

.username {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin: 0;
  font-size: 24px;
  font-weight: bold;
}

.admin-badge {
  background-color: #ffdf02;
  color: #000;
  font-size: 12px;
  font-weight: bold;
  border-radius: 12px;
  padding: 2px 8px;
  line-height: 1;
}

.admin-badge {
  background-color: #ffdf02;
  color: #000;
  font-size: 12px;
  font-weight: bold;
  border-radius: 12px;
  padding: 2px 8px;
  line-height: 1; /* 避免徽章高度不一致 */
}

.user-remark {
  font-size: 14px;
  color: #555;
  margin: 4px 0;
  max-width: 100%;
  display: inline-block;

  /* 核心：最多显示一行并省略号 */
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;

  /* 美化：视觉上更柔和 */
  background-color: #f8f8f8;
  padding: 4px 8px;
  border-radius: 6px;
  font-style: italic;
  max-width: 300px; /* 或者你设置 profile-details 的固定宽度 */
}


.userip {
  color: #888;
  margin: 4px 0;
}

.stats {
  margin-top: 8px;
  font-size: 14px;
  color: #666;
  display: flex;
  gap: 24px;
}

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

.follow-button {
  margin-top: 12px;
  transition: all 0.3s ease;
}

.section-card {
  margin-bottom: 30px;
  padding: 16px;
}

.section-title {
  font-size: 18px;
  margin-bottom: 16px;
  font-weight: bold;
}

.blog-card {
  transition: 0.3s;
  cursor: pointer;
}

.blog-card:hover {
  transform: scale(1.01);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.blog-summary {
  display: -webkit-box;
  -webkit-line-clamp: 8; /* 最多显示 8 行 */
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.6; /* 可调，建议和字体配合设置 */
  color: #666;
  margin-top: 8px;
}

.file-table :deep(.el-button) {
  font-weight: bold;
}

.collapse-toggle {
  margin-top: 12px;
  text-align: center;
}

/* 弹窗固定高度 */
:deep(.fixed-dialog-height .el-dialog) {
  max-height: 500px;
  display: flex;
  flex-direction: column;
}

/* 弹窗 body 内部滚动 */
:deep(.fixed-dialog-height .el-dialog__body) {
  flex: 1;
  overflow: hidden; /* ⚠️ 禁止 body 滚动 */
  padding-top: 0;
}

.card-list-scroll {
  max-height: 560px;
  overflow-y: auto;
  padding: 12px 16px; /* 上下12px，左右16px */
  box-sizing: border-box;
}

/* 卡片列表 */
.card-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.list-card {
  height: 160px;
  padding: 12px 16px;
  background-color: #edf8f4;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  transition: all 0.3s ease;
  cursor: pointer;
}

.list-card:hover {
  transform: scale(1.01);
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
}

.card-top {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: #666;
  margin-bottom: 4px;
}

.card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  margin: 6px 0;
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow: hidden;
  color: #333;
}

.card-summary {
  color: #666;
  font-size: 13px;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.list-card-follow {
  height: 60px;
  padding: 12px 16px;
  background-color: #ffffff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  transition: all 0.3s ease;
  cursor: pointer;

  /* 加这两行让内容垂直居中 */
  display: flex;
  align-items: center;
}

.follower-card-content {
  display: flex;
  align-items: center; /* 头像 + 姓名垂直居中 */
  gap: 12px; /* 控制头像和名字间距 */
}

.author-avatar-follow {
  width: 60px !important;
  height: 60px !important;
  font-size: 20px;
  background-color: #f2f2f2;
  flex-shrink: 0; /* 避免头像被压缩 */
}

.author-name {
  font-size: 14px;
  font-weight: bold;
  color: #333;
}

@media (max-width: 768px) {
  /* 时间线整体去掉左边距和内边距 */
  .person-info :deep(.el-timeline) {
    padding-left: 0 !important;
    margin-left: 0 !important;
  }

  /* 时间线每一项去除左内边距和左边距 */
  .person-info :deep(.el-timeline-item) {
    padding-left: 0 !important;
    margin-left: 0 !important;
  }

  /* 去除内容wrapper左边距，宽度撑满 */
  .person-info :deep(.el-timeline-item__wrapper) {
    margin-left: 0 !important;
    width: 100% !important;
    padding-left: 0 !important;
  }

  /* 隐藏竖线和节点 */
  .person-info :deep(.el-timeline-item__tail),
  .person-info :deep(.el-timeline-item__node) {
    display: none !important;
  }

  /* 个人信息头部布局调整 */
  .profile-header {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .left-info {
    flex-direction: column;
    align-items: center;
  }

  .profile-details {
    margin-left: 0;
    margin-top: 12px;
  }

  .right-info {
    align-items: center;
    justify-content: center;
    margin-top: 12px;
  }

  .stats {
    justify-content: center;
  }

  .interaction-buttons {
    flex-direction: column;
    align-items: center;
  }

  .follow-button {
    width: 100%;
  }

  .section-title {
    text-align: center;
  }

  .list-card, .list-card-follow {
    height: auto;
    flex-direction: column;
    align-items: flex-start;
  }

  .card-top {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .card-body {
    width: 100%;
  }

  .card-summary {
    -webkit-line-clamp: 4;
  }
}

</style>
