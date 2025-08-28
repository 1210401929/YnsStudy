import {encrypt, sendAxiosRequest} from "@/utils/common.js";

export const pubOpenOneBlog = (router, blogGuid) => {
    const routeUrl = router.resolve({name: "oneBlog", params: {g: blogGuid}}).href;
    window.open(routeUrl, blogGuid);
}

export const pubOpenUser = (router, userCode) => {
    const routeUrl = router.resolve({name: 'personInfomation', params: {u: encrypt(userCode)}}).href;
    window.open(routeUrl, userCode);
}

export const getAnnouncementByRouterName = async (routerName) => {
    let result = await sendAxiosRequest("/blog-api/sso/getAnnouncementByType", {type: routerName});
    return result.result;
}