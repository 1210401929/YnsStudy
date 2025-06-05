window["button_component"] = defineComponent({
    props:["info"],
    setup(props) {
        // 使用 onMounted 定义生命周期钩子
        onMounted(() => {
            console.log('Component mounted');
        });
        // 定义 showMessage 方法
        function showMessage() {
            alert(props.info.btnName);
        }
        // 返回模板中需要的数据和方法
        return {
            showMessage,
        };
    },
    template: `
            <div>
                <button @click="showMessage">{{ info.btnName }}</button>
            </div>
        `
});