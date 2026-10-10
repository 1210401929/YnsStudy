// 文章正文展示用的 HTML 处理：过滤、补全表格容器、代码高亮。
// 文章页直接渲染保存好的 HTML，不再为“只读”加载整个 wangEditor。
import createDOMPurify from 'dompurify'

// 单独的实例，下面的样式过滤钩子不影响 common.js 里的 sanitizeHtml
const purify = createDOMPurify(window)

// 与后台 html_sanitize.go 的样式白名单保持一致
const ALLOWED_STYLES = [
    'color', 'background-color', 'text-align', 'line-height', 'text-indent',
    'font-size', 'font-weight', 'font-style', 'text-decoration',
    'width', 'height', 'max-width', 'vertical-align', 'white-space',
    'margin-left', 'padding-left', 'border', 'border-collapse', 'font-family'
]

purify.addHook('afterSanitizeAttributes', (node) => {
    // 只保留编辑器会产生的行内样式，避免旧数据里的 position: fixed 之类影响页面
    if (node.hasAttribute && node.hasAttribute('style')) {
        const kept = []
        for (const name of ALLOWED_STYLES) {
            const value = node.style.getPropertyValue(name)
            if (value && !/url\s*\(/i.test(value)) kept.push(`${name}: ${value}`)
        }
        if (kept.length) node.setAttribute('style', kept.join('; '))
        else node.removeAttribute('style')
    }
    // 新窗口打开的链接不把当前页暴露给对方
    if (node.tagName === 'A' && node.getAttribute('target') === '_blank') {
        node.setAttribute('rel', 'noopener noreferrer')
    }
})

const PURIFY_CONFIG = {
    // 编辑器不允许插入视频，这里也不渲染任何媒体嵌入
    FORBID_TAGS: ['style', 'form', 'video', 'audio', 'source', 'iframe', 'object', 'embed'],
    FORBID_ATTR: ['contenteditable'],
    RETURN_DOM_FRAGMENT: true
}

/**
 * 把保存的正文 HTML 处理成可直接 v-html 的字符串
 */
export function prepareArticleHtml(html) {
    if (!html) return ''
    const fragment = purify.sanitize(String(html), PURIFY_CONFIG)

    // 去掉块级标签之间的换行缩进，正文按原样保留空格时不会多出空行
    const blockParents = 'ul, ol, table, thead, tbody, tfoot, tr'
    const strip = (parent) => {
        for (const node of Array.from(parent.childNodes)) {
            if (node.nodeType === Node.TEXT_NODE && !node.textContent.trim()) node.remove()
        }
    }
    strip(fragment)
    fragment.querySelectorAll(blockParents).forEach(strip)

    // 宽表格横向滚动，和编辑器一样包一层 table-container
    fragment.querySelectorAll('table').forEach((table) => {
        if (table.parentElement?.classList.contains('table-container')) return
        const wrapper = document.createElement('div')
        wrapper.className = 'table-container'
        table.replaceWith(wrapper)
        wrapper.appendChild(table)
    })

    // 待办只用于展示，勾选框一律禁用
    fragment.querySelectorAll('input').forEach((input) => {
        if (input.getAttribute('type') !== 'checkbox') input.remove()
        else input.setAttribute('disabled', '')
    })

    const box = document.createElement('div')
    box.appendChild(fragment)
    return box.innerHTML
}

/* ---------- 代码高亮（Prism.js，与 wangEditor 相同的高亮库，按需加载） ---------- */

// 编辑器“选择语言”里的语言及其依赖；markup/css/clike/javascript 已内置在 Prism 核心里
const LANGUAGE_LOADERS = {
    typescript: [() => import('prismjs/components/prism-typescript')],
    jsx: [() => import('prismjs/components/prism-jsx')],
    tsx: [() => import('prismjs/components/prism-jsx'), () => import('prismjs/components/prism-typescript'), () => import('prismjs/components/prism-tsx')],
    go: [() => import('prismjs/components/prism-go')],
    php: [() => import('prismjs/components/prism-markup-templating'), () => import('prismjs/components/prism-php')],
    c: [() => import('prismjs/components/prism-c')],
    cpp: [() => import('prismjs/components/prism-c'), () => import('prismjs/components/prism-cpp')],
    python: [() => import('prismjs/components/prism-python')],
    java: [() => import('prismjs/components/prism-java')],
    csharp: [() => import('prismjs/components/prism-csharp')],
    'visual-basic': [() => import('prismjs/components/prism-visual-basic')],
    sql: [() => import('prismjs/components/prism-sql')],
    ruby: [() => import('prismjs/components/prism-ruby')],
    swift: [() => import('prismjs/components/prism-swift')],
    bash: [() => import('prismjs/components/prism-bash')],
    lua: [() => import('prismjs/components/prism-lua')],
    groovy: [() => import('prismjs/components/prism-groovy')],
    markdown: [() => import('prismjs/components/prism-markdown')],
    json: [() => import('prismjs/components/prism-json')],
    yaml: [() => import('prismjs/components/prism-yaml')]
}
const LANGUAGE_ALIAS = {html: 'markup', xml: 'markup', js: 'javascript', ts: 'typescript', shell: 'bash', sh: 'bash', vb: 'visual-basic', yml: 'yaml'}

let prismPromise = null
const loadedLanguages = new Set()

function loadPrism() {
    if (!prismPromise) {
        // 关闭 Prism 加载后自动高亮整页，只高亮正文里的代码块
        window.Prism = window.Prism || {}
        window.Prism.manual = true
        prismPromise = import('prismjs').then((module) => {
            const Prism = module.default || module
            // 语言包通过全局 Prism 注册自己
            window.Prism = Prism
            return Prism
        })
    }
    return prismPromise
}

async function loadLanguage(Prism, lang) {
    if (Prism.languages[lang] || loadedLanguages.has(lang)) return
    loadedLanguages.add(lang)
    for (const loader of LANGUAGE_LOADERS[lang] || []) {
        try {
            await loader()
        } catch (e) {
            console.error('代码高亮语言加载失败', lang, e)
        }
    }
}

function getLanguage(codeEl) {
    const match = /\blang(?:uage)?-([\w-]+)/i.exec(codeEl.className || '')
    if (!match) return ''
    const lang = match[1].toLowerCase()
    return LANGUAGE_ALIAS[lang] || lang
}

/**
 * 高亮容器里带语言的代码块，没有代码块时不会下载 Prism
 */
export async function highlightCodeBlocks(root) {
    if (!root) return
    const blocks = Array.from(root.querySelectorAll('pre > code')).filter(getLanguage)
    if (!blocks.length) return

    const Prism = await loadPrism()
    for (const code of blocks) {
        const lang = getLanguage(code)
        await loadLanguage(Prism, lang)
        // 渲染期间内容可能已被替换，节点不在页面上就跳过
        if (!code.isConnected || !Prism.languages[lang]) continue
        if (!code.classList.contains(`language-${lang}`)) code.classList.add(`language-${lang}`)
        Prism.highlightElement(code)
    }
}
