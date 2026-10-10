// title / description 用于各栏目页的搜索结果标题和摘要
const menuItems = [
    {
        name: '首页', router: 'Home', path: '/ynsStudy/Home',
        title: '最新文章与热门推荐 - YnsStudy',
        description: 'YnsStudy 最新发布的技术文章、学习笔记与热门推荐，涵盖编程学习、技术实践与生活思考。'
    },
    {
        name: '发表', router: 'MyBlog', path: '/ynsStudy/MyBlog',
        title: '发表文章 - YnsStudy',
        description: '在 YnsStudy 发表和管理自己的文章。'
    },
    {
        name: '资源', router: 'Resources', path: '/ynsStudy/Resources',
        title: '学习资源分享 - YnsStudy',
        description: 'YnsStudy 用户分享的学习资料、工具与文档资源，按分类整理，方便查找和下载。'
    },
    {
        name: '社区', router: 'Community', path: '/ynsStudy/Community',
        title: '社区动态 - YnsStudy',
        description: 'YnsStudy 社区：分享日常、交流学习心得，与其他热爱写作和编程的人互动讨论。'
    },
    {
        name: '友链', router: 'FriendLink', path: '/ynsStudy/FriendLink',
        title: '友情链接 - YnsStudy',
        description: 'YnsStudy 的友情链接，收录优质的个人博客与技术站点，欢迎交换友链。'
    },
    {
        name: '关于', router: 'About', path: '/ynsStudy/About',
        title: '关于 YnsStudy',
        description: 'YnsStudy 是一个集文章创作、评论互动与资源分享于一体的知识社区，了解本站的由来、功能与联系方式。'
    }
]

export function getMenuItems(){
    return menuItems;
}

