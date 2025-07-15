window["descriptions_component"] = defineComponent({
    props: ["info"],
    setup(props) {
        let isShowComputed = (item) => {

            if (window.top.userInfo.guid == item.USERID)
                return true;
            else
                return false;
        };
        // 返回模板中需要的数据和方法
        return {
            isShowComputed
        };
    },
    template: `
    <template v-for="item in info.descriptions_data">
      <el-descriptions :title="item.BLOG_TITLE">
        <el-descriptions-item >
            <!--pre标签可以显示内容的换行符以及空白区域-->
            <pre v-html="item.MAINTEXT"></pre>
            <el-tag size="small">{{item.USERNAME}}</el-tag> &nbsp;
            <el-button type="primary" plain size="small" v-if=isShowComputed(item) @click="info.editData(item)">编辑</el-button>
            <el-button  type="danger" plain size="small" v-if=isShowComputed(item) @click="info.deleteData(item)">删除</el-button>
        </el-descriptions-item>
    </el-descriptions>
    </template>
    `
});
