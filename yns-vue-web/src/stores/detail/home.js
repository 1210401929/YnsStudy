import {defineStore} from "pinia";
import {ref} from "vue";
import {sendAxiosRequest} from "@/utils/common.js";

export const useHomeStore = defineStore("Home", () => {
    //博客热门文章
    const homeData = ref({});

    const initHomeData = async () => {
        let result = await sendAxiosRequest("/blog-api/home/getHomeData");
        if (result && !result.isError) {
            homeData.value = result.result;
        }
    }

    return {homeData, initHomeData};
})