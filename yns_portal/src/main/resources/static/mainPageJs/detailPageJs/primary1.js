//初始化数据
getData2Descriptions();

const app = createApp({
    setup() {
        const buttonInfo = {
            btnName: "按钮组件Vue"
        };
        const descriptionsInfo = reactive({
            descriptions_data: window["descriptions_data"],
            editData: function (item) {
                Y.util.showPage(`primary5.html?paperGuid=${item["GUID"]}`, `编辑内容`, `80%`, `80%`);

            },
            deleteData: function (item) {
                Y.util.confirm(vm, `是否确认删除`, function () {
                    let result = Y.util.ajax("/operation/deletePaper", {paperGuid: item["GUID"]});
                    if (result && !result.isError) {
                        getData2Descriptions();
                        descriptionsInfo.descriptions_data = window["descriptions_data"];
                        Y.util.showMessage("执行成功!");
                        return true;
                    } else {
                        alert("执行失败!" + result.errMsg);
                        return false;
                    }
                });
            }
        });
        return {
            buttonInfo,
            descriptionsInfo
        };
    }
});
registrationComponent(app, [`button`, `descriptions`]);
var vm = app.use(ElementPlus).mount('#app');


function getData2Descriptions() {
    debugger;
    let pagePms = Y.util.getUrlPms();
    let result = Y.util.ajax("/operation/getPaperByType", {type: pagePms["type"] || "Privacy"});
    if (result && result.result) {
        window["descriptions_data"] = result.result;
    }
}