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
    const initBlogContent = async () => {
        //获取用户的所有文章
        let result = await sendAxiosRequest("/blog-api/blog/getUserBlog", {});
        if (result && !result.isError) {
            blogContents.value = result.result;
        }
    }

    return {
        blogContents, getBlogTitles, setBlogContent, clearBlogContent, initBlogContent
    }
});