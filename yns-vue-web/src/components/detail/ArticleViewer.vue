<!--文章正文展示组件：直接渲染保存的 HTML，不加载编辑器-->
<template>
  <div class="editor-container is-reading article-viewer">
    <div ref="bodyRef" class="article-body" @dblclick="onDblClick" v-html="safeHtml"></div>

    <!-- 图片查看器 -->
    <el-image-viewer
        v-if="showViewer"
        :url-list="[viewerUrl]"
        @close="showViewer = false"
    />
  </div>
</template>

<script setup>
import {computed, nextTick, onMounted, ref, watch} from 'vue'
import {highlightCodeBlocks, prepareArticleHtml} from '@/utils/articleHtml.js'

const props = defineProps({
  content: String
})

const bodyRef = ref(null)
const safeHtml = computed(() => prepareArticleHtml(props.content))

/* ---------- 代码高亮 ---------- */
const highlight = () => nextTick(() => highlightCodeBlocks(bodyRef.value))
onMounted(highlight)
watch(safeHtml, highlight)

/* ---------- 双击查看图片 ---------- */
const showViewer = ref(false)
const viewerUrl = ref('')
const onDblClick = (e) => {
  if (e.target.tagName === 'IMG') {
    viewerUrl.value = e.target.src
    showViewer.value = true
  }
}
</script>

<style scoped>
/*
 * 排版与 ArticleEditor 只读模式保持一致：
 * 前半部分对应 wangEditor 自带的正文样式，后半部分是手账风格的阅读样式。
 */
.article-viewer {
  position: relative;
  min-height: 300px;
}

.article-body {
  padding: 0 10px;
  color: var(--w-e-textarea-color);
  word-wrap: break-word;
  overflow-x: auto;
  border-top: 1px solid transparent;
}

.article-body :deep(*) {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

.article-body :deep(p),
.article-body :deep(blockquote),
.article-body :deep(td),
.article-body :deep(th) {
  line-height: 1.5;
}

/* 编辑器内文字按原样保留空格 */
.article-body :deep(p),
.article-body :deep(li),
.article-body :deep(h1),
.article-body :deep(h2),
.article-body :deep(h3),
.article-body :deep(h4),
.article-body :deep(h5),
.article-body :deep(td),
.article-body :deep(th),
.article-body :deep(blockquote),
.article-body :deep([data-w-e-type="todo"]) {
  white-space: pre-wrap;
}

.article-body :deep(.table-container) {
  white-space: pre-wrap;
}

.article-body :deep(p) {
  margin: 0.6em 0;
}

.article-body :deep(span) {
  text-indent: 0;
}

.article-body :deep(h1),
.article-body :deep(h2),
.article-body :deep(h3),
.article-body :deep(h4),
.article-body :deep(h5) {
  margin: 1.6em 0 0.6em;
  line-height: 1.4;
  color: var(--j-ink);
}

.article-body :deep(h1) {
  font-size: 2em;
}

.article-body :deep(h2) {
  font-size: 22px;
  padding-bottom: 6px;
  border-bottom: 1px dashed var(--j-rule-strong);
}

.article-body :deep(h3) {
  font-size: 19px;
}

.article-body :deep(h4) {
  font-size: 1em;
}

.article-body :deep(h5) {
  font-size: 0.83em;
}

.article-body :deep(ul),
.article-body :deep(ol) {
  margin: 0;
  padding-left: 0;
  list-style-position: inside;
}

.article-body :deep(li) {
  margin: 5px 0;
  line-height: normal;
}

/* 多级列表逐级缩进 */
.article-body :deep(li > ul),
.article-body :deep(li > ol) {
  padding-left: 2em;
}

.article-body :deep(a) {
  color: var(--j-pen);
  text-decoration: underline;
  text-underline-offset: 3px;
}

.article-body :deep(blockquote) {
  display: block;
  margin: 1em 0;
  padding: 10px 16px;
  border-left: 4px solid #e8c95b;
  background: #fdf6d8;
  color: var(--j-ink-soft);
  font-size: 100%;
  line-height: 1.8;
}

.article-body :deep(code) {
  padding: 1px 6px;
  border-radius: 3px;
  background: #f3ecdc;
  color: #8a4b2f;
  font-family: monospace;
  font-size: 0.9em;
  white-space: pre-wrap;
  word-break: break-word;
}

.article-body :deep(pre) {
  white-space: pre-wrap;
  word-break: break-word;
}

.article-body :deep(pre > code) {
  display: block;
  margin: 0.5em 0;
  padding: 14px 16px;
  overflow: auto;
  border: 1px solid var(--j-rule);
  border-radius: 3px;
  background: #f7f3ea;
  color: var(--j-ink);
  font-family: Consolas, Monaco, "Andale Mono", "Ubuntu Mono", monospace;
  font-size: 13px;
  line-height: 1.7;
  text-align: left;
  text-indent: 0;
  text-shadow: none;
  white-space: pre;
  word-break: normal;
  word-spacing: normal;
  word-wrap: normal;
  tab-size: 4;
  hyphens: none;
}

/* Prism 配色，与编辑器相同 */
.article-body :deep(pre > code .token.comment),
.article-body :deep(pre > code .token.prolog),
.article-body :deep(pre > code .token.doctype),
.article-body :deep(pre > code .token.cdata) {
  color: #708090;
}

.article-body :deep(pre > code .token.punctuation) {
  color: #999;
}

.article-body :deep(pre > code .token.namespace) {
  opacity: 0.7;
}

.article-body :deep(pre > code .token.property),
.article-body :deep(pre > code .token.tag),
.article-body :deep(pre > code .token.boolean),
.article-body :deep(pre > code .token.number),
.article-body :deep(pre > code .token.constant),
.article-body :deep(pre > code .token.symbol),
.article-body :deep(pre > code .token.deleted) {
  color: #905;
}

.article-body :deep(pre > code .token.selector),
.article-body :deep(pre > code .token.attr-name),
.article-body :deep(pre > code .token.string),
.article-body :deep(pre > code .token.char),
.article-body :deep(pre > code .token.builtin),
.article-body :deep(pre > code .token.inserted) {
  color: #690;
}

.article-body :deep(pre > code .token.operator),
.article-body :deep(pre > code .token.entity),
.article-body :deep(pre > code .token.url),
.article-body :deep(pre > code .language-css .token.string),
.article-body :deep(pre > code .style .token.string) {
  color: #9a6e3a;
}

.article-body :deep(pre > code .token.atrule),
.article-body :deep(pre > code .token.attr-value),
.article-body :deep(pre > code .token.keyword) {
  color: #07a;
}

.article-body :deep(pre > code .token.function),
.article-body :deep(pre > code .token.class-name) {
  color: #dd4a68;
}

.article-body :deep(pre > code .token.regex),
.article-body :deep(pre > code .token.important),
.article-body :deep(pre > code .token.variable) {
  color: #e90;
}

.article-body :deep(pre > code .token.important),
.article-body :deep(pre > code .token.bold) {
  font-weight: 700;
}

.article-body :deep(pre > code .token.italic) {
  font-style: italic;
}

.article-body :deep(img) {
  display: inline !important;
  max-width: 100%;
  min-width: 20px;
  min-height: 20px;
  padding: 6px;
  background: #fff;
  box-shadow: 0 1px 3px rgba(60, 50, 30, 0.2);
  cursor: zoom-in;
}

.article-body :deep(.table-container) {
  width: 100%;
  margin-top: 10px;
  padding: 10px;
  overflow-x: auto;
  border: 1px dashed var(--w-e-textarea-border-color);
  border-radius: 5px;
}

.article-body :deep(table) {
  border-collapse: collapse;
}

.article-body :deep(th),
.article-body :deep(td) {
  min-width: 30px;
  padding: 6px 10px;
  border: 1px solid var(--j-rule-strong);
  line-height: 1.5;
  text-align: left;
}

.article-body :deep(th) {
  background: var(--j-paper-warm);
  font-weight: 700;
  text-align: center;
}

.article-body :deep(hr) {
  display: block;
  height: 1px;
  margin: 40px 20px;
  border: 0;
  background-color: var(--w-e-textarea-border-color);
}

.article-body :deep([data-w-e-type="todo"]) {
  margin: 5px 0;
  line-height: normal;
}

.article-body :deep([data-w-e-type="todo"] input) {
  margin-right: 0.5em;
}
</style>
