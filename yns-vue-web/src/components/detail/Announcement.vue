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
.custom-alert {
  display: flex;
  justify-content: center;
  align-items: center;
  text-align: center;

  margin-top: 6px;
  border: 1px solid #dcdfe6; /* 浅灰色边框 */
  border-radius: 6px;        /* 稍微圆角，让边框柔和 */
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05); /* 可选：轻微阴影 */
}

.alert-content {
  text-align: center;
}

.alert-link {
  color: #409EFF;
  text-decoration: underline;
  margin-left: 4px;
}
</style>
