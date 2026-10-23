<!-- 话题输入：回车、逗号或空格确认一个话题，最多 5 个；下面列出常用话题，点一下就加上 -->
<template>
  <div class="tag-input">
    <div class="tag-box" :class="{ 'is-full': modelValue.length >= MAX_TAGS }" @click="focusInput">
      <span v-for="tag in modelValue" :key="tag" class="tag-chip">
        #{{ tag }}
        <button type="button" class="tag-remove" :title="`移除 #${tag}`" @click.stop="removeTag(tag)">×</button>
      </span>
      <input
          v-if="modelValue.length < MAX_TAGS"
          ref="inputRef"
          v-model="draft"
          class="tag-draft"
          maxlength="21"
          :placeholder="modelValue.length ? `继续添加${noun}` : `添加${noun}，例如 Go、踩坑（回车确认）`"
          @keydown.enter.prevent="commit"
          @keydown="onKeydown"
          @blur="commit"
          @compositionstart="composing = true"
          @compositionend="composing = false"
      >
      <span v-else class="tag-full">最多 {{ MAX_TAGS }} 个{{ noun }}</span>
    </div>
    <div v-if="suggestions.length" class="tag-suggest">
      <span class="suggest-label">常用：</span>
      <button v-for="item in suggestions" :key="item" type="button" class="suggest-item" @click="addTag(item)">#{{ item }}</button>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'

const MAX_TAGS = 5
const MAX_RUNES = 20
const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  // 常用话题（名字数组）
  popular: { type: Array, default: () => [] },
  // 提示文字里的叫法：便签墙叫“话题”，文章叫“标签”
  noun: { type: String, default: '话题' }
})
const emit = defineEmits(['update:modelValue'])

const inputRef = ref(null)
const draft = ref('')
// 中文输入法拼写过程中按空格、回车不算确认
const composing = ref(false)

// 和后端 normalizeTagName 一致：去掉开头的 #，不能有空白和 , # / ? 这类分隔符，最长 20 个字
const normalize = (name) => {
  const value = String(name || '').trim().replace(/^[#＃]+/, '').trim()
  if (!value || [...value].length > MAX_RUNES || /[\s,，#＃;；、<>"'/\\?]/.test(value)) return ''
  return value
}

const has = (name) => props.modelValue.some(tag => tag.toLowerCase() === name.toLowerCase())

const suggestions = computed(() => props.popular.filter(name => !has(name)).slice(0, 10))

// 一次加入多个话题：先算好完整列表再更新一次，避免连续更新时后一次覆盖前一次
function addTags(rawList) {
  const next = [...props.modelValue]
  let invalid = false
  for (const raw of rawList) {
    const name = normalize(raw)
    if (!name) {
      invalid = invalid || !!String(raw || '').trim()
      continue
    }
    if (next.length >= MAX_TAGS) break
    if (!next.some(tag => tag.toLowerCase() === name.toLowerCase())) next.push(name)
  }
  if (invalid) ElMessage.warning(`${props.noun}不能包含标点，最长 20 个字`)
  if (next.length !== props.modelValue.length) emit('update:modelValue', next)
}

const addTag = (raw) => addTags([raw])

function commit() {
  if (composing.value) return
  // 一次粘贴多个时按空格、逗号拆开
  const parts = draft.value.split(/[\s,，]+/).filter(Boolean)
  draft.value = ''
  addTags(parts)
}

function onKeydown(event) {
  if (composing.value) return
  if (event.key === ',' || event.key === '，' || event.key === ' ') {
    event.preventDefault()
    commit()
  } else if (event.key === 'Backspace' && !draft.value && props.modelValue.length) {
    removeTag(props.modelValue[props.modelValue.length - 1])
  }
}

function removeTag(tag) {
  emit('update:modelValue', props.modelValue.filter(item => item !== tag))
}

function focusInput() {
  inputRef.value?.focus()
}
</script>

<style scoped>
.tag-box {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  min-height: 34px;
  padding: 4px 8px;
  border: 1px solid var(--el-border-color, var(--j-rule-strong));
  border-radius: 4px;
  background: #fff;
  cursor: text;
  box-sizing: border-box;
}

.tag-box:focus-within {
  border-color: var(--j-pen);
}

.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 1px 4px 1px 8px;
  border-radius: 10px;
  background: #eef3f8;
  font-size: 13px;
  color: var(--j-pen);
}

.tag-remove {
  width: 16px;
  height: 16px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--j-muted);
  font-size: 14px;
  line-height: 16px;
  cursor: pointer;
}

.tag-remove:hover {
  background: var(--j-stamp);
  color: #fff;
}

.tag-draft {
  flex: 1;
  min-width: 160px;
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  color: var(--j-ink);
}

.tag-full {
  font-size: 12px;
  color: var(--j-muted);
}

.tag-suggest {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px 8px;
  margin-top: 6px;
  font-size: 12px;
}

.suggest-label {
  color: var(--j-muted);
}

.suggest-item {
  padding: 0;
  border: none;
  background: none;
  font-size: 12px;
  color: var(--j-pen);
  cursor: pointer;
}

.suggest-item:hover {
  text-decoration: underline;
}
</style>
