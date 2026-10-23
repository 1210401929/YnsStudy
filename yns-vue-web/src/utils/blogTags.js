// 文章标签：和便签墙的话题共用一套标签，但统计和标签页只算公开文章
import { sendAxiosRequest } from '@/utils/common.js'

// 标签页地址（与 Go 端 service.BlogTagPath 一致）
export const blogTagHref = (tag) => `/tag/${encodeURIComponent(tag)}`

// 热门文章标签（名字和公开文章数）。同一页面里多处使用时只请求一次；发布、修改文章后调用 refresh 重新获取
let cache = null
export async function loadBlogTags({ refresh = false } = {}) {
    if (!cache || refresh) {
        cache = sendAxiosRequest('/blog-api/blog/getBlogTags', { limit: 30 })
            .then(res => (res && !res.isError && Array.isArray(res.result) ? res.result : []))
            .catch(() => [])
    }
    return cache
}
