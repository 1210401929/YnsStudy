import { computed, onBeforeUnmount, reactive, unref, watch } from 'vue'
import { useHead } from '@vueuse/head'
import { adminUserCode } from '@/config/vue-config.js'

export const SITE_NAME = 'YnsStudy'

/**
 * 站长信息：首页标题和结构化数据用它把网站与站长本人关联起来，让搜索站长名字时也能找到本站。
 * 与 Go 端 config.SEOConfig.Owner() 的默认值保持一致；站长账号即超级管理员 adminUserCode。
 */
export const SITE_OWNER = {
    code: adminUserCode,
    name: '余南松',
    alternateNames: ['YuNanSong'],
    sameAs: ['https://github.com/1210401929'],
    profilePath: '/user/101'
}
// “余南松（YuNanSong）”形式，用于介绍文字
export const SITE_OWNER_DISPLAY = `${SITE_OWNER.name}（${SITE_OWNER.alternateNames[0]}）`

// 与 Go 端 homeTitle / homeDescription 一致
export const DEFAULT_TITLE = `${SITE_NAME} - ${SITE_OWNER.name}的技术分享与个人博客`
export const DEFAULT_DESCRIPTION = `${SITE_NAME} 是${SITE_OWNER_DISPLAY}的个人博客与技术社区，记录编程学习、技术实践与生活思考。`
// 1200×630 的分享卡片图，用于搜索结果和社交平台预览
export const DEFAULT_IMAGE_PATH = '/og-image.png'

export const siteOrigin = () => window.location.origin

/**
 * 当前页面是否声明了 noindex（例如文章或用户不存在）。
 * App.vue 的全站默认值据此不再输出 canonical，避免 noindex 与 canonical 两个信号互相矛盾。
 */
// 每个页面单独登记，避免路由切换时旧页面卸载把新页面的状态清掉
const noindexOwners = reactive(new Set())
export const pageNoindex = computed(() => noindexOwners.size > 0)

/** 页面组件调用：把自身的 noindex 状态同步给全站默认值，卸载时自动移除。 */
export function usePageNoindex(source) {
    const owner = Symbol('noindex')
    watch(source, (value) => {
        if (value) noindexOwners.add(owner)
        else noindexOwners.delete(owner)
    }, { immediate: true })
    onBeforeUnmount(() => noindexOwners.delete(owner))
}

// 页面地址只保留路径，查询参数和 hash 不参与 canonical，避免同一页面出现多个收录地址
export const absoluteUrl = (path = '/') => siteOrigin() + (path.startsWith('/') ? path : '/' + path)

/**
 * Go 服务端和 index.html 输出的 SEO 标签不带 Vue 的跟踪标记，Vue 接管后会和 useHead 生成的标签重复。
 * 挂载前统一移除，之后由 App.vue 的默认值和各页面的 useSeo 重新生成。
 */
export function removeStaticSeoTags() {
    const selectors = [
        'meta[name="description"]',
        'meta[name="robots"]',
        'meta[name="keywords"]',
        'link[rel="canonical"]',
        'meta[property^="og:"]',
        'meta[property^="article:"]',
        'meta[name^="twitter:"]',
        'script[type="application/ld+json"]'
    ]
    document.head.querySelectorAll(selectors.join(',')).forEach(el => el.remove())
}

/**
 * 统一生成 title、description、canonical、robots、Open Graph 和结构化数据。
 * 参数可以是对象，也可以是返回对象的函数（用于依赖接口数据的页面）。
 *
 * @param {object|Function} source
 *   title        完整标题
 *   description  页面描述
 *   path         canonical 路径，例如 /ynsStudy/About
 *   noindex      true 时输出 noindex,follow
 *   type         og:type，默认 website
 *   image        分享图片绝对地址
 *   jsonLd       结构化数据对象
 */
export function useSeo(source, { root = false } = {}) {
    const seo = computed(() => {
        const raw = typeof source === 'function' ? source() : unref(source)
        return raw || {}
    })
    // App.vue 的全站默认值（root）只读取页面的 noindex 状态，其他调用方负责上报
    if (!root) usePageNoindex(() => seo.value.noindex)

    useHead(() => {
        const { title, description, path, noindex, type, image, jsonLd } = seo.value
        const pageTitle = title || DEFAULT_TITLE
        const pageDescription = description || DEFAULT_DESCRIPTION
        const pageImage = image || absoluteUrl(DEFAULT_IMAGE_PATH)
        const meta = [
            { name: 'description', content: pageDescription },
            {
                name: 'robots',
                content: noindex ? 'noindex,follow' : 'index,follow,max-image-preview:large,max-snippet:-1'
            },
            { property: 'og:type', content: type || 'website' },
            { property: 'og:site_name', content: SITE_NAME },
            { property: 'og:title', content: pageTitle },
            { property: 'og:description', content: pageDescription },
            { property: 'og:image', content: pageImage },
            { name: 'twitter:card', content: 'summary_large_image' },
            { name: 'twitter:title', content: pageTitle },
            { name: 'twitter:description', content: pageDescription },
            { name: 'twitter:image', content: pageImage }
        ]
        const link = []
        if (path && !noindex) {
            const canonical = absoluteUrl(path)
            link.push({ rel: 'canonical', href: canonical })
            meta.push({ property: 'og:url', content: canonical })
        }
        const head = { title: pageTitle, meta, link }
        if (jsonLd) {
            head.script = [{ type: 'application/ld+json', children: JSON.stringify(jsonLd) }]
        }
        return head
    })
}

/** 站长的 schema.org Person，url 默认指向站长主页 */
export function ownerPerson(url = absoluteUrl(SITE_OWNER.profilePath)) {
    return {
        '@type': 'Person',
        name: SITE_OWNER.name,
        alternateName: SITE_OWNER.alternateNames,
        url,
        sameAs: SITE_OWNER.sameAs
    }
}

/**
 * 手机号形式的昵称替换为“手机用户”+后四位，避免在页面和搜索结果中泄露手机号。
 * 与 Go 端 model.PublicDisplayName 一致。
 */
export function publicDisplayName(name) {
    const text = String(name || '').trim()
    return /^1\d{10}$/.test(text) ? `手机用户${text.slice(-4)}` : text
}

// 与 Go 端 collapseWhitespace + truncateRunes 一致
function collapseAndTruncate(text, maximum) {
    const chars = Array.from(String(text || '').split(/\s+/).filter(Boolean).join(' '))
    return chars.length <= maximum ? chars.join('') : chars.slice(0, maximum).join('') + '…'
}

/**
 * 作者主页 /user/:u 的标题、描述和 ProfilePage 结构化数据。
 * 必须与 Go 端 seo_profile.go 的 profileTitle / profileDescription / renderSEOProfileHTML 保持一致，
 * 否则 Vue 接管后标题会和服务端首屏不同。
 */
export function buildProfileSeo(user, userNumber) {
    const name = publicDisplayName(user?.name) || '用户'
    const remark = String(user?.remark || '').trim()
    const isOwner = !!user?.code && user.code === SITE_OWNER.code
    const canonical = absoluteUrl(`/user/${encodeURIComponent(userNumber)}`)

    const title = `${isOwner ? SITE_OWNER_DISPLAY : name}的个人主页 - ${SITE_NAME}`
    const rawDescription = isOwner
        ? `${SITE_OWNER_DISPLAY}，${SITE_NAME} 站长。${remark}`
        : (remark || `${name} 在 ${SITE_NAME} 发表的文章、动态与分享的资源。`)
    const description = collapseAndTruncate(rawDescription, 160)

    let avatar = ''
    if (user?.avatar) {
        try {
            const resolved = new URL(user.avatar, siteOrigin())
            if (resolved.protocol === 'http:' || resolved.protocol === 'https:') avatar = resolved.href
        } catch { /* 头像地址无效时不输出 */ }
    }

    const person = isOwner
        ? { ...ownerPerson(canonical), description }
        : { '@type': 'Person', name, url: canonical, ...(remark ? { description: remark } : {}) }
    if (avatar) person.image = avatar

    // 收录规则与 Go 的 ProfileIndexable 一致：有公开文章或是站长本人才收录；
    // 拿不到公开文章数（旧版后端）时按可以收录处理
    const publicArticles = Number(user?.publicArticles)
    const noindex = !isOwner && Number.isFinite(publicArticles) && publicArticles === 0

    return {
        title,
        description,
        image: avatar || undefined,
        noindex,
        jsonLd: {
            '@context': 'https://schema.org',
            '@type': 'ProfilePage',
            url: canonical,
            name: title,
            mainEntity: person
        }
    }
}
