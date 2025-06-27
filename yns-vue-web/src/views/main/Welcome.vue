<template>
  <div class="container" :class="currentTheme">
    <!-- 顶部导航栏 -->
    <header class="navbar glassy">
      <div class="logo toDefault">YnsStudy</div>
      <ul class="menu">
        <li @click="menuClick(item)" v-for="item in menuItems" :key="item.router">{{ item.name }}</li>
      </ul>
      <div class="login toHand">
        <LoginDialog/>
      </div>
    </header>

    <!-- 中间内容 -->
    <main class="main-content">
      <div class="brand toDefault">🌴 YnsStudy</div>
      <h2 class="toDefault">少一点迷茫，多一点引导，学习之路不再孤单</h2>
      <p class="toDefault">永远相信美好的事情即将发生</p>
      <div class="buttons">
        <el-button type="success" @click="buttonClick('ai')" size="large" round>AI</el-button>
        <el-button type="primary" @click="buttonClick('article')" size="large" round>文章</el-button>
        <el-button type="danger" @click="buttonClick('aboutWe')" size="large" round>关于我们</el-button>
      </div>
    </main>
    <!-- 泡泡背景 -->
    <div class="bubbles">
      <span v-for="n in 20" :key="n" class="bubble"></span>
    </div>
    <!-- 波浪背景 -->
    <div class="wave-box">
      <!-- 第1层波浪 -->
      <svg class="wave-svg" viewBox="0 0 1200 200" preserveAspectRatio="none">
        <path
            d="M0,100 C300,200 900,0 1200,100 L1200,200 L0,200 Z"
            fill="rgba(255,255,255,0.5)"
        >
          <animate
              attributeName="d"
              dur="8s"
              repeatCount="indefinite"
              values="
              M0,100 C300,200 900,0 1200,100 L1200,200 L0,200 Z;
              M0,100 C300,0 900,200 1200,100 L1200,200 L0,200 Z;
              M0,100 C300,200 900,0 1200,100 L1200,200 L0,200 Z"
          />
        </path>
      </svg>

      <!-- 第2层波浪 -->
      <svg class="wave-svg" viewBox="0 0 1200 200" preserveAspectRatio="none" style="bottom: -20px">
        <path
            d="M0,100 C400,150 800,50 1200,100 L1200,200 L0,200 Z"
            fill="rgba(255,255,255,0.3)"
        >
          <animate
              attributeName="d"
              dur="12s"
              repeatCount="indefinite"
              values="
              M0,100 C400,150 800,50 1200,100 L1200,200 L0,200 Z;
              M0,100 C400,50 800,150 1200,100 L1200,200 L0,200 Z;
              M0,100 C400,150 800,50 1200,100 L1200,200 L0,200 Z"
          />
        </path>
      </svg>
    </div>
  </div>
</template>

<script setup>
import {ref, onMounted, onBeforeUnmount} from 'vue'
import {useRouter} from "vue-router";
import {ElMessage} from 'element-plus';
import LoginDialog from "@/components/main/LoginDialog.vue";
import * as menuUtil from "@/utils/menu.js"
import YnsStudyAi from "@/views/detail/ai/YnsStudyAi.vue";
import {encrypt} from "@/utils/common.js";
//Router 实例
const router = useRouter();
//class样式
const themes = ['theme-sky', 'theme-ocean', 'theme-purple']
const currentTheme = ref(themes[0])
const menuItems = ref(menuUtil.getMenuItems());

//菜单点击
function menuClick(menu) {
  router.push({name: menu.router});
}

//按钮点击
function buttonClick(typeName) {
  switch (typeName) {
      //导航
    case "ai": {
      const routeUrl = router.resolve({name: 'YnsStudyAi'}).href;
      window.open(routeUrl, "YnsStudyAi");
      break;
    }
      //文章
    case "article": {
      router.push({name: "Community"});
      break;
    }
      //关于我们
    case "aboutWe": {
      router.push({name: "About"});
      break;
    }
  }
}

//切换背景
const switchTheme = () => {
  const index = themes.indexOf(currentTheme.value)
  currentTheme.value = themes[(index + 1) % themes.length]
}
//监听键盘
const handleKey = (e) => {
  if (e && e.key && e.key.toLowerCase() === 'b') {
    switchTheme()
  }
}
//添加监听
onMounted(() => {
  window.addEventListener('keydown', handleKey)
})
onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKey)
})
</script>

<style scoped>

.toDefault {
  cursor: default;
}

.toHand {
  cursor: pointer;
}

.theme-sky {
  background: linear-gradient(to bottom, #38bdf8, #0ea5e9);
}

.theme-ocean {
  background: linear-gradient(to bottom, #0f2027, #203a43, #2c5364);
}

.theme-purple {
  background: linear-gradient(to bottom, #8e2de2, #4a00e0);
}

.container {
  min-height: 100vh;
  font-family: "Segoe UI", sans-serif;
  color: white;
  overflow: hidden;
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 5vw;
  transition: background 0.6s ease;
}

/* 玻璃拟态 */
.glassy {
  background: rgba(255, 255, 255, 0.1);
  box-shadow: 0 4px 30px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border-radius: 10px;
}

.navbar {
  width: 100%;
  max-width: 1200px;
  padding: 20px 30px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  z-index: 10;
  margin-top: 20px;
}

.logo {
  font-size: 1.8rem;
  font-weight: bold;
  font-style: italic;
}

.menu {
  display: flex;
  list-style: none;
  gap: 20px;
  padding: 0;
  margin: 10px 0;
}

.menu li {
  cursor: pointer;
  transition: color 0.2s;
}

.menu li:hover {
  color: #ffd700;
}

.login {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 1rem;
}

.main-content {
  text-align: center;
  padding: 60px 10px 160px;
  z-index: 2;
  position: relative;
  width: 100%;
  max-width: 900px;
}

.brand {
  font-size: 2.4rem;
  font-weight: bold;
  margin-bottom: 20px;
}

h2 {
  font-size: 1.2rem;
  margin-bottom: 10px;
  font-weight: normal;
}

p {
  font-size: 1rem;
  color: #d9f1ff;
  margin-bottom: 40px;
}

.buttons {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 20px;
}

/* 波浪样式 */
.wave-box {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 600px;
  z-index: 1;
  overflow: hidden;
}

.wave-svg {
  position: absolute;
  width: 100%;
  height: 100%;
  display: block;
}

/* 泡泡样式 */
.bubbles {
  position: absolute;
  width: 100%;
  height: 100%;
  top: 0;
  left: 0;
  z-index: 0;
  overflow: hidden;
}

.bubble {
  position: absolute;
  bottom: -50px;
  width: 20px;
  height: 20px;
  background-color: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  animation: floatUp 15s infinite ease-in;

}

.bubble:nth-child(odd) {
  width: 12px;
  height: 12px;
  background-color: rgba(255, 255, 255, 0.15);
}

@keyframes floatUp {
  0% {
    transform: translateY(0);
    opacity: 0.3;
  }
  50% {
    opacity: 1;
  }
  100% {
    transform: translateY(-120vh);
    opacity: 0;
  }
}

@media (max-width: 768px) {
  .menu {
    justify-content: center;
    flex-wrap: wrap;
  }

  .navbar {
    flex-direction: column;
    align-items: center;
  }

  .brand {
    font-size: 2rem;
  }

  h2 {
    font-size: 1rem;
  }

  p {
    font-size: 0.9rem;
  }

  .buttons {
    flex-direction: column;
  }

  .buttons .el-button {
    margin-left: 12px;
  }

  .login {
    margin-top: 10px;
  }
}
</style>