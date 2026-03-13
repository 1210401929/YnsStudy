<template>
  <div class="person-info" :style="currentBgStyle">
    <WelcomeOverlay
        :visible="showWelcome"
        :is-self="isSelf"
        :user="user"
        @enter="enterHomepage"
    />

    <BackgroundAndMusic
        ref="bgMusicComponentRef"
        :is-self="isSelf"
        :user-name="user.name"
        :init-bg-image="serverBgImage"
        :init-bg-audio="serverBgAudio"
        @update-bg-style="handleBgStyleUpdate"
    />

    <div class="inner-container">
      <div class="toolbar">
        <el-button type="success" size="small" @click="scrollToFiles">
          📁 跳转到文件列表
        </el-button>
      </div>

      <el-card class="profile-card">
        <div class="profile-header">
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
                <span v-if="user.code === adminUserCode" class="public-badge superAdmin-badge">超级管理员</span>
                <span v-if="user.role === 'admin'" class="public-badge admin-badge">管理员</span>
                <span v-if="user.isban === '1'" class="public-badge ban-badge">已封禁</span>
              </h2>
              <p>邮箱: {{ user.email || "未知" }}</p>
              <p class="user-remark" :title="user.remark">个性签名: {{ user.remark || "这个人很神秘~" }}</p>
              <p class="userip">IP: {{ user.loginaddress || "未知" }}</p>
            </div>
          </div>

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
              <el-button size="small" class="follow-button" round @click="messageAuthor">私信</el-button>
            </div>
            <div class="interaction-buttons">
              <el-button type="primary" link @click="goToLiked">👍 点赞数:{{ likeNum }}</el-button>
              <el-button type="warning" link @click="goToFavorites">⭐ 收藏数:{{ collectNum }}</el-button>
            </div>
          </div>
        </div>
      </el-card>

      <el-card class="section-card">
        <div class="section-header">
          <h3 class="section-title">📚 发表内容</h3>
          <el-switch
              v-model="onlyArticle"
              size="small"
              active-text="仅文章"
              style="margin-left: auto;"
          />
        </div>
        <el-timeline v-if="showBlogs.length>0">
          <el-timeline-item
              v-for="(blog, index) in showBlogs"
              :key="blog.GUID"
              :timestamp="(blog.TYPE=='blog' ? '发表文章：' : '社区发起：') + formatDate(blog.CREATE_TIME)"
              placement="top"
              type="primary"
          >
            <el-card shadow="hover" class="blog-card" @click="blogMainClick(blog)">
              <h4>{{ blog.BLOG_TITLE }}</h4>
              <p class="blog-summary" v-html="stripImages(blog.MAINTEXT)"></p>
            </el-card>
          </el-timeline-item>
        </el-timeline>

        <el-empty v-if="!showBlogs.length && !loading" description="暂无内容"/>
        <el-button
            v-if="!noMore && !loading"
            type="primary"
            link
            @click="fetchArticles(null)"
            style="margin: 20px auto; display: block;"
        >
          加载更多
        </el-button>
        <div v-if="loading" class="loading-text">加载中...</div>
        <div v-if="noMore" class="end-text">没有更多内容了</div>
      </el-card>

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
              <el-button type="primary" link @click="downloadFile(scope.row)">下载</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
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
    <InteractionListDialog
        v-model="showLikeDialog"
        title="👍 点赞列表"
        :list-data="likeList"
        type="blog"
        @item-click="blogMainClick"
    />

    <InteractionListDialog
        v-model="showCollectDialog"
        title="⭐ 收藏列表"
        :list-data="collectList"
        type="blog"
        @item-click="blogMainClick"
    />

    <InteractionListDialog
        v-model="showFollowersUser"
        title="🙋粉丝列表"
        :list-data="followersUser"
        type="user"
        @item-click="openFollowersUser"
    />

    <InteractionListDialog
        v-model="showFollowingUser"
        title="👀关注列表"
        :list-data="followingUser"
        type="user"
        @item-click="openFollowingUser"
    />
  </div>
</template>

<script setup>
import {ref, onMounted, computed, defineAsyncComponent} from 'vue'
import {ElMessage} from 'element-plus'
import {useRoute, useRouter} from 'vue-router'
import {
  pubFormatDate,
  decrypt,
  sendAxiosRequest,
  stripImages,
  downloadFileByUrl,
  sendNotifications, getUserInfoByCode
} from '@/utils/common.js'

//引入背景和音乐相关组件
import BackgroundAndMusic from "@/components/detail/personInformation/BackgroundAndMusic.vue";
//引入欢迎页相关组件
import WelcomeOverlay from "@/components/detail/personInformation/WelcomeOverlay.vue";
//引入正文和评论相关组件  改成动态获取组件=>defineAsyncComponent：当真正调用这个组件时，才会调用
const ContentAndComment = defineAsyncComponent(() => import('@/views/detail/blog/ContentAndComment.vue'))
//引入列表组件（收藏，点赞列表；关注，被关注列表）
const InteractionListDialog = defineAsyncComponent(() => import("@/components/detail/personInformation/InteractionListDialog.vue"))
import {adminUserCode} from '@/config/vue-config.js'
import {useUserStore} from '@/stores/main/user.js'
import {pubOpenUser} from "@/utils/blogUtil.js";

const userStore = useUserStore()
userStore.initFromLocal()

const showLikeDialog = ref(false)
const showCollectDialog = ref(false)
const likeNum = ref(0)
const likeList = ref([])
const collectNum = ref(0)
const collectList = ref([])

const route = useRoute()
const router = useRouter()
const user = ref({})

//当前打开用户code
const targetUserCode = ref('');
try {
  const raw = route.params.u;
  targetUserCode.value = raw ? decrypt(raw) : '';
} catch (e) {
  targetUserCode.value = '';
}

// 是否访问自己的主页
const isSelf = computed(() =>
    !!userStore.userBean.code &&
    !!targetUserCode.value &&
    targetUserCode.value === userStore.userBean.code
);

// ==== 背景与音乐：父组件状态对接 ====
const bgMusicComponentRef = ref(null);
const serverBgImage = ref('');
const serverBgAudio = ref('');
const currentBgStyle = ref({});

const handleBgStyleUpdate = (style) => {
  currentBgStyle.value = style;
}

// 关注相关
const isFollowing = ref(false)
const followersNum = ref(0)
const followersUser = ref([])
const followingNum = ref(0)
const followingUser = ref([])

const showFollowersUser = ref(false)
const showFollowingUser = ref(false)
const toggleFollow = () => {
  if (!userStore.userBean.code) {
    ElMessage.error('用户过期,请返回主页面重新登录!')
    return false
  }
  isFollowing.value = !isFollowing.value
  if (isFollowing.value) {
    followersNum.value++
    sendAxiosRequest('/blog-api/userInformation/followUser', {
      followUserCode: user.value.code,
      followUserName: user.value.name
    })
    sendNotifications(userStore.userBean.code, user.value.code, "followUser", null, `${userStore.userBean.name}关注了你`)

  } else {
    followersNum.value--
    sendAxiosRequest('/blog-api/userInformation/noFollowUser', {followUserCode: user.value.code})
  }
  ElMessage.success(isFollowing.value ? '已关注' : '已取消关注')
}

// 获取关注列表
const loadFollowingList = async () => {
  if (followingUser.value.length > 0) return;
  const result = await sendAxiosRequest('/blog-api/userInformation/getFollowUser', {userCode: targetUserCode.value});
  if (result && !result.isError) {
    followingUser.value = result.result.followingUser;
  }
}

// 获取粉丝列表
const loadFollowersList = async () => {
  if (followersUser.value.length > 0) return; // 已经有数据就不再请求（缓存）
  // 注意：真实场景中，如果数据极多，这里还需要加上 page 和 pageSize 做分页
  const result = await sendAxiosRequest('/blog-api/userInformation/getFollowUser', {userCode: targetUserCode.value});
  if (result && !result.isError) {
    followersUser.value = result.result.followersUser;
  }
}


const followingUserClick = async () => {
  // if (followingUser.value.length === 0) {
  //ElMessage.success('他很高冷,没有关注任何人')
  //   return false
  // }
  showFollowingUser.value = true;
  await loadFollowingList();
}

const followersUserClick = async () => {
  // if (followersUser.value.length === 0) {
  //ElMessage.success('他没有粉丝,仍需努力')
  //   return false
  // }
  showFollowersUser.value = true;
  await loadFollowersList();
}

const openFollowersUser = (item) => {
  pubOpenUser(router, item.CODE);
}
const openFollowingUser = (item) => {
  pubOpenUser(router, item.CODE);
}

const page = ref(1)
const pageSize = 2
const loading = ref(false)
const noMore = ref(false)

const fetchArticles = async (userCode) => {
  if (loading.value || noMore.value) return;
  if (!userCode) userCode = decrypt(route.params.u);
  loading.value = true
  try {
    const res = await sendAxiosRequest('/blog-api/userInformation/getBlogAndCommunityByUserCode', {
      userCode,
      page: page.value,
      pageSize,
      keyword: ""
    })
    const newData = res.result.data
    if (newData.length < pageSize) {
      noMore.value = true
    }
    blogs.value.push(...newData)
    page.value++
  } catch (e) {
    console.error('获取内容失败', e)
  } finally {
    loading.value = false
  }
}

//获取用户信息
async function getUserInfo2Data() {
  const userCode = targetUserCode.value;
  if (!userCode) return;


  // 1. 获取核心用户信息 (独立运行)
  const fetchBasicInfo = async () => {
    const result = await getUserInfoByCode(userCode);
    if (result && !result.isError) {
      user.value = result.result;

      // 用户名拿到后再去修改浏览器标签页标题
      document.title = user.value.name + "的主页";
      const metaDesc = document.querySelector('meta[name="description"]');
      if (metaDesc) {
        metaDesc.setAttribute("content", document.title);
      } else {
        const desc = document.createElement('meta');
        desc.name = "description";
        desc.content = document.title;
        document.head.appendChild(desc);
      }
    }
  };

  // 2. 获取上传的文件列表 (独立运行)
  const fetchFilesList = async () => {
    const result = await sendAxiosRequest('/blog-api/userInformation/getResourceByUserCode', {userCode});
    if (result && !result.isError) {
      files.value = result.result;
    }
  };

  // 3. 获取点赞和收藏数据 (独立运行)
  const fetchInteractions = async () => {
    const result = await sendAxiosRequest('/blog-api/blog/getLikeAndCollectByUserCode', {
      userCode,
      isCountOnly: "true"
    });
    if (result && !result.isError) {
      result.result.forEach(item => {
        //点赞数
        if (item.TYPE === 'like') {
          likeNum.value = item.TOTAL;
          //收藏数
        } else if (item.TYPE === 'collect') {
          collectNum.value = item.TOTAL;
        }
      });
    }
  };

  // 4. 获取粉丝和关注数量(独立运行)
  const fetchFollows = async () => {
    const result = await sendAxiosRequest('/blog-api/userInformation/getFollowUser', {userCode, isCountOnly: "true"});
    if (result && !result.isError) {
      // 1. 赋值数量（这里假设你的底层把 COUNT(1) 包装成了数组里的对象，具体需要根据你控制台打印的实际结构调整）
      followersNum.value = result.result.followersNum || 0;
      followingNum.value = result.result.followingNum || 0;

      // 2. ✨ 判断当前用户是否关注
      const isFollowCount = result.result.isFollowingCount || 0;
      isFollowing.value = isFollowCount > 0;
    }
  };
  //5.获取背景和音乐
  const setPersonInfo = async () => {
    let result = await sendAxiosRequest("/blog-api/userInformation/getPersonInfo", {userCode: userCode});
    if (result && !result.isError) {
      result = result.result[0] || {};
      // 直接赋值给响应式变量，子组件的 watch 会接收到
      serverBgAudio.value = result.BGMUSICURL || "";
      serverBgImage.value = result.BGIMAGEURL || "";
    }
  }
  // 【核心改变】：同时发射所有请求！不要加 await 让他们排队了
  fetchBasicInfo();
  fetchArticles(userCode); // 文章分页本来就是独立的
  fetchFilesList();
  fetchInteractions();
  fetchFollows();
  setPersonInfo(); // 获取背景和音乐也是独立的
}



const blogs = ref([])
const files = ref([])
const onlyArticle = ref(false)
const showDialog = ref(false)
const selectedBlogId = ref('')

const showBlogs = computed(() => {
  if (!onlyArticle.value)
    return blogs.value;
  else {
    return blogs.value.filter(item => item.TYPE === "blog");
  }
})

function blogMainClick(blog) {
  if (blog.TYPE === "community") {
    ElMessage.success('社区发起内容不用查看详情')
    return false
  }
  selectedBlogId.value = blog.GUID
  showDialog.value = true
}

const onDialogClose = () => {
  selectedBlogId.value = ''
  showDialog.value = false
}

const formatDate = (dateStr) => pubFormatDate(dateStr)

const downloadFile = (file) => {
  ElMessage.success(`准备下载：${file.ORIGINALFILENAME}`)
  downloadFileByUrl(file.FILEVIEWURL, file.ORIGINALFILENAME)
}

const fileSection = ref(null)
const scrollToFiles = () => {
  if (fileSection.value?.$el) {
    fileSection.value.$el.scrollIntoView({behavior: 'smooth'})
  }
}

// 获取关注或收藏列表
const loadLikeAndFavoritesList = async (type) => {
  const result = await sendAxiosRequest('/blog-api/blog/getLikeAndCollectByUserCode', {
    userCode: targetUserCode.value,
    isCountOnly: "false",
    type
  });
  if (result && !result.isError) {
    return result.result || []; // 防御
  }
  return [];
}

//显示点赞列表
const goToLiked = async () => {
  showLikeDialog.value = true;
  //有值则直接返回
  if (likeList.value.length > 0) return;
  likeList.value = await loadLikeAndFavoritesList("like");
}

//显示收藏列表
const goToFavorites = async () => {
  showCollectDialog.value = true;
  //有值则直接返回
  if (collectList.value.length > 0) return;
  collectList.value = await loadLikeAndFavoritesList("collect");
}

const messageAuthor = () => {
  ElMessage.success('私信作者:' + user.value.name)
}

const showWelcome = ref(true);


const enterHomepage = async () => {
  showWelcome.value = false;
  // 调用子组件的播放方法
  if (bgMusicComponentRef.value) {
    bgMusicComponentRef.value.playMusicForce();
  }
};


onMounted(() => {
  // 正常调用
  getUserInfo2Data();
});
</script>

<style scoped>
.person-info.bg-fade {
  animation: bgfade .25s ease;
}

@keyframes bgfade {
  from {
    opacity: .6;
  }
  to {
    opacity: 1;
  }
}

.person-info {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  max-height: 100%;
  padding: 24px 0;
  box-sizing: border-box;
  position: relative;
  background: url('/picture/user/default.webp') no-repeat center center fixed;
  background-size: cover;
  background-attachment: fixed;
  transition: background-image .001s;
}


.inner-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 0 16px;
  box-sizing: border-box;
}

.toolbar {
  text-align: right;
  margin-bottom: 12px;
}

.section-card,
.profile-card {
  background-color: rgba(255, 255, 255) !important;
  border-radius: 12px;
  box-shadow: 0 4px 18px rgba(0, 0, 0, 0.08);
  opacity: 0.9;
}

.profile-card {
  margin-bottom: 24px;
  padding: 24px;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.05);
  background-color: #ffffff;
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

.profile-details {
  margin-left: 24px;
}

.username {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: 24px;
  font-weight: bold;
}

.public-badge {
  color: #000;
  font-size: 12px;
  font-weight: bold;
  border-radius: 10px;
  padding: 2px 8px;
  line-height: 1.2;
  margin-left: 4px;
  flex-shrink: 0;
  position: relative;
  top: 0px;
}

.superAdmin-badge {
  background-color: #ffdf02;
}

.admin-badge {
  background-color: #86ff93;
}

.ban-badge {
  background-color: #ff5a5a;
}

.user-remark {
  font-size: 14px;
  color: #555;
  margin: 4px 0;
  max-width: 300px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  background-color: #f8f8f8;
  padding: 4px 8px;
  border-radius: 6px;
  font-style: italic;
}

.userip {
  color: #888;
  margin: 4px 0;
}

.right-info {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: flex-end;
  gap: 10px;
}

.stats {
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
  padding: 16px 20px;
  border-radius: 12px;
  background-color: #ffffff;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.06);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.section-title {
  font-size: 18px;
  font-weight: bold;
  color: #333;
}

.blog-card {
  transition: 0.3s;
  cursor: pointer;
}

.blog-card:hover {
  transform: scale(1.01);
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
}

.blog-summary {
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.6;
  color: #666;
  margin-top: 8px;
}

.file-table :deep(.el-button) {
  font-weight: bold;
}

.loading-text,
.end-text {
  text-align: center;
  color: #999;
  margin: 16px 0;
}


@media (max-width: 768px) {
  .person-info :deep(.el-timeline),
  .person-info :deep(.el-timeline-item),
  .person-info :deep(.el-timeline-item__wrapper) {
    padding-left: 0 !important;
    margin-left: 0 !important;
  }

  .person-info :deep(.el-timeline-item__wrapper) {
    width: 100% !important;
  }

  .person-info :deep(.el-timeline-item__tail),
  .person-info :deep(.el-timeline-item__node) {
    display: none !important;
  }

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

}
</style>