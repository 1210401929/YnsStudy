//该文件是自己定义的,只是想把配置的变量统一管理

//使用这个方式调用,固定多拼一个/api,到时候使用nginx跳转到网关.
export const produceDevIpPort = '/api';
//是否开启前后端数据交互加密
export const isSendCrypto = true;
//加密key和iv  随意配置,用于数据加密(如果用户前后端数据加密,必须保证值一样)
export const crypCfg = {
    key:"1234567890123456",
    iv:"6543210987654321"
}
//网站管理员
export const adminUserCode = "yulei";