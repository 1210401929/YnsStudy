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
            <p>作者：YuNanSong | 联系方式：<a href="mailto:yvnansong@qq.com">yvnansong@qq.com</a> |
              <a href="https://github.com/1210401929/YnsStudy" target="_blank">访问作者 GitHub</a></p>
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
    <div class="scroll-to-top" :class="{ hidden: !showScrollButton }" @click="scrollToTop" @mouseenter="hover = true" @mouseleave="hover = false">
      <el-icon><Top/></el-icon>
    </div>
  </div>
</template>




<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { Top } from "@element-plus/icons-vue"; // 引入 Top 图标

const showScrollButton = ref(false);
const hover = ref(false); // 用于控制鼠标移入移出时的透明度变化

// 判断是否显示回到顶部按钮
const handleScroll = () => {
  const scrollY = window.scrollY;  // 获取页面滚动的垂直距离

  // 当页面滚动超过300px时，显示回到顶部按钮
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

/* 回到顶部按钮 */
.scroll-to-top {
  position: fixed;
  left: 20px;
  bottom: 20px;
  width: 30px; /* 设置按钮宽度 */
  height: 80px; /* 设置按钮高度，使其纵向 */
  background-color: rgba(64, 158, 255, 0.4); /* 半透明 */
  color: #fff;
  border-radius: 20px; /* 圆柱形 */
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  box-shadow: 0px 4px 6px rgba(0, 0, 0, 0.1);
  opacity: 0.4; /* 初始透明度 */
  transition: opacity 0.3s ease, transform 0.3s ease;
  z-index: 999;
  font-size: 14px;
}

.scroll-to-top .el-icon {
  font-size: 24px; /* 设置图标的大小 */
}

.scroll-to-top:hover {
  opacity: 1; /* 鼠标悬停时完全显示 */
  transform: translateY(-5px); /* 轻微上升 */
}

.scroll-to-top .scroll-text {
  display: inline-block;
  transition: opacity 0.3s ease;
}

.scroll-to-top.hidden {
  display: none;
}

.scroll-to-top .scroll-text {
  opacity: 0; /* 初始文本不可见 */
}

.scroll-to-top:hover .scroll-text {
  opacity: 1; /* 鼠标悬停时显示文本 */
}

/* 手机端样式 */
@media (max-width: 768px) {
  .scroll-to-top {
    width: 35px;
    height: 100px;
  }

  .scroll-to-top .el-icon {
    font-size: 18px;
  }
}
</style>




