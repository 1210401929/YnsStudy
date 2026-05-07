<template>
  <div id="app">
    <!-- 页面内容 -->
    <div class="page-wrapper">
      <router-view/>
    </div>

    <!-- 备案信息区域 -->
    <footer class="beian-info">
      <div class="beian-container">
        <div class="beian-text">
          <div class="footer-contact">
            <p>作者：YuNanSong | 联系方式：
              <a href="mailto:yvnansong@qq.com">yvnansong@qq.com</a> |
              <a href="https://github.com/1210401929/YnsStudy" target="_blank">访问作者 GitHub</a> |
              <a href="/rss">RSS订阅</a>
            </p>
          </div>
          <div class="footer-records">
            <a href="https://beian.miit.gov.cn/" target="_blank" rel="noreferrer">
              冀ICP备2025120557号-1
            </a>
            &nbsp;|&nbsp;
            <a href="https://beian.mps.gov.cn/#/query/webSearch?code=11010502057144" target="_blank" rel="noreferrer">
              <img src="/picture/beian.png" alt="公安备案图标" class="beian-icon"/>
              京公网安备11010502057144号
            </a>
          </div>
          <div class="beian-copy">©2025 YnsStudy. All rights reserved. 保留所有权利</div>
        </div>
      </div>
    </footer>

    <!-- 回到顶部按钮 -->
    <div class="scroll-to-top" :class="{ hidden: !showScrollButton }" @click="scrollToTop" >
      <el-icon><Top/></el-icon>
    </div>
  </div>
</template>


<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { Top } from "@element-plus/icons-vue";// 引入 Top 图标
import {getSendAxiosUrl} from "@/utils/common.js";

const showScrollButton = ref(false);
const hover = ref(false); // 用于控制鼠标移入移出时的透明度变化

// 判断是否显示回到顶部按钮
const handleScroll = () => {
  const scrollY = window.scrollY;

  // 当页面滚动超过300px时
  if (scrollY > 300) {
    showScrollButton.value = true;
  } else {
    showScrollButton.value = false;
  }
};

// 回到顶部函数
const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: 'smooth' });
};

// 监听滚动事件
onMounted(() => {
  window.addEventListener('scroll', handleScroll);
});

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll);
});
</script>




<style scoped>
html, body {
  margin: 0;
  padding: 0;
  min-height: 100vh;
  background-color: #f4f6f8;
}

#app {
  min-height: 100vh;
  display: block;
}

/* 页面内容容器 */
.page-wrapper {
  min-height: calc(100vh - 160px); /* 根据备案信息高度大致预留空间 */
}

/* 备案信息区域 */
.beian-info {
  background-color: #ffffff;
  color: #333;
  padding: 24px 0 16px;
  font-size: 13px;
  border-top: 1px solid #eee;
}

.beian-container {
  max-width: 960px;
  margin: 0 auto;
  padding: 0 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  text-align: center;
}

/* 联系方式 */
.footer-contact p {
  margin: 4px 0;
  color: #888888;
}

.footer-contact a {
  color: #409EFF;
  text-decoration: none;
}

.footer-contact a:hover {
  text-decoration: underline;
}

/* 备案链接 */
.footer-records {
  margin-top: 8px;
}

.footer-records a {
  color: #888;
  font-size: 12px;
  text-decoration: none;
}

.footer-records a:hover {
  color: #000;
  text-decoration: underline;
}

.beian-icon {
  vertical-align: middle;
  width: 16px;
  height: 16px;
  margin-right: 4px;
}

.beian-copy {
  font-size: 11px;
  color: #aaa;
  margin-top: 8px;
}

/* 回到顶部按钮 - 简约现代风 */
.scroll-to-top {
  position: fixed;
  left: 20px; /* 保持在左侧 */
  bottom: 30px; /* 距离底部稍远一点 */
  width: 48px;  /* 固定的圆形 */
  height: 48px;
  background-color: #ffffff; /* 默认白色背景 */
  color: #606266; /* 默认图标颜色 */
  border-radius: 50%; /* 完美的圆形 */
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;

  /* 柔和的阴影，使其悬浮感更强 */
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);

  /* 初始状态微透明 */
  opacity: 0.9;

  /* 过渡动画：所有属性，平滑，时间适中 */
  transition: all 0.3s cubic-bezier(0.645, 0.045, 0.355, 1);

  z-index: 999;
}

.scroll-to-top .el-icon {
  font-size: 22px; /* 调整图标大小 */
}

/* 鼠标悬停样式 */
.scroll-to-top:hover {
  opacity: 1; /* 完全不透明 */
  background-color: #ecf5ff; /* Element Plus 的浅蓝色背景 */
  color: #409EFF; /* 品牌蓝色图标 */
  box-shadow: 0 6px 16px rgba(64, 158, 255, 0.2); /* 蓝色调的阴影 */
  transform: translateY(-3px); /* 稍稍向上浮动 */
}

/* 隐藏状态（通过 JS 控制） */
.scroll-to-top.hidden {
  opacity: 0;
  pointer-events: none; /* 隐藏时不可点击 */
  transform: translateY(20px); /* 隐藏时向下移动，出场动画 */
}

/* 手机端样式 - 稍微调小一点 */
@media (max-width: 768px) {
  .scroll-to-top {
    left: 15px; /* 离边缘更近 */
    bottom: 20px;
    width: 40px;
    height: 40px;
  }
  .scroll-to-top .el-icon {
    font-size: 18px;
  }
}
</style>




