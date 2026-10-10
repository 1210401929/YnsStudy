//该文件是自己定义的,只是想把配置的变量统一管理

//使用这个方式调用,固定多拼一个/api,到时候使用nginx跳转到网关.

export const produceDevIpPort = '/api';
//是否开启前后端数据交互加密
export const isSendCrypto = true;
//加密key和iv  随意配置,用于数据加密(如果用户前后端数据加密,必须保证值一样)
export const crypCfg = {
    key: "7546455674406856",
    iv: "7513956994549178"
}
//网站管理员
export const adminUserCode = "yulei";


//文章正文的展示方式:
//  'html'   直接渲染保存的 HTML(默认,文章页不加载编辑器,打开更快)
//  'editor' 用 wangEditor 只读模式展示(旧方式,如发现显示问题可临时切回)
export const articleRenderMode = 'html';
