import {defineStore} from 'pinia';
import {ref} from 'vue';
import {sendAxiosRequest} from "@/utils/common.js";

export const useUserStore = defineStore('user', () => {
    //用户信息
    const userBean = ref({});
    //用户未读通知
    const userUnreadArr = ref([]);

    const setUser = (userObj) => {
        const user = userObj.user;
        const userToken = userObj.userToken;
        userBean.value = user;
        localStorage.setItem('userBean', JSON.stringify(user));
        localStorage.setItem('userToken', userToken);
    };

    const resetLocalUser = () => {
        userBean.value = {};
        userUnreadArr.value = [];
        localStorage.removeItem('userBean');
        localStorage.removeItem('userToken');
    };

    const clearUser = async () => {
        try {
            await sendAxiosRequest('/pub-api/login/logout');
        } finally {
            resetLocalUser();
        }
    };

    const initFromLocal = async () => {
        try {
            let result = await sendAxiosRequest('/pub-api/login/checkUserLogin');
            if (result && result.result && !result.isError) {
                setUser(result.result);
                result = await sendAxiosRequest("/pub-api/notice/getNotice", {userCode: userBean.value.code});
                userUnreadArr.value = result.result || [];
                return;
            }
        } catch (error) {
            console.warn('恢复登录状态失败:', error);
        }
        resetLocalUser();
    };

    return {
        userBean,
        userUnreadArr,
        setUser,
        resetLocalUser,
        clearUser,
        initFromLocal
    };
});
