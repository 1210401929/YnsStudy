<template>
  <div class="sso-page j-desk">
    <div class="sso-container" v-if="getCurrentUserAdminObject().adminLevel==='superAdmin'">
      <el-card class="sync-card">
        <template #header>
          <div class="card-header">后台管理</div>
        </template>
        <!-- lazy：切到对应标签页时才加载数据 -->
        <el-tabs v-model="activeTab" class="admin-tabs" :tab-position="isWide ? 'left' : 'top'">
          <el-tab-pane label="数据概览" name="dashboard" lazy><AdminDashboard/></el-tab-pane>
          <el-tab-pane label="内容管理" name="content" lazy><AdminContent/></el-tab-pane>
          <el-tab-pane label="评论管理" name="comments" lazy><AdminComments/></el-tab-pane>
          <el-tab-pane label="评论限制" name="commentBans" lazy><AdminCommentBans/></el-tab-pane>
          <el-tab-pane label="用户管理" name="users" lazy><user-list ISOPERATIONUSER="true"></user-list></el-tab-pane>
          <el-tab-pane label="全站公告" name="announcements" lazy>
        <div class="card-header-row announcement-toolbar">
          <span class="tab-hint">公告会显示在对应栏目页的顶部</span>
          <el-button type="primary" size="small" @click="addAnnouncement">新增</el-button>
        </div>
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
          </el-tab-pane>
          <el-tab-pane label="违禁词" name="bannedWords" lazy><AdminBannedWords/></el-tab-pane>
          <el-tab-pane label="SEO 工具" name="seo" lazy><AdminSeoTools/></el-tab-pane>
          <el-tab-pane label="文件清理" name="files" lazy><AdminFiles/></el-tab-pane>
          <el-tab-pane label="操作日志" name="logs" lazy><AdminLogs/></el-tab-pane>
        </el-tabs>
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
import {ele_confirm, getCurrentUserAdminObject, getGuid, pubFormatDate, sendAxiosRequest} from "@/utils/common.js";
import {getMenuItems} from "@/utils/menu.js";
import {onBeforeUnmount, onMounted, ref} from "vue";
import UserList from "@/components/detail/UserList.vue";
import AdminDashboard from "@/components/admin/AdminDashboard.vue";
import AdminComments from "@/components/admin/AdminComments.vue";
import AdminFiles from "@/components/admin/AdminFiles.vue";
import AdminContent from "@/components/admin/AdminContent.vue";
import AdminBannedWords from "@/components/admin/AdminBannedWords.vue";
import AdminSeoTools from "@/components/admin/AdminSeoTools.vue";
import AdminLogs from "@/components/admin/AdminLogs.vue";
import AdminCommentBans from "@/components/admin/AdminCommentBans.vue";

const activeTab = ref('dashboard');
// 标签较多，宽屏放在左侧，窄屏放在顶部
const isWide = ref(window.innerWidth >= 1000);
const updateWidth = () => { isWide.value = window.innerWidth >= 1000; };
onMounted(() => window.addEventListener('resize', updateWidth));
onBeforeUnmount(() => window.removeEventListener('resize', updateWidth));
import { useHead } from '@vueuse/head'; // 1. 引入 useHead
// 2. 定义响应式的 SEO 数据源，给定默认值
const seoTitle = ref('后台管理 - ynsStudy');
const seoDescription = ref('后台管理 - ynsStudy');

useHead({
  title: seoTitle,
  meta: [
    {
      name: 'description',
      content: seoDescription
    }
  ]
});

const userStore = useUserStore();
userStore.initFromLocal();
// 文件清理已移到 AdminFiles 组件：先预览，确认后再删除
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
//--------------------------------用户权限设置-------------------------------------------------------

//--------------------------------公共-------------------------------------------------------

</script>

<style scoped>
.sso-page {
  min-height: 100vh;
  padding: 40px 20px;
  box-sizing: border-box;
}

.sso-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 28px;
}

.sync-card {
  width: min(1200px, 100%);
  border: 1px solid var(--j-rule);
  border-radius: 2px;
  background-color: var(--j-paper);
  box-shadow: var(--j-shadow) !important;
}

.sync-card :deep(.el-card__header) {
  border-bottom: 1px dashed var(--j-rule-strong);
}

.card-header {
  padding: 4px 0;
  font-family: var(--j-hand);
  font-size: 21px;
  color: var(--j-ink);
}

.button-row {
  display: flex;
  justify-content: flex-start;
  flex-wrap: wrap;
  gap: 16px;
  padding-top: 8px;
}

.announcement-toolbar {
  margin-bottom: 12px;
}

.tab-hint {
  color: var(--j-muted);
  font-size: 13px;
}

.admin-tabs :deep(.el-tabs__item) {
  font-size: 15px;
}

/* 左侧标签栏时，内容区不能被撑出卡片 */
.admin-tabs :deep(.el-tabs__content) {
  min-width: 0;
}

.admin-tabs.el-tabs--left :deep(.el-tabs__content) {
  padding-left: 20px;
}

.sync-card :deep(.admin-table) {
  --el-table-header-bg-color: var(--j-paper-warm);
  --el-table-row-hover-bg-color: rgba(250, 216, 96, 0.18);
  --el-table-tr-bg-color: #fffefb;
  --el-table-bg-color: #fffefb;
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

.sync-card :deep(.el-table) {
  --el-table-header-bg-color: var(--j-paper-warm);
  --el-table-row-hover-bg-color: rgba(250, 216, 96, 0.18);
  --el-table-tr-bg-color: #fffefb;
  --el-table-bg-color: #fffefb;
}

@media (max-width: 768px) {
  .sync-card {
    width: 100%;
  }
}
</style>
