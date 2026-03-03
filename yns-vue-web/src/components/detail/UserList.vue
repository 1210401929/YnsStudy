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
                <el-dropdown-item command="setAdmin">设为管理员</el-dropdown-item>
                <el-dropdown-item command="removeAdmin">取消管理员</el-dropdown-item>
                <el-dropdown-item divided command="ban" class="text-danger">封禁用户</el-dropdown-item>
                <el-dropdown-item command="removeBan">取消封禁</el-dropdown-item>
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
</template>

<script setup>
import {ArrowDown, MoreFilled} from "@element-plus/icons-vue";
import {onMounted, ref} from "vue";
import {ele_confirm, getCurrentUserAdminObject, sendAxiosRequest} from "@/utils/common.js";
import {ElMessage} from "element-plus";
import {pubOpenUser} from "@/utils/blogUtil.js";
import {useRouter} from "vue-router";
import debounce from "lodash/debounce.js";
import {adminUserCode} from "@/config/vue-config.js";

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
const handleUserAction = (command, userInfo) => {
  debugger;
  let type = command;
  switch (command) {
    case 'setAdmin':
      // 这里替换为你的实际 API 请求
      ele_confirm(`确定要将【${userInfo.NAME}】设为管理员吗？`, () => {
        userInfo.ROLE = "admin";
        ElMessage.success(`已将 ${userInfo.NAME} 设为管理员`);
      });
      break;
    case 'removeAdmin':
      ele_confirm(`确定要取消【${userInfo.NAME}】的管理员权限吗？`, () => {
        userInfo.ROLE = "1";
        ElMessage.success(`已取消 ${userInfo.NAME} 的管理员权限`);
      });
      break;
    case 'ban':
      ele_confirm(`确定要封禁【${userInfo.NAME}】吗？`, () => {
        userInfo.ISBAN = "1";
        ElMessage.success(`已封禁用户 ${userInfo.NAME}`);
      });
      break;
    case 'removeBan':
      ele_confirm(`确定要取消【${userInfo.NAME}】的封禁吗？`, () => {
        userInfo.ISBAN = "";
        ElMessage.success(`已取消用户 ${userInfo.NAME}封禁`);
      });
      break;
  }
  const result = sendAxiosRequest("/pub-api/login/operationUser", {userId: userInfo.GUID, type});
}

onMounted(() => {
  //默认加载用户
  getAllUserInfo();
})

</script>
<style scoped>
/* --- 用户网格布局 --- */
.user-grid {
  display: grid;
  /* 自动响应式：每个卡片最小280px，放不下自动换行 */
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.user-card {
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid #ebeef5;
}

.user-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* 新增：包裹名字和标签的容器 */
.name-badge-wrapper {
  display: flex;
  align-items: flex-start; /* 让内部元素顶部对齐 */
  margin-bottom: 4px;
  width: 100%;
}
.public-badge{

  color: #000;
  font-size: 9px;
  font-weight: bold;
  border-radius: 10px;
  padding: 2px 5px; /* 稍微增加了点左右内边距，更好看 */
  line-height: 1.2;

  /* --- 核心位置微调 --- */
  margin-left: 4px; /* 和名字拉开一点距离 */
  flex-shrink: 0; /* 关键：保证名字再长，标签都不会被压缩变形 */
  position: relative;
  top: 0px; /* 往上提一点，制造右上角上标的视觉效果 */
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
  background-color: #409eff; /* 给没有头像的文字头像加个默认底色 */
  color: #fff;
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
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  /* margin-bottom: 4px; 之前这里的 margin-bottom 移到了父容器 wrapper 上 */
}

.user-remark {
  font-size: 12px;
  color: #909399;
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
  color: #f56c6c;
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
  color: #c0c4cc;
  font-size: 13px;
}
</style>