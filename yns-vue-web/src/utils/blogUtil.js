import {encrypt, sendAxiosRequest} from "@/utils/common.js";

export const pubOpenOneBlog = (router, blogGuid) => {
    const routeUrl = router.resolve({name: "oneBlog", params: {g: blogGuid}}).href;
    window.open(routeUrl, blogGuid);
}

export const pubOpenUser = async (router, userCode) => {
    debugger;
    let result = await sendAxiosRequest("/pub-api/login/getUserInfoByCode", {userCode});
    if(result && result.result){
        //const routeUrl = router.resolve({name: 'user', params: {u: encrypt(userCode)}}).href;
        const routeUrl = router.resolve({name: 'user', params: {u: result.result.usernum}}).href;
        window.open(routeUrl, userCode);
    }
}

export const getAnnouncementByRouterName = async (routerName) => {
    let result = await sendAxiosRequest("/blog-api/sso/getAnnouncementByType", {type: routerName});
    return result.result;
}