import axios from 'axios'
import {ElMessage, ElMessageBox} from 'element-plus'
import CryptoJS from 'crypto-js'
//获取配置的生产环境ip端口
import {produceDevIpPort,crypCfg} from "@/config/vue-config.js";

/**
 *  sendAxiosRequest        发送后台请求统一入口
 *  getGuid                 获取随机32位码
 *  ele_confirm             弹出确认对话框
 *  buildChildrenData       构造上下级数组
 *  downloadFileByUrl       根据路径下载文件
 *  pubFormatDate           格式化日期字符串
 *  encrypt                 加密字符串
 *  decrypt                 解密字符串
 */



//调用后台方法
export const sendAxiosRequest = async (url, data, type, isReturnAll) => {
    // 环境判断：开发使用相对路径(开发环境的配置在vite.config.js)，生产使用真实地址
    url = import.meta.env.MODE === 'development'
        ? url
        : produceDevIpPort + url;
    data = data || {};
    type = type || "post";
    type = type.toLowerCase();
    let response;
    // 配置对象
    let config = {
        withCredentials: true // 允许携带 Cookie（包括 SESSION）  为了实现后端session共享
    };
    // 4. 如果是 POST/PUT 并且 data 是 FormData，则添加 multipart header
    if (typeof FormData !== 'undefined' && data instanceof FormData) {
        config.headers = {
            'Content-Type': 'multipart/form-data'
        };
    }
    if (type === "get") {
        response = await axios.get(url, {params: data});
    } else {
        response = await axios[type](url, data, config);
    }
    return isReturnAll ? response : response.data;
}


//获取随机32码
export const getGuid = () => {
    const generateUUID = () => {
        let d = new Date().getTime();
        if (window.performance && typeof window.performance.now === "function") {
            d += performance.now(); //use high-precision timer if available
        }
        let uuid = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
            let r = (d + Math.random() * 16) % 16 | 0;
            d = Math.floor(d / 16);
            return (c == 'x' ? r : (r & 0x3 | 0x8)).toString(16);
        });
        return uuid;
    }
    let guid = generateUUID();
    return guid.replaceAll("-", "").toUpperCase();
}

export const ele_confirm = (tip, fun) => {
    ElMessageBox.confirm(
        tip,
        '提示',
        {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning',
        }
    )
        .then(() => {
            if (typeof fun === 'function') {
                fun()
            }
        })
        .catch(() => {
            ElMessage.info("已取消");
        })
}

//数组转换成上下级结构   option:上下级 字段的依赖
export const buildChildrenData = (data, options = {}) => {
    const {
        idKey = 'GUID',
        parentKey = 'SUPERGUID',
        childrenKey = 'children'
    } = options;

    const itemMap = new Map();
    const result = [];

    // 初始化映射
    data.forEach(item => {
        const newItem = {...item};
        itemMap.set(newItem[idKey], newItem);
    });

    // 构建树
    data.forEach(item => {
        const currentItem = itemMap.get(item[idKey]);
        const parentItem = itemMap.get(item[parentKey]);

        if (parentItem) {
            parentItem[childrenKey] = parentItem[childrenKey] || [];
            parentItem[childrenKey].push(currentItem);
        } else {
            result.push(currentItem);
        }
    });

    return result;
}

//根据路径下载文件
export const downloadFileByUrl = (url,fileName)=>{
    fetch(url, { mode: 'cors' }) // mode: 'cors' 如果资源是跨域的，目标服务器需要设置 CORS
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok')
            }
            return response.blob()
        })
        .then(blob => {
            const blobUrl = window.URL.createObjectURL(blob)
            const a = document.createElement('a')
            a.href = blobUrl
            a.download = fileName || 'download'
            document.body.appendChild(a)
            a.click()
            document.body.removeChild(a)
            window.URL.revokeObjectURL(blobUrl)
        })
        .catch(error => {
            console.error('Download failed:', error)
        })
}

export function pubFormatDate(dateStr) {
    const timestamp = Date.parse(dateStr);
    if (isNaN(timestamp)) {
        // 非法日期字符串，直接返回原始输入
        return dateStr;
    }
    const date = new Date(timestamp);
    return date.toLocaleString();
}



export function encrypt(text) {
    const key = CryptoJS.enc.Utf8.parse(crypCfg.key)
    const iv  = CryptoJS.enc.Utf8.parse(crypCfg.iv)

    let srcs = CryptoJS.enc.Utf8.parse(text);
    let encrypted = CryptoJS.AES.encrypt(srcs, key, { iv: iv, mode: CryptoJS.mode.CBC, padding: CryptoJS.pad.Pkcs7 });
    return encrypted.ciphertext.toString().toUpperCase();
}

export function decrypt(cipherText) {
    const key = CryptoJS.enc.Utf8.parse(crypCfg.key)
   const iv  = CryptoJS.enc.Utf8.parse(crypCfg.iv)

    let encryptedHexStr = CryptoJS.enc.Hex.parse(cipherText);
    let srcs = CryptoJS.enc.Base64.stringify(encryptedHexStr);
    let decrypt = CryptoJS.AES.decrypt(srcs, key, { iv: iv, mode: CryptoJS.mode.CBC, padding: CryptoJS.pad.Pkcs7 });
    let decryptedStr = decrypt.toString(CryptoJS.enc.Utf8);
    return decryptedStr.toString();
}