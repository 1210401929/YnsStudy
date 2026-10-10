<!--把文字里的搜索关键词标出来（按纯文本拆分，不使用 v-html）-->
<template>
  <template v-for="(part, index) in parts" :key="index">
    <mark v-if="part.hit" class="search-hit">{{ part.text }}</mark>
    <template v-else>{{ part.text }}</template>
  </template>
</template>

<script setup>
import {computed} from 'vue'

const props = defineProps({
  text: {type: String, default: ''},
  keyword: {type: String, default: ''}
})

const parts = computed(() => {
  const text = props.text || ''
  const keyword = (props.keyword || '').trim()
  if (!keyword || !text) return [{text, hit: false}]
  // 不区分大小写，关键词里的正则特殊字符按普通字符处理
  const escaped = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return text.split(new RegExp(`(${escaped})`, 'gi'))
      .filter(Boolean)
      .map(piece => ({text: piece, hit: piece.toLowerCase() === keyword.toLowerCase()}))
})
</script>

<style scoped>
.search-hit {
  padding: 0 1px;
  color: inherit;
  background-image: linear-gradient(transparent 55%, var(--j-highlight) 55%);
  background-color: transparent;
}
</style>
