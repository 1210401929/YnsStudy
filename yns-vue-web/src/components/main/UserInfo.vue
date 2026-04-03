<template>
  <div class="left-sidebar">
    <div class="sticky-left">
      <div class="align-spacer"></div>

      <el-card class="profile-card">
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

.profile-card {
  background-color: rgba(255, 255, 255, 0.5) !important;
  border: 1px solid rgba(255, 255, 255, 0.5);
  border-radius: 12px;
  box-shadow: 0 4px 18px rgba(0, 0, 0, 0.08);
  padding: 24px 20px;
}

.profile-header-vertical {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.author-avatar {
  width: 90px !important;
  height: 90px !important;
  font-size: 24px;
  background-color: #f2f2f2;
  border: 3px solid rgba(255,255,255,0.8);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  margin-bottom: 12px;
  cursor: pointer;
}

.profile-details-vertical {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.username-vertical {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  margin: 0 0 12px 0;
}

.username-vertical .name {
  font-size: 22px;
  font-weight: bold;
  color: #2c3e50;
}

.badges-container {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 6px;
}

.user-info-text, .userip-text {
  font-size: 13px;
  color: #666;
  margin: 4px 0;
}

.user-remark-vertical {
  font-size: 13px;
  color: #555;
  margin: 12px 0;
  background-color: rgba(248, 248, 248, 0.6);
  padding: 8px 12px;
  border-radius: 6px;
  font-style: italic;
  width: 100%;
  box-sizing: border-box;
  word-break: break-all;
}

.stats-vertical {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  gap: 20px;
  margin: 16px 0;
  padding: 12px 0;
  border-top: 1px dashed rgba(0,0,0,0.06);
  border-bottom: 1px dashed rgba(0,0,0,0.06);
}

.stat-item {
  font-size: 14px;
  color: #606266;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  transition: color 0.2s;
}

.stat-item:hover { color: #409eff; }
.stat-item strong { font-size: 18px; color: #303133; }
.stat-divider { width: 1px; height: 24px; background-color: rgba(0,0,0,0.08); }

.author-actions-vertical {
  display: flex;
  width: 100%;
  gap: 12px;
  margin-bottom: 16px;
}

.action-btn {
  flex: 1;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
}

.interaction-buttons-vertical {
  display: flex;
  width: 100%;
  justify-content: space-around;
  background: rgba(0,0,0,0.02);
  border-radius: 8px;
  padding: 8px 0;
}

.public-badge {
  color: #000;
  font-size: 12px;
  font-weight: bold;
  border-radius: 10px;
  padding: 2px 8px;
  line-height: 1.2;
}

.superAdmin-badge { background-color: #ffdf02; }
.admin-badge { background-color: #86ff93; }
.ban-badge { background-color: #ff5a5a; }

/* 响应式调整 */
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