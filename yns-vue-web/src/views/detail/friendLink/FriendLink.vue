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
    <section v-if="!isEmbed" class="link-hero">
      <div class="hero-glow hero-glow-one"></div>
      <div class="hero-glow hero-glow-two"></div>
      <div class="hero-copy">
        <span class="hero-kicker"><i></i> 朋友和精选</span>
        <h1>把散落在互联网里的<br><em>好地方</em>，认真地放在一起</h1>
        <p>这里有长期同行的朋友，也有值得专程拜访的内容站点。沿着链接出发，认识更大的互联网。</p>
        <div class="hero-stats">
          <div class="hero-stat">
            <strong>{{ normalFriendLinks.length }}</strong>
            <span>同行伙伴</span>
          </div>
          <span class="stat-divider"></span>
          <div class="hero-stat">
            <strong>{{ recommendLinks.length }}</strong>
            <span>站长精选</span>
          </div>
        </div>
      </div>
      <div class="hero-side">
        <div class="orbit-mark" aria-hidden="true">
          <span class="orbit-core">Y</span>
          <span class="orbit-dot orbit-dot-one"></span>
          <span class="orbit-dot orbit-dot-two"></span>
        </div>
        <el-button
            type="primary"
            round
            size="large"
            class="apply-btn hero-apply-btn"
            @click="pushFriendLink"
        >
          <el-icon class="el-icon--left">
            <Plus/>
          </el-icon>
          加入朋友墙
        </el-button>
      </div>
    </section>

    <!-- ================= 友情链接模块 ================= -->
    <section :class="['section-shell', 'friend-section', { 'embed-section': isEmbed }]">
      <div class="section-heading friend-heading">
        <div class="heading-copy">
          <span class="section-index" v-if="!isEmbed">01</span>
          <div>
            <div class="title-line">
              <span class="title-mark friend-mark">↗</span>
              <h2>{{ isEmbed ? '发现宝藏站点' : '友情链接' }}</h2>
              <span class="count-pill">{{ normalFriendLinks.length }} 位朋友</span>
            </div>
            <p>{{ isEmbed ? '从这里拜访更多有趣的创作者' : '认真交换的不只是链接，也是彼此对内容的长期关注。' }}</p>
          </div>
        </div>
      </div>

      <div class="friend-grid">
        <div
            class="friend-grid-item"
            v-for="(item, index) in normalFriendLinks"
            :key="item.GUID"
        >
          <a
              :href="item.LINK"
              target="_blank"
              rel="noopener noreferrer"
              class="friend-card fade-in-up"
              :style="{ animationDelay: `${index * 0.06}s` }"
          >
            <div class="friend-avatar-wrap">
              <el-avatar :src="item.AVATAR" :size="isEmbed ? 46 : 52" class="site-avatar friend-avatar">
                {{ item.NAME?.charAt(0) }}
              </el-avatar>
            </div>
            <div class="card-content">
              <h3 class="site-name" :title="item.NAME">{{ item.NAME }}</h3>
              <p class="site-desc" :title="item.REMARK">{{ item.REMARK || '这个朋友很低调，还没有留下介绍' }}</p>
            </div>
            <span class="friend-card-arrow">↗</span>

            <div class="card-action-menu" @click.prevent.stop v-if="isAdmin || item.USERCODE===userStore.userBean.code">
              <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, item)">
                <span class="action-btn">
                  <el-icon><MoreFilled/></el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="edit">
                      <el-icon><Edit/></el-icon>编辑
                    </el-dropdown-item>
                    <el-dropdown-item command="delete" class="danger-item">
                      <el-icon><Delete/></el-icon>删除
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </a>
        </div>
      </div>
    </section>

    <!-- ================= 推荐好站模块 (仅非嵌入模式显示) ================= -->
    <section class="section-shell recommend-section" v-if="!isEmbed && (recommendLinks.length > 0 || isAdmin)">
      <div class="section-heading recommend-heading">
        <div class="heading-copy">
          <span class="section-index">02</span>
          <div>
            <div class="title-line">
              <span class="title-mark recommend-mark">✦</span>
              <h2>推荐好站</h2>
              <span class="count-pill dark-count">{{ recommendLinks.length }} 个精选</span>
            </div>
            <p>不是简单收录，而是我愿意主动推荐给你的站点与资源。</p>
          </div>
        </div>
        <el-button
            v-if="isAdmin"
            round
            class="recommend-add-btn"
            @click="pushRecommendLink"
        >
          <el-icon class="el-icon--left"><Plus/></el-icon>
          添加推荐
        </el-button>
      </div>

      <div class="featured-grid">
        <div
            class="featured-grid-item"
            v-for="(item, index) in recommendLinks"
            :key="item.GUID"
        >
          <a
              :href="item.LINK"
              target="_blank"
              rel="noopener noreferrer"
              class="featured-card fade-in-up"
              :style="{ animationDelay: `${index * 0.08}s` }"
          >
            <span class="featured-number">{{ String(index + 1).padStart(2, '0') }}</span>
            <div class="featured-avatar-wrap">
              <el-avatar :src="item.AVATAR" :size="48" class="site-avatar featured-avatar">
                {{ item.NAME?.charAt(0) }}
              </el-avatar>
            </div>
            <div class="featured-content">
              <span class="featured-label">EDITOR'S PICK</span>
              <h3 :title="item.NAME">{{ item.NAME }}</h3>
              <p :title="item.REMARK">{{ item.REMARK || '一个值得花时间探索的好站点' }}</p>
              <div class="featured-footer">
                <span>站长精选</span>
                <span class="visit-link">访问站点 <b>↗</b></span>
              </div>
            </div>

            <div class="card-action-menu featured-action-menu" @click.prevent.stop v-if="isAdmin">
              <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, item)">
                <span class="action-btn">
                  <el-icon><MoreFilled/></el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="edit">
                      <el-icon><Edit/></el-icon>编辑
                    </el-dropdown-item>
                    <el-dropdown-item command="delete" class="danger-item">
                      <el-icon><Delete/></el-icon>删除
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </a>
        </div>
      </div>
    </section>

    <!-- ================= 弹窗表单 ================= -->
    <el-dialog
        v-model="dialogVisible"
        :title="isEditMode ? (isRecommendMode ? '编辑推荐' : '编辑友链') : (isRecommendMode ? '添加推荐' : '申请加入友链')"
        width="480px"
        :close-on-click-modal="false"
        destroy-on-close
        class="custom-dialog"
    >
      <div class="dialog-tip" v-if="!isEditMode && !isRecommendMode">
        欢迎互换友链！请确保您的站点能够正常访问，且包含本站链接。
      </div>
      <div class="dialog-tip" style="color: #e6a23c; background-color: #fdf6ec;" v-if="isRecommendMode && !isEditMode">
        添加的内容将展示在“推荐好站”专区。
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
  --page-ink: #14213d;
  --muted-ink: #6e7b91;
  --soft-line: #e8edf5;
  --brand-blue: #3b6ff5;
  --brand-cyan: #4fd1c5;
  padding: 34px 20px 72px;
  max-width: 1200px;
  margin: 0 auto;
  color: var(--page-ink);
}

.link-hero {
  position: relative;
  min-height: 250px;
  display: grid;
  grid-template-columns: minmax(0, 1.55fr) minmax(220px, 0.45fr);
  align-items: center;
  gap: 28px;
  padding: 32px 52px;
  margin-bottom: 8px;
  overflow: hidden;
  border-radius: 30px;
  background:
      linear-gradient(120deg, rgba(255, 255, 255, 0.98), rgba(247, 250, 255, 0.93)),
      radial-gradient(circle at 85% 20%, rgba(79, 209, 197, 0.2), transparent 35%);
  border: 1px solid rgba(59, 111, 245, 0.12);
  box-shadow: 0 24px 70px rgba(45, 70, 112, 0.11);
}

.link-hero::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image: radial-gradient(rgba(59, 111, 245, 0.13) 1px, transparent 1px);
  background-size: 22px 22px;
  mask-image: linear-gradient(90deg, transparent 48%, #000 100%);
}

.hero-glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(2px);
  pointer-events: none;
}

.hero-glow-one {
  width: 230px;
  height: 230px;
  right: -70px;
  top: -90px;
  background: rgba(79, 209, 197, 0.18);
}

.hero-glow-two {
  width: 180px;
  height: 180px;
  left: 44%;
  bottom: -130px;
  background: rgba(59, 111, 245, 0.13);
}

.hero-copy,
.hero-side {
  position: relative;
  z-index: 1;
}

.hero-kicker {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  margin-bottom: 12px;
  color: #54709e;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.18em;
}

.hero-kicker i {
  width: 22px;
  height: 2px;
  border-radius: 2px;
  background: linear-gradient(90deg, var(--brand-blue), var(--brand-cyan));
}

.hero-copy h1 {
  margin: 0;
  color: #172642;
  font-size: clamp(30px, 3.4vw, 44px);
  font-weight: 850;
  letter-spacing: -0.045em;
  line-height: 1.16;
}

.hero-copy h1 em {
  position: relative;
  color: var(--brand-blue);
  font-style: normal;
}

.hero-copy h1 em::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 2px;
  height: 8px;
  z-index: -1;
  border-radius: 8px;
  background: rgba(79, 209, 197, 0.33);
}

.hero-copy > p {
  max-width: 590px;
  margin: 14px 0 18px;
  color: var(--muted-ink);
  font-size: 15px;
  line-height: 1.7;
}

.hero-stats {
  display: flex;
  align-items: center;
  gap: 22px;
}

.hero-stat {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.hero-stat strong {
  color: #1f3152;
  font-size: 24px;
  line-height: 1;
}

.hero-stat span {
  color: #8490a3;
  font-size: 12px;
}

.stat-divider {
  width: 1px;
  height: 22px;
  background: #dfe6f0;
}

.hero-side {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 18px;
}

.orbit-mark {
  position: relative;
  width: 118px;
  height: 118px;
  display: grid;
  place-items: center;
  border: 1px solid rgba(59, 111, 245, 0.22);
  border-radius: 50%;
  box-shadow: inset 0 0 0 24px rgba(255, 255, 255, 0.45);
}

.orbit-mark::before,
.orbit-mark::after {
  content: '';
  position: absolute;
  border-radius: 50%;
  border: 1px solid rgba(79, 209, 197, 0.22);
}

.orbit-mark::before { inset: 14px; }
.orbit-mark::after { inset: 31px; }

.orbit-core {
  width: 52px;
  height: 52px;
  display: grid;
  place-items: center;
  z-index: 1;
  border-radius: 18px;
  color: #fff;
  background: linear-gradient(145deg, #3b6ff5, #3157bf);
  box-shadow: 0 14px 30px rgba(59, 111, 245, 0.3);
  font-family: Georgia, serif;
  font-size: 25px;
  font-weight: 700;
}

.orbit-dot {
  position: absolute;
  width: 12px;
  height: 12px;
  z-index: 2;
  border: 3px solid #fff;
  border-radius: 50%;
  box-shadow: 0 4px 10px rgba(20, 33, 61, 0.16);
}

.orbit-dot-one {
  top: 12px;
  right: 17px;
  background: #4fd1c5;
}

.orbit-dot-two {
  left: 9px;
  bottom: 27px;
  background: #ffbd59;
}

.apply-btn {
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.apply-btn:hover {
  transform: translateY(-2px);
}

.hero-apply-btn {
  min-width: 148px;
  height: 42px;
  border: 0;
  background: linear-gradient(135deg, #3b6ff5, #315bc9);
  box-shadow: 0 12px 26px rgba(59, 111, 245, 0.27);
}

.section-shell {
  position: relative;
  padding: 38px;
  margin-top: 28px;
  border-radius: 26px;
}

.section-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 28px;
}

.heading-copy {
  display: flex;
  align-items: flex-start;
  gap: 18px;
  min-width: 0;
}

.section-index {
  padding-top: 5px;
  color: #9aa7ba;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.12em;
}

.title-line {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

.title-line h2 {
  margin: 0;
  font-size: 25px;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.title-mark {
  width: 30px;
  height: 30px;
  display: inline-grid;
  place-items: center;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 900;
}

.section-heading p {
  margin: 8px 0 0;
  color: #8a96a9;
  font-size: 13px;
  line-height: 1.65;
}

.count-pill {
  padding: 5px 9px;
  border: 1px solid #dfe6f0;
  border-radius: 999px;
  color: #718096;
  background: #f8fafc;
  font-size: 11px;
  font-weight: 650;
}

.recommend-section {
  overflow: hidden;
  color: #fff;
  background:
      radial-gradient(circle at 92% 8%, rgba(79, 209, 197, 0.13), transparent 28%),
      linear-gradient(145deg, #14213d 0%, #1d2e52 55%, #172642 100%);
  box-shadow: 0 24px 55px rgba(20, 33, 61, 0.18);
}

.recommend-section::after {
  content: 'YnsStudy';
  position: absolute;
  right: -16px;
  top: 56px;
  color: rgba(255, 255, 255, 0.025);
  font-size: 78px;
  font-weight: 900;
  letter-spacing: 0.05em;
  pointer-events: none;
}

.recommend-heading,
.featured-grid {
  position: relative;
  z-index: 1;
}

.featured-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.featured-grid-item,
.friend-grid-item {
  min-width: 0;
}

.recommend-heading .section-index { color: rgba(255, 255, 255, 0.42); }
.recommend-heading .title-line h2 { color: #fff; }
.recommend-heading p { color: #9eacc3; }
.recommend-mark { color: #1d2e52; background: #ffd276; }

.dark-count {
  color: #d7e0ee;
  border-color: rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.06);
}

.recommend-add-btn {
  color: #e8eef8;
  border-color: rgba(255, 255, 255, 0.18);
  background: rgba(255, 255, 255, 0.07);
}

.recommend-add-btn:hover {
  color: #1d2e52;
  border-color: #ffd276;
  background: #ffd276;
}

.featured-card {
  position: relative;
  min-height: 176px;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 16px 14px;
  box-sizing: border-box;
  overflow: hidden;
  color: inherit;
  text-decoration: none;
  border: 1px solid rgba(255, 255, 255, 0.11);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.065);
  backdrop-filter: blur(10px);
  transition: transform 0.28s ease, border-color 0.28s ease, background 0.28s ease, box-shadow 0.28s ease;
}

.featured-card::before {
  content: '';
  position: absolute;
  width: 130px;
  height: 130px;
  left: -80px;
  bottom: -86px;
  border-radius: 50%;
  background: rgba(79, 209, 197, 0.14);
  transition: transform 0.35s ease;
}

.featured-card:hover {
  transform: translateY(-5px);
  border-color: rgba(255, 210, 118, 0.48);
  background: rgba(255, 255, 255, 0.105);
  box-shadow: 0 18px 34px rgba(5, 14, 32, 0.22);
}

.featured-card:hover::before { transform: scale(1.4); }

.featured-number {
  position: absolute;
  right: 12px;
  bottom: 9px;
  color: rgba(255, 255, 255, 0.09);
  font-family: Georgia, serif;
  font-size: 28px;
  font-style: italic;
}

.featured-avatar-wrap {
  position: relative;
  flex-shrink: 0;
  padding: 3px;
  border: 1px solid rgba(255, 210, 118, 0.38);
  border-radius: 16px;
}

.featured-avatar {
  border-radius: 12px;
  background: linear-gradient(145deg, #fff4d8, #ffd276);
  color: #45371a;
}

.featured-content {
  min-width: 0;
  flex: 1;
}

.featured-label {
  display: block;
  margin-bottom: 5px;
  color: #ffd276;
  font-size: 8px;
  font-weight: 800;
  letter-spacing: 0.16em;
}

.featured-content h3 {
  margin: 0;
  overflow: hidden;
  color: #fff;
  font-size: 14px;
  font-weight: 750;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.featured-content > p {
  min-height: 36px;
  margin: 6px 0 10px;
  overflow: hidden;
  color: #aebbd0;
  display: -webkit-box;
  font-size: 10.5px;
  line-height: 1.7;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.featured-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  padding-top: 8px;
  color: #8494ae;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  font-size: 9px;
}

.visit-link {
  color: #eaf0fa;
  font-size: 9px;
  transition: color 0.2s ease;
}

.visit-link b {
  display: inline-block;
  margin-left: 3px;
  color: #ffd276;
  transition: transform 0.2s ease;
}

.featured-card:hover .visit-link b { transform: translate(2px, -2px); }

.friend-section {
  border: 1px solid rgba(59, 111, 245, 0.1);
  background: transparent;
  box-shadow: none;
}

.friend-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.friend-mark {
  color: #2f63db;
  background: #eaf1ff;
}

.friend-card {
  position: relative;
  height: 112px;
  width: 100%;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px;
  box-sizing: border-box;
  overflow: hidden;
  color: inherit;
  text-decoration: none;
  border: 1px solid #e9edf4;
  border-radius: 17px;
  background: linear-gradient(145deg, #fff, #fbfcff);
  transition: transform 0.25s ease, border-color 0.25s ease, box-shadow 0.25s ease;
}

.friend-card::after {
  content: '';
  position: absolute;
  left: 0;
  top: 22%;
  bottom: 22%;
  width: 3px;
  border-radius: 0 4px 4px 0;
  background: linear-gradient(180deg, var(--brand-blue), var(--brand-cyan));
  opacity: 0;
  transform: scaleY(0.4);
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.friend-card:hover {
  transform: translateY(-4px);
  border-color: rgba(59, 111, 245, 0.26);
  box-shadow: 0 14px 26px rgba(47, 79, 135, 0.11);
}

.friend-card:hover::after {
  opacity: 1;
  transform: scaleY(1);
}

.friend-avatar-wrap {
  flex-shrink: 0;
  padding: 3px;
  border: 1px solid #e1e8f4;
  border-radius: 15px;
  background: #fff;
}

.friend-avatar {
  border-radius: 12px;
  background: #edf3ff;
  color: #3b6ff5;
  font-weight: 750;
}

.card-content {
  min-width: 0;
  flex: 1;
  padding-right: 18px;
}

.site-name {
  margin: 0 0 6px;
  overflow: hidden;
  color: #273753;
  font-size: 15px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.site-desc {
  margin: 0;
  overflow: hidden;
  color: #8a96a9;
  display: -webkit-box;
  font-size: 11px;
  line-height: 1.55;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.friend-card-arrow {
  position: absolute;
  right: 14px;
  bottom: 13px;
  color: #b5c0d1;
  font-size: 14px;
  transition: color 0.2s ease, transform 0.2s ease;
}

.friend-card:hover .friend-card-arrow {
  color: var(--brand-blue);
  transform: translate(2px, -2px);
}

.card-action-menu {
  position: absolute;
  top: 9px;
  right: 9px;
  z-index: 3;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.featured-action-menu { top: 10px; right: 10px; }
.featured-card:hover .card-action-menu,
.friend-card:hover .card-action-menu { opacity: 1; }

.action-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ba8bb;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 5px 14px rgba(20, 33, 61, 0.12);
}

.action-btn:hover {
  color: var(--brand-blue);
  background: #fff;
}

.featured-action-menu .action-btn {
  color: #d9e2ef;
  background: rgba(9, 19, 39, 0.58);
  box-shadow: none;
}

.danger-item { color: #f56c6c !important; }

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}

.fade-in-up {
  opacity: 0;
  animation: fadeInUp 0.52s cubic-bezier(0.22, 0.8, 0.24, 1) forwards;
}

/* --- 嵌入模式适配样式 --- */
.friend-link-container.is-embed {
  padding: 0;
  max-width: 100%;
  margin: 0;
}

.is-embed .embed-section {
  padding: 0;
  margin: 0;
  border: 0;
  border-radius: 0;
  box-shadow: none;
}

.is-embed .section-heading { margin-bottom: 18px; }
.is-embed .title-line h2 { font-size: 19px; }
.is-embed .section-heading p { margin-top: 4px; }
.is-embed .friend-grid { grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; }
.is-embed .friend-card { height: 92px; padding: 13px; border-radius: 14px; }
.is-embed .card-content { padding-right: 12px; }

.dialog-tip {
  padding: 11px 15px;
  margin-bottom: 20px;
  color: #4c8f2f;
  background: #f1f9ed;
  border: 1px solid #e2f1da;
  border-radius: 10px;
  font-size: 13px;
  line-height: 1.6;
}

@media (max-width: 1100px) {
  .featured-grid,
  .friend-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
}

@media (max-width: 900px) {
  .link-hero {
    grid-template-columns: 1fr auto;
    padding: 30px 36px;
  }

  .orbit-mark { width: 108px; height: 108px; }
  .orbit-core { width: 48px; height: 48px; }
  .hero-side { min-width: 130px; }
  .section-shell { padding: 30px; }
  .featured-grid,
  .friend-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}

@media (max-width: 700px) {
  .friend-link-container { padding: 20px 14px 48px; }

  .link-hero {
    min-height: auto;
    grid-template-columns: 1fr;
    gap: 18px;
    padding: 26px 22px;
    border-radius: 22px;
  }

  .link-hero::before { mask-image: linear-gradient(180deg, transparent 42%, #000 100%); }
  .hero-copy h1 { font-size: 30px; }
  .hero-copy > p { font-size: 14px; }
  .hero-side { align-items: flex-start; }
  .orbit-mark { display: none; }

  .section-shell {
    padding: 24px 18px;
    border-radius: 20px;
  }

  .section-heading {
    align-items: flex-start;
    flex-direction: column;
    margin-bottom: 22px;
  }

  .heading-copy { gap: 10px; }
  .title-line h2 { font-size: 22px; }
  .featured-grid,
  .friend-grid,
  .is-embed .friend-grid { grid-template-columns: 1fr; }
  .featured-card { min-height: 160px; padding: 20px; }
  .featured-avatar-wrap { align-self: flex-start; }
  .friend-card { height: 102px; }
  .card-action-menu { opacity: 1; }
  .friend-card-arrow { display: none; }
}

@media (max-width: 420px) {
  .hero-copy h1 { font-size: 30px; }
  .hero-stats { gap: 14px; }
  .featured-card { align-items: flex-start; gap: 14px; }
  .featured-avatar { --el-avatar-size: 52px !important; }
  .featured-content h3 { font-size: 16px; }
  .featured-content > p { margin-bottom: 9px; }
}

@media (prefers-reduced-motion: reduce) {
  .fade-in-up { opacity: 1; animation: none; }
  .featured-card,
  .friend-card,
  .apply-btn { transition: none; }
}
</style>
