// 便签墙的随记类型：名称与 Go 后端 noteTypeNames 一致。
// 纸张统一用米白色，类型靠胶带和小圆点的颜色区分，避免整面墙都是一种颜色。
export const NOTE_TYPES = [
    { value: 'note', label: '随笔', paper: 'paper-note', color: '#e0b93f', hint: '碎碎念、项目进度、随手的想法' },
    { value: 'tip', label: '技巧', paper: 'paper-tip', color: '#5f9e78', hint: '一个小知识点、一条踩坑记录' },
    { value: 'share', label: '分享', paper: 'paper-share', color: '#5b87b5', hint: '推荐一个链接、工具或资源' },
    { value: 'question', label: '提问', paper: 'paper-question', color: '#d4766d', hint: '遇到的问题，请大家帮忙看看' }
]

const TYPE_MAP = Object.fromEntries(NOTE_TYPES.map(item => [item.value, item]))

export const noteTypeOf = (value) => TYPE_MAP[value] || TYPE_MAP.note

const stripTags = (html) => String(html || '').replace(/<[^>]+>/g, ' ').replace(/&nbsp;/g, ' ').replace(/\s+/g, ' ').trim()

// 随记没有标题：正文以标题开头时用这个标题，否则取正文（不含代码块）前 40 个字，和后端 NoteTitle 一致
export const noteTitle = (note) => {
    if (note?.TITLE?.trim()) return note.TITLE.trim()
    const heading = String(note?.TEXT || '').match(/^\s*<h[1-6]\b[^>]*>([\s\S]*?)<\/h[1-6]>/i)
    if (heading && stripTags(heading[1])) return stripTags(heading[1]).slice(0, 40)
    const text = stripTags(String(note?.TEXT || '').replace(/<pre\b[\s\S]*?<\/pre>/gi, ' '))
    if (!text) return '一张便签'
    return text.length > 40 ? text.slice(0, 40) : text
}

// 随记地址：数字编号
export const notePath = (guid) => `/note/${encodeURIComponent(guid)}`
