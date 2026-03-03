<template>
  <!-- 公告横幅 -->
  <Announcement v-for="al in topAlert" :key="al.GUID" :TEXT="al.TEXT" :URL="al.URL" :URLNAME="al.URLNAME"/>
  <div class="friend-link-container">
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">✨ 发现宝藏站点</h2>
        <span class="page-subtitle">与优秀的人同行，让互联网更开放</span>
      </div>
      <el-button type="primary" round size="large" class="apply-btn" @click="pushFriendLink">
        <el-icon class="el-icon--left">
          <Plus/>
        </el-icon>
        申请友链
      </el-button>
    </div>

    <el-row :gutter="24" class="link-grid">
      <el-col
          :xs="24" :sm="12" :md="8" :lg="6"
          v-for="(item, index) in friendLinks"
          :key="index"
      >
        <a
            :href="item.LINK"
            target="_blank"
            class="modern-card fade-in-up"
            :style="{ animationDelay: `${index * 0.1}s` }"
        >
          <div class="card-avatar-box">
            <el-avatar :src="item.AVATAR" :size="56" class="site-avatar">
              {{ item.NAME?.charAt(0) }}
            </el-avatar>
          </div>
          <div class="card-content">
            <h3 class="site-name" :title="item.NAME">{{ item.NAME }}</h3>
            <p class="site-desc" :title="item.REMARK">{{ item.REMARK }}</p>
          </div>

          <div class="card-action-menu" @click.prevent.stop v-if="getCurrentUserAdminObject().isAdmin || item.USERCODE===userStore.userBean.code">
            <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, item)">
              <span class="action-btn">
                <el-icon><MoreFilled/></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="edit">
                    <el-icon>
                      <Edit/>
                    </el-icon>
                    编辑
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" class="danger-item">
                    <el-icon>
                      <Delete/>
                    </el-icon>
                    删除
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </a>
      </el-col>
    </el-row>

    <el-dialog
        v-model="dialogVisible"
        :title="isEditMode ? '编辑友链' : '申请加入友链'"
        width="480px"
        :close-on-click-modal="false"
        destroy-on-close
        class="custom-dialog"
    >
      <div class="dialog-tip" v-if="!isEditMode">
        欢迎互换友链！请确保您的站点能够正常访问，且包含本站链接。
      </div>
      <el-form :model="applyForm" label-width="80px" class="apply-form">
        <el-form-item label="网站名称" required>
          <el-input v-model="applyForm.NAME" placeholder="请输入您的网站名称" clearable/>
        </el-form-item>
        <el-form-item label="网站地址" required>
          <el-input v-model="applyForm.LINK" placeholder="例如：https://www.example.com" clearable/>
        </el-form-item>
        <el-form-item label="网站头像">
          <el-input v-model="applyForm.AVATAR" placeholder="图片 URL，建议正方形比例" clearable/>
        </el-form-item>
        <el-form-item label="一句话描述" required>
          <el-input
              v-model="applyForm.REMARK"
              type="textarea"
              :rows="3"
              placeholder="简短地介绍一下您的站点吧（50字以内）"
              maxlength="50"
              show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false" round>取消</el-button>
          <el-button type="primary" @click="submitApply" round>
            {{ isEditMode ? '保存修改' : '提交申请' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import {ref, reactive, onMounted} from "vue";
import {useUserStore} from "@/stores/main/user.js";
import {ElMessage, ElMessageBox} from "element-plus";
import {Plus, MoreFilled, Edit, Delete} from '@element-plus/icons-vue';
import {ele_confirm, getCurrentUserAdminObject, getGuid, sendAxiosRequest} from "@/utils/common.js";
import Announcement from "@/components/detail/Announcement.vue";
import {getAnnouncementByRouterName} from "@/utils/blogUtil.js";

const userStore = useUserStore();

// --- 状态控制 ---
const dialogVisible = ref(false);
const isEditMode = ref(false); // 区分是新增申请还是编辑

// 友链数据
const friendLinks = ref([]);

// --- 表单数据 ---
const applyForm = reactive({
  NAME: "",
  LINK: "",
  AVATAR: "",
  REMARK: ""
});

// --- 方法 ---
const pushFriendLink = () => {
  if (!userStore?.userBean?.code) {
    ElMessage.error('请先登录后再尝试发布友链吧!');
    return false;
  }
  isEditMode.value = false;
  Object.assign(applyForm, {NAME: "", LINK: "", AVATAR: "", REMARK: ""}); // 清空表单
  dialogVisible.value = true;
};

// 处理卡片下拉菜单点击
const handleCommand = (command, item) => {
  if (command === 'edit') {
    isEditMode.value = true;
    // 将当前项数据回显到表单（深拷贝以免直接修改列表数据）
    Object.assign(applyForm, JSON.parse(JSON.stringify(item)));
    dialogVisible.value = true;
  } else if (command === 'delete') {
    ele_confirm(`是否确认删除该友链?`, () => {
      sendAxiosRequest("/blog-api/friendLink/deleteFriendLink", {friendLinkId: item.GUID});
      //删除前台显示数据
      const index = friendLinks.value.findIndex(link => link.GUID === item.GUID);
      if (index !== -1) {
        friendLinks.value.splice(index, 1);
      }
      ElMessage.success("删除成功!");
    })
  }
};

const submitApply = () => {
  if (!userStore?.userBean?.code) {
    ElMessage.error('请先登录后再尝试发布友链吧!');
    return false;
  }
  if (!applyForm.NAME || !applyForm.LINK || !applyForm.REMARK) {
    ElMessage.warning('请将必填项填写完整');
    return false;
  }
  //编辑
  if (isEditMode.value) {
    sendAxiosRequest("/blog-api/friendLink/updateFriendLink", {friendLink: applyForm});
    //修改前台展示数据
    const target = friendLinks.value.find(oneLink => oneLink.GUID === applyForm.GUID);
    if (target) {
      // 将 applyForm 的所有属性同步到 target 对象中
      Object.assign(target, applyForm);
      ElMessage.success("修改成功！");
    }
    //新增
  } else {
    //如果不是管理员,则控制一个账户只能公开3个友链
    if (!getCurrentUserAdminObject().isAdmin) {
      let currentUserLinks = friendLinks.value.filter(oneLink => oneLink.USERCODE === userStore.userBean.code);
      if (currentUserLinks.length >= 3) {
        ElMessage.warning('为保证友链质量,只允许发布三个友链');
        return false;
      }
    }
    let addData = {...applyForm};
    addData.GUID = getGuid();
    addData.USERCODE = userStore.userBean.code;
    addData.USERNAME = userStore.userBean.name;
    //调用后台插入数据库
    sendAxiosRequest("/blog-api/friendLink/addFriendLink", {friendLink: addData})
    friendLinks.value.push(addData);
    ElMessage.success("提交成功！");
  }

  dialogVisible.value = false;
};

const getFriendLinks = async () => {
  const result = await sendAxiosRequest("/blog-api/friendLink/getFriendLinks");
  friendLinks.value = result.result;
}
onMounted(() => {
  getFriendLinks();
});
//公告横幅内容
const topAlert = ref([]);
const setTopAlert = async () => {
  topAlert.value = await getAnnouncementByRouterName("FriendLink");
}
setTopAlert();
</script>

<style scoped>
.friend-link-container {
  padding: 30px 20px;
  max-width: 1200px;
  margin: 0 auto;
}

/* --- 头部区域 --- */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 40px;
  padding-bottom: 15px;
  border-bottom: 1px solid #ebeef5;
}

.header-left .page-title {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 700;
  color: #303133;
}

.header-left .page-subtitle {
  font-size: 14px;
  color: #909399;
}

.apply-btn {
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
  transition: transform 0.2s;
}

.apply-btn:hover {
  transform: translateY(-2px);
}

/* --- 卡片网格区域 --- */
.link-grid {
  margin-bottom: 30px;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.fade-in-up {
  opacity: 0;
  animation: fadeInUp 0.6s cubic-bezier(0.25, 0.8, 0.25, 1) forwards;
}

.modern-card {
  display: flex;
  align-items: center;
  padding: 20px;
  margin-bottom: 24px;
  background: #ffffff;
  border-radius: 16px;
  border: 1px solid #ebeef5;
  text-decoration: none;
  transition: box-shadow 0.3s cubic-bezier(0.25, 0.8, 0.25, 1), border-color 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
  height: 110px;
  box-sizing: border-box;
  position: relative;
  top: 0;
}

.modern-card:hover {
  top: -6px;
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.08);
  border-color: #c0c4cc;
}

.card-avatar-box {
  flex-shrink: 0;
  margin-right: 16px;
  border-radius: 50%;
  border: 2px solid #f5f7fa;
  padding: 2px;
}

.site-avatar {
  background-color: #ecf5ff;
  color: #409eff;
  font-weight: bold;
  font-size: 24px;
}

.card-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.site-name {
  margin: 0 0 6px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color 0.3s;
  padding-right: 24px; /* 防止标题过长与右上角菜单重叠 */
}

.modern-card:hover .site-name {
  color: #409eff;
}

.site-desc {
  margin: 0;
  font-size: 13px;
  color: #909399;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  padding-right: 10px;
}

/* --- 卡片右上角操作菜单 --- */
.card-action-menu {
  position: absolute;
  top: 12px;
  right: 12px;
  opacity: 0; /* 默认隐藏，悬浮显示。如果在移动端你需要它常显，把这行改为 opacity: 1; color: #c0c4cc; */
  transition: opacity 0.3s;
}

.modern-card:hover .card-action-menu {
  opacity: 1;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  color: #909399;
  font-size: 18px;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn:hover {
  background-color: #f4f4f5;
  color: #409eff;
}

/* 删除按钮标红 */
.danger-item {
  color: #f56c6c !important;
}

.danger-item:hover {
  background-color: #fef0f0 !important;
  color: #f56c6c !important;
}

/* --- 弹窗美化 --- */
.dialog-tip {
  background-color: #f0f9eb;
  color: #67c23a;
  padding: 10px 15px;
  border-radius: 6px;
  font-size: 13px;
  margin-bottom: 20px;
}
</style>