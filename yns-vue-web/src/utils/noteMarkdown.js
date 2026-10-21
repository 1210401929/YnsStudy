// 便签的 Markdown 转 HTML
import { Marked } from 'marked'
import { sanitizeHtml } from '@/utils/common.js'

// 裸网址只取 ASCII 字符，结尾的英文标点不算：
// marked 默认会把“https://go.dev，很好”整个当成网址，生成打不开的链接，链接卡片也抓不到
const BARE_URL = /^https?:\/\/[\w\-.~:/?#@!$&*+,;=%]+/i

export const firstUrl = (text) =>
    ((String(text || '').match(/https?:\/\/[\w\-.~:/?#@!$&*+,;=%]+/i) || [])[0] || '').replace(/[.,;:!?]+$/, '')

const noteMarked = new Marked({
    // 单个回车也换行（标准 Markdown 要空一行才分段，写便签时不符合习惯）
    breaks: true,
    gfm: true,
    tokenizer: {
        url(src) {
            const match = BARE_URL.exec(src)
            if (!match) return false // 其他情况（如 www. 开头、邮箱）交给 marked 默认处理
            const url = match[0].replace(/[.,;:!?]+$/, '')
            return { type: 'link', raw: url, text: url, href: url, tokens: [{ type: 'text', raw: url, text: url }] }
        }
    }
})

export const renderNoteMarkdown = (text) => sanitizeHtml(noteMarked.parse(String(text || '')))
