var sourceObj={};
var typeCode=getTypeCode();
var dataSource=getDataSource();
/**
 * 初始化函数
 */
window.onload=function() {
    debugger
    YCDCommon.loading("on_load")
}
/**
 * 初始化函数
 */
function on_load() {
    registerParent();
}

/**
 * 注册父组件
 */
function registerParent() {
    debugger
    window.vm = new Vue({
        //绑定元素
        el: "#fastbuild",
        //数据对象
        data:{
            modelType:"FORM",//模板类型
            modelOptions:[{//模板项
                value:"FORM",
                label:"表单"
            },{
                value:"BJ",
                label:"布局"
            }],
            tableColumObj:{//列对象[{prop:"",label:"",editor:"",code:"",group:false,visible:"",formatter:""}]
                FORM:[{
                    prop:"type",
                    label:"控件类型",
                    editor:"el-select",
                    code:"typeCode",
                    group:true
                },{
                    prop:"ds_ctrlname",
                    label:"控件名",
                    editor:"el-input",
                    code:""
                },{
                    prop:"ds_labeltxt",
                    label:"标签名",
                    editor:"el-input",
                    code:""
                },{
                    prop:"ds_placeholder",
                    label:"描述",
                    editor:"el-input",
                    code:""
                },{
                    prop:"ds_width",
                    label:"宽度(0-24)",
                    editor:"el-input",
                    code:""
                },{
                    prop:"ds_datasource",
                    label:"数据源",
                    editor:"el-select",
                    code:"dataSource"
                },{
                    prop:"ds_name",
                    label:"字段名",
                    editor:"el-select",
                    code:"fieldCode",
                    visible:"getField"

                }]
            },
            tableData: [],//表数据
            currRowIndex: "",//行号
            currColField: "",//当前列字段
            currRow:{},
            codeJson:{//码表数据
                typeCode:typeCode,
                dataSource:dataSource,
                fieldCode:[],
            },

        },
        //被创建后执行的方法
        created:function () {
        },
        //计算属性
        computed:{

        },
        //侦听属性 newval,oldval
        watch:{
        },
        //方法
        methods:{
            commonFunction:$DS.util.commonFunction,
            selectionChange:function (selection) {//行变更事件

            },
            loseInputFcous: function (scope,flag) {//单元格输入框失去焦点事件,
                scope.row.seen = false;
                this.currRowIndex = -1;
                this.currColField='';
                var field = scope.column.columnKey;
                var editValue = scope.row[field];
                scope['val'] = editValue;
                var obj={};
                obj[field]=editValue;
                this.onAfterEdit(this,scope.row,obj);
            },
            rowClassName:function ({row, rowIndex}) {//行的 className 的回调方法
                //把每一行的索引放进row
                row.index = rowIndex;
            },
            cellClick:function (row, column, cell, event) {//单元格点击事件
                this.currColField = column.columnKey
                row.seen = true;
                this.$nextTick(function () {
                    $(cell).find("input").focus();
                    $(cell).find("input").keydown(function(e){
                        var curKey = e.which;
                        if (curKey == 13) {
                            $(cell).find("input").blur();
                            row.seen = false;
                        }
                    });
                });
            },
            clickGridRow:function (row) {//行点击事件 || 行变更事件
                if(this.currRow.seen){
                    this.currRow.seen=false;
                }
                this.getIndex = row.index
                this.currRowIndex = row.index;
                this.currRow=row;
            },
            selectRow:function (selection, row) {//行选择事件
                //自定义函数 行选择事件
            },
            onAfterEdit:function (vue,row,change) {//结束编辑事件
                //getField(vue,row,change)

            },
            selectChange:function(val){
                var obj={};
                this.currRow.seen=true;
                obj[this.currColField]=val;
                if(this.currRow.seen){
                    this.currRow.seen=false;
                    this.currRowIndex = -1;
                    this.currColField='';
                }
                this.onAfterEdit(this,this.currRow,obj);
            },
            addData:function () {
                if(this.modelType=="FORM"){
                    var row={};
                    for(var d=0;d<this.tableColumObj.FORM.length;d++){
                        row[this.tableColumObj.FORM[d].prop]="";
                    }
                    this.tableData.push(row);
                }
            },
            stateFormat:function (value,col) {
                debugger
                if(col.code&&col.editor=="el-select"){
                    var codeArr=this.$data.codeJson[col.code];
                    if(col.group){
                        for(var g=0;g<codeArr.length;g++){
                            var options=codeArr[g].options;
                            for(var o=0;o<options.length;o++){
                                if(options[o].value==value){
                                    return options[o].label;
                                }
                            }
                        }
                    }else{
                        for(var c=0;c<codeArr.length;c++){
                            if(codeArr[c].value==value){
                                return codeArr[c].label;
                            }
                        }
                    }
                }
                return value;
            }
        }
    })
}

/**
 * 获取控件类型数据
 * @returns {*}
 */
function getTypeCode() {
    var except={};
    var typeArr=parent.getAllCtrl();
    for(var t=0;t<typeArr.length;t++){
        if(typeArr[t].label.trim()=="布局型组件"){
            typeArr.splice(t,1);
            continue;
        };
        for(var o=0;o<typeArr[t].options.length;o++){
            if(except[typeArr[t].options[o].value]){
                typeArr[t].options.splice(o,1);
            }
        }
    }
    return typeArr;
}

/**
 * 获取数据源
 * @returns {[]}
 */
function getDataSource() {
    var dataSourceArr=[];
    var sourceArr=parent.VUECFG.sourceArr;
    for(var s=0;s<sourceArr.length;s++){
        sourceObj[sourceArr[s].sourceId]=sourceArr[s];
        dataSourceArr.push({
            label:sourceArr[s].sourceName,
            value:sourceArr[s].sourceId
        })
    }
    return dataSourceArr;
}

/**
 * 获取字段
 * @param params
 */
function getField(params) {
    var row=params.vue.$data.currRow;
    if(row&&row.ds_datasource){
        var fieldArr=[];
        var source=sourceObj[row.ds_datasource];
        var fieldInfo=source.fieldInfo;
        for(key in fieldInfo){
            fieldArr.push({
                label:fieldInfo[key].FIELD_NAMECN,
                value:key
            })
        }
        params.vue.$data.codeJson.fieldCode=fieldArr;
    }else{
        params.vue.$data.codeJson.fieldCode=[];
    }

}