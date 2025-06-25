import {createRouter, createWebHistory} from 'vue-router'
import Welcome from '../views/main/Welcome.vue'
import Index from '../views/main/index.vue'
import Home from '../views/detail/home/Home.vue'
import Blog from '../views/detail/blog/Blog.vue'
import Resources from '../views/detail/resources/Resources.vue'
import Community from '../views/detail/community/Community.vue'
import YnsStudyAi from '../views/detail/ai/YnsStudyAi.vue'

import About from '../views/detail/about/About.vue'
import ContentAndComment from "@/views/detail/blog/ContentAndComment.vue";

import personalCenter from "@/views/main/user/personalCenter.vue";

import LoveStory from "@/views/detail/love/LoveStory.vue";

const routes = [
    //临时
    {path:'/loveLY' ,name:'love',component: LoveStory},
    //根路径重定向到welcome
    {path: '', redirect: '/welcome'},
    {path: '/welcome', name: 'Welcome', component: Welcome},
    {path: '/personalCenter', name: 'personalCenter', component: personalCenter},
    {path: '/YnsStudyAi', name: 'YnsStudyAi', component: YnsStudyAi},
    {
        path: '/ynsStudy',
        component: Index,  // Index 作为父组件
        children: [
            {path: '', redirect: '/ynsStudy/Home'},
            {path: 'Home', name: 'Home', component: Home},
            {
                path: 'Blog',
                name: 'Blog',
                component: Blog,
                children: [{
                    path: 'content', name: 'BlogContent', component: ContentAndComment
                }]
            },
            {path: 'Resources', name: 'Resources', component: Resources},
            {path: 'Community', name: 'Community', component: Community},
            {path: 'About', name: 'About', component: About}
        ]
    }
]

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes
})

export default router
