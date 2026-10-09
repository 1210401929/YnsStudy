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
          <span class="j-tape j-tape--pink letter-tape"></span>
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
        <span class="j-tape j-tape--green tech-tape"></span>
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
.about-wrapper {
  min-height: 100vh;
  padding: 40px 20px 72px;
  color: var(--j-ink);
}

.about-container {
  max-width: 920px;
  margin: 0 auto;
}

/* ============ 顶部 ============ */
.hero {
  text-align: center;
}

.hero h1 {
  margin: 0;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 34px;
}

.hero-description {
  margin: 12px 0 0;
  font-size: 15px;
  color: var(--j-ink-soft);
}

/* 站长想说：一封写在横线信纸上的信 */
.site-owner-message {
  position: relative;
  margin: 40px auto 0;
  padding: 34px 44px 34px 64px;
  text-align: left;
  background-color: var(--j-paper);
  background-image: repeating-linear-gradient(transparent 0 33px, #ece4d4 33px 34px);
  background-position: 0 37px;
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
  transform: rotate(-0.4deg);
}

/* 信纸的页边红线 */
.site-owner-message::before {
  content: "";
  position: absolute;
  top: 0;
  bottom: 0;
  left: 42px;
  width: 1px;
  background: var(--j-margin-red);
  opacity: 0.6;
}

.letter-tape {
  top: -11px;
  left: 50%;
  margin-left: -42px;
  transform: rotate(-2deg);
}

.owner-title {
  margin: 0;
  font-family: var(--j-hand);
  font-size: 22px;
  line-height: 34px;
}

.site-owner-message p {
  margin: 0;
  font-family: var(--j-hand);
  font-size: 17px;
  line-height: 34px;
  color: var(--j-ink);
  text-indent: 2em;
}

/* ============ 数据统计 ============ */
.stats-section {
  margin-top: 56px;
  text-align: center;
}

.stats-section h2,
.tech-section h2 {
  margin: 0 0 22px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 24px;
}

.stats {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 22px;
}

/* 每个数字盖成一枚印章 */
.stat-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 128px;
  height: 128px;
  border: 2px solid rgba(47, 93, 138, 0.7);
  border-radius: 50%;
  outline: 1px solid rgba(47, 93, 138, 0.35);
  outline-offset: 4px;
  color: var(--j-pen);
  background: rgba(255, 253, 248, 0.6);
}

.stat-card:nth-child(odd) {
  border-color: rgba(194, 72, 62, 0.7);
  outline-color: rgba(194, 72, 62, 0.35);
  color: var(--j-stamp);
  transform: rotate(-6deg);
}

.stat-card:nth-child(even) {
  transform: rotate(5deg);
}

.stat-card h3 {
  margin: 0;
  font-family: var(--j-num);
  font-weight: normal;
  font-size: 30px;
  line-height: 1.1;
}

.stat-card p {
  margin: 6px 0 0;
  padding: 0 8px;
  font-family: var(--j-hand);
  font-size: 14px;
  line-height: 1.3;
}

/* ============ 功能：便签 ============ */
.features {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 22px;
  margin-top: 60px;
}

.feature-card {
  padding: 22px 18px 18px;
  background: var(--j-note);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.1), 0 12px 20px -14px rgba(60, 50, 30, 0.45);
  transition: transform 0.2s ease;
}

.feature-card:nth-child(4n+1) { background: var(--j-note); transform: rotate(-1.5deg); }
.feature-card:nth-child(4n+2) { background: #d4ead9; transform: rotate(1deg); }
.feature-card:nth-child(4n+3) { background: #f6d9d6; transform: rotate(-0.6deg); }
.feature-card:nth-child(4n) { background: #d6e4f0; transform: rotate(1.4deg); }

.feature-card:hover {
  transform: rotate(0deg) translateY(-3px);
}

.feature-card .icon {
  font-size: 28px;
}

.feature-card h3 {
  margin: 10px 0 6px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 19px;
}

.feature-card p {
  margin: 0;
  font-size: 13px;
  line-height: 1.7;
  color: var(--j-ink-soft);
}

/* ============ 使用技术：清单卡 ============ */
.tech-section {
  position: relative;
  max-width: 560px;
  margin: 60px auto 0;
  padding: 30px 32px 24px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
}

.tech-tape {
  top: -10px;
  left: 26px;
  transform: rotate(-5deg);
}

.tech-section ul {
  margin: 0;
  padding: 0;
  list-style: none;
}

.tech-section li {
  position: relative;
  padding: 10px 0 10px 30px;
  border-bottom: 1px dashed var(--j-rule);
  font-size: 15px;
  color: var(--j-ink);
}

.tech-section li:last-child {
  border-bottom: none;
}

/* 打勾的方框 */
.tech-section li::before {
  content: "✓";
  position: absolute;
  left: 0;
  top: 50%;
  width: 16px;
  height: 16px;
  margin-top: -9px;
  border: 1px solid var(--j-ink-soft);
  font-family: var(--j-hand);
  font-size: 15px;
  line-height: 14px;
  text-align: center;
  color: var(--j-stamp);
}

@media (max-width: 860px) {
  .features {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .about-wrapper {
    padding: 28px 14px 56px;
  }

  .hero h1 {
    font-size: 27px;
  }

  .site-owner-message {
    padding: 34px 18px 34px 34px;
    transform: none;
  }

  .site-owner-message::before {
    left: 22px;
  }

  .stat-card {
    width: 104px;
    height: 104px;
  }

  .stat-card h3 {
    font-size: 24px;
  }

  .features {
    grid-template-columns: 1fr;
  }

  .feature-card {
    transform: none !important;
  }
}
</style>
