const app = createApp({
    setup(){
        const articles = ref([
            { id: 1, title: '热门文章1' },
            { id: 2, title: '热门文章2' },
            { id: 3, title: '热门文章3' }
        ]);

        const comments = ref([
            { id: 1, content: '评论1' },
            { id: 2, content: '评论2' },
            { id: 3, content: '评论3' }
        ]);

        return { articles, comments };
    }
});

var vm = app.use(ElementPlus).mount('#app');