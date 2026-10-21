// 便签墙的话题标签
import { sendAxiosRequest } from '@/utils/common.js'

// 话题筛选的地址：便签墙上只看这个话题
export const tagHref = (tag) => `/ynsStudy/Notes?tag=${encodeURIComponent(tag)}`

// 热门话题（名字和使用次数）。同一页面里多处使用时只请求一次；发布、编辑后调用 refresh 重新获取
let cache = null
export async function loadPopularTags({ refresh = false } = {}) {
    if (!cache || refresh) {
        cache = sendAxiosRequest('/blog-api/community/getTags', { limit: 30 })
            .then(res => (res && !res.isError && Array.isArray(res.result) ? res.result : []))
            .catch(() => [])
    }
    return cache
}
