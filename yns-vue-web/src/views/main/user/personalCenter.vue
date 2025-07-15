<template>
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

    <el-card shadow="always" class="card">
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
</template>


<script setup>
import {ref} from 'vue'
import {ElMessage} from 'element-plus'
import {useUserStore} from "@/stores/main/user.js";
import {Edit} from "@element-plus/icons-vue";
import {ele_confirm, sendAxiosRequest,encrypt} from "@/utils/common.js";
import {useRouter} from "vue-router";

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
  router.push({name:"personInfomation",query:{c: encrypt(userStore.userBean.code)}});
}
</script>

<style scoped>
.personal-center {
  max-width: 800px;
  margin: 40px auto;
  padding: 20px;
  background: #f9f9f9;
  border-radius: 16px;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.05);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.left-header h2 {
  font-size: 26px;
  color: #333;
  margin: 0;
}

.city {
  font-size: 14px;
  color: #888;
  margin-top: 5px;
}

.career-btn {
  background-color: #67C23A;
  color: white;
  font-weight: bold;
  border-radius: 20px;
}

.card {
  padding: 30px;
  border-radius: 20px;
  background-color: #fff;
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.05);
}

.avatar-section {
  display: flex;
  justify-content: center;
  margin-bottom: 30px;
  position: relative; /* 用于按钮定位 */
}

.avatar-container {
  position: relative;
  width: 120px;
  height: 120px;
  border-radius: 50%;
  overflow: visible; /* 改这里，不再裁剪按钮 */
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.avatar {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
  transition: transform 0.3s;
}

.avatar:hover {
  transform: scale(1.05);
}

.avatar-edit-button {
  position: absolute;
  bottom: -10px;  /* 放在容器外部，避免被圆形裁切 */
  right: -10px;
  z-index: 10;
  background-color: white;
  border: 1px solid #dcdfe6;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
  width: 28px;
  height: 28px;
  display: flex;
  justify-content: center;
  align-items: center;
}


.info-form {
  margin-top: 10px;
}

.el-form-item {
  margin-bottom: 22px;
}

.el-input {
  width: 100%;
}

</style>
