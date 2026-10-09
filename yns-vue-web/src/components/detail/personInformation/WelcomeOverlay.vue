<!--欢迎页-->
<template>
  <transition name="fade">
    <div v-if="visible && !isSelf" class="welcome-overlay">
      <div class="welcome-content">
        <span class="j-tape j-tape--top"></span>
        <el-avatar
            :src="user.avatar"
            size="large"
            class="welcome-avatar"
        >
          {{ user.name?.charAt(0) }}
        </el-avatar>
        <h2 class="welcome-title">
          欢迎来到
          <span class="random-name" :style="{ color: randomNameColor }">
            {{ user.name || '' }}
          </span>
          的主页
        </h2>
        <p class="welcome-desc">个性签名：{{ user.remark }}</p>
        <el-button type="primary" size="large" round @click="handleEnter" class="enter-btn">
          进入主页
        </el-button>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  visible: {
    type: Boolean,
    default: true
  },
  isSelf: {
    type: Boolean,
    default: false
  },
  user: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['enter'])

// 生成随机的亮色 (高饱和度，中高亮度)
const getRandomBrightColor = () => {
  const h = Math.floor(Math.random() * 360)
  const s = Math.floor(Math.random() * 30) + 70
  const l = Math.floor(Math.random() * 20) + 65
  return `hsl(${h}, ${s}%, ${l}%)`
}

// 组件自己维护名字的颜色，不占用父组件的逻辑
const randomNameColor = ref(getRandomBrightColor())

const handleEnter = () => {
  // 向父组件广播：用户点击了进入主页！
  emit('enter')
}
</script>

<style scoped>
.welcome-overlay {
  position: fixed;
  top: 0;
  left: 0;
  z-index: 9999;
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100vw;
  height: 100vh;
  background: rgba(60, 50, 30, 0.35);
  backdrop-filter: blur(6px);
}

/* 欢迎卡：一张贴着胶带的明信片 */
.welcome-content {
  position: relative;
  width: min(420px, calc(100vw - 40px));
  padding: 40px 32px 30px;
  box-sizing: border-box;
  text-align: center;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: 0 30px 60px -24px rgba(40, 32, 20, 0.6);
  color: var(--j-ink);
  transform: rotate(-1deg);
  animation: slideUp 0.6s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.welcome-avatar {
  width: 96px !important;
  height: 96px !important;
  margin-bottom: 18px;
  border: 5px solid #fff;
  border-bottom-width: 14px;
  border-radius: 0 !important;
  box-shadow: 0 1px 3px rgba(60, 50, 30, 0.3);
  background: #ece4d3;
  color: var(--j-ink-soft);
  font-family: var(--j-hand);
  font-size: 32px;
  transform: rotate(3deg);
}

.welcome-title {
  margin: 0 0 12px;
  font-family: var(--j-hand);
  font-weight: normal;
  font-size: 26px;
  line-height: 1.5;
}

.random-name {
  padding: 0 4px;
  background-image: linear-gradient(transparent 58%, var(--j-highlight) 58%, var(--j-highlight) 92%, transparent 92%);
}

.welcome-desc {
  margin: 0 0 26px;
  padding-top: 12px;
  border-top: 1px dashed var(--j-rule-strong);
  font-size: 14px;
  color: var(--j-ink-soft);
}

.enter-btn {
  padding: 12px 36px;
  font-size: 16px;
  transition: transform 0.2s;
}

.enter-btn:hover {
  transform: translateY(-2px);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.5s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@keyframes slideUp {
  from { opacity: 0; transform: rotate(-1deg) translateY(30px); }
  to { opacity: 1; transform: rotate(-1deg) translateY(0); }
}
</style>
