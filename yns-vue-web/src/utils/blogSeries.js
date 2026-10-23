// 文章系列：作者把自己的几篇文章编成一个有顺序的合集
import { ElMessage, ElMessageBox } from 'element-plus'
import { sendAxiosRequest, sendAxiosRequestChecked } from '@/utils/common.js'

// 与 Go 端 service.SeriesPath 一致
export const seriesHref = (id) => `/series/${encodeURIComponent(id)}`
export const SERIES_LIST_PATH = '/series'
export const MAX_SERIES_TITLE = 60
export const MAX_SERIES_DESCRIPTION = 300

// 当前用户的全部系列（含私密文章数）。写作页和“我的系列”共用一份；新建、改名、删除后调用 refresh
let cache = null
export async function loadMySeries({ refresh = false } = {}) {
    if (!cache || refresh) {
        cache = sendAxiosRequest('/blog-api/series/getMySeries')
            .then(res => (res && !res.isError && Array.isArray(res.result) ? res.result : []))
            .catch(() => [])
    }
    return cache
}

// 弹窗输入名称新建系列，成功返回新系列（含 ID），取消或失败返回 null
export async function promptCreateSeries() {
    let title
    try {
        const { value } = await ElMessageBox.prompt('给系列起个名字，例如“Go 入门”', '新建系列', {
            confirmButtonText: '创建',
            cancelButtonText: '取消',
            inputPlaceholder: `最多 ${MAX_SERIES_TITLE} 个字`,
            inputValidator: (text) => {
                const value = String(text || '').trim()
                if (!value) return '系列名称不能为空'
                if ([...value].length > MAX_SERIES_TITLE) return `最多 ${MAX_SERIES_TITLE} 个字`
                return true
            }
        })
        title = String(value).trim()
    } catch {
        return null
    }
    const result = await sendAxiosRequestChecked('/blog-api/series/addSeries', { title, description: '' }, '新建系列失败')
    if (!result?.result) return null
    ElMessage.success(`已新建系列《${title}》`)
    await loadMySeries({ refresh: true })
    return result.result
}
