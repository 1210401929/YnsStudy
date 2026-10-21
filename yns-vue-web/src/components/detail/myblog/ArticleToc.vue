<!-- 文章目录：文章详情页放在左栏用户信息下方，其他页面放在右栏，小屏放在抽屉里 -->
<template>
  <section class="toc-fixed-card" :class="{ 'is-plain': plain }">
    <span v-if="!plain" class="j-tape j-tape--green toc-tape"></span>
    <div class="toc-title">文章目录</div>
    <el-scrollbar :max-height="maxHeight">
      <div v-if="items.length === 0" class="toc-empty">✎ 这篇文章没有小标题</div>
      <div v-for="item in items" :key="item.id"
           class="toc-item"
           :class="['toc-level-' + item.level, { 'is-active': activeId === item.id }]"
           :title="item.text"
           @click="$emit('select', item.id)">
        <span>{{ item.text }}</span>
      </div>
    </el-scrollbar>
  </section>
</template>

<script setup>
defineProps({
  items: { type: Array, default: () => [] },
  activeId: { type: String, default: '' },
  // 吸附在侧栏时限制高度，目录很长时在卡片内部滚动
  maxHeight: { type: String, default: 'calc(100vh - 140px)' },
  // 放在抽屉里时不需要纸张卡片的胶带和阴影
  plain: { type: Boolean, default: false }
})
defineEmits(['select'])
</script>

<style scoped>
/* 一张索引卡 */
.toc-fixed-card {
  position: relative;
  padding: 24px 18px 16px;
  background: var(--j-paper);
  border: 1px solid var(--j-rule);
  box-shadow: var(--j-shadow);
  box-sizing: border-box;
}

.toc-fixed-card.is-plain {
  padding: 4px 0 0;
  background: transparent;
  border: 0;
  box-shadow: none;
}

.toc-tape {
  top: -10px;
  left: 22px;
  transform: rotate(-4deg);
}

.toc-title {
  margin-bottom: 12px;
  font-family: var(--j-hand);
  font-size: 19px;
  color: var(--j-ink);
}

.toc-empty {
  padding: 10px 4px;
  font-family: var(--j-hand);
  font-size: 15px;
  color: var(--j-muted);
}

.toc-item {
  padding: 7px 4px;
  border-bottom: 1px dashed var(--j-rule);
  font-size: 14px;
  line-height: 20px;
  color: var(--j-ink-soft);
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.toc-item span {
  background-repeat: no-repeat;
  background-image: linear-gradient(transparent 55%, var(--j-highlight) 55%, var(--j-highlight) 95%, transparent 95%);
  background-size: 0 100%;
  transition: background-size 0.3s ease;
}

.toc-item:hover,
.toc-item.is-active {
  color: var(--j-ink);
}

.toc-item.is-active span {
  background-size: 100% 100%;
}

.toc-level-2 { padding-left: 4px; }
.toc-level-3 { padding-left: 18px; font-size: 13px; }
.toc-level-4 { padding-left: 32px; font-size: 12px; }
</style>
