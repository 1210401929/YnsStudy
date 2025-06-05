const app = Vue.createApp({
    setup() {
        let refDate = ref(0);
        let existenceFileList = Y.util.ajax("/upload/getUserFile");
        existenceFileList = ref(existenceFileList.result);

        watch(refDate, () => {
            debugger;
            let result = Y.util.ajax("/upload/getUserFile");
            existenceFileList.value = result.result;
        });

        const fileList = ref([]);
        const handlePreview = (file) => {
            debugger;
        };
        const handleRemove = (file, fileList) => {
            debugger;
        };
        const handleSuccess = (response, file, fileList) => {
            refDate.value++;
            Y.util.showMessage("上传成功!");
        };
        const handleExceed = (files, fileList) => {
            debugger;

        };
        const downloadFile = (item) => {
            debugger;
            const oriFileName = item["ORIGINALFILENAME"]
            const fileName = item["STOREDFILENAME"];
            const fileDate = item["CREATE_TIME"];
            const date = formatDate(fileDate);
            const url = `/upload/downloadFile/${date}/${fileName}/${oriFileName}`;
            //window.location.href = url;
            // 创建一个隐藏的a标签
            const a = document.createElement('a');
            a.href = url;
            a.download = oriFileName; // 设置下载文件的名称
            // 将a标签添加到文档中，然后点击它
            document.body.appendChild(a);
            a.click();
            // 点击后移除a标签
            document.body.removeChild(a);
        };
        const deleteFile = (item) => {
            debugger;
            Y.util.confirm(vm, "是否确认删除?", (() => {
                let result = Y.util.ajax("/upload/deleteFile", {guid: item["GUID"]});
                if (result && !result.isError) {
                    refDate.value++;
                    Y.util.showMessage("删除成功!");
                }
            }));
        };
        const actionUrl = computed(() => {
            return "/upload/uploadFile";
        })
        return {
            existenceFileList,
            fileList,
            handlePreview,
            handleRemove,
            handleSuccess,
            handleExceed,
            deleteFile,
            downloadFile,
            actionUrl
        };
    }
})

var vm = app.use(ElementPlus).mount("#upload");

function formatDate(fileDate) {
    const date = new Date(fileDate);
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
}