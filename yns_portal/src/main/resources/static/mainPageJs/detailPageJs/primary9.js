const app = Vue.createApp({
    setup() {
        let messages = ref([
            {role: 'user', content: '你好！'},
            {role: 'assistant', content: '你好！有什么可以帮您的吗？您可以尽情提问!'}
        ]);
        let newMessage = ref('');
        let sendMessage = () => {

            if (newMessage.value.trim() !== '') {
                messages.value.push({ role: 'user', content: newMessage.value });
                let newMessage_ = new String(newMessage.value);
                newMessage.value = '';
                $.ajax({
                    url:"/AI/BaiduAi",
                    type:"post",
                    data:{
                        content: newMessage_
                    },
                    async: true,
                    success:(result)=>{
                        messages.value.push({ role: 'assistant', content: result.result });
                        Vue.nextTick(() => {
                            const messagesContainer = document.querySelector('.messages');
                            messagesContainer.scrollTop = messagesContainer.scrollHeight;
                        });
                    }
                });
            }
        }

        return {messages, newMessage, sendMessage};
    }
});

var vm = app.use(ElementPlus).mount("#release");