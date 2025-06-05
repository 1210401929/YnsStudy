const {createApp, ref,reactive, onMounted,computed,nextTick, defineProps, defineComponent,watch} = Vue;

function mainInit() {


}

mainInit();

//注册组件
function registrationComponent(vm, types) {
    types.forEach(type => {
        if (window[`${type}_component`])
            vm.component(`${type}_component`, window[`${type}_component`]);
    });
}
