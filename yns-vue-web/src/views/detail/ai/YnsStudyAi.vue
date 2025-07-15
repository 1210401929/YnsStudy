<template>
  <iframe src="/html/ai.html" class="ai-frame"></iframe>
</template>

<style scoped>
html, body {
  width: 100%;
  height: 100%;
  margin: 0;
  padding: 0;
}

.ai-frame {
  width: 100%;
  height: 100%;
  border: none;
  position: absolute;
  top: 0;
  left: 0;
}

/* 针对手机端的优化 */
@media only screen and (max-width: 768px) {
  .ai-frame {
    height: 100vh;  /* 设置为视口高度，确保在小屏幕上不超出 */
    width: 100%;
  }
}
</style>
