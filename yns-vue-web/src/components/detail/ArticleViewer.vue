<!--文章正文展示组件：直接渲染保存的 HTML，不加载编辑器-->
<template>
  <div class="editor-container is-reading article-viewer">
    <div ref="bodyRef" class="article-body" @click="onClick" @dblclick="onDblClick" v-html="safeHtml"></div>

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
import {copyText, highlightCodeBlocks, prepareArticleHtml} from '@/utils/articleHtml.js'

const props = defineProps({
  content: String
})

const bodyRef = ref(null)
const safeHtml = computed(() => prepareArticleHtml(props.content))

/* ---------- 单独成段的大图像照片一样贴一条胶带 ---------- */
// 和文字混排的小图、并排的多张图、链接或表格里的图不处理；图片加载后宽度够大才显示胶带，避免贴在小图标上
const framePhotos = (root) => {
  root?.querySelectorAll('img').forEach((img) => {
    if (img.closest('.photo-frame, a, td, th, li')) return
    const block = img.parentElement
    if (!block || block.textContent.trim() || block.querySelectorAll('img').length > 1) return
    const frame = document.createElement('span')
    frame.className = 'photo-frame'
    img.replaceWith(frame)
    frame.appendChild(img)
    const mark = () => frame.classList.toggle('has-tape', img.naturalWidth >= 160)
    if (img.complete) mark()
    else img.addEventListener('load', mark, {once: true})
  })
}

/* ---------- 代码高亮 ---------- */
const highlight = () => nextTick(() => {
  highlightCodeBlocks(bodyRef.value)
  framePhotos(bodyRef.value)
})
onMounted(highlight)
watch(safeHtml, highlight)

/* ---------- 复制代码 ---------- */
const onClick = async (e) => {
  const button = e.target.closest?.('.code-copy')
  if (!button) return
  const code = button.closest('.code-block')?.querySelector('pre')
  if (!code) return
  const ok = await copyText(code.innerText.replace(/\n$/, ''))
  button.textContent = ok ? '已复制' : '复制失败'
  button.classList.toggle('is-done', ok)
  clearTimeout(button._resetTimer)
  button._resetTimer = setTimeout(() => {
    button.textContent = '复制'
    button.classList.remove('is-done')
  }, 1500)
}

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

/* 二级标题：荧光笔划过（按行重复，标题折行时每一行都有） */
.article-body :deep(h2) {
  width: fit-content;
  max-width: 100%;
  padding: 0 4px;
  font-size: 22px;
  background-image: linear-gradient(transparent 58%, rgba(250, 216, 96, 0.55) 58%, rgba(250, 216, 96, 0.55) 90%, transparent 90%);
  background-size: 100% 1.4em;
}

/* 三级标题：左边一小段绿色胶带 */
.article-body :deep(h3) {
  padding-left: 10px;
  border-left: 4px solid var(--j-tape-green);
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

/*
 * 多级列表逐级缩进，与编辑器一致（每深一级缩进 20px）。
 * wangEditor 保存二级列表时，里层的 ul/ol 直接写在外层 ul/ol 里（不在 li 里），两种写法都要覆盖
 */
.article-body :deep(ul ul),
.article-body :deep(ul ol),
.article-body :deep(ol ul),
.article-body :deep(ol ol) {
  padding-left: 20px;
}

.article-body :deep(a) {
  color: var(--j-pen);
  text-decoration: underline;
  text-underline-offset: 3px;
}

/* 引用：一张贴了胶带的黄色便签 */
.article-body :deep(blockquote) {
  position: relative;
  display: block;
  margin: 1.6em 0 1.2em;
  padding: 16px 18px 12px;
  background: var(--j-note);
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.1), 0 10px 18px -12px rgba(60, 50, 30, 0.4);
  color: var(--j-ink-soft);
  font-size: 100%;
  line-height: 1.8;
}

.article-body :deep(blockquote)::before {
  content: "";
  position: absolute;
  top: -9px;
  left: 50%;
  width: 64px;
  height: 18px;
  margin-left: -32px;
  background: var(--j-tape-green);
  transform: rotate(-2deg);
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

/* 代码块：顶部一条栏显示语言和复制按钮，下面是代码 */
.article-body :deep(.code-block) {
  margin: 0.5em 0;
  border: 1px solid var(--j-rule);
  border-radius: 3px;
  background: #f7f3ea;
  overflow: hidden;
}

/* 代码栏用牛皮纸色，像夹在本子里的一张卡片；代码本身保持清楚，不做装饰 */
.article-body :deep(.code-bar) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-height: 28px;
  padding: 0 8px 0 14px;
  background: rgba(231, 214, 181, 0.45);
  border-bottom: 1px solid #ddcca9;
  font-size: 12px;
  line-height: 1.5;
  color: var(--j-muted);
  white-space: normal;
}

.article-body :deep(.code-lang) {
  font-family: Consolas, Monaco, "Andale Mono", "Ubuntu Mono", monospace;
}

.article-body :deep(.code-copy) {
  padding: 2px 10px;
  border: 1px solid transparent;
  border-radius: 3px;
  background: transparent;
  font-size: 12px;
  color: var(--j-ink-soft);
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s;
}

.article-body :deep(.code-copy:hover) {
  border-color: var(--j-rule-strong);
  color: var(--j-pen);
}

.article-body :deep(.code-copy.is-done) {
  color: #4f8a5b;
}

.article-body :deep(.code-block pre > code) {
  margin: 0;
  border: none;
  border-radius: 0;
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

.article-body :deep(.photo-frame) {
  position: relative;
  display: inline-block;
  max-width: 100%;
}

.article-body :deep(.photo-frame.has-tape)::before {
  content: "";
  position: absolute;
  top: -9px;
  left: 50%;
  z-index: 1;
  width: 72px;
  height: 20px;
  margin-left: -36px;
  background: var(--j-tape-yellow);
  transform: rotate(-3deg);
  pointer-events: none;
}

.article-body :deep(.photo-frame img) {
  vertical-align: bottom;
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

/* 表格：铅笔画的格子，表头用楷体 */
.article-body :deep(th),
.article-body :deep(td) {
  min-width: 30px;
  padding: 6px 10px;
  border: 1px solid rgba(77, 73, 67, 0.32);
  line-height: 1.5;
  text-align: left;
}

.article-body :deep(th) {
  background: rgba(231, 214, 181, 0.3);
  font-family: var(--j-hand);
  font-size: 15px;
  font-weight: normal;
  text-align: center;
}

/* 分割线：手画的波浪线 */
.article-body :deep(hr) {
  display: block;
  height: 8px;
  margin: 40px 20px;
  border: 0;
  background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='8'%3E%3Cpath d='M0 4 Q6 0.5 12 4 T24 4' fill='none' stroke='%23c9bea8' stroke-width='1.4'/%3E%3C/svg%3E") repeat-x center;
}

.article-body :deep([data-w-e-type="todo"]) {
  margin: 5px 0;
  line-height: normal;
}

.article-body :deep([data-w-e-type="todo"] input) {
  margin-right: 0.5em;
}
</style>
