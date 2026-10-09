<template>
  <div class="personal-page j-desk">
  <div class="personal-center">
    <div class="header">
      <div class="left-header">
        <h2>个人中心</h2>
        <div class="city">📍 {{ userStore.userBean.loginaddress }}</div>
      </div>
      <el-button type="success" size="small" class="career-btn" @click="personalCareer">
        🙂 个人生涯
      </el-button>
    </div>

    <el-card shadow="never" class="card">
      <span class="j-tape j-tape--top"></span>
      <div class="avatar-section">
        <div class="avatar-container">
          <img :src="userStore.userBean.avatar" class="avatar"/>
          <el-button
              class="avatar-edit-button"
              size="small"
              title="编辑头像"
              @click="triggerFileSelect"
              :icon="Edit"
              circle
          />
        </div>
      </div>

      <el-form label-width="90px" autocomplete="off" class="info-form">
        <el-form-item label="用户名">
          <el-input v-model="userStore.userBean.name" name="nouser" />
        </el-form-item>

        <el-form-item label="账号">
          <el-input :value="userStore.userBean.code" disabled name="noaccount" />
        </el-form-item>

        <el-form-item label="手机号">
          <el-input v-model="userStore.userBean.phone" name="nophone" />
        </el-form-item>

        <el-form-item label="邮箱">
          <el-input
              v-model="userStore.userBean.email"
              name="noemail"
              :readonly="readonly.email"
              @focus="readonly.email = false"
          />
        </el-form-item>

        <el-form-item label="个性签名">
          <el-input type="textarea" v-model="userStore.userBean.remark" name="remark" />
        </el-form-item>

        <el-form-item label="新密码">
          <el-input
              v-model="userStore.userBean.newPassWord"
              name="newpass"
              show-password
              placeholder="不修改请留空"
              :readonly="readonly.password"
              @focus="readonly.password = false"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="submit">保存</el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <input
        ref="fileInputRef"
        type="file"
        accept="image/*"
        style="display: none"
        @change="handleFileChange"
    />
  </div>
  </div>
</template>


<script setup>
import {ref} from 'vue'
import {ElMessage} from 'element-plus'
import {useUserStore} from "@/stores/main/user.js";
import {Edit} from "@element-plus/icons-vue";
import {ele_confirm, sendAxiosRequest,encrypt} from "@/utils/common.js";
import {useRouter} from "vue-router";
import {pubOpenUser} from "@/utils/blogUtil.js";

const router = useRouter();

const userStore = useUserStore()
userStore.initFromLocal();



// 防止自动填充 readonly 开关
const readonly = ref({
  email: true,
  password: true
})

// 保存表单数据
const submit = () => {
  if (!userStore.userBean.code) {
    ElMessage.error("用户过期,请返回主页面重新登录!");
    return false;
  }
  if (!userStore.userBean.name || !userStore.userBean.name.trim()) {
    ElMessage.error("用户名不允许为空!");
    return false;
  }
  let tip = `是否保存当前信息!`;
  if (userStore.userBean.newPassWord) {
    tip = `当前信息包含修改密码,保存后需重新登录,是否确认!`;
  }
  ele_confirm(tip, () => {
    let userInfo = {};
    Object.keys(userStore.userBean).forEach(key => {
      userInfo[key.toUpperCase()] = userStore.userBean[key];
    });
    sendAxiosRequest("/pub-api/login/changeUserInfo", {userInfo});
    ElMessage.success('保存成功');
  })
}

// 重置密码字段
const resetForm = () => {
  form.value.password = ''
  readonly.value.email = true
  readonly.value.password = true
}

// 上传逻辑
const customUploadRequest = async (options) => {
  const {file} = options;
  const formData = new FormData();
  formData.append('file', file);
  formData.append('spliceUrl', "userAvatar");
  try {
    const res = await sendAxiosRequest('/pub-api/upload/uploadFile', formData);
    // 删除旧头像
    if (userStore.userBean.avatar) {
      sendAxiosRequest("/pub-api/login/deleteUserAvatarFile", {userCode: userStore.userBean.code});
    }
    userStore.userBean.avatar = res.result.fileViewUrl;
    submit();
  } catch (error) {
    ElMessage.error("上传失败: " + error);
  }
};

// 上传前校验
const beforeUploadCheck = (file) => {
  const maxSizeMB = 50;
  const isLtMaxSize = file.size / 1024 / 1024 < maxSizeMB;
  if (!isLtMaxSize) {
    ElMessage.error(`图片太大!`);
    return false;
  }

  const isImage = file.type.startsWith('image/');
  if (!isImage) {
    ElMessage.error('只允许上传图片!');
    return false;
  }
  return true;
}

// ========== 头像上传新逻辑 ==========
const fileInputRef = ref(null);

// 点击按钮触发文件选择
const triggerFileSelect = () => {
  fileInputRef.value?.click();
}

// 用户选择文件
const handleFileChange = async (e) => {
  const file = e.target.files[0];
  if (!file) return false;
  if (!beforeUploadCheck(file)) return false;
  await customUploadRequest({file});

  e.target.value = ''; // 清空选择，避免下次同图无效
};

const personalCareer = ()=>{
  if (!userStore.userBean.code) {
    ElMessage.error("用户过期,请返回主页面重新登录!");
    return false;
  }
  pubOpenUser(router,userStore.userBean.code);
}
</script>

<style scoped>
.personal-page {
  min-height: 100vh;
  padding: 40px 20px 64px;
  box-sizing: border-box;
}

.personal-center {
  max-width: 760px;
  margin: 0 auto;
  color: var(--j-ink);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 28px;
}

.left-header h2 {
  margin: 0;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 32px;
  color: var(--j-ink);
}

.city {
  margin-top: 6px;
  font-size: 14px;
  color: var(--j-muted);
}

/* 个人生涯：一张绿色便签 */
.career-btn {
  height: auto;
  padding: 8px 16px;
  border: none;
  border-radius: 2px;
  background-color: #d4ead9;
  color: var(--j-ink);
  font-family: var(--j-hand);
  font-size: 15px;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.1), 0 10px 16px -10px rgba(60, 50, 30, 0.45);
  transform: rotate(2deg);
}

.career-btn:hover,
.career-btn:focus {
  background-color: #c4e2cc;
  color: var(--j-ink);
  transform: rotate(0deg);
}

/* 信息页：一张纸 */
.card {
  position: relative;
  overflow: visible;
  border: 1px solid var(--j-rule);
  border-radius: 2px;
  background: var(--j-paper);
  box-shadow: var(--j-shadow) !important;
}

.card :deep(.el-card__body) {
  padding: 40px 48px 28px;
}

.avatar-section {
  position: relative;
  display: flex;
  justify-content: center;
  margin-bottom: 34px;
}

/* 头像：贴上去的照片 */
.avatar-container {
  position: relative;
  width: 116px;
  height: 116px;
  padding: 6px 6px 18px;
  background: #fff;
  box-shadow: 0 1px 3px rgba(60, 50, 30, 0.25);
  transform: rotate(-2deg);
}

.avatar {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  background: #ece4d3;
}

.avatar-edit-button {
  position: absolute;
  bottom: -10px;
  right: -14px;
  z-index: 10;
  display: flex;
  justify-content: center;
  align-items: center;
  width: 30px;
  height: 30px;
  border: 1px solid var(--j-rule-strong);
  background-color: var(--j-note);
  box-shadow: 0 1px 3px rgba(60, 50, 30, 0.2);
}

.info-form {
  margin-top: 10px;
}

.el-form-item {
  margin-bottom: 20px;
}

.info-form :deep(.el-form-item__label) {
  font-family: var(--j-hand);
  font-size: 16px;
  color: var(--j-ink-soft);
}

/* 输入框：一条下划线 */
.info-form :deep(.el-input__wrapper) {
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  border-bottom: 1px solid var(--j-rule-strong);
  padding-left: 2px;
}

.info-form :deep(.el-input__wrapper.is-focus) {
  border-bottom-color: var(--j-ink);
}

.info-form :deep(.el-input.is-disabled .el-input__wrapper) {
  background: transparent;
  box-shadow: none;
  border-bottom-style: dashed;
}

.info-form :deep(.el-textarea__inner) {
  border-radius: 0;
  box-shadow: none;
  border: 1px solid var(--j-rule);
  background: #fffefb;
}

.el-input {
  width: 100%;
}

@media (max-width: 640px) {
  .personal-page {
    padding: 24px 12px 48px;
  }

  .card :deep(.el-card__body) {
    padding: 32px 18px 20px;
  }

  .info-form :deep(.el-form-item) {
    display: block;
  }

  .info-form :deep(.el-form-item__label) {
    justify-content: flex-start;
  }

  .info-form :deep(.el-form-item__content) {
    margin-left: 0 !important;
  }
}
</style>
