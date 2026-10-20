// 评论限制的规则类型，与 Go 端 comment_ban_service.go 的类型和匹配方式一致
export const BAN_TYPES = {
    user: { label: '注册用户', field: '账号', placeholder: '用户账号，例如 li', tip: '账号完全一致才会拦截' },
    ip: { label: 'IP', field: 'IP', placeholder: '1.2.3.4 或 1.2.3.0/24', tip: '支持单个 IP 或网段；只有执行迁移 008 后的新评论记录了 IP' },
    website: { label: '网址', field: '网址包含', placeholder: '例如 spam-site.com', tip: '评论填写的网址只要包含这段文字就会被拦截，不区分大小写' },
    email: { label: '邮箱', field: '邮箱包含', placeholder: '例如 @junk.cn', tip: '评论填写的邮箱包含这段文字就会被拦截' },
    nickname: { label: '昵称', field: '昵称包含', placeholder: '例如 代刷', tip: '昵称包含这段文字就会被拦截，请避免填写太短的常见字' }
}

/** 从网址中取出域名作为“网址包含”的默认值，例如 https://www.spam.com/a → spam.com */
export function websiteHost(website) {
    try {
        return new URL(website).hostname.replace(/^www\./, '')
    } catch {
        return String(website || '').trim()
    }
}
