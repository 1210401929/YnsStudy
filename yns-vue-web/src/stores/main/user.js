import {defineStore} from 'pinia';
import {ref} from 'vue';
import {sendAxiosRequest} from "@/utils/common.js";

export const useUserStore = defineStore('user', () => {
    const userBean = ref({});

    const setUser = (userObj) => {
        userBean.value = userObj;
        localStorage.setItem('userBean', JSON.stringify(userObj));
    };

    const clearUser = async () => {
        await sendAxiosRequest('/pub-api/login/logout');
        userBean.value = {};
        localStorage.removeItem('userBean');
    };

    const initFromLocal = async () => {
        const result = await sendAxiosRequest('/pub-api/login/checkUserLogin');
        if (result && result.result) {
            setUser(result.result);
        }
    };

    return {
        userBean,
        setUser,
        clearUser,
        initFromLocal
    };
});
