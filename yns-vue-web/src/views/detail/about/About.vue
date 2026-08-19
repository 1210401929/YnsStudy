<template>
  <!-- 公告横幅 -->
  <Announcement
      v-for="al in topAlert"
      :key="al.GUID"
      :TEXT="al.TEXT"
      :URL="al.URL"
      :URLNAME="al.URLNAME"
  />

  <div class="about-wrapper">
    <div class="about-container">

      <!-- 顶部欢迎区 -->
      <div class="hero">
        <h1>欢迎来到 YnsStudy</h1>
        <p class="hero-description">
          一个集文章创作、评论互动与资源分享于一体的知识社区
        </p>

        <!-- 站长想说 -->
        <section class="site-owner-message">
          <div class="owner-title">站长想说</div>

          <p>
            正如上面的介绍，这个网站从建立之初，就不只是想做成我一个人的博客，
            而是想为每一个热爱写文章的人提供一个可以自由创作的平台。
            在这里，每个人都可以发表属于自己的文章。
          </p>

          <p>
            也正因如此，很多博客聚合网站拒绝收录本站，
            原因也很简单：这个网站并不算一个纯粹的个人博客。
          </p>

          <p>
            建站到现在，网站的流量确实不多，也确实没有太多能让路人驻足的理由。
            因为没有多少流量，也没有多少用户，它慢慢地几乎又变成了我自己的个人博客。
          </p>

          <p>
            我的技术分享、生活吐槽，还有这些年留下的一点一滴，都记录在了这里。
            我只希望自己还能坚守当初建站时的本心，让这个网站一直活下去。
          </p>
        </section>
      </div>

      <!-- 数据统计框 -->
      <div class="stats-section">
        <h2>数据统计</h2>

        <div class="stats">
          <div class="stat-card">
            <h3>{{ stats.ARTICLENUM }}</h3>
            <p>文章</p>
          </div>

          <div class="stat-card">
            <h3>{{ stats.COMMUNITYNUM }}</h3>
            <p>讨论</p>
          </div>

          <div class="stat-card">
            <h3>{{ stats.VIEW_PAGE }}</h3>
            <p>阅读</p>
          </div>

          <div class="stat-card">
            <h3>{{ stats.USERNUM }}</h3>
            <p>用户</p>
          </div>

          <div class="stat-card">
            <h3>{{ stats.USERLOGINNUM }}</h3>
            <p>正式用户访问量</p>
          </div>
        </div>
      </div>

      <!-- 功能区域 -->
      <div class="features">
        <div
            class="feature-card"
            v-for="feature in features"
            :key="feature.title"
        >
          <div class="icon">{{ feature.icon }}</div>
          <h3>{{ feature.title }}</h3>
          <p>{{ feature.description }}</p>
        </div>
      </div>

      <!-- 技术栈 -->
      <div class="tech-section">
        <h2>使用技术</h2>

        <ul>
          <li>前端：Vue 3 + Element Plus</li>
          <li>后端：Go（2026年8月弃用 Spring Cloud + MyBatis）</li>
          <li>数据库：MySQL + Redis</li>
          <li>存储：本地 / 云端</li>
        </ul>
      </div>

    </div>
  </div>
</template>

<script setup>
import {ref, onMounted} from 'vue'
import {sendAxiosRequest} from "@/utils/common.js"
import Announcement from "@/components/detail/Announcement.vue"
import {getAnnouncementByRouterName} from "@/utils/blogUtil.js"

const features = [
  {
    title: '✍️ 发布文章',
    description: '记录学习心得与技术经验，支持富文本编辑。',
    icon: '📝'
  },
  {
    title: '💬 评论互动',
    description: '查看并参与技术讨论，促进社区交流。',
    icon: '💬'
  },
  {
    title: '⬆️ 上传资源',
    description: '分享你收集的学习资料。',
    icon: '📂'
  },
  {
    title: '⬇️ 下载工具',
    description: '下载他人分享的学习资源和开发工具。',
    icon: '📥'
  }
]

// 网站统计信息
const stats = ref({})

// 公告横幅内容
const topAlert = ref([])

onMounted(() => {
  getWebsiteStatistics()
})

// 获取网站统计数据
const getWebsiteStatistics = async () => {
  let result = await sendAxiosRequest("/blog-api/home/getWebsiteStatistics")

  if (result && !result.isError) {
    stats.value = result.result
  }
}

// 获取公告
const setTopAlert = async () => {
  topAlert.value = await getAnnouncementByRouterName("About")
}

setTopAlert()
</script>

<style scoped>

/* 外层容器 */
.about-wrapper {
  display: flex;
  flex-direction: column;
  height: 100%;
  max-height: 100%;
  overflow: hidden;
  background: #f9f9f9;
}

/* 主体内容区域 */
.about-container {
  flex: 1;
  overflow-y: auto;
  padding: 0 16px 16px 16px;
  box-sizing: border-box;
  width: 100%;
  max-width: 1400px;
  margin-left: auto;
  margin-right: auto;
}

/* =========================
   顶部欢迎区
   ========================= */

.hero {
  text-align: center;
  padding: 12px 8px;
}

.hero h1 {
  font-size: 22px;
  margin-bottom: 6px;
  color: #409eff;
}

.hero-description {
  margin: 0;
  font-size: 14px;
  color: #666;
}

/* =========================
   站长想说
   ========================= */

.site-owner-message {
  max-width: 900px;
  margin: 18px auto 0;
  padding: 12px 16px;

  text-align: left;

  background: #f5f6f7;
  border-left: 3px solid #dcdfe6;
  border-radius: 4px;
}

.owner-title {
  margin-bottom: 8px;

  font-size: 13px;
  font-weight: 600;
  color: #777;
}

.hero .site-owner-message p {
  margin: 6px 0;

  font-size: 13px;
  line-height: 1.8;
  color: #888;
}

/* =========================
   数据统计区域
   ========================= */

.stats-section {
  background: #f3f3f3;
  padding: 16px;
  margin-top: 20px;
  border-radius: 8px;
}

.stats-section h2 {
  font-size: 18px;
  color: #48e3a7;
  margin-bottom: 16px;
  text-align: center;
}

.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
  margin-top: 12px;
}

.stat-card {
  background: #ffffff;
  border-radius: 10px;

  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);

  padding: 16px;
  text-align: center;

  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-5px);

  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.1);
}

.stat-card h3 {
  font-size: 22px;
  margin: 8px 0;
  color: #409eff;
}

.stat-card p {
  font-size: 14px;
  color: #888;
}

/* =========================
   功能卡片区域
   ========================= */

.features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.feature-card {
  background: white;
  border-radius: 10px;

  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);

  padding: 12px;
  text-align: center;
}

.feature-card .icon {
  font-size: 24px;
  margin-bottom: 6px;
}

.feature-card h3 {
  font-size: 16px;
  margin: 4px 0;
}

.feature-card p {
  font-size: 13px;
  color: #555;
}

/* =========================
   技术栈介绍
   ========================= */

.tech-section {
  background: white;
  padding: 12px;
  margin-top: 12px;
  border-radius: 8px;
}

.tech-section h2 {
  font-size: 16px;
  margin-bottom: 8px;
}

.tech-section ul {
  padding-left: 18px;
  font-size: 13px;
  color: #444;
}

/* =========================
   底部区域
   ========================= */

.footer {
  text-align: center;
  font-size: 12px;
  color: #888;

  padding: 12px;

  background: #fff;
  border-top: 1px solid #eee;
}

.footer a {
  color: #409eff;
  text-decoration: none;
}

.footer a:hover {
  text-decoration: underline;
}

</style>

