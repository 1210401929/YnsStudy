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
      <div class="heading-label">
        <span class="j-tape heading-tape"></span>
        <h1>友链</h1>
        <p class="page-intro">互联网很大，很高兴在这里遇见你们。</p>
        <p class="page-stats">{{ normalFriendLinks.length }} 位朋友 · {{ recommendLinks.length }} 个推荐站点</p>
      </div>
      <el-button type="primary" plain @click="pushFriendLink">
        <el-icon class="el-icon--left"><Plus/></el-icon>加入友链
      </el-button>
    </header>

    <!-- 本站信息：方便别人互加友链，每项都能一键复制 -->
    <section v-if="!isEmbed" class="site-info" aria-label="本站信息">
      <div class="site-info-head">
        <h2>本站信息</h2>
        <p>欢迎互换友链，申请后站长审核通过即会显示。</p>
      </div>
      <dl class="site-info-list">
        <div v-for="item in siteInfo" :key="item.label" class="site-info-row">
          <dt>{{ item.label }}</dt>
          <dd :title="item.value">{{ item.value }}</dd>
          <button type="button" class="copy-btn" @click="copyText(item.value, item.label)">复制</button>
        </div>
      </dl>
    </section>

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
            class="link-card" :class="{ 'has-actions': isAdmin || item.USERCODE === userStore.userBean.code, 'is-pending': item.STATUS === 'pending' }">
          <span v-if="item.STATUS === 'pending'" class="pending-stamp" title="站长审核通过后，所有人才能看到">待审核</span>
          <a :href="safeLink(item.LINK)" target="_blank" rel="noopener noreferrer" class="card-link">
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
                    <el-dropdown-item v-if="isAdmin && item.STATUS === 'pending'" command="approve"><el-icon><Check/></el-icon>通过审核</el-dropdown-item>
                    <el-dropdown-item v-if="isAdmin && item.STATUS === 'pending'" command="reject" class="danger-item"><el-icon><CloseBold/></el-icon>拒绝</el-dropdown-item>
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
          <a :href="safeLink(item.LINK)" target="_blank" rel="noopener noreferrer" class="card-link">
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
      <div class="dialog-tip" v-if="!isRecommendMode && !isAdmin">
        欢迎互换友链！请确保您的站点能够正常访问，且包含本站链接。提交{{ isEditMode ? '修改' : '' }}后需要站长审核，通过后才会公开显示。
      </div>
      <div class="dialog-tip dialog-tip--warm" v-if="isRecommendMode && !isEditMode">
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
          <div class="avatar-row">
            <el-input v-model="applyForm.AVATAR" placeholder="图片 URL，建议正方形比例" clearable/>
            <el-button :loading="fetchingInfo" :disabled="!applyForm.LINK" @click="fetchSiteInfo">自动获取</el-button>
          </div>
          <p class="form-hint">填好网站地址后点“自动获取”，会读取对方网站的图标，名称和描述为空时也会一并填上。</p>
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
import {Plus, MoreFilled, Edit, Delete, Check, CloseBold} from '@element-plus/icons-vue';
import {ele_confirm, getCurrentUserAdminObject, sendAxiosRequest, sendAxiosRequestChecked} from "@/utils/common.js";
import {absoluteUrl, DEFAULT_DESCRIPTION, SITE_NAME} from "@/utils/seo.js";
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

// 只允许 http/https 链接：后端已校验，这里再兜底库里已有的旧数据，防止 javascript: 链接被点击执行
const safeLink = (link) => /^https?:\/\//i.test(String(link || '').trim()) ? String(link).trim() : undefined;

// 本站信息：别人互加友链时需要的几项
const siteInfo = [
  {label: '网站名称', value: SITE_NAME},
  {label: '网站地址', value: absoluteUrl('/')},
  {label: '网站头像', value: absoluteUrl('/icon-512.png')},
  {label: '一句话描述', value: DEFAULT_DESCRIPTION}
];

async function copyText(text, label) {
  try {
    await navigator.clipboard.writeText(text);
    ElMessage.success(`已复制${label}`);
  } catch {
    ElMessage.warning('浏览器不允许自动复制，请手动选中复制');
  }
}

// 填好网址后自动读取对方网站的图标、名称和简介（复用便签墙的链接卡片接口，需要登录）
const fetchingInfo = ref(false);
async function fetchSiteInfo() {
  if (!userStore?.userBean?.code) {
    ElMessage.warning('登录后才能自动获取');
    return;
  }
  let url = String(applyForm.LINK || '').trim();
  if (url && !/^https?:\/\//i.test(url)) url = 'https://' + url;
  fetchingInfo.value = true;
  const result = await sendAxiosRequestChecked('/blog-api/community/previewLink', {url}, '没能读取这个网站的信息，可以手动填写');
  fetchingInfo.value = false;
  const info = result?.result;
  if (!info) return;
  if (info.icon || info.image) applyForm.AVATAR = info.icon || info.image;
  if (!applyForm.NAME && (info.siteName || info.title)) applyForm.NAME = String(info.siteName || info.title).slice(0, 40);
  if (!applyForm.REMARK && info.description) applyForm.REMARK = String(info.description).slice(0, 50);
  ElMessage.success(info.icon || info.image ? '已获取网站图标' : '这个网站没有提供图标，可以手动填写');
}

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

const handleCommand = async (command, item) => {
  if (command === 'approve' || command === 'reject') {
    const result = await sendAxiosRequestChecked("/blog-api/friendLink/reviewFriendLink", {friendLinkId: item.GUID, action: command}, "审核失败");
    if (!result) return;
    if (command === 'approve') {
      item.STATUS = 'approved';
      ElMessage.success(`已通过「${item.NAME}」，申请人会收到通知`);
    } else {
      friendLinks.value = friendLinks.value.filter(link => link.GUID !== item.GUID);
      ElMessage.success(`已拒绝「${item.NAME}」，申请人会收到通知`);
    }
    return;
  }
  if (command === 'edit') {
    isEditMode.value = true;
    isRecommendMode.value = item.LINK_TYPE === 2;
    Object.assign(applyForm, JSON.parse(JSON.stringify(item)));
    dialogVisible.value = true;
  } else if (command === 'delete') {
    ele_confirm(`是否确认删除该记录?`, async () => {
      const result = await sendAxiosRequestChecked("/blog-api/friendLink/deleteFriendLink", {friendLinkId: item.GUID}, "删除失败");
      if (!result) return;
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

  // 以后台结果为准：网址不合法、超过数量上限等会在这里提示，不会出现“显示成功其实没保存”
  if (isEditMode.value) {
    const result = await sendAxiosRequestChecked("/blog-api/friendLink/updateFriendLink", {friendLink: applyForm}, "修改失败");
    if (!result) return false;
    const saved = result.result?.[0] || applyForm;
    const target = friendLinks.value.find(oneLink => oneLink.GUID === applyForm.GUID);
    if (target) Object.assign(target, saved);
    ElMessage.success(saved.STATUS === 'pending' ? "已提交修改，站长审核通过后显示" : "修改成功！");
  } else {
    // 编号、作者由后台生成；普通用户每人最多 3 个友链，也由后台检查
    const result = await sendAxiosRequestChecked("/blog-api/friendLink/addFriendLink", {friendLink: {...applyForm}}, "提交失败");
    if (!result) return false;
    const saved = result.result?.[0];
    if (saved) friendLinks.value.push(saved);
    ElMessage.success(saved?.STATUS === 'pending' ? "已提交申请，站长审核通过后会显示在这里" : "添加成功！");
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
  topAlert.value = await getAnnouncementByRouterName("FriendLink");
}

onMounted(() => {
  getFriendLinks();
  setTopAlert();
});
</script>

<style scoped>
.friend-link-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 36px 20px 56px;
  color: var(--j-ink);
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
  align-items: flex-end;
  padding-bottom: 8px;
}

/* 页头：一张贴胶带的标签 */
.heading-label {
  position: relative;
  padding: 20px 30px 18px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
  transform: rotate(-0.6deg);
}

.heading-tape {
  top: -10px;
  left: -18px;
  transform: rotate(-30deg);
}

.page-heading h1 {
  margin: 0;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 30px;
  line-height: 1.3;
}

.page-intro {
  margin: 8px 0 0;
  font-size: 15px;
  line-height: 1.6;
  color: var(--j-ink-soft);
}

.page-stats {
  margin: 8px 0 0;
  padding-top: 8px;
  border-top: 1px dashed var(--j-rule);
  font-family: var(--j-hand);
  font-size: 14px;
  color: var(--j-muted);
}

.link-section {
  margin-top: 36px;
}

.section-heading {
  margin-bottom: 18px;
}

.title-line {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 12px;
}

.title-line h2 {
  margin: 0;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 23px;
  line-height: 1.5;
}

.section-count {
  font-size: 13px;
  color: var(--j-muted);
}

.section-heading p {
  margin: 4px 0 0;
  font-size: 14px;
  line-height: 1.6;
  color: var(--j-ink-soft);
}

.friend-grid,
.recommend-grid {
  display: grid;
  gap: 18px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

/* 友情链接：名片 */
.link-card {
  position: relative;
  min-width: 0;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.07);
  transition: transform 200ms ease, box-shadow 200ms ease;
}

.friend-grid .link-card:nth-child(4n+1) { transform: rotate(-0.8deg); }
.friend-grid .link-card:nth-child(4n+2) { transform: rotate(0.6deg); }
.friend-grid .link-card:nth-child(4n+3) { transform: rotate(-0.3deg); }
.friend-grid .link-card:nth-child(4n) { transform: rotate(0.9deg); }

/* 推荐好站：牛皮纸书签，左侧有一个穿孔 */
.recommend-grid .link-card {
  background: var(--j-kraft);
  border-color: #d8c49d;
}

.recommend-grid .link-card::before {
  content: "";
  position: absolute;
  top: 50%;
  left: 10px;
  width: 8px;
  height: 8px;
  margin-top: -4px;
  border-radius: 50%;
  background: var(--j-desk);
  box-shadow: inset 0 1px 2px rgba(60, 50, 30, 0.35);
}

.recommend-grid .card-link {
  padding-left: 28px;
}

.card-link {
  display: flex;
  align-items: center;
  gap: 12px;
  box-sizing: border-box;
  height: 100%;
  min-height: 104px;
  padding: 16px;
  color: inherit;
  text-decoration: none;
}

.site-avatar {
  flex-shrink: 0;
  border: 3px solid #fff;
  border-radius: 2px;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.25);
  background: #ece4d3;
  color: var(--j-ink-soft);
  font-family: var(--j-hand);
  font-size: 20px;
}

.card-content {
  flex: 1;
  min-width: 0;
}

.site-name {
  overflow: hidden;
  margin: 0 0 5px;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.5;
  color: var(--j-ink);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.site-desc {
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--j-ink-soft);
  overflow-wrap: anywhere;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.recommend-grid .site-desc {
  color: #6b5330;
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
  border-radius: 3px;
  background: transparent;
  color: var(--j-muted);
  cursor: pointer;
}

.action-btn:hover {
  background: rgba(250, 216, 96, 0.35);
  color: var(--j-ink);
}

.card-link:focus-visible,
.action-btn:focus-visible {
  outline: 2px solid var(--j-pen);
  outline-offset: 2px;
}

.danger-item {
  color: var(--j-stamp);
}

.friend-link-container.is-embed {
  max-width: 100%;
  margin: 0;
  padding: 0;
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
  background: var(--j-paper-warm);
  border-left: 3px solid var(--j-rule-strong);
  font-size: 14px;
  line-height: 1.6;
  color: var(--j-ink-soft);
}

.dialog-tip--warm {
  background: #fdf3c4;
  border-left-color: #e8c95b;
  color: #7a5c12;
}

/* ================= 本站信息 ================= */
.site-info {
  margin: 28px 0 8px;
  padding: 18px 22px;
  background: var(--j-note);
  box-shadow: var(--j-shadow);
}

.site-info-head h2 {
  margin: 0 0 4px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 20px;
}

.site-info-head p {
  margin: 0 0 12px;
  font-size: 13px;
  color: var(--j-muted);
}

.site-info-list {
  margin: 0;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 24px;
}

.site-info-row {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  font-size: 14px;
}

.site-info-row dt {
  flex-shrink: 0;
  width: 5em;
  color: var(--j-muted);
}

.site-info-row dd {
  flex: 1;
  min-width: 0;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--j-ink);
}

.copy-btn {
  flex-shrink: 0;
  border: 1px dashed var(--j-rule-strong);
  background: var(--j-paper);
  padding: 0 8px;
  font-size: 12px;
  line-height: 1.8;
  color: var(--j-pen);
  cursor: pointer;
}

.copy-btn:hover {
  border-color: var(--j-pen);
}

/* ================= 待审核 ================= */
.link-card.is-pending {
  opacity: 0.75;
  border-style: dashed;
}

.pending-stamp {
  position: absolute;
  top: -8px;
  left: 10px;
  z-index: 1;
  padding: 0 6px;
  border: 1.5px solid var(--j-stamp);
  background: var(--j-paper);
  color: var(--j-stamp);
  font-size: 12px;
  line-height: 1.6;
  transform: rotate(-4deg);
}

.avatar-row {
  display: flex;
  gap: 8px;
  width: 100%;
}

.form-hint {
  margin: 4px 0 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--j-muted);
}

@media (hover: hover) {
  .link-card:hover {
    transform: rotate(0deg) translateY(-3px) !important;
    box-shadow: 0 1px 2px rgba(60, 50, 30, 0.07), 0 12px 20px -12px rgba(60, 50, 30, 0.45);
  }

  .link-card:hover .site-name {
    color: var(--j-pen);
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
    padding: 24px 14px 36px;
  }

  .heading-label {
    transform: none;
  }

  .site-info {
    padding: 14px 16px;
  }

  .site-info-list {
    grid-template-columns: minmax(0, 1fr);
  }

  .friend-grid,
  .recommend-grid {
    grid-template-columns: 1fr;
  }

  .link-card {
    transform: none !important;
  }

  .card-link {
    min-height: 92px;
    padding: 14px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .link-card {
    transition: none;
  }
}
</style>
