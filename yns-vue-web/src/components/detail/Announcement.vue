<!--横幅公共组件-->
<template>
  <el-alert
      type="success"
      class="custom-alert"
      closable
  >
    <template #title>
      <div class="alert-content">
        {{TEXT}}
        <a
            :href="URL"
            target="_blank"
            class="alert-link"
        >{{URLNAME}}</a>
      </div>
    </template>
  </el-alert>
</template>

<script setup>
const props = defineProps({
  TEXT: String,
  URL: String,
  URLNAME: String
})
</script>

<style scoped>
/* 公告：一张贴在页面顶部的黄色便签 */
.custom-alert {
  display: flex;
  justify-content: center;
  align-items: center;
  text-align: center;
  max-width: 1120px;
  margin: 12px auto 0;
  border: none;
  border-radius: 2px;
  background: var(--j-note);
  color: var(--j-ink);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.1), 0 8px 14px -12px rgba(60, 50, 30, 0.45);
}

.custom-alert :deep(.el-alert__close-btn) {
  color: var(--j-ink-soft);
}

.alert-content {
  text-align: center;
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-ink);
}

.alert-content::before {
  content: "便签 · ";
  color: var(--j-stamp);
}

.alert-link {
  color: var(--j-pen);
  text-decoration: underline;
  text-underline-offset: 3px;
  margin-left: 4px;
}
</style>
