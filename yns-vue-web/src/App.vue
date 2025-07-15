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
  </div>
</template>

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
</style>
<script setup>
</script>