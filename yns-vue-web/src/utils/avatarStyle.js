// 没有头像时，根据用户名生成固定的头像底色（取自胶带配色），同一个名字每次颜色都一样
const AVATAR_COLORS = [
    { bg: '#f6e3a1', text: '#7a5c12' },
    { bg: '#c7e2cf', text: '#2f6040' },
    { bg: '#f3cccc', text: '#8a3a34' },
    { bg: '#c9dbeb', text: '#2f5d8a' },
    { bg: '#e2d5c0', text: '#6b5330' }
]

export const getAvatarStyle = (name) => {
    if (!name) return {}
    let hash = 0
    for (let i = 0; i < name.length; i++) {
        hash = name.charCodeAt(i) + ((hash << 5) - hash)
    }
    const color = AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length]
    return { backgroundColor: color.bg, color: color.text, fontWeight: '600' }
}
