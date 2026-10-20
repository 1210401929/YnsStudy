<!--用户列表组件-->
<template>
  <el-input
      v-model="searchKeyword"
      placeholder="输入用户名"
      clearable
      prefix-icon="el-icon-search"
      size="small"
      @clear="resetAndLoad"
      @input="debouncedSearch"
      style="margin-bottom: 12px"
  />
  <div class="user-grid">
    <el-card
        v-for="userInfo in allUserArr"
        :key="userInfo.CODE"
        class="user-card"
        shadow="hover"
        @click="userInfoCLick(userInfo)"
    >
      <div class="user-card-content">
        <el-avatar
            :src="userInfo.AVATAR"
            size="default"
            class="author-avatar"
        >
          {{ userInfo.NAME?.charAt(0) }}
        </el-avatar>

        <div class="user-info-text">
          <div class="name-badge-wrapper">
            <span class="user-name" :title="userInfo.NAME">{{ userInfo.NAME }}</span>
            <span v-if="userInfo.CODE === adminUserCode" class="public-badge superAdmin-badge">超级管理员</span>
            <span v-if="userInfo.ROLE === 'admin'" class="public-badge admin-badge">管理员</span>
            <span v-if="userInfo.ISBAN === '1'" class="public-badge ban-badge">已封禁</span>
          </div>

          <span class="user-remark" :title="userInfo.REMARK">{{ userInfo.REMARK || '暂无备注' }}</span>
        </div>

        <div v-if="getCurrentUserAdminObject().adminLevel==='superAdmin' && ISOPERATIONUSER==='true'"
             class="user-actions" @click.stop>
          <el-dropdown trigger="click" @command="(command) => handleUserAction(command, userInfo)">
            <el-button link type="info" :icon="MoreFilled">操作
              <el-icon class="el-icon--right">
                <arrow-down/>
              </el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="detail">查看详情</el-dropdown-item>
                <el-dropdown-item divided command="setAdmin">设为管理员</el-dropdown-item>
                <el-dropdown-item command="removeAdmin">取消管理员</el-dropdown-item>
                <el-dropdown-item divided command="ban" class="text-danger">封禁用户</el-dropdown-item>
                <el-dropdown-item command="removeBan">取消封禁</el-dropdown-item>
                <template v-if="userInfo.CODE !== adminUserCode">
                  <el-dropdown-item divided command="banComment">禁止评论</el-dropdown-item>
                  <el-dropdown-item command="forceLogout">强制下线</el-dropdown-item>
                  <el-dropdown-item command="resetPassword" class="text-danger">重置密码</el-dropdown-item>
                </template>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </el-card>
  </div>

  <div class="list-footer">
    <el-button
        v-if="!noMore && !loading"
        type="primary"
        link
        @click="getAllUserInfo"
        class="load-more-btn"
    >
      加载更多用户
    </el-button>
    <div v-if="noMore && allUserArr.length > 0" class="end-text-v2">- 到底了 -</div>
  </div>

  <CommentBanDialog v-model="banVisible" preset-type="user" :preset-value="banUserCode" lock-type/>

  <!-- 用户详情 -->
  <el-dialog v-model="detailVisible" :title="detailTitle" width="640px" append-to-body>
    <div v-loading="detailLoading">
      <template v-if="detail">
        <div class="detail-grid">
          <div><span class="k">账号</span>{{ detail.user.code }}</div>
          <div><span class="k">用户编号</span>{{ detail.user.usernum }}</div>
          <div><span class="k">手机</span>{{ detail.user.phone || '—' }}</div>
          <div><span class="k">邮箱</span>{{ detail.user.email || '—' }}</div>
          <div><span class="k">最近登录地区</span>{{ detail.user.loginaddress || '—' }}</div>
          <div><span class="k">最近登录 IP</span>{{ detail.user.loginip || '—' }}</div>
        </div>
        <div class="detail-stats">
          <span>公开文章 <b>{{ show(detail.stats.publicArticles) }}</b></span>
          <span>私密文章 <b>{{ show(detail.stats.privateArticles) }}</b></span>
          <span>社区帖子 <b>{{ show(detail.stats.communityPosts) }}</b></span>
          <span>评论 <b>{{ show(detail.stats.comments) }}</b></span>
          <span>登录次数 <b>{{ show(detail.stats.logins) }}</b></span>
        </div>
        <h4 class="detail-sub">最近登录记录</h4>
        <el-table v-if="detail.recentLogins.length" :data="detail.recentLogins" size="small" max-height="260">
          <el-table-column label="时间" min-width="150">
            <template #default="{ row }">{{ pubFormatDate(row.LOGIN_TIME) }}</template>
          </el-table-column>
          <el-table-column prop="LOGINADDRESS" label="地区" min-width="100"/>
          <el-table-column prop="LOGINIP" label="IP" min-width="130"/>
        </el-table>
        <el-empty v-else description="暂无登录记录" :image-size="50"/>
      </template>
    </div>
  </el-dialog>

  <!-- 重置密码结果：临时密码只显示这一次 -->
  <el-dialog v-model="passwordVisible" title="密码已重置" width="420px" append-to-body :close-on-click-modal="false">
    <p>已为 <b>{{ passwordResult.name }}</b>（账号 {{ passwordResult.account }}）生成临时密码，该用户所有设备已退出登录：</p>
    <div class="temp-password">{{ passwordResult.password }}</div>
    <p class="warn">关闭后将无法再次查看，请告知用户登录后尽快在个人中心修改密码。</p>
    <template #footer>
      <el-button @click="copyPassword">复制密码</el-button>
      <el-button type="primary" @click="passwordVisible = false">我已记下</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import {ArrowDown, MoreFilled} from "@element-plus/icons-vue";
import {onMounted, ref} from "vue";
import {ele_confirm, getCurrentUserAdminObject, pubFormatDate, sendAxiosRequest, sendAxiosRequestChecked} from "@/utils/common.js";
import {ElMessage} from "element-plus";
import {pubOpenUser} from "@/utils/blogUtil.js";
import {useRouter} from "vue-router";
import debounce from "lodash/debounce.js";
import {adminUserCode} from "@/config/vue-config.js";
import CommentBanDialog from "@/components/admin/CommentBanDialog.vue";

const router = useRouter();
const props = defineProps({
  ISOPERATIONUSER: String,
})

const searchKeyword = ref("");
const page = ref(1)
const pageSize = 5 //
const loading = ref(false)
const noMore = ref(false)
const allUserArr = ref([]);
//获取用户方法
const getAllUserInfo = async () => {
  if (loading.value || noMore.value) return
  loading.value = true
  try {
    const res = await sendAxiosRequest('/pub-api/login/getAllUserInfo', {
      page: page.value,
      pageSize,
      keyword: searchKeyword.value
    })
    const newData = res.result.data;
    if (newData.length < pageSize) {
      noMore.value = true
    }
    allUserArr.value.push(...newData);
    //超级管理员始终排在最前面,找到超级管理员所在的索引
    const adminIndex = allUserArr.value.findIndex(oneUser => oneUser.CODE === adminUserCode);
    //如果找到了，并且它不在第一位（index > 0）
    if (adminIndex > 0) {
      // 把该条数据从数组中截取出来
      const [adminUser] = allUserArr.value.splice(adminIndex, 1);
      // 插入到数组最前面
      allUserArr.value.unshift(adminUser);
    }
    //用户名长度大于10 将隐藏后几位  1是当用户是手机号时,保护隐私   2是太长也不美观
    allUserArr.value.forEach(oneUser=>{
      if(oneUser.NAME.length>10){
          oneUser.NAME = oneUser.NAME.substring(0,8)+"**";
      }
    });
    page.value++
  } catch (e) {
    console.error('获取用户失败', e)
    ElMessage.error("获取用户列表失败")
  } finally {
    loading.value = false
  }
}

//搜索用户输入框事件
const resetAndLoad = () => {
  page.value = 1
  noMore.value = false
  allUserArr.value = []
  getAllUserInfo()
}

const debouncedSearch = debounce(() => {
  resetAndLoad()
}, 1000)

const userInfoCLick = (userInfo) => {
  pubOpenUser(router, userInfo.CODE);
}

//操作用户 例如:设置管理员,取消管理员,禁用或删除用户
//先确认，确认后再请求后台，以后台结果为准更新页面
const userActions = {
  setAdmin: {
    confirm: (name) => `确定要将【${name}】设为管理员吗？`,
    apply: (userInfo) => { userInfo.ROLE = "admin"; },
    success: (name) => `已将 ${name} 设为管理员`
  },
  removeAdmin: {
    confirm: (name) => `确定要取消【${name}】的管理员权限吗？`,
    apply: (userInfo) => { userInfo.ROLE = "1"; },
    success: (name) => `已取消 ${name} 的管理员权限`
  },
  ban: {
    confirm: (name) => `确定要封禁【${name}】吗？`,
    apply: (userInfo) => { userInfo.ISBAN = "1"; },
    success: (name) => `已封禁用户 ${name}`
  },
  removeBan: {
    confirm: (name) => `确定要取消【${name}】的封禁吗？`,
    apply: (userInfo) => { userInfo.ISBAN = ""; },
    success: (name) => `已取消用户 ${name}封禁`
  }
};

// ---------- 后台：用户详情、强制下线、重置密码 ----------
const detailVisible = ref(false);
const detailLoading = ref(false);
const detail = ref(null);
const detailTitle = ref('用户详情');
const show = (value) => (value === null || value === undefined ? '无数据' : value);

const openDetail = async (userInfo) => {
  detailTitle.value = `用户详情 - ${userInfo.NAME}`;
  detail.value = null;
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    const result = await sendAxiosRequestChecked('/blog-api/admin/getUserDetail', {userId: userInfo.GUID}, '获取用户详情失败');
    if (result) detail.value = result.result;
  } finally {
    detailLoading.value = false;
  }
};

const forceLogout = (userInfo) => {
  ele_confirm(`确定让【${userInfo.NAME}】在所有设备上退出登录吗？`, async () => {
    const result = await sendAxiosRequestChecked('/blog-api/admin/forceLogout', {userId: userInfo.GUID}, '强制下线失败');
    if (result) ElMessage.success(`${userInfo.NAME} 已被强制下线`);
  });
};

const passwordVisible = ref(false);
const passwordResult = ref({});
const resetPassword = (userInfo) => {
  ele_confirm(`确定重置【${userInfo.NAME}】的密码吗？原密码将失效，该用户所有设备会退出登录。`, async () => {
    const result = await sendAxiosRequestChecked('/blog-api/admin/resetPassword', {userId: userInfo.GUID}, '重置密码失败');
    if (!result) return;
    passwordResult.value = {name: userInfo.NAME, ...result.result};
    passwordVisible.value = true;
  });
};

const copyPassword = async () => {
  try {
    await navigator.clipboard.writeText(passwordResult.value.password);
    ElMessage.success('已复制');
  } catch {
    ElMessage.warning('复制失败，请手动选择密码复制');
  }
};

const banVisible = ref(false);
const banUserCode = ref('');
const banComment = (userInfo) => {
  banUserCode.value = userInfo.CODE;
  banVisible.value = true;
};

const extraActions = {detail: openDetail, forceLogout, resetPassword, banComment};

const handleUserAction = (command, userInfo) => {
  if (extraActions[command]) return extraActions[command](userInfo);
  const action = userActions[command];
  if (!action) return;
  ele_confirm(action.confirm(userInfo.NAME), async () => {
    const result = await sendAxiosRequestChecked("/pub-api/login/operationUser", {userId: userInfo.GUID, type: command}, "操作失败");
    if (!result) return;
    action.apply(userInfo);
    ElMessage.success(action.success(userInfo.NAME));
  });
}

onMounted(() => {
  //默认加载用户
  getAllUserInfo();
})

</script>
<style scoped>
/* --- 后台：用户详情与临时密码 --- */
.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px 20px;
  font-size: 14px;
  overflow-wrap: anywhere;
}

.detail-grid .k {
  display: inline-block;
  min-width: 96px;
  color: var(--j-muted);
}

.detail-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin: 16px 0;
  padding: 10px 12px;
  background: var(--j-paper-warm);
  font-size: 13px;
}

.detail-sub {
  margin: 0 0 8px;
  font-weight: normal;
  font-family: var(--j-hand);
  font-size: 16px;
}

.temp-password {
  margin: 12px 0;
  padding: 12px;
  text-align: center;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 22px;
  letter-spacing: 2px;
  background: var(--j-note);
  user-select: all;
}

.warn {
  color: var(--j-stamp);
  font-size: 13px;
}

@media (max-width: 640px) {
  .detail-grid {
    grid-template-columns: 1fr;
  }
}

/* --- 用户网格布局 --- */
.user-grid {
  display: grid;
  /* 自动响应式：每个卡片最小280px，放不下自动换行 */
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.user-card {
  border: 1px solid var(--j-rule);
  border-radius: 2px;
  background: #fffefb;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.user-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.08), 0 10px 18px -12px rgba(60, 50, 30, 0.45) !important;
}

/* 新增：包裹名字和标签的容器 */
.name-badge-wrapper {
  display: flex;
  align-items: flex-start; /* 让内部元素顶部对齐 */
  margin-bottom: 4px;
  width: 100%;
}
.public-badge{
  border: 1px solid currentColor;
  font-size: 10px;
  border-radius: 3px;
  padding: 1px 5px;
  line-height: 1.2;

  /* --- 核心位置微调 --- */
  margin-left: 4px; /* 和名字拉开一点距离 */
  flex-shrink: 0; /* 关键：保证名字再长，标签都不会被压缩变形 */
  position: relative;
  top: 0px; /* 往上提一点，制造右上角上标的视觉效果 */
}

.superAdmin-badge {
  color: var(--j-stamp);
}

.admin-badge {
  color: var(--j-pen);
}

.ban-badge {
  color: #fff;
  background-color: var(--j-stamp);
  border-color: var(--j-stamp);
}
/* 覆盖 el-card 默认的内边距，使其更紧凑 */
:deep(.user-card .el-card__body) {
  padding: 16px;
}

.user-card-content {
  display: flex;
  align-items: center;
  width: 100%;
}

.author-avatar {
  flex-shrink: 0;
  margin-right: 12px;
  background-color: #ece4d3; /* 没有头像时的文字头像底色 */
  color: var(--j-ink-soft);
}

.user-info-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden; /* 防止文字过长撑破布局 */
}

.user-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--j-ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  /* margin-bottom: 4px; 之前这里的 margin-bottom 移到了父容器 wrapper 上 */
}

.user-remark {
  font-size: 12px;
  color: var(--j-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-actions {
  margin-left: 8px;
  flex-shrink: 0;
}

/* 红色警告文字 */
.text-danger {
  color: var(--j-stamp);
}

.list-footer {
  margin-top: 20px;
  text-align: center;
  padding-bottom: 10px;
  /* 占满整行 */
  grid-column: 1 / -1;
}

.load-more-btn {
  width: 120px;
}

.end-text-v2 {
  font-family: var(--j-hand);
  font-size: 14px;
  color: var(--j-muted);
}
</style>