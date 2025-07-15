const menuItems = [
    {name: '首页', router: 'Home', path: '/ynsStudy/Home'},
    {name: '发表', router: 'MyBlog', path: '/ynsStudy/MyBlog'},
    {name: '资源', router: 'Resources', path: '/ynsStudy/Resources'},
    {name: '社区', router: 'Community', path: '/ynsStudy/Community'},
    {name: '关于', router: 'About', path: '/ynsStudy/About'}
]

export function getMenuItems(){
    return menuItems;
}