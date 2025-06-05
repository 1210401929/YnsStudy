//是否关闭调试功能(只是屏蔽了F12和鼠标右键  仍有办法破解)
let prohibitDebugger = false;
let Y = new Object();
Y.util = new Object();

//----------------------------------------文档------------------------------------------------------------------------------
/**
 *  Y.util.showMessage  弹出消息提示   1秒后消失
 *  Y.util.getUrlPms  获取路径参数对象
 *  Y.util.ajax 请求后台
 *  Y.util.close 关闭弹窗
 *  Y.util.confirm 确认函数
 *  Y.util.showPage 弹窗
 *  Y.util.flattenChildren  层级数组转平级数组
 */
//----------------------------------------实现------------------------------------------------------------------------------
Y.util.flattenChildren = function (nodes) {
    const result = [];
    function traverse(nodeList) {
        nodeList.forEach(node => {
            result.push(node);
            if (node.children && node.children.length > 0) {
                traverse(node.children);
            }
        });
    }
    traverse(nodes);
    return result;
}
/**
 * @param content  正文-必填
 * @param backgroundColor  背景颜色-默认黑色
 * @param color 字体颜色-默认白色
 * @param millisecond 几秒后消失-默认一秒
 */
Y.util.showMessage = function (content, backgroundColor, color, millisecond) {
    if (!content) return false;
    // 创建消息框元素
    var message = document.createElement("div");
    message.textContent = content;
    message.style.position = "fixed";
    message.style.top = "50%";
    message.style.left = "50%";
    message.style.transform = "translate(-50%, -50%)";
    message.style.padding = "10px 20px";
    message.style.backgroundColor = backgroundColor || "rgba(0, 0, 0, 0.7)";
    message.style.color = color || "#fff";
    message.style.borderRadius = "5px";
    message.style.zIndex = "9999";
    // 将消息框添加到页面中
    document.body.appendChild(message);
    // 3秒后移除消息框
    setTimeout(function () {
        document.body.removeChild(message);
    }, millisecond || 1000);
}

Y.util.getUrlPms = function () {
    let params = new Object(), url, pmsArr, i;
    try {
        url = window.location.href;
        url = url.split("?")[1];
        pmsArr = url.split("&");
        pmsArr.forEach(item => {
            i = item.split("=");
            params[i[0]] = i[1];
        });
        return params;
    } catch (e) {
        return params;
    }
}

Y.util.ajax = function (_url, _data, ajaxType = "post", async = false) {
    try {
        _data = _data || {};
        _data.isAjax = true;
        var result = $.ajax({
            type: ajaxType, data: _data, cache: false,
            xhrFields: {
                withCredentials: true  // 允许跨域携带 cookies
            }, url: _url, async: async
        }).responseJSON;
        return result;
    } catch (err) {
        return null;
    }
};

Y.util.close = function () {
    if (!window.name) {
        if (parent.$("[id^='layui-layer-iframe']").length > 0)
            window.name = parent.$("[id^='layui-layer-iframe']")[0].name;
    }
    var index = parent.layer.getFrameIndex(window.name); // 先得到当前iframe层的索引
    if (!index) {
        index = parent.layer.getFrameIndex(parent.$("[id^='layui-layer-iframe']")[0].name);
    }
    parent.layer.close(index); // 再执行关闭
}

Y.util.confirm = function (vue, tip, sureFun, cancleMsg, pms, cancleFun, topWin) {
    if (!cancleMsg) cancleMsg = "已取消!"
    var win = vue;
    if (topWin) win = topWin;
    let opt = {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        cancelButtonClass: 'btn-custom-cancel',
        type: 'warning'
    }
    if (pms && pms.removeCancle) {
        opt.cancelButtonClass = 'y-hide'
    }
    win.$confirm(tip, '提示', opt).then(() => {
        sureFun(vue, pms);
    }).catch((e) => {
        if (cancleFun)
            cancleFun(vue, pms);
        win.$message({
            type: 'info',
            message: cancleMsg
        });
        console.error("【确认函数执行异常】", {vue, tip, sureFun, cancleMsg, pms, cancleFun, topWin, e})
    });
}


var pageCloseFlag = true;
Y.util.showPage = function (url, title, width, height, args, isOpen, win) {
    if (!win) win = window;
    if (!url) {
        alert("请传入url地址");
        return false;
    }
    url = encodeURI(url);
    if (!args) args = {};
    if (isOpen) {
        window.open(url, title)
    } else {
        showMyDialog(title, width, height, url, args.callBack, args.hiddTitle, args.time, args.beginClose || '_layBeginClose')
    }

    function showMyDialog(title, width, height, url, fn, hiddTitle, time, beginClose) {
        if (typeof width == 'string') {
            if (width.indexOf("%") > -1 || width.indexOf("px") > -1) {
            } else
                width = parseInt(width) + 'px';
        } else if (typeof width === 'number')
            width = width + "px";

        if (typeof height == 'string') {
            if (height.indexOf("%") > -1 || height.indexOf("px") > -1) {
            } else
                height = parseInt(height) + 'px';
        } else if (typeof height === 'number')
            height = height + "px";
        if (!time) time = 0;
        var options = {
            type: 2,
            title: [
                title,
                "font-size:15px;color:#fff;background:#2e62bf;background:-moz-linear-gradient(top, #73a6fe 0%, #2e62bf 100%);background:-webkit-gradient(linear, left top, left bottom, color-stop(0%,#73a6fe), color-stop(100%,#2e62bf));background:-webkit-linear-gradient(top, #73a6fe 0%,#2e62bf 100%);background:-o-linear-gradient(top, #73a6fe 0%,#2e62bf 100%);background:-ms-linear-gradient(top, #73a6fe 0%,#2e62bf 100%);background:linear-gradient(to bottom, #73a6fe 0%,#2e62bf 100%);filter: progid:DXImageTransform.Microsoft.gradient(startColorstr='#73a6fe',endColorstr='#2e62bf',GradientType=0);"
            ],
            area: [width, height],
            content: [url],
            skin: 'layer-ext-myskin',
            time: time,
            resize: true,
            cancel: function () {
                if (beginClose && $("[id^='layui-layer-iframe']")[0].contentWindow[beginClose]) {
                    pageCloseFlag = $("[id^='layui-layer-iframe']")[0].contentWindow[beginClose]();
                }
                var last_Index = layer.getFrameIndex($("[id^='layui-layer-iframe']")[0].name);
                if (!pageCloseFlag) {
                    layer.confirm('数据未保存,确定要关闭么?', function (index, layero) {
                        layer.close(index);
                        layer.close(last_Index);

                    });
                    return false;
                }
            },
            end: function () {
                if (fn || fn != null)
                    fn();
            }
        };
        if (hiddTitle) {
            //隐藏标题
            options.closeBtn = 0;
            options.title[1] = "height:0px;font-size:15px;color:#fff;background:#2e62bf;background:-moz-linear-gradient(top, #73a6fe 0%, #2e62bf 100%);background:-webkit-gradient(linear, left top, left bottom, color-stop(0%,#73a6fe), color-stop(100%,#2e62bf));background:-webkit-linear-gradient(top, #73a6fe 0%,#2e62bf 100%);background:-o-linear-gradient(top, #73a6fe 0%,#2e62bf 100%);background:-ms-linear-gradient(top, #73a6fe 0%,#2e62bf 100%);background:linear-gradient(to bottom, #73a6fe 0%,#2e62bf 100%);filter: progid:DXImageTransform.Microsoft.gradient(startColorstr='#73a6fe',endColorstr='#2e62bf',GradientType=0);";
        } else {
            //显示标题
            options.closeBtn = 1;
            options.title[1] = "font-size:15px;color:#fff;background:#2e62bf;background:-moz-linear-gradient(top, #73a6fe 0%, #2e62bf 100%);background:-webkit-gradient(linear, left top, left bottom, color-stop(0%,#73a6fe), color-stop(100%,#2e62bf));background:-webkit-linear-gradient(top, #73a6fe 0%,#2e62bf 100%);background:-o-linear-gradient(top, #73a6fe 0%,#2e62bf 100%);background:-ms-linear-gradient(top, #73a6fe 0%,#2e62bf 100%);background:linear-gradient(to bottom, #73a6fe 0%,#2e62bf 100%);filter: progid:DXImageTransform.Microsoft.gradient(startColorstr='#73a6fe',endColorstr='#2e62bf',GradientType=0);";
        }
        layer.config({
            //extend: 'myskin/style.css'
        });
        layer.open(options);
    }
}


//屏蔽F12 ,
document.onkeydown = function (event) {
    event = event || window.event;
    if (event.keyCode == 123 && prohibitDebugger) {
        Y.util.showMessage("禁止使用F12查看源代码！");
        return false;
    }
};

document.oncontextmenu = function (event) {
    if (prohibitDebugger) {
        Y.util.showMessage("禁止鼠标右键！");
        event.preventDefault();
    }
};


