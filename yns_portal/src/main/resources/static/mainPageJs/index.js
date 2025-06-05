function initIndex() {
    //检查页面
    if (!checkPageInit()) return false;
    //获取菜单数据
    window["menuData"] = Y.util.ajax("/pub-api/authority/getUserMenu");
    window["menuData"] = window["menuData"].result;
    //获取菜单对应路径数据
    let menuDataArr = Y.util.flattenChildren(window["menuData"]);
    window["menuIndex2Url"] = new Object();
    menuDataArr.forEach(item=>window["menuIndex2Url"][item["index"]] = item["url"]);
    //初始化vue组件
    initVueComponents();
}

initIndex();

function initVueComponents() {
    const {createApp, ref, onMounted, defineExpose, defineComponent} = Vue;
    //Vue.defineComponent('')
    createApp({
        name: "menu_Y",
        setup() {
            //默认选择节点
            let active = ref("1");
            //记录菜单实例
            let menuRef = ref(null);
            //菜单数据
            let menuData = ref(window["menuData"]);
            //iframe的src
            let iframeUrlY = ref("");
            //用户信息
            let userInfo = ref(window["userInfo"]);

            //fun
            function handleSelect(index, indexPath, item) {
                if (window["menuIndex2Url"][index]) {
                    iframeUrlY.value = window["menuIndex2Url"][index];
                } else {
                    iframeUrlY.vlaue = null;
                }
            }

            function logout() {
                Y.util.ajax("/logout");
                window.location.href = `/login`;
            }

            function personCenterCommand(command) {
                switch (command) {
                    case "1": {
                        //个人中心
                        break;
                    }
                    case "2": {
                        //修改密码
                        Y.util.showPage(`../mainPage/changePassword.html`,`修改密码`,`80%`,`80%`);
                        break;
                    }
                    case "3":{
                        //注册账号
                        Y.util.showPage(`../mainPage/register.html`,`修改密码`,`80%`,`80%`);
                        break;
                    }
                    case "4":{
                        //登录
                        window.location.href = "/login";
                        break;
                    }
                }
            }

            // onMounted钩子函数
            onMounted(() => {
                //记录菜单实例
                window["ref_menu"] = menuRef.value;
                //默认执行选中第一个
                handleSelect(active.value);
            });

            return {active, menuRef, menuData, iframeUrlY, userInfo, handleSelect, logout, personCenterCommand};
            //暴露属性,参数为允许访问的属性
            defineExpose({active, menuRef, menuData, iframeUrlY, handleSelect});
        }
    }).use(ElementPlus).mount('#app');
}

function checkPageInit() {
    //检查用户是否登录
    let result = Y.util.ajax("/checkUserLogin");
    if (!result || result.isError) {
        console.error(result.errMsg);
        window.location.href = `/login`;
        return false;
    }
    window["userInfo"] = result.result;
    window.top["userInfo"] = result.result;
    return true;
}

