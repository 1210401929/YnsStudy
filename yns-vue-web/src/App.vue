<template>
  <div id="app">
    <router-view></router-view>  <!-- 用来渲染匹配的组件 -->
  </div>
</template>
<script>

export default {
  name: 'App'
};
</script>

<style scoped>

#app {
  display: block;
  height: 100%;
  width: 100%;
  /*超出部分隐藏*/
  overflow: hidden;
}

</style>
