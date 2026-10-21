<!-- 便签墙的留言板：访客不登录也能留言，留言下面可以回复（一层）；站长和管理员的留言盖对应的章 -->
<template>
  <div class="guestbook">
    <!-- 写留言：一张横线信纸 -->
    <form class="write-card" @submit.prevent="submitMessage">
      <span class="j-tape j-tape--top"></span>
      <p class="write-title">路过的朋友，留张便签再走吧 ✎</p>
      <div v-if="!isLoggedIn" class="guest-fields">
        <el-input v-model="guestInfo.nickname" name="author" autocomplete="name" :prefix-icon="User" placeholder="昵称（选填）" clearable/>
        <el-input v-model="guestInfo.email" name="mail" type="email" autocomplete="email" :prefix-icon="Message" placeholder="邮箱（选填，有人回复时通知你，不会公开）" clearable/>
        <el-input v-model="guestInfo.website" v-url-hint="guestInfo.website" name="url" type="url" autocomplete="on" :prefix-icon="Link" placeholder="网址（选填）http://" clearable/>
      </div>
      <p v-else class="write-as">以 <b>{{ userStore.userBean.name }}</b> 的身份留言</p>
      <el-input
          v-model="newMessage"
          type="textarea"
          :autosize="{ minRows: 3, maxRows: 10 }"
          maxlength="1000"
          show-word-limit
          placeholder="想说点什么？建议、问候、交换友链都可以…"
          class="lined-paper"
          @keyup.enter.ctrl="submitMessage"
      />
      <div class="write-actions">
        <span class="hint-text">Ctrl + Enter 发送</span>
        <el-button type="primary" native-type="submit" :loading="sending">
          <el-icon class="el-icon--left"><EditPen/></el-icon>贴上留言
        </el-button>
      </div>
    </form>

    <p class="board-count">共 {{ total }} 条留言</p>

    <div class="message-list" :class="{ 'is-refreshing': loading && messages.length }">
      <article v-for="(message, index) in messages" :key="message.GUID" class="message-card">
        <span v-if="message.ISTOP === '1'" class="note-pin" title="置顶"></span>
        <span v-else class="j-tape message-tape" :class="[tapeClasses[index % tapeClasses.length], 'tape-' + (index % 3)]"></span>

        <header class="message-head">
          <el-avatar :src="message.AVATAR" :size="32" class="message-avatar" @click="openPerson(message)"
                     :style="!message.AVATAR ? getAvatarStyle(message.USERNAME) : {}">
            {{ message.USERNAME?.charAt(0) }}
          </el-avatar>
          <div class="message-who">
            <span class="message-name" :class="{ 'has-link': hasLink(message) }" @click="openPerson(message)">{{ message.USERNAME || '访客' }}</span>
            <span v-if="roleStamp(message)" class="j-stamp owner-stamp" :class="{ 'j-stamp--admin': roleStamp(message) === '管理员' }">{{ roleStamp(message) }}</span>
            <time class="message-time">{{ pubFormatDate(message.CREATE_TIME) }}</time>
          </div>
          <div class="message-tools">
            <button type="button" class="tool-link" @click="toggleReply(message, message)">回复</button>
            <button v-if="isSuperAdmin" type="button" class="tool-link" @click="toggleTop(message)">{{ message.ISTOP === '1' ? '取消置顶' : '置顶' }}</button>
            <button v-if="canDelete(message)" type="button" class="tool-link danger" @click="removeMessage(message)">删除</button>
          </div>
        </header>

        <p class="message-text">{{ message.TEXT }}</p>

        <div v-if="message.children?.length" class="reply-list">
          <div v-for="reply in message.children" :key="reply.GUID" class="reply-item">
            <span class="reply-name" :class="{ 'has-link': hasLink(reply) }" @click="openPerson(reply)">{{ reply.USERNAME || '访客' }}</span>
            <span v-if="roleStamp(reply)" class="j-stamp owner-stamp" :class="{ 'j-stamp--admin': roleStamp(reply) === '管理员' }">{{ roleStamp(reply) }}</span>
            <span v-if="reply.RECEIVE_USERNAME && reply.RECEIVE_USERNAME !== message.USERNAME" class="reply-to">
              回复 <b>@{{ reply.RECEIVE_USERNAME }}</b>
            </span>：
            <span class="reply-text">{{ reply.TEXT }}</span>
            <span class="reply-meta">
              <time>{{ pubFormatDate(reply.CREATE_TIME) }}</time>
              <button type="button" class="tool-link" @click="toggleReply(message, reply)">回复</button>
              <button v-if="canDelete(reply)" type="button" class="tool-link danger" @click="removeMessage(reply, message)">删除</button>
            </span>
          </div>
        </div>

        <!-- 回复框：用真正的表单，浏览器才会记住昵称、邮箱、网址 -->
        <form v-if="replyFor === message.GUID" class="reply-form" @submit.prevent="submitReply(message)">
          <div v-if="!isLoggedIn" class="guest-fields small">
            <el-input v-model="guestInfo.nickname" name="author" autocomplete="name" size="small" placeholder="昵称（选填）" clearable/>
            <el-input v-model="guestInfo.email" name="mail" type="email" autocomplete="email" size="small" placeholder="邮箱（选填）" clearable/>
          </div>
          <div class="reply-row">
            <el-input v-model="replyText" size="small" maxlength="1000" :placeholder="`回复 @${replyTarget?.USERNAME || '访客'}：`" clearable/>
            <el-button size="small" type="primary" native-type="submit" :loading="sending">发送</el-button>
          </div>
        </form>
      </article>
    </div>

    <p v-if="!loading && messages.length === 0" class="board-empty">还没有人留言，来贴第一张吧</p>
    <div class="board-more">
      <span v-if="loading" class="more-text">正在翻找…</span>
      <button v-else-if="!noMore" type="button" class="j-button" @click="load()">看更早的留言</button>
      <span v-else-if="messages.length" class="more-text">留言都在这里了</span>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { EditPen, Link, Message, User } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/main/user.js'
import { adminUserCode } from '@/config/vue-config.js'
import { ele_confirm, getCurrentUserAdminObject, getGuid, pubFormatDate, sendAxiosRequest, sendAxiosRequestChecked } from '@/utils/common.js'
import { getAvatarStyle } from '@/utils/avatarStyle.js'
import { pubOpenUser } from '@/utils/blogUtil.js'
import '@/css/notes.css'

const PAGE_SIZE = 10
const router = useRouter()
const userStore = useUserStore()
const isLoggedIn = computed(() => !!userStore.userBean?.code)
const isSuperAdmin = computed(() => getCurrentUserAdminObject().adminLevel === 'superAdmin')
const tapeClasses = ['', 'j-tape--green', 'j-tape--blue', 'j-tape--pink']

// 网址格式不对时给出明确提示（与评论区相同）
const URL_HINT = '网址需要以 http:// 或 https:// 开头，例如 https://example.com'
const urlInputOf = (el) => (el.tagName === 'INPUT' ? el : el.querySelector('input'))
const vUrlHint = {
  mounted(el) {
    const input = urlInputOf(el)
    if (!input) return
    input.addEventListener('invalid', () => {
      if (input.validity.typeMismatch) input.setCustomValidity(URL_HINT)
    })
  },
  updated(el, binding) {
    if (binding.value !== binding.oldValue) urlInputOf(el)?.setCustomValidity('')
  }
}

// 访客上次填写的昵称、邮箱、网址，和文章评论共用
const loadGuest = () => {
  try {
    const cached = localStorage.getItem('blog_guest_info')
    if (cached) return { nickname: '', email: '', website: '', ...JSON.parse(cached) }
  } catch (e) {}
  return { nickname: '', email: '', website: '' }
}
const guestInfo = ref(loadGuest())

const messages = ref([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const noMore = ref(false)
const sending = ref(false)
const newMessage = ref('')
const replyFor = ref('')
const replyTarget = ref(null)
const replyText = ref('')

const load = async () => {
  if (loading.value || noMore.value) return
  loading.value = true
  try {
    const res = await sendAxiosRequest('/blog-api/guestbook/getMessages', { page: page.value, pageSize: PAGE_SIZE })
    if (!res || res.isError) {
      ElMessage.error(res?.errMsg || '获取留言失败')
      return
    }
    const rows = res.result?.data || []
    messages.value.push(...rows.map(row => ({ ...row, children: row.children || [] })))
    total.value = res.result?.total || 0
    noMore.value = rows.length < PAGE_SIZE
    page.value++
  } finally {
    loading.value = false
  }
}
load()

// 给访客看的身份：超级管理员账号就是站长本人，盖“站长”；管理员盖“管理员”（USERROLE 由后端返回）
const roleStamp = (item) => {
  if (!item.USERCODE) return ''
  if (item.USERCODE === adminUserCode) return '站长'
  return item.USERROLE === 'admin' ? '管理员' : ''
}
const hasLink = (item) => /^https?:\/\//i.test(String(item.USERWEBSITE || '')) || !!item.USERCODE

// 访客留了网址的打开网址（只接受 http/https），注册用户打开个人主页
function openPerson(item) {
  const website = String(item.USERWEBSITE || '').trim()
  if (/^https?:\/\//i.test(website)) window.open(website, '_blank', 'noopener')
  else if (item.USERCODE) pubOpenUser(router, item.USERCODE)
}

// 删除权限：本人、管理员（不能删超级管理员的）、超级管理员；匿名留言只有管理员能删
function canDelete(item) {
  const admin = getCurrentUserAdminObject()
  if (admin.adminLevel === 'superAdmin') return true
  if (userStore.userBean?.code && item.USERCODE === userStore.userBean.code) return true
  return !!admin.isAdmin && item.USERCODE !== adminUserCode
}

function commenter() {
  const user = userStore.userBean
  if (user?.code) return { USERNAME: user.name, USERCODE: user.code, USERROLE: user.role || '', AVATAR: user.avatar || '', USEREMAIL: '', USERWEBSITE: '' }
  localStorage.setItem('blog_guest_info', JSON.stringify(guestInfo.value))
  return {
    USERNAME: guestInfo.value.nickname.trim() || '访客',
    USERCODE: '',
    AVATAR: '',
    USEREMAIL: guestInfo.value.email.trim(),
    USERWEBSITE: guestInfo.value.website.trim()
  }
}

async function post(message, extra = {}) {
  const payload = { ...message, ...extra }
  delete payload.AVATAR
  delete payload.CREATE_TIME
  delete payload.children
  delete payload.USERCODE
  delete payload.USERROLE
  return sendAxiosRequest('/blog-api/guestbook/addMessage', { message: payload })
}

async function submitMessage() {
  const text = newMessage.value.trim()
  if (!text) {
    ElMessage.warning('写点什么再贴吧')
    return
  }
  if (sending.value) return
  sending.value = true
  try {
    const message = { GUID: getGuid(), TEXT: text, ...commenter() }
    const res = await post(message)
    if (!res || res.isError) {
      ElMessage.error(res?.errMsg || '留言失败')
      return
    }
    messages.value.unshift({ ...message, CREATE_TIME: '刚刚', children: [] })
    total.value++
    newMessage.value = ''
    ElMessage.success('留言已贴上，谢谢你来过')
  } finally {
    sending.value = false
  }
}

function toggleReply(message, target) {
  if (replyFor.value === message.GUID && replyTarget.value?.GUID === target.GUID) {
    replyFor.value = ''
    replyTarget.value = null
    return
  }
  replyFor.value = message.GUID
  replyTarget.value = target
  replyText.value = ''
}

async function submitReply(message) {
  const text = replyText.value.trim()
  if (!text || sending.value) return
  sending.value = true
  try {
    const target = replyTarget.value || message
    const reply = {
      GUID: getGuid(), SUPERGUID: message.GUID, TEXT: text,
      RECEIVE_USERCODE: target.USERCODE || '', RECEIVE_USERNAME: target.USERNAME || '访客',
      ...commenter()
    }
    // 告诉后台回复的是哪一条，用于通知对方（匿名访客靠留下的邮箱）
    const res = await post(reply, { REPLY_TO: target.GUID })
    if (!res || res.isError) {
      ElMessage.error(res?.errMsg || '回复失败')
      return
    }
    message.children.push({ ...reply, USERCODE: userStore.userBean?.code || '', CREATE_TIME: '刚刚' })
    replyFor.value = ''
    replyTarget.value = null
    replyText.value = ''
    ElMessage.success('回复成功')
  } finally {
    sending.value = false
  }
}

function removeMessage(item, parent = null) {
  const tip = parent ? '确定删除这条回复吗？' : '确定撕下这条留言吗？它下面的回复会一起删除。'
  ele_confirm(tip, async () => {
    const result = await sendAxiosRequestChecked('/blog-api/guestbook/deleteMessage', { guid: item.GUID }, '删除失败')
    if (!result) return
    if (parent) {
      parent.children = parent.children.filter(reply => reply.GUID !== item.GUID)
    } else {
      messages.value = messages.value.filter(message => message.GUID !== item.GUID)
      total.value = Math.max(0, total.value - 1)
    }
    ElMessage.success('已删除')
  })
}

async function toggleTop(message) {
  const isTop = message.ISTOP === '1' ? '0' : '1'
  const result = await sendAxiosRequestChecked('/blog-api/guestbook/setTopMessage', { guid: message.GUID, isTop }, '置顶操作失败')
  if (!result) return
  message.ISTOP = isTop === '1' ? '1' : null
  ElMessage.success(isTop === '1' ? '已置顶，刷新后排到最前' : '已取消置顶')
}
</script>

<style scoped>
.guestbook {
  max-width: 780px;
  margin: 0 auto;
  min-height: 70vh;
}

/* ---------- 写留言 ---------- */
.write-card {
  position: relative;
  padding: 24px 22px 16px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
}

.write-title {
  margin: 0 0 12px;
  font-family: var(--j-hand);
  font-size: 18px;
  color: var(--j-ink);
}

.write-as {
  margin: 0 0 10px;
  font-size: 13px;
  color: var(--j-muted);
}

.guest-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 10px;
}

.guest-fields > :last-child:nth-child(3) {
  grid-column: span 2;
}

.guest-fields.small {
  margin-bottom: 8px;
}

/* 横线信纸，横线跟着文字一起滚动 */
.lined-paper :deep(.el-textarea__inner) {
  padding-top: 4px;
  line-height: 26px;
  background-color: #fff;
  background-image: repeating-linear-gradient(transparent 0 25px, var(--j-rule) 25px 26px);
  background-position: 0 4px;
  background-attachment: local;
}

.write-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 12px;
}

.hint-text {
  font-size: 12px;
  color: var(--j-muted);
}

.board-count {
  margin: 26px 0 14px;
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-muted);
}

/* ---------- 留言列表：一张张略微歪着的便签 ---------- */
.message-list {
  display: flex;
  flex-direction: column;
  gap: 26px;
}

.message-list.is-refreshing {
  opacity: 0.6;
}

.message-card {
  position: relative;
  padding: 22px 20px 14px;
  background: #fffdf8;
  border: 1px solid var(--j-rule);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.06), 0 10px 20px -14px rgba(60, 50, 30, 0.35);
}

/*
 * 卡片本身不旋转：旋转后的文字和输入框不再对齐像素，看起来发虚（尤其是回复框里的小字）。
 * “贴在墙上”的感觉交给胶带：每张的胶带位置和角度略有不同。
 */
.message-tape {
  top: -10px;
  width: 70px;
}

.tape-0 { left: 28px; transform: rotate(-4deg); }
.tape-1 { left: 50%; margin-left: -35px; transform: rotate(2deg); }
.tape-2 { left: auto; right: 34px; transform: rotate(-2deg); }

.message-head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.message-avatar {
  flex-shrink: 0;
  cursor: pointer;
}

.message-who {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px 8px;
  flex: 1;
  min-width: 0;
}

.message-name,
.reply-name {
  font-weight: 600;
  color: var(--j-ink);
}

.message-name.has-link,
.reply-name.has-link {
  color: var(--j-pen);
  cursor: pointer;
}

.message-time {
  font-size: 12px;
  color: var(--j-muted);
}

.owner-stamp {
  font-size: 11px;
  line-height: 15px;
}

.message-tools {
  display: flex;
  flex-shrink: 0;
  gap: 2px;
}

.tool-link {
  padding: 2px 6px;
  border: none;
  background: none;
  font-size: 12px;
  color: var(--j-pen);
  cursor: pointer;
}

.tool-link.danger {
  color: var(--j-stamp);
}

.message-text {
  margin: 12px 0 0;
  font-size: 15px;
  line-height: 1.8;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

/* 回复：左边一道铅笔线 */
.reply-list {
  margin: 12px 0 0 6px;
  padding-left: 12px;
  border-left: 1.5px solid rgba(110, 100, 85, 0.35);
}

.reply-item {
  padding: 6px 0;
  font-size: 14px;
  line-height: 1.7;
  overflow-wrap: anywhere;
}

.reply-to {
  margin-left: 4px;
  color: var(--j-muted);
}

.reply-to b {
  color: var(--j-pen);
  font-weight: normal;
}

.reply-text {
  white-space: pre-wrap;
}

.reply-meta {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  margin-left: 8px;
  font-size: 12px;
  color: var(--j-muted);
}

.reply-form {
  margin-top: 10px;
}

.reply-row {
  display: flex;
  gap: 8px;
}

.board-empty {
  padding: 40px 0 10px;
  text-align: center;
  font-family: var(--j-hand);
  font-size: 17px;
  color: var(--j-muted);
}

.board-more {
  display: flex;
  justify-content: center;
  margin-top: 30px;
}

.more-text {
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-muted);
}

@media (max-width: 640px) {
  .guest-fields {
    grid-template-columns: 1fr;
  }

  .guest-fields > :last-child:nth-child(3) {
    grid-column: auto;
  }

  .hint-text {
    display: none;
  }

  .message-card {
    padding: 20px 14px 12px;
  }

  .message-head {
    flex-wrap: wrap;
  }
}
</style>
