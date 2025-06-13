<template>
  <div class="personal-center">
    <div class="header">
      <h2>个人中心</h2>
      <div class="city">📍 {{ userStore.userBean.loginaddress }}</div>
    </div>
    <el-card shadow="always" class="card">
      <div class="avatar-section">
        <div class="avatar-container">
          <img :src="userStore.userBean.avatar" class="avatar"/>
          <el-button
              class="avatar-edit-button"
              size="small"
              @click="triggerFileSelect"
              :icon="Edit"
              circle
          />
        </div>
      </div>
      <el-form label-width="90px" autocomplete="off" class="info-form">
        <el-form-item label="用户名">
          <el-input
              v-model="userStore.userBean.name"
              name="nouser"
              autocomplete="off"
          />
        </el-form-item>
        <el-form-item label="账号">
          <el-input
              :value="userStore.userBean.code"
              disabled
              name="noaccount"
              autocomplete="off"
          />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input
              v-model="userStore.userBean.phone"
              name="nophone"
              autocomplete="off"
          />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input
              v-model="userStore.userBean.email"
              name="noemail"
              autocomplete="off"
              :readonly="readonly.email"
              @focus="readonly.email = false"
          />
        </el-form-item>

        <el-form-item label="新密码">
          <el-input
              v-model="userStore.userBean.newPassWord"
              name="newpass"
              autocomplete="new-password"
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

    <!-- 隐藏的文件上传 input -->
    <input
        ref="fileInputRef"
        type="file"
        accept="image/*"
        style="display: none"
        @change="handleFileChange"
    />
  </div>
</template>

<script setup>
import {ref} from 'vue'
import {storeToRefs} from 'pinia'
import {ElMessage} from 'element-plus'
import {useUserStore} from "@/stores/main/user.js";
import {Edit} from "@element-plus/icons-vue";
import {ele_confirm, sendAxiosRequest} from "@/utils/common.js";

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
</script>

<style scoped>
.personal-center {
  max-width: 800px;
  margin: 40px auto;
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h2 {
  font-size: 24px;
  color: #333;
}

.city {
  font-size: 14px;
  color: #999;
}

.card {
  padding: 30px;
  border-radius: 16px;
}

.avatar-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.avatar-container {
  position: relative;
  width: 100px;
  height: 100px;
}

.avatar {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid #ccc;
}

.avatar-edit-button {
  position: absolute;
  bottom: 0;
  right: 0;
  background: #fff;
  border: 1px solid #ccc;
}

.stats {
  display: flex;
  gap: 40px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 22px;
  font-weight: bold;
  color: #409EFF;
}

.stat-label {
  font-size: 14px;
  color: #666;
}

.info-form {
  margin-top: 20px;
}
</style>
