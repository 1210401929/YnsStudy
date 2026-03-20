import {defineStore} from "pinia";
import {ref} from 'vue';
import {sendAxiosRequest} from "@/utils/common.js";

export const useBlogContentStore = defineStore('BlogContent', () => {
    //正文
    const blogContents = ref([]);
    //评论
    const blogComments = ref({});
    const setBlogContent = (content) => {
        blogContents.value = content;
    }

    const getBlogTitles = (info) => {
        return blogContents.value.map(item => item[info]);
    }
    const clearBlogContent = () => {
        blogContents.value = [];
    }
    const initBlogContent = async (userCode) => {
        //没有传递用户编码，查询当前用户
        if (!userCode) {
            //获取用户的所有文章
            let result = await sendAxiosRequest("/blog-api/blog/getUserBlog", {});
            if (result && !result.isError) {
                blogContents.value = result.result;
            }
            //查询指定用户文章数据
        } else {
            const res = await sendAxiosRequest('/blog-api/blog/getUserBlogByUserCode', {userCode});
            if (res && !res.isError) {
                blogContents.value = res.result;
            }
        }
    }

    return {
        blogContents, getBlogTitles, setBlogContent, clearBlogContent, initBlogContent
    }
});