<template>
  <div class="sso-page">
    <div class="sso-container" v-if="userStore.userBean.code === adminUserCode">
      <el-card class="sync-card">
        <template #header>
          <div class="card-header">检查一致性文件</div>
        </template>
        <div class="button-row">
          <el-button type="primary" @click="checkEditorFiles">检查编辑器文件</el-button>
          <el-button type="warning" @click="checkResourceFiles">检查资源文件</el-button>
          <el-button type="success" @click="checkAvatarFiles">检查用户头像文件</el-button>
        </div>
      </el-card>
    </div>
    <div class="no-access" v-else>
      <el-empty description="禁止!只允许超级管理员访问"/>
    </div>
  </div>
</template>

<script setup>
import {ElMessage} from 'element-plus'
import {useUserStore} from "@/stores/main/user.js";
import {adminUserCode} from "@/config/vue-config.js";
import {sendAxiosRequest} from "@/utils/common.js";

const userStore = useUserStore();
userStore.initFromLocal();

const checkEditorFiles = async () => {
  ElMessage.success('正在验证编辑器文件...请等待');
  let result = await sendAxiosRequest("/blog-api/file/consistencyFileCheck", {type: "editorImage"});
  if (result && !result.isError) {
    ElMessage.success(`核对成功:删除文件总数:${result.result.length},文件名:${result.result}`);
  }else{
    ElMessage.error(result.errMsg);
  }
};

const checkResourceFiles = async () => {
  ElMessage.success('正在验证资源文件...');
  let result = await sendAxiosRequest("/blog-api/file/consistencyFileCheck", {type: "resourcesFile"});
  if (result && !result.isError) {
    ElMessage.success(`核对成功:删除文件总数:${result.result.length},文件名:${result.result}`);
  }else{
    ElMessage.error(result.errMsg);
  }
};

const checkAvatarFiles = async () => {
  ElMessage.success('正在验证用户头像文件...');
  let result = await sendAxiosRequest("/blog-api/file/consistencyFileCheck", {type: "userAvatar"});
  if (result && !result.isError) {
    ElMessage.success(`核对成功:删除文件总数:${result.result.length},文件名:${result.result}`);
  }else{
    ElMessage.error(result.errMsg);
  }
};
</script>

<style scoped>
.sso-page {
  min-height: 100vh;
  background-color: #f5f7fa;
  padding: 40px 20px;
  box-sizing: border-box;
}

.sso-container {
  display: flex;
  justify-content: center;
}

.sync-card {
  width: 80%;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  border-radius: 8px;
  background-color: #ffffff;
}

.card-header {
  font-weight: 600;
  font-size: 18px;
  padding: 4px 0;
  color: #333;
}

.button-row {
  display: flex;
  justify-content: flex-start;
  gap: 16px;
  flex-wrap: wrap;
  padding-top: 8px;
}

.no-access {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
}
</style>
