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
      <el-card class="sync-card">
        <template #header>
          <div class="card-header-row">
            <div class="card-header">发送全网站通知公告</div>
            <el-button type="primary" size="small" @click="addAnnouncement">新增</el-button>
          </div>
        </template>
        <div>
          <el-table :data="tableData" height="250" stripe border style="width: 100%">
            <el-table-column prop="TEXT" label="内容" min-width="300">
              <template #default="{ row }">
                <el-input v-model="row.TEXT" size="small" placeholder="请输入内容"/>
              </template>
            </el-table-column>
            <el-table-column prop="URL" label="跳转路径" min-width="280">
              <template #default="{ row }">
                <el-input v-model="row.URL" size="small" placeholder="请输入内容"/>
              </template>
            </el-table-column>
            <el-table-column prop="URLNAME" label="跳转名称" min-width="100">
              <template #default="{ row }">
                <el-input v-model="row.URLNAME" size="small" placeholder="请输入内容"/>
              </template>
            </el-table-column>
            <el-table-column prop="TYPE" label="类型" min-width="100">
              <template #default="{ row }">
                <el-select v-model="row.TYPE" placeholder="请选择" size="small" style="width: 100%">
                  <el-option :label="menu.name" :value="menu.router" v-for="menu in getMenuItems()"></el-option>
                </el-select>
              </template>
            </el-table-column>
            <el-table-column prop="CREATE_TIME" label="时间" min-width="200"/>
            <el-table-column prop="ISENABLE" label="是否启用" min-width="70">
              <template #default="{ row }">
                <el-select v-model="row.ISENABLE" placeholder="请选择" size="small" style="width: 100%">
                  <el-option label="是" value="1"></el-option>
                  <el-option label="否" value="0"></el-option>
                </el-select>
              </template>
            </el-table-column>
            <el-table-column fixed="right" label="操作" min-width="100">
              <template #default="{row}">
                <el-button link type="primary" size="small" @click="saveAnnouncement(row)">保存</el-button>
                <el-button link type="primary" size="small" @click="deleteAnnouncement(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
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
import {ele_confirm, getGuid, pubFormatDate, sendAxiosRequest} from "@/utils/common.js";
import {getMenuItems} from "@/utils/menu.js";
import {ref} from "vue";

document.title = "后台管理 - ynsStudy";
const userStore = useUserStore();
userStore.initFromLocal();
//------------------------------检查文件一致性---------------------------------------------------------
const checkEditorFiles = async () => {
  ElMessage.success('正在验证编辑器文件...请等待');
  let result = await sendAxiosRequest("/blog-api/file/consistencyFileCheck", {type: "editorImage"});
  if (result && !result.isError) {
    ElMessage.success(`核对成功:删除文件总数:${result.result.length},文件名:${result.result}`);
  } else {
    ElMessage.error(result.errMsg);
  }
};

const checkResourceFiles = async () => {
  ElMessage.success('正在验证资源文件...');
  let result = await sendAxiosRequest("/blog-api/file/consistencyFileCheck", {type: "resourcesFile"});
  if (result && !result.isError) {
    ElMessage.success(`核对成功:删除文件总数:${result.result.length},文件名:${result.result}`);
  } else {
    ElMessage.error(result.errMsg);
  }
};

const checkAvatarFiles = async () => {
  ElMessage.success('正在验证用户头像文件...');
  let result = await sendAxiosRequest("/blog-api/file/consistencyFileCheck", {type: "userAvatar"});
  if (result && !result.isError) {
    ElMessage.success(`核对成功:删除文件总数:${result.result.length},文件名:${result.result}`);
  } else {
    ElMessage.error(result.errMsg);
  }
};
//--------------------------------发送全网站通知公告-------------------------------------------------------
const tableData = ref([]);
const setTableData = async () => {
  let result = await sendAxiosRequest("/blog-api/sso/getAllAnnouncement");
  if (result && !result.isError) {
    tableData.value = result.result.map(item => {
      item.CREATE_TIME = pubFormatDate(item.CREATE_TIME);
      return item;
    });
  }
}
setTableData();
const addAnnouncement = () => {
  let announcement = {
    GUID: getGuid(),
    TEXT: "新公告内容",
    URL: "",
    URLNAME: "",
    TYPE: "Home",
    ISENABLE: "0"
  }
  tableData.value.push(announcement);
  sendAxiosRequest("/blog-api/sso/addAnnouncement", {announcement});
};
//保存
const saveAnnouncement = async (row) => {
  debugger;
  let announcement = {...row};
  delete announcement.CREATE_TIME;
  let result = await sendAxiosRequest("/blog-api/sso/editAnnouncement", {announcement});
  if (result && !result.isError) {
    ElMessage.success("保存成功");
  } else {
    ElMessage.error(result.errMsg);
  }
}

//删除
const deleteAnnouncement = async (row) => {
  debugger;
  ele_confirm("是否确认删除?", async () => {
    let result = await sendAxiosRequest("/blog-api/sso/deleteAnnouncement", {guid: row.GUID});
    if (result && !result.isError) {
      tableData.value = tableData.value.filter(item => item.GUID !== row.GUID);
      ElMessage.success("删除成功");
    } else {
      ElMessage.error(result.errMsg);
    }
  })
}

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
  flex-direction: column; /* 纵向排列 */
  align-items: center; /* 横向居中 */
  gap: 20px; /* 两个卡片之间的间距 */
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

.card-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

</style>
