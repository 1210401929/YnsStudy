<!--欢迎页-->
<template>
  <transition name="fade">
    <div v-if="visible && !isSelf" class="welcome-overlay">
      <div class="welcome-content">
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
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(12px);
  z-index: 9999;
  display: flex;
  justify-content: center;
  align-items: center;
}

.welcome-content {
  text-align: center;
  color: #ffffff;
  animation: slideUp 0.6s ease cubic-bezier(0.2, 0.8, 0.2, 1);
}

.welcome-avatar {
  width: 100px !important;
  height: 100px !important;
  font-size: 32px;
  border: 4px solid rgba(255, 255, 255, 0.3);
  margin-bottom: 20px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.welcome-title {
  font-size: 28px;
  font-weight: bold;
  margin: 0 0 12px 0;
  letter-spacing: 1px;
}

.welcome-desc {
  font-size: 16px;
  color: rgba(255, 255, 255, 0.85);
  margin-bottom: 30px;
}

.enter-btn {
  font-size: 16px;
  padding: 12px 36px;
  box-shadow: 0 4px 15px rgba(64, 158, 255, 0.4);
  transition: transform 0.2s;
}

.enter-btn:hover {
  transform: scale(1.05);
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
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>