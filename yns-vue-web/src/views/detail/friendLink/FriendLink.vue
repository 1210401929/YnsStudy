<template>
  <Announcement
      v-if="!isEmbed"
      v-for="al in topAlert"
      :key="al.GUID"
      :TEXT="al.TEXT"
      :URL="al.URL"
      :URLNAME="al.URLNAME"
  />

  <div :class="['friend-link-container', { 'is-embed': isEmbed }]">
    <header v-if="!isEmbed" class="page-heading">
      <div>
        <h1>友链</h1>
        <p class="page-intro">互联网很大，很高兴在这里遇见你们。</p>
        <p class="page-stats">{{ normalFriendLinks.length }} 位朋友 · {{ recommendLinks.length }} 个推荐站点</p>
      </div>
      <el-button type="primary" plain @click="pushFriendLink">
        <el-icon class="el-icon--left"><Plus/></el-icon>加入友链
      </el-button>
    </header>

    <section class="link-section">
      <div class="section-heading">
        <div>
          <div class="title-line">
            <h2>友情链接</h2>
            <span class="section-count">{{ normalFriendLinks.length }} 位朋友</span>
          </div>
          <p>一些经常访问，也值得认识的朋友。</p>
        </div>
      </div>
      <div class="friend-grid">
        <article v-for="item in normalFriendLinks" :key="item.GUID"
            class="link-card" :class="{ 'has-actions': isAdmin || item.USERCODE === userStore.userBean.code }">
          <a :href="item.LINK" target="_blank" rel="noopener noreferrer" class="card-link">
            <el-avatar :src="item.AVATAR" :size="48" shape="square" class="site-avatar">
              {{ item.NAME?.trim().charAt(0) || '站' }}
            </el-avatar>
            <div class="card-content">
              <h3 class="site-name" :title="item.NAME">{{ item.NAME }}</h3>
              <p class="site-desc" :title="item.REMARK">{{ item.REMARK || '这个朋友很低调，还没有留下介绍' }}</p>
            </div>
          </a>
            <div v-if="isAdmin || item.USERCODE === userStore.userBean.code" class="card-action-menu">
              <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, item)">
                <button type="button" class="action-btn" :aria-label="'管理' + item.NAME">
                  <el-icon><MoreFilled/></el-icon>
                </button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="edit"><el-icon><Edit/></el-icon>编辑</el-dropdown-item>
                    <el-dropdown-item command="delete" class="danger-item"><el-icon><Delete/></el-icon>删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
        </article>
      </div>
    </section>

    <section v-if="!isEmbed && (recommendLinks.length > 0 || isAdmin)" class="link-section">
      <div class="section-heading">
        <div>
          <div class="title-line">
            <h2>推荐好站</h2>
            <span class="section-count">{{ recommendLinks.length }} 个站点</span>
          </div>
          <p>一些我自己用过，或者觉得不错的网站。</p>
        </div>
        <el-button v-if="isAdmin" @click="pushRecommendLink">
          <el-icon class="el-icon--left"><Plus/></el-icon>添加推荐
        </el-button>
      </div>
      <div class="recommend-grid">
        <article v-for="item in recommendLinks" :key="item.GUID"
            class="link-card" :class="{ 'has-actions': isAdmin }">
          <a :href="item.LINK" target="_blank" rel="noopener noreferrer" class="card-link">
            <el-avatar :src="item.AVATAR" :size="48" shape="square" class="site-avatar">
              {{ item.NAME?.trim().charAt(0) || '站' }}
            </el-avatar>
            <div class="card-content">
              <h3 class="site-name" :title="item.NAME">{{ item.NAME }}</h3>
              <p class="site-desc" :title="item.REMARK">{{ item.REMARK || '一个值得花时间探索的好站点' }}</p>
            </div>
          </a>
            <div v-if="isAdmin" class="card-action-menu">
              <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, item)">
                <button type="button" class="action-btn" :aria-label="'管理' + item.NAME">
                  <el-icon><MoreFilled/></el-icon>
                </button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="edit"><el-icon><Edit/></el-icon>编辑</el-dropdown-item>
                    <el-dropdown-item command="delete" class="danger-item"><el-icon><Delete/></el-icon>删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
        </article>
      </div>
    </section>

    <!-- ================= 弹窗表单 ================= -->
    <el-dialog
        v-model="dialogVisible"
        :title="isEditMode ? (isRecommendMode ? '编辑推荐' : '编辑友链') : (isRecommendMode ? '添加推荐' : '申请加入友链')"
        width="min(480px, calc(100vw - 32px))"
        :close-on-click-modal="false"
        destroy-on-close
    >
      <div class="dialog-tip" v-if="!isEditMode && !isRecommendMode">
        欢迎互换友链！请确保您的站点能够正常访问，且包含本站链接。
      </div>
      <div class="dialog-tip" style="color: #e6a23c; background-color: #fdf6ec;" v-if="isRecommendMode && !isEditMode">
        添加的内容将展示在“推荐好站”专区。
      </div>
      <el-form :model="applyForm" label-width="80px">
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
        <span>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitApply">
            {{ isEditMode ? '保存修改' : '提交' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import {ref, reactive, onMounted, defineProps, computed} from "vue";
import {useUserStore} from "@/stores/main/user.js";
import {ElMessage} from "element-plus";
import {Plus, MoreFilled, Edit, Delete} from '@element-plus/icons-vue';
import {ele_confirm, getCurrentUserAdminObject, getGuid, sendAxiosRequest} from "@/utils/common.js";
import Announcement from "@/components/detail/Announcement.vue";
import {getAnnouncementByRouterName} from "@/utils/blogUtil.js";

const props = defineProps({
  isEmbed: {
    type: Boolean,
    default: false
  }
});

const userStore = useUserStore();
const dialogVisible = ref(false);
const isEditMode = ref(false);
const isRecommendMode = ref(false); // 标识当前是否在操作"推荐好站"

const friendLinks = ref([]);
const topAlert = ref([]);

const isAdmin = computed(() => getCurrentUserAdminObject().isAdmin);

// 过滤出推荐好站
const recommendLinks = computed(() => {
  return friendLinks.value.filter(item => item.LINK_TYPE == 2);
});

// 过滤出普通的友链
const normalFriendLinks = computed(() => {
  return friendLinks.value.filter(item => item.LINK_TYPE == 1);
});

const applyForm = reactive({
  NAME: "",
  LINK: "",
  AVATAR: "",
  REMARK: "",
  LINK_TYPE: 0 // 新增字段，1代表友链，2代表推荐好站
});

// 点击添加推荐好站 (管理员)
const pushRecommendLink = () => {
  isEditMode.value = false;
  isRecommendMode.value = true;
  Object.assign(applyForm, {NAME: "", LINK: "", AVATAR: "", REMARK: "", LINK_TYPE: 2});
  dialogVisible.value = true;
};

// 点击申请友链
const pushFriendLink = () => {
  if (!userStore?.userBean?.code) {
    ElMessage.error('请先登录后再尝试发布友链吧!');
    return false;
  }
  isEditMode.value = false;
  isRecommendMode.value = false;
  Object.assign(applyForm, {NAME: "", LINK: "", AVATAR: "", REMARK: "", LINK_TYPE: 1});
  dialogVisible.value = true;
};

const handleCommand = (command, item) => {
  if (command === 'edit') {
    isEditMode.value = true;
    isRecommendMode.value = item.LINK_TYPE === 2;
    Object.assign(applyForm, JSON.parse(JSON.stringify(item)));
    dialogVisible.value = true;
  } else if (command === 'delete') {
    ele_confirm(`是否确认删除该记录?`, async () => {
      await sendAxiosRequest("/blog-api/friendLink/deleteFriendLink", {friendLinkId: item.GUID});
      const index = friendLinks.value.findIndex(link => link.GUID === item.GUID);
      if (index !== -1) friendLinks.value.splice(index, 1);
      ElMessage.success("删除成功!");
    })
  }
};

const submitApply = async () => {
  // 普通用户申请友链需要拦截，管理员添加推荐不拦截
  if (!userStore?.userBean?.code && !isAdmin.value) {
    ElMessage.error('请先登录后再尝试发布吧!');
    return false;
  }
  if (!applyForm.NAME || !applyForm.LINK || !applyForm.REMARK) {
    ElMessage.warning('请将必填项填写完整');
    return false;
  }

  if (isEditMode.value) {
    await sendAxiosRequest("/blog-api/friendLink/updateFriendLink", {friendLink: applyForm});
    const target = friendLinks.value.find(oneLink => oneLink.GUID === applyForm.GUID);
    if (target) {
      Object.assign(target, applyForm);
      ElMessage.success("修改成功！");
    }
  } else {
    // 限制普通用户最多发布3条友链
    if (!isAdmin.value) {
      let currentUserLinks = friendLinks.value.filter(oneLink => oneLink.USERCODE === userStore.userBean.code && oneLink.LINK_TYPE === 2);
      if (currentUserLinks.length >= 3) {
        ElMessage.warning('为保证友链质量,只允许发布三个友链');
        return false;
      }
    }
    let addData = {...applyForm};
    addData.GUID = getGuid();
    addData.USERCODE = userStore.userBean.code || 'admin';
    addData.USERNAME = userStore.userBean.name || '管理员';

    await sendAxiosRequest("/blog-api/friendLink/addFriendLink", {friendLink: addData})
    friendLinks.value.push(addData);
    ElMessage.success("提交成功！");
  }
  dialogVisible.value = false;
};

const getFriendLinks = async () => {
  const result = await sendAxiosRequest("/blog-api/friendLink/getFriendLinks");
  friendLinks.value = result.result || [];
}

const setTopAlert = async () => {
  // 嵌入模式下不重复显示全局公告
  if (props.isEmbed) return;
  debugger;
  topAlert.value = await getAnnouncementByRouterName("FriendLink");
}

onMounted(() => {
  getFriendLinks();
  setTopAlert();
});
</script>

<style scoped>
.friend-link-container {
  --link-ink: var(--el-text-color-primary, #303133);
  --link-muted: var(--el-text-color-secondary, #909399);
  --link-border: var(--el-border-color-light, #e4e7ed);
  --link-blue: var(--el-color-primary, #409eff);
  max-width: 1200px;
  margin: 0 auto;
  padding: 28px 20px 48px;
  color: var(--link-ink);
}

.page-heading,
.section-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
}

.page-heading {
  padding-bottom: 22px;
  border-bottom: 1px solid var(--link-border);
}

.page-heading h1 {
  margin: 0;
  font-size: 26px;
  font-weight: 600;
  line-height: 1.4;
}

.page-intro {
  margin: 8px 0 0;
  color: var(--el-text-color-regular, #606266);
  font-size: 15px;
  line-height: 1.6;
}

.page-stats {
  margin: 8px 0 0;
  color: var(--link-muted);
  font-size: 13px;
  line-height: 1.5;
}

.link-section {
  margin-top: 28px;
}

.section-heading {
  margin-bottom: 14px;
}

.title-line {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 12px;
}

.title-line h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  line-height: 1.5;
}

.section-count {
  color: var(--link-muted);
  font-size: 13px;
}

.section-heading p {
  margin: 4px 0 0;
  color: var(--el-text-color-regular, #606266);
  font-size: 14px;
  line-height: 1.6;
}

.friend-grid,
.recommend-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.link-card {
  position: relative;
  min-width: 0;
  border: 1px solid var(--link-border);
  border-radius: 10px;
  background: var(--el-bg-color, #fff);
  transition: border-color 200ms ease, background-color 200ms ease, transform 200ms ease, box-shadow 200ms ease;
}

.recommend-grid .link-card::after {
  content: '';
  position: absolute;
  right: 16px;
  bottom: 0;
  left: 16px;
  height: 2px;
  border-radius: 2px;
  background: var(--el-color-primary-light-5, #a0cfff);
  pointer-events: none;
  transform: scaleX(0);
  transform-origin: left;
  transition: transform 220ms ease;
}

.card-link {
  display: flex;
  align-items: center;
  gap: 12px;
  box-sizing: border-box;
  height: 100%;
  min-height: 104px;
  padding: 16px;
  border-radius: inherit;
  color: inherit;
  text-decoration: none;
}

.site-avatar {
  flex-shrink: 0;
  border-radius: 10px;
  background: var(--el-color-primary-light-9, #ecf5ff);
  color: var(--link-blue);
  font-size: 20px;
  font-weight: 500;
}

.card-content {
  flex: 1;
  min-width: 0;
}

.site-name {
  overflow: hidden;
  margin: 0 0 5px;
  color: var(--link-ink);
  font-size: 16px;
  font-weight: 600;
  line-height: 1.5;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: color 200ms ease;
}

.site-desc {
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  color: var(--el-text-color-regular, #606266);
  font-size: 14px;
  line-height: 1.6;
  overflow-wrap: anywhere;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.has-actions .site-name {
  padding-right: 16px;
}

.card-action-menu {
  position: absolute;
  top: 8px;
  right: 6px;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  padding: 0;
  border: 0;
  border-radius: 6px;
  color: var(--link-muted);
  background: transparent;
  cursor: pointer;
}

.action-btn:hover {
  color: var(--link-blue);
  background: var(--el-color-primary-light-9, #ecf5ff);
}

.card-link:focus-visible,
.action-btn:focus-visible {
  outline: 2px solid var(--link-blue);
  outline-offset: 2px;
}

.danger-item {
  color: var(--el-color-danger, #f56c6c);
}

.friend-link-container.is-embed {
  max-width: 100%;
  padding: 0;
  margin: 0;
}

.is-embed .link-section {
  margin-top: 0;
}

.is-embed .friend-grid {
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 240px), 1fr));
}

.dialog-tip {
  margin-bottom: 20px;
  padding: 10px 12px;
  border-radius: 8px;
  color: var(--el-text-color-regular, #606266);
  background: var(--el-fill-color-light, #f5f7fa);
  font-size: 14px;
  line-height: 1.6;
}

@media (hover: hover) {
  .friend-grid .link-card:hover {
    transform: translateY(-2px);
    border-color: var(--el-color-primary-light-5, #a0cfff);
    box-shadow: 0 4px 12px rgba(35, 55, 80, 0.06);
  }

  .recommend-grid .link-card:hover {
    border-color: var(--el-color-primary-light-7, #c6e2ff);
    background: var(--el-color-primary-light-9, #ecf5ff);
  }

  .recommend-grid .link-card:hover::after {
    transform: scaleX(1);
  }

  .link-card:hover .site-name {
    color: var(--link-blue);
  }
}

@media (max-width: 1100px) {
  .friend-grid,
  .recommend-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 900px) {
  .friend-grid,
  .recommend-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .friend-link-container {
    padding: 20px 14px 32px;
  }

  .page-heading {
    align-items: flex-start;
  }

  .friend-grid,
  .recommend-grid {
    grid-template-columns: 1fr;
  }

  .link-section {
    margin-top: 24px;
  }

  .card-link {
    min-height: 96px;
    padding: 14px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .link-card,
  .site-name,
  .recommend-grid .link-card::after {
    transition: none;
  }

  .friend-grid .link-card:hover {
    transform: none;
  }
}
</style>
