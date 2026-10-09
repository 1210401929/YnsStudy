<template>
  <div class="left-sidebar">
    <div class="sticky-left">
      <div class="align-spacer"></div>

      <el-card class="profile-card" shadow="never">
        <span class="j-tape j-tape--top"></span>
        <div class="profile-header-vertical">
          <el-avatar
              :src="user.avatar"
              size="large"
              class="author-avatar"
              alt="用户头像"
              @click="avatarClick(user)"
              title="进入主页"
          >
            {{ user.name?.charAt(0) }}
          </el-avatar>

          <div class="profile-details-vertical">
            <h2 class="username-vertical">
              <span class="name">{{ user.name }}</span>
              <div class="badges-container">
                <span v-if="user.code === adminUserCode" class="public-badge superAdmin-badge">超级管理员</span>
                <span v-if="user.role === 'admin'" class="public-badge admin-badge">管理员</span>
                <span v-if="user.isban === '1'" class="public-badge ban-badge">已封禁</span>
              </div>
            </h2>
            <p class="user-info-text">邮箱: {{ user.email || "未知" }}</p>
            <p class="user-remark-vertical" :title="user.remark">{{ user.remark || "这个人很神秘~" }}</p>
            <p class="userip-text">IP: {{ user.loginaddress || "未知" }}</p>
          </div>

          <div class="stats-vertical">
            <span @click="followingUserClick" class="stat-item">关注 <strong>{{ followingNum }}</strong></span>
            <div class="stat-divider"></div>
            <span @click="followersUserClick" class="stat-item">粉丝 <strong>{{ followersNum }}</strong></span>
          </div>

          <div class="author-actions-vertical">
            <el-button
                :type="isFollowing ? 'danger' : 'primary'"
                size="default"
                round
                @click="toggleFollow"
                class="follow-button action-btn"
            >
              {{ isFollowing ? '取消关注' : '关注作者' }}
            </el-button>
            <el-button size="default" round @click="messageAuthor" class="action-btn">私聊沟通</el-button>
          </div>

          <div class="interaction-buttons-vertical">
            <el-button type="primary" link @click="goToLiked" class="icon-btn">👍 点赞数: {{ likeNum }}</el-button>
            <el-button type="warning" link @click="goToFavorites" class="icon-btn">⭐ 收藏数: {{ collectNum }}</el-button>
          </div>
        </div>
      </el-card>
    </div>

    <InteractionListDialog v-model="showLikeDialog" title="👍 点赞列表" :list-data="likeList" type="blog" @item-click="handleBlogClick"/>
    <InteractionListDialog v-model="showCollectDialog" title="⭐ 收藏列表" :list-data="collectList" type="blog" @item-click="handleBlogClick"/>
    <InteractionListDialog v-model="showFollowersUser" title="🙋粉丝列表" :list-data="followersUser" type="user" @item-click="openFollowersUser"/>
    <InteractionListDialog v-model="showFollowingUser" title="👀关注列表" :list-data="followingUser" type="user" @item-click="openFollowingUser"/>
    <Chat v-if="chatVisible" :title="'与' + user.name + '的聊天'" @closeChat="toggleChatWindow" />
  </div>

</template>

<script setup>
import { ref, watch, defineAsyncComponent } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { sendAxiosRequest, sendNotifications } from '@/utils/common.js'
import { pubOpenUser } from "@/utils/blogUtil.js"
import { adminUserCode } from '@/config/vue-config.js'
import { useUserStore } from '@/stores/main/user.js'

const Chat = defineAsyncComponent(() => import("@/components/detail/Chat.vue"))
const InteractionListDialog = defineAsyncComponent(() => import("@/components/detail/personInformation/InteractionListDialog.vue"))

const props = defineProps({
  user: {
    type: Object,
    required: true,
    default: () => ({})
  },
  targetUserCode: {
    type: [String, Number],
    required: true
  }
})

const emit = defineEmits(['blog-click'])

const router = useRouter()
const userStore = useUserStore()

const showLikeDialog = ref(false)
const showCollectDialog = ref(false)
const likeNum = ref(0)
const likeList = ref([])
const collectNum = ref(0)
const collectList = ref([])

const isFollowing = ref(false)
const followersNum = ref(0)
const followersUser = ref([])
const followingNum = ref(0)
const followingUser = ref([])
const showFollowersUser = ref(false)
const showFollowingUser = ref(false)

// 获取关注数据
const fetchFollows = async () => {
  if (!props.targetUserCode) return;
  const result = await sendAxiosRequest('/blog-api/userInformation/getFollowUser', { userCode: props.targetUserCode, isCountOnly: "true" });
  if (result && !result.isError) {
    followersNum.value = result.result.followersNum || 0;
    followingNum.value = result.result.followingNum || 0;
    const isFollowCount = result.result.isFollowingCount || 0;
    isFollowing.value = isFollowCount > 0;
  }
}

// 获取点赞收藏统计
const fetchInteractions = async () => {
  if (!props.targetUserCode) return;
  const result = await sendAxiosRequest('/blog-api/blog/getLikeAndCollectByUserCode', {
    userCode: props.targetUserCode,
    isCountOnly: "true"
  });
  if (result && !result.isError) {
    result.result.forEach(item => {
      if (item.TYPE === 'like') {
        likeNum.value = item.TOTAL;
      } else if (item.TYPE === 'collect') {
        collectNum.value = item.TOTAL;
      }
    });
  }
}

// 监听 targetUserCode 变化，初始化数据
watch(() => props.targetUserCode, (newVal) => {
  if (newVal) {
    fetchFollows();
    fetchInteractions();
  }
}, { immediate: true })

// 关注/取消关注
const toggleFollow = () => {
  if (!userStore.userBean.code) {
    ElMessage.error('用户过期,请返回主页面重新登录!')
    return false
  }
  isFollowing.value = !isFollowing.value
  if (isFollowing.value) {
    followersNum.value++
    sendAxiosRequest('/blog-api/userInformation/followUser', {
      followUserCode: props.user.code,
      followUserName: props.user.name
    })
    sendNotifications(userStore.userBean.code, props.user.code, "followUser", null, `${userStore.userBean.name}关注了你`)
  } else {
    followersNum.value--
    sendAxiosRequest('/blog-api/userInformation/noFollowUser', { followUserCode: props.user.code })
  }
  ElMessage.success(isFollowing.value ? '已关注' : '已取消关注')
}

//用户头像点击
const avatarClick = (userInfo)=>{
  pubOpenUser(router, userInfo.code);
}

// 列表数据加载
const loadFollowingList = async () => {
  if (followingUser.value.length > 0) return;
  const result = await sendAxiosRequest('/blog-api/userInformation/getFollowUser', { userCode: props.targetUserCode });
  if (result && !result.isError) {
    followingUser.value = result.result.followingUser;
  }
}

const loadFollowersList = async () => {
  if (followersUser.value.length > 0) return;
  const result = await sendAxiosRequest('/blog-api/userInformation/getFollowUser', { userCode: props.targetUserCode });
  if (result && !result.isError) {
    followersUser.value = result.result.followersUser;
  }
}

const followingUserClick = async () => {
  showFollowingUser.value = true;
  await loadFollowingList();
}

const followersUserClick = async () => {
  showFollowersUser.value = true;
  await loadFollowersList();
}

const openFollowersUser = (item) => pubOpenUser(router, item.CODE);
const openFollowingUser = (item) => pubOpenUser(router, item.CODE);

const loadLikeAndFavoritesList = async (type) => {
  const result = await sendAxiosRequest('/blog-api/blog/getLikeAndCollectByUserCode', {
    userCode: props.targetUserCode,
    isCountOnly: "false",
    type
  });
  if (result && !result.isError) {
    return result.result || [];
  }
  return [];
}

const goToLiked = async () => {
  showLikeDialog.value = true;
  if (likeList.value.length > 0) return;
  likeList.value = await loadLikeAndFavoritesList("like");
}

const goToFavorites = async () => {
  showCollectDialog.value = true;
  if (collectList.value.length > 0) return;
  collectList.value = await loadLikeAndFavoritesList("collect");
}

// 新增聊天窗口显示状态
const chatVisible = ref(false);

// 替换掉原来的 messageAuthor 方法，接管聊天逻辑
const messageAuthor = () => {
  toggleChatWindow();
}

// 聊天窗口开关逻辑，自带身份校验
const toggleChatWindow = () => {
  if (!userStore.userBean.code) {
    ElMessage.error("用户过期,请返回主页面重新登录!");
    return false;
  }
  if (userStore.userBean.code === props.user.code) {
    ElMessage.error("和自己就别聊了");
    return false;
  }
  chatVisible.value = !chatVisible.value;
};

const handleBlogClick = (blog) => {
  emit('blog-click', blog)
}
</script>

<style scoped>
.left-sidebar {
  width: 280px;
  flex-shrink: 0;
}

/* 资料卡：一张贴着胶带的卡片 */
.profile-card {
  position: relative;
  overflow: visible;
  border: 1px solid var(--j-rule);
  border-radius: 2px;
  background: var(--j-paper);
  box-shadow: var(--j-shadow) !important;
}

.profile-card :deep(.el-card__body) {
  padding: 30px 22px 20px;
}

.profile-header-vertical {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  text-align: center;
}

/* 头像像一张拍立得 */
.author-avatar {
  width: 92px !important;
  height: 92px !important;
  border: 5px solid #fff;
  border-bottom-width: 14px;
  border-radius: 0 !important;
  box-shadow: 0 1px 3px rgba(60, 50, 30, 0.25);
  background: #ece4d3;
  color: var(--j-ink-soft);
  font-family: var(--j-hand);
  font-size: 34px;
  transform: rotate(-3deg);
  cursor: pointer;
  transition: transform 0.2s ease;
}

.author-avatar:hover {
  transform: rotate(0deg);
}

.profile-details-vertical {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  width: 100%;
}

.username-vertical {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  margin: 0;
}

.username-vertical .name {
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 26px;
  color: var(--j-ink);
}

.badges-container {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 6px;
}

.user-info-text,
.userip-text {
  margin: 0;
  font-size: 12px;
  color: var(--j-muted);
}

.user-remark-vertical {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  width: 100%;
  margin: 4px 0;
  padding: 8px 12px;
  box-sizing: border-box;
  background: #fdf3c4;
  font-family: var(--j-hand);
  font-size: 15px;
  line-height: 1.6;
  color: var(--j-ink-soft);
  transform: rotate(-0.5deg);
}

.stats-vertical {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 18px;
  width: 100%;
  padding: 12px 0;
  border-top: 1px dashed var(--j-rule-strong);
  border-bottom: 1px dashed var(--j-rule-strong);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  font-family: var(--j-hand);
  font-size: 14px;
  color: var(--j-ink-soft);
  cursor: pointer;
}

.stat-item:hover {
  color: var(--j-pen);
}

.stat-item strong {
  font-family: var(--j-num);
  font-weight: normal;
  font-size: 22px;
  color: var(--j-ink);
}

.stat-divider {
  width: 1px;
  height: 28px;
  background-color: var(--j-rule-strong);
}

.author-actions-vertical {
  display: flex;
  justify-content: center;
  gap: 10px;
  width: 100%;
}

.action-btn {
  flex: 1;
  margin: 0 !important;
}

.interaction-buttons-vertical {
  display: flex;
  justify-content: space-around;
  width: 100%;
  padding: 8px 0 0;
}

.icon-btn {
  font-size: 13px;
}

/* 身份：小印章 */
.public-badge {
  padding: 1px 8px;
  border: 1.5px solid currentColor;
  border-radius: 3px;
  font-family: var(--j-hand);
  font-size: 13px;
  line-height: 18px;
  background: transparent;
}

.superAdmin-badge {
  color: var(--j-stamp);
  transform: rotate(-4deg);
}

.admin-badge {
  color: var(--j-pen);
  transform: rotate(3deg);
}

.ban-badge {
  color: #fff;
  background: var(--j-stamp);
  border-color: var(--j-stamp);
}

@media (max-width: 1024px) {
  .left-sidebar {
    width: 100%;
    position: static;
  }

  .align-spacer {
    display: none;
  }

  .profile-header-vertical {
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: space-between;
    text-align: left;
  }

  .profile-details-vertical {
    align-items: flex-start;
    width: auto;
    flex: 1;
    margin-left: 20px;
  }

  .username-vertical {
    flex-direction: row;
  }

  .author-actions-vertical,
  .interaction-buttons-vertical {
    width: auto;
  }
}

@media (max-width: 768px) {
  .profile-header-vertical {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .profile-details-vertical {
    margin-left: 0;
    align-items: center;
  }

  .username-vertical {
    flex-direction: column;
  }
}
</style>
