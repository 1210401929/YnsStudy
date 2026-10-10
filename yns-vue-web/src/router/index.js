import {createRouter, createWebHistory} from 'vue-router'

// 页面组件全部按路由懒加载，首屏只下载当前页面需要的代码
const Welcome = () => import('../views/main/Welcome.vue')
const Index = () => import('../views/main/index.vue')
const Home = () => import('../views/detail/home/Home.vue')
const MyBlog = () => import('../views/detail/blog/MyBlog.vue')
const OneBlog = () => import('../views/detail/blog/OneBlog.vue')
const Resources = () => import('../views/detail/resources/Resources.vue')
const Community = () => import('../views/detail/community/Community.vue')
const FriendLink = () => import('../views/detail/friendLink/FriendLink.vue')
const YnsStudyAi = () => import('../views/detail/ai/YnsStudyAi.vue')
const About = () => import('../views/detail/about/About.vue')
const ContentAndComment = () => import('@/views/detail/blog/ContentAndComment.vue')
const personalCenter = () => import('@/views/main/user/personalCenter.vue')
const personInfomation = () => import('@/views/main/user/personInformation.vue')
const personInfomationV2 = () => import('@/views/main/user/personInformationV2.vue')
const sso = () => import('@/views/main/sso/sso.vue')
const RssDetailView = () => import('@/views/detail/rss/RssDetailView.vue')
const LuLu = () => import('@/views/detail/z_lulu/LuLu.vue')
const NotFound = () => import('@/views/main/NotFound.vue')

const routes = [
    //噜噜
    {path:'/lulu' ,name:'lulu',component: LuLu},

    //后台管理
    {path:'/sso' ,name:'sso',component: sso, meta: {noindex: true}},
    //根路径重定向到welcome
    //{path: '/', redirect: '/welcome'},
    {path: '/', name: 'Welcome', component: Welcome},
    //欢迎页面
    {path: '/welcome', name: 'WelcomePage', component: Welcome},
    //个人信息
    {path: '/personalCenter', name: 'personalCenter', component: personalCenter, meta: {noindex: true}},
    //账号主页   u为传递的参数
    {path:'/personInfomation/:userId',name:'personInfomation',component: personInfomation},
    //简短路径
    {path:'/user/:u/:blogId?',name:'user',component: personInfomation},
    //账号主页   u为传递的参数
    {path:'/userV2/:u/:blogId?',name:'userV2',component: personInfomationV2},
    //Ai
    {path: '/YnsStudyAi', name: 'YnsStudyAi', component: YnsStudyAi},
    //展示一条博客内容
    {path:'/oneBlog/:g', name: 'oneBlog', component: OneBlog},
    //展示一条博客内容
    {path:'/rss', name: 'rss', component: RssDetailView, meta: {noindex: true}},
    //导航栏菜单
    {
        path: '/ynsStudy',
        component: Index,  // Index 作为父组件
        children: [
            {path: '', redirect: '/ynsStudy/Home'},
            {path: 'Home', name: 'Home', component: Home},
            {
                path: 'MyBlog',
                name: 'MyBlog',
                component: MyBlog,
                // 登录用户自己的写作空间，未登录时为空页面，不参与收录
                meta: {noindex: true},
                children: [{
                    path: 'content', name: 'BlogContent', component: ContentAndComment, meta: {noindex: true}
                }]
            },
            {path: 'Resources', name: 'Resources', component: Resources},
            {path: 'Community', name: 'Community', component: Community},
            {path: 'FriendLink', name: 'FriendLink', component: FriendLink},
            {path: 'About', name: 'About', component: About}
        ]
    },
    //未匹配的地址，Nginx 同时返回 404 状态码
    {path: '/:pathMatch(.*)*', name: 'NotFound', component: NotFound, meta: {noindex: true}}
]

// 开发环境专用：对比文章两种展示方式，打包后不包含
if (import.meta.env.DEV) {
    routes.splice(routes.length - 1, 0, {
        path: '/dev/article-compare/:id?',
        name: 'devArticleCompare',
        component: () => import('@/views/dev/ArticleCompare.vue'),
        meta: {noindex: true}
    })
}

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes
})

export default router
