<template>
  <!-- 公告横幅 -->
  <Announcement
      v-for="al in topAlert"
      :key="al.GUID"
      :TEXT="al.TEXT"
      :URL="al.URL"
      :URLNAME="al.URLNAME"
  />

  <!-- 文字内容来自后台 getAboutPage（与 Go 输出的首屏同一份，见 about_service.go） -->
  <div class="about-wrapper">
    <div class="about-container">
      <!-- 顶部欢迎区 -->
      <div class="hero">
        <h1>{{ content.title || '欢迎来到 YnsStudy' }}</h1>
        <p class="hero-description">{{ content.intro }}</p>

        <!-- 站长卡片 -->
        <section v-if="owner.name" class="owner-card" aria-label="站长">
          <el-avatar :src="owner.avatar" :size="72" class="owner-avatar">{{ owner.name.charAt(0) }}</el-avatar>
          <div class="owner-body">
            <div class="owner-name">
              {{ owner.name }}<span v-if="owner.alternateNames?.length" class="owner-alt">{{ owner.alternateNames[0] }}</span>
              <span class="owner-badge">站长</span>
            </div>
            <p v-if="owner.remark" class="owner-remark">{{ owner.remark }}</p>
            <div class="owner-links">
              <router-link v-if="owner.profilePath" :to="owner.profilePath">站长主页</router-link>
              <router-link :to="{ path: '/ynsStudy/Notes', query: { tab: 'guestbook' } }">给站长留言</router-link>
              <a v-if="content.contactEmail" :href="`mailto:${content.contactEmail}`">{{ content.contactEmail }}</a>
              <a v-for="link in owner.sameAs" :key="link" :href="link" target="_blank" rel="noopener me">{{ linkLabel(link) }}</a>
            </div>
          </div>
        </section>

        <!-- 站长想说 -->
        <section v-if="content.ownerMessage?.length" class="site-owner-message">
          <span class="j-tape j-tape--pink letter-tape"></span>
          <div class="owner-title">站长想说</div>
          <p v-for="(paragraph, index) in content.ownerMessage" :key="index">{{ paragraph }}</p>
        </section>
      </div>

      <!-- 数据统计 -->
      <div class="stats-section">
        <h2>数据统计</h2>
        <div class="stats">
          <div v-for="item in statItems" :key="item.key" class="stat-card" :title="item.title || item.label">
            <h3>{{ formatCount(stats[item.key]) }}</h3>
            <p>{{ item.label }}</p>
          </div>
        </div>
      </div>

      <!-- 功能 -->
      <div class="features">
        <component
            :is="feature.link ? 'router-link' : 'div'"
            v-for="feature in content.features"
            :key="feature.title"
            :to="feature.link || undefined"
            class="feature-card"
        >
          <div class="icon">{{ feature.icon }}</div>
          <h3>{{ feature.title }}</h3>
          <p>{{ feature.text }}</p>
        </component>
      </div>

      <!-- 建站历程 -->
      <section v-if="content.timeline?.length" class="timeline-section">
        <h2>建站历程</h2>
        <ol class="timeline">
          <li v-for="item in content.timeline" :key="item.date + item.title" class="timeline-item">
            <time class="timeline-date">{{ item.date }}</time>
            <div class="timeline-body">
              <h3>{{ item.title }}</h3>
              <p>{{ item.text }}</p>
            </div>
          </li>
        </ol>
        <router-link v-if="content.changelogPath" :to="content.changelogPath" class="changelog-link">查看完整更新日志 →</router-link>
      </section>

      <!-- 常见问题 -->
      <section v-if="content.faq?.length" class="faq-section">
        <h2>常见问题</h2>
        <details v-for="item in content.faq" :key="item.title" class="faq-item">
          <summary>{{ item.title }}</summary>
          <p>
            {{ item.text }}
            <router-link v-if="item.link" :to="item.link" class="faq-link">去看看 →</router-link>
          </p>
        </details>
      </section>

      <!-- 使用技术 -->
      <div v-if="content.techStack?.length" class="tech-section">
        <span class="j-tape j-tape--green tech-tape"></span>
        <h2>使用技术</h2>
        <ul>
          <li v-for="item in content.techStack" :key="item.title">{{ item.title }}：{{ item.text }}</li>
        </ul>
        <p v-if="content.sourceUrl" class="tech-source">
          网站源码：<a :href="content.sourceUrl" target="_blank" rel="noopener">{{ linkLabel(content.sourceUrl) }}</a>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import {ref} from 'vue'
import {sendAxiosRequest} from "@/utils/common.js"
import Announcement from "@/components/detail/Announcement.vue"
import {getAnnouncementByRouterName} from "@/utils/blogUtil.js"
import {absoluteUrl, ownerPerson, SITE_NAME, SITE_OWNER_DISPLAY, useSeo} from "@/utils/seo.js"

const content = ref({})
const owner = ref({})
const stats = ref({})

// 统计项：随记是便签墙的便签；登录次数是登录用户的累计登录次数
const statItems = [
  {key: 'ARTICLENUM', label: '文章'},
  {key: 'COMMUNITYNUM', label: '随记', title: '便签墙上的便签数'},
  {key: 'SERIESNUM', label: '系列', title: '有公开文章的系列数'},
  {key: 'TAGNUM', label: '标签', title: '公开文章用到的标签数'},
  {key: 'VIEW_PAGE', label: '阅读'},
  {key: 'USERNUM', label: '用户'},
  {key: 'USERLOGINNUM', label: '登录次数', title: '注册用户的累计登录次数'}
]

const formatCount = (value) => {
  const number = Number(value) || 0
  if (number >= 10000) return `${(number / 10000).toFixed(number >= 100000 ? 0 : 1).replace(/\.0$/, '')}万`
  return String(number)
}

// 外部链接显示成“github.com/xxx”这种短形式
const linkLabel = (link) => String(link || '').replace(/^https?:\/\//, '').replace(/\/$/, '')

const loadAboutPage = async () => {
  const result = await sendAxiosRequest("/blog-api/home/getAboutPage").catch(() => null)
  if (result && !result.isError && result.result) {
    content.value = result.result.content || {}
    owner.value = result.result.owner || {}
    stats.value = result.result.stats || {}
  }
}
loadAboutPage()

const topAlert = ref([])
const setTopAlert = async () => {
  topAlert.value = await getAnnouncementByRouterName("About")
}
setTopAlert()

// 与 Go 端 renderSEOAboutHTML 一致的标题、描述和结构化数据
useSeo(() => {
  const intro = content.value.intro || '一个集文章创作、评论互动与资源分享于一体的知识社区'
  const description = `${SITE_NAME}：${intro}。站长 ${SITE_OWNER_DISPLAY}，介绍网站的由来、功能、建站历程和常见问题。`.slice(0, 180)
  const person = ownerPerson()
  if (owner.value.avatar) person.image = absoluteUrl(owner.value.avatar)
  return {
    title: `关于本站 - ${SITE_NAME}`,
    description,
    path: '/ynsStudy/About',
    type: 'profile',
    jsonLd: {
      '@context': 'https://schema.org',
      '@type': 'AboutPage',
      name: `关于本站 - ${SITE_NAME}`,
      url: absoluteUrl('/ynsStudy/About'),
      description,
      inLanguage: 'zh-CN',
      mainEntity: person
    }
  }
})
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
  grid-template-columns: repeat(3, minmax(0, 1fr));
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

a.feature-card {
  display: block;
  color: inherit;
  text-decoration: none;
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

/* ============ 站长卡片：一张名片 ============ */
.owner-card {
  display: flex;
  align-items: center;
  gap: 20px;
  max-width: 600px;
  margin: 32px auto 0;
  padding: 20px 24px;
  text-align: left;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
  transform: rotate(0.4deg);
}

.owner-avatar {
  flex-shrink: 0;
  border: 3px solid #fff;
  box-shadow: 0 2px 6px rgba(60, 50, 30, 0.2);
}

.owner-body {
  min-width: 0;
}

.owner-name {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 8px;
  font-family: var(--j-hand);
  font-size: 22px;
}

.owner-alt {
  font-family: var(--j-sans);
  font-size: 13px;
  color: var(--j-muted);
}

.owner-badge {
  padding: 0 6px;
  border: 1.5px solid var(--j-stamp);
  color: var(--j-stamp);
  font-family: var(--j-sans);
  font-size: 12px;
  line-height: 1.6;
  transform: rotate(-4deg);
}

.owner-remark {
  margin: 4px 0 0;
  font-size: 14px;
  color: var(--j-ink-soft);
}

.owner-links {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 16px;
  margin-top: 8px;
  font-size: 13px;
}

.owner-links a {
  color: var(--j-pen);
  text-decoration: none;
}

.owner-links a:hover {
  text-decoration: underline;
}

/* ============ 建站历程：日记里的时间线 ============ */
.timeline-section,
.faq-section {
  max-width: 720px;
  margin: 60px auto 0;
}

.timeline-section h2,
.faq-section h2 {
  margin: 0 0 22px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 24px;
  text-align: center;
}

.timeline {
  position: relative;
  margin: 0;
  padding: 0 0 0 28px;
  list-style: none;
}

.timeline::before {
  content: "";
  position: absolute;
  top: 6px;
  bottom: 6px;
  left: 8px;
  border-left: 2px dashed var(--j-rule-strong);
}

.timeline-item {
  position: relative;
  display: flex;
  gap: 16px;
  padding: 0 0 22px;
}

.timeline-item::before {
  content: "";
  position: absolute;
  top: 6px;
  left: -25px;
  width: 12px;
  height: 12px;
  border: 2px solid var(--j-stamp);
  border-radius: 50%;
  background: var(--j-paper);
}

.changelog-link {
  display: block;
  margin-top: 4px;
  text-align: center;
  font-size: 14px;
  color: var(--j-pen);
  text-decoration: none;
}

.changelog-link:hover {
  text-decoration: underline;
}

.timeline-date {
  flex-shrink: 0;
  width: 96px;
  white-space: nowrap;
  font-family: var(--j-num);
  color: var(--j-stamp);
}

.timeline-body h3 {
  margin: 0 0 4px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 18px;
}

.timeline-body p {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
  color: var(--j-ink-soft);
}

/* ============ 常见问题：可以展开的便签条 ============ */
.faq-item {
  margin-bottom: 10px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.07);
}

.faq-item summary {
  padding: 12px 16px;
  font-size: 15px;
  color: var(--j-ink);
  cursor: pointer;
  list-style-position: inside;
}

.faq-item[open] summary {
  border-bottom: 1px dashed var(--j-rule);
}

.faq-item p {
  margin: 0;
  padding: 12px 16px 14px;
  font-size: 14px;
  line-height: 1.8;
  color: var(--j-ink-soft);
}

.faq-link {
  margin-left: 6px;
  color: var(--j-pen);
  text-decoration: none;
}

.tech-source {
  margin: 14px 0 0;
  font-size: 13px;
  color: var(--j-muted);
}

.tech-source a {
  color: var(--j-pen);
  text-decoration: none;
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

  .owner-card {
    flex-direction: column;
    text-align: center;
    transform: none;
  }

  .owner-name,
  .owner-links {
    justify-content: center;
  }

  .timeline-item {
    flex-direction: column;
    gap: 2px;
  }
}
</style>
