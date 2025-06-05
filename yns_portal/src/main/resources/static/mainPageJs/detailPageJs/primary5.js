const app = Vue.createApp({
    setup() {
        let submitBtnName = ref("发布");
        const options = [{label: "私密", value: "Privacy"}, {label: "公开", value: "open"}];
        const form = reactive({
            blog_title: '',
            region: '',
            brieText: '',
            mainText: '',
        });
        const primaryRef = ref("");

        function onSubmit() {
            let urlPms = Y.util.getUrlPms();
            let result = Y.util.ajax("/operation/savePaper", {
                paperData: JSON.stringify(form),
                paperGuid: urlPms["paperGuid"]
            });
            if (result && !result.isError) {
                if (urlPms["paperGuid"]) {
                    parent.getData2Descriptions();
                    parent.vm.descriptionsInfo.descriptions_data = parent.window["descriptions_data"];
                    Y.util.close();
                } else {
                    Y.util.showMessage("执行成功!");
                    form.blog_title = "", form.region = "", form.brieText = "", form.mainText = "";
                }
            } else {
                alert(result.errMsg);
            }
        }

        onMounted(() => {
            let urlPms = Y.util.getUrlPms();
            if (urlPms["paperGuid"]) {
                submitBtnName.value = "修改";
                let result = Y.util.ajax("/operation/getPaperOne", {paperGuid: urlPms["paperGuid"]});
                if (result.result) {
                    let paper = result.result[0];
                    form.blog_title = paper.BLOG_TITLE, form.region = paper.REGION, form.brieText = paper.BRIETEXT, form.mainText = paper.MAINTEXT;
                }
            }
        });
        return {submitBtnName, primaryRef, options, form, onSubmit};
    }
});

var vm = app.use(ElementPlus).mount("#release");