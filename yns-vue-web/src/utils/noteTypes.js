// 便签墙的随记类型：名称、纸张样式与 Go 后端 noteTypeNames 一致
export const NOTE_TYPES = [
    { value: 'note', label: '随笔', paper: 'paper-note', hint: '碎碎念、项目进度、随手的想法' },
    { value: 'tip', label: '技巧', paper: 'paper-tip', hint: '一个小知识点、一条踩坑记录' },
    { value: 'share', label: '分享', paper: 'paper-share', hint: '推荐一个链接、工具或资源' },
    { value: 'question', label: '提问', paper: 'paper-question', hint: '遇到的问题，请大家帮忙看看' }
]

const TYPE_MAP = Object.fromEntries(NOTE_TYPES.map(item => [item.value, item]))

export const noteTypeOf = (value) => TYPE_MAP[value] || TYPE_MAP.note

const stripTags = (html) => String(html || '').replace(/<[^>]+>/g, ' ').replace(/&nbsp;/g, ' ').replace(/\s+/g, ' ').trim()

// 随记没有标题：正文以标题开头时用这个标题，否则取正文（不含代码块）前 40 个字，和后端 NoteTitle 一致
export const noteTitle = (note) => {
    if (note?.TITLE?.trim()) return note.TITLE.trim()
    const heading = String(note?.TEXT || '').match(/^\s*<h[1-6]\b[^>]*>([\s\S]*?)<\/h[1-6]>/i)
    if (heading && stripTags(heading[1])) return stripTags(heading[1]).slice(0, 40)
    const text = String(note?.TEXT || '').replace(/<pre\b[\s\S]*?<\/pre>/gi, ' ').replace(/<[^>]+>/g, ' ').replace(/&nbsp;/g, ' ').replace(/\s+/g, ' ').trim()
    if (!text) return '一张便签'
    return text.length > 40 ? text.slice(0, 40) : text
}

// 随记地址：新随记是数字编号，旧随记是一长串字符，都可以直接放进地址
export const notePath = (guid) => `/note/${encodeURIComponent(guid)}`
