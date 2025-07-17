import axios from 'axios'
import {ElMessage, ElMessageBox} from 'element-plus'
import CryptoJS from 'crypto-js'
//获取配置的生产环境ip端口
import {produceDevIpPort, crypCfg, isSendCrypto} from "@/config/vue-config.js";

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

/**
 * 统一请求方法（纯 JS）
 * @param url         接口地址
 * @param data        请求参数或 body
 * @param method      'get' | 'post' | 'put' | 'delete'
 * @param needCrypto  是否走加/解密，默认用全局 isSendCrypto
 * @param isReturnAll 是否返回完整响应对象
 */
export const sendAxiosRequest = async function (
    url,
    data = {},
    method = 'post',
    needCrypto = isSendCrypto,
    isReturnAll = false
) {
    // 环境判断
    url = import.meta.env.MODE === 'development'
        ? url
        : produceDevIpPort + url

    method = method.toLowerCase()
    const config = {withCredentials: true}
    const isForm = typeof FormData !== 'undefined' && data instanceof FormData

    // 构造 payload
    let payload
    if (needCrypto && !isForm && (method === 'post' || method === 'put')) {
        payload = {cipherText: encrypt(data)}
    } else {
        payload = data
        if (isForm && (method === 'post' || method === 'put')) {
            config.headers = {'Content-Type': 'multipart/form-data'}
        }
    }

    // 发请求
    let resp
    if (method === 'get' || method === 'delete') {
        resp = await axios[method](url, {params: payload, ...config})
    } else {
        resp = await axios[method](url, payload, config)
    }

    if (isReturnAll) {
        return resp
    }

    const respData = resp.data
    // 只有 POST/PUT 且开启加密时才解密
    if (needCrypto && (method === 'post' || method === 'put') && respData) {
        const cipher = (typeof respData === 'object' && respData.data)
            ? respData.data
            : respData
        return decrypt(cipher)
    }
    // 其它直接返回 data
    return respData
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
export const downloadFileByUrl = (url, fileName) => {
    fetch(url, {mode: 'cors'}) // mode: 'cors' 如果资源是跨域的，目标服务器需要设置 CORS
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

/**
 * AES 加密函数，支持字符串或对象，返回 Base64 编码字符串
 * @param {string|object} data - 原始数据
 * @returns {string} - 加密后 Base64 字符串
 */
export function encrypt(data) {
    const key = CryptoJS.enc.Utf8.parse(crypCfg.key)
    const iv = CryptoJS.enc.Utf8.parse(crypCfg.iv)

    const plaintext =
        typeof data === 'object' ? JSON.stringify(data) : String(data)

    const encrypted = CryptoJS.AES.encrypt(plaintext, key, {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7,
    })

    // 默认输出即为 Base64
    return encrypted.toString()
}

/**
 * AES 解密函数，输入为 Base64 字符串，返回原始字符串或对象
 * @param {string} cipherText - 加密后的 Base64 字符串
 * @returns {string|object} - 解密后的字符串或对象
 */
export function decrypt(cipherText) {
    const key = CryptoJS.enc.Utf8.parse(crypCfg.key)
    const iv = CryptoJS.enc.Utf8.parse(crypCfg.iv)

    const decrypted = CryptoJS.AES.decrypt(cipherText, key, {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7,
    })
    const result = decrypted.toString(CryptoJS.enc.Utf8)
    try {
        return JSON.parse(result)
    } catch (e) {
        return result
    }
}