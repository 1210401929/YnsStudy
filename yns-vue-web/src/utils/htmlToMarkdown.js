/**
 * 把随记的 HTML 转回 Markdown，用于编辑旧随记（迁移 010 之前发布的随记没有保存 Markdown 原文）。
 * 只处理便签里常见的写法：标题、段落、加粗、斜体、行内代码、代码块、链接、图片、列表、引用、分割线。
 */
export function htmlToMarkdown(html) {
    if (!html) return ''
    const doc = new DOMParser().parseFromString(`<body>${html}</body>`, 'text/html')

    const inline = (node) => [...node.childNodes].map(convert).join('')

    function convert(node) {
        if (node.nodeType === Node.TEXT_NODE) return node.textContent
        if (node.nodeType !== Node.ELEMENT_NODE) return ''
        const tag = node.tagName.toLowerCase()
        switch (tag) {
            case 'h1': case 'h2': case 'h3': case 'h4': case 'h5': case 'h6':
                return `${'#'.repeat(Number(tag[1]))} ${inline(node).trim()}\n\n`
            case 'p':
                return `${inline(node).trim()}\n\n`
            case 'br':
                return '\n'
            case 'strong': case 'b':
                return `**${inline(node)}**`
            case 'em': case 'i':
                return `*${inline(node)}*`
            case 'del': case 's':
                return `~~${inline(node)}~~`
            case 'code':
                return node.parentElement?.tagName.toLowerCase() === 'pre' ? node.textContent : `\`${node.textContent}\``
            case 'pre': {
                const code = node.querySelector('code')
                const lang = (code?.className.match(/language-([\w-]+)/) || [])[1] || ''
                return `\`\`\`${lang}\n${(code || node).textContent.replace(/\n$/, '')}\n\`\`\`\n\n`
            }
            case 'a':
                return `[${inline(node)}](${node.getAttribute('href') || ''})`
            case 'img':
                return `![${node.getAttribute('alt') || ''}](${node.getAttribute('src') || ''})`
            case 'ul': case 'ol': {
                const items = [...node.children].filter(child => child.tagName.toLowerCase() === 'li')
                return items.map((li, index) => `${tag === 'ol' ? `${index + 1}.` : '-'} ${inline(li).trim()}`).join('\n') + '\n\n'
            }
            case 'blockquote':
                return inline(node).trim().split('\n').map(line => `> ${line}`).join('\n') + '\n\n'
            case 'hr':
                return '---\n\n'
            default:
                return inline(node)
        }
    }

    return inline(doc.body).replace(/\n{3,}/g, '\n\n').trim()
}
