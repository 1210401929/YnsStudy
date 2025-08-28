<template>
  <el-alert
      type="success"
      show-icon
      style="display: flex;justify-content: center;align-items: center;text-align: center;"
      closable
  >
    <template #title>
      <div style="text-align: center;">
        {{TEXT}}
        <a
            :href="URL"
            target="_blank"
            style="color: #409EFF; text-decoration: underline; margin-left: 4px;"
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

</style>