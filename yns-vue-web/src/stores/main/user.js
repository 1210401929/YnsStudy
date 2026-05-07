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

    const clearUser = async () => {
        await sendAxiosRequest('/pub-api/login/logout');
        userBean.value = {};
        localStorage.removeItem('userBean');
        localStorage.removeItem('userToken');
    };

    const initFromLocal = async () => {
        let result = await sendAxiosRequest('/pub-api/login/checkUserLogin');
        if (result && result.result) {
            setUser(result.result);
            result = await sendAxiosRequest("/pub-api/notice/getNotice", {userCode: userBean.value.code});
            userUnreadArr.value = result.result;
        }
    };

    return {
        userBean,
        userUnreadArr,
        setUser,
        clearUser,
        initFromLocal
    };
});
