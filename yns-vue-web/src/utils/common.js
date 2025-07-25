import axios from 'axios'
import {ElMessage, ElMessageBox, ElLoading} from 'element-plus'
import CryptoJS from 'crypto-js'
//获取配置的生产环境ip端口
import {produceDevIpPort, crypCfg, isSendCrypto} from "@/config/vue-config.js";

/**
 *  sendAxiosRequest        发送后台请求统一入口
 *  uploadFileWithProgress  上传文件调用后台接口
 *  getGuid                 获取随机32位码
 *  ele_confirm             弹出确认对话框
 *  buildChildrenData       构造上下级数组
 *  downloadFileByUrl       根据路径下载文件
 *  pubFormatDate           格式化日期字符串
 *  encrypt                 加密字符串
 *  decrypt                 解密字符串
 *  pubLoading              loading动画
 *  loadScript              动态加载外部脚本
 *  isProbablyCipher        判断字符串是否“看起来像” Base64 密文
 *  stripImages             删除html里的图片等内容
 *  extractFirstImage       提取html中的第一个图片
 *  extractPlainTextFromHTML 提取html中的纯文本
 *  sendNotifications       系统通用发送消息
 */

/**
 * 统一发送 Axios 请求，支持可选加密 / 解密
 *
 * @param {string}  url           请求地址
 * @param {Object}  data          请求数据 (默认 {})
 * @param {string}  method        HTTP 方法 (默认 'post')
 * @param {boolean} needCrypto    是否对请求 & 响应做加密 (默认全局 isSendCrypto)
 * @param {boolean} isReturnAll   是否返回整个 Axios 响应对象 (默认 false)
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

    /* ---------- 构造 payload ---------- */
    let payload
    if (needCrypto && !isForm && (method === 'post' || method === 'put')) {
        // 普通 JSON 请求且需要加密
        payload = {cipherText: encrypt(data)}
    } else {
        // GET / DELETE / multipart / 不需要加密 —— 原样发送
        payload = data
        if (isForm && (method === 'post' || method === 'put')) {
            config.headers = {'Content-Type': 'multipart/form-data'}
        }
    }

    /* ---------- 发请求 ---------- */
    let resp
    if (method === 'get' || method === 'delete') {
        resp = await axios[method](url, {params: payload, ...config})
    } else {
        resp = await axios[method](url, payload, config)
    }

    if (isReturnAll) {
        return resp
    }

    /* ---------- 处理响应 ---------- */
    const respData = resp.data

    // 仅当 needCrypto=true 并且响应看上去是密文才尝试解密
    if (needCrypto && (method === 'post' || method === 'put') && respData) {
        // 兼容后端 {"data": "..."} 或直接返回字符串
        const cipherCandidate =
            typeof respData === 'object' && respData !== null && 'data' in respData
                ? respData.data
                : respData

        if (isProbablyCipher(cipherCandidate)) {
            try {
                return decrypt(cipherCandidate)
            } catch (e) {
                // 解密失败说明并非我们的密文，直接返回原数据
                console.warn('Decrypt failed, return raw response:', e)
            }
        }
    }
    // 走到这里说明不需要解密 / 不是密文 / 解密失败
    return respData
}

//上传文件调用后台接口  sendAxiosRequest方法的扩展,轻易不要用这个方法,使用场景,需要获取上传文件的进度
export function uploadFileWithProgress({
                                           url,
                                           file,
                                           fieldName = "file",
                                           extraData = {},
                                           onProgress = () => {
                                           },
                                           onSuccess = () => {
                                           },
                                           onError = () => {
                                           }
                                       }) {
    const formData = new FormData();
    formData.append(fieldName, file);
    Object.entries(extraData).forEach(([key, value]) => {
        formData.append(key, value);
    });
    const xhr = new XMLHttpRequest();
    // 加上这个！
    xhr.withCredentials = true;
    // 自动拼接生产路径（模仿你的封装）
    const realUrl =
        import.meta.env.MODE === "development" ? url : produceDevIpPort + url;

    xhr.open("POST", realUrl, true);

    xhr.upload.onprogress = (event) => {
        if (event.lengthComputable) {
            const percent = Math.round((event.loaded / event.total) * 100);
            onProgress(percent);
        }
    };

    xhr.onload = () => {
        if (xhr.status === 200) {
            try {
                const res = JSON.parse(xhr.responseText);
                onSuccess(res);
            } catch (e) {
                onError("返回结果解析失败");
            }
        } else {
            onError(`上传失败：${xhr.statusText}`);
        }
    };
    xhr.onerror = () => {
        onError("上传失败，请检查网络连接");
    };
    xhr.send(formData);
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
    // 开启loading
    pubLoading("start", {text: "下载中，请稍候..."});

    fetch(url, {mode: 'cors'})
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            const reader = response.body.getReader();
            const contentLength = +response.headers.get('Content-Length');
            let receivedLength = 0;
            const chunks = [];
            const pump = () => {
                reader.read().then(({done, value}) => {
                    if (done) {
                        // 创建 Blob 对象，下载文件
                        const blob = new Blob(chunks);
                        const blobUrl = window.URL.createObjectURL(blob);
                        const a = document.createElement('a');
                        a.href = blobUrl;
                        a.download = fileName || 'download';
                        document.body.appendChild(a);
                        a.click();
                        document.body.removeChild(a);
                        window.URL.revokeObjectURL(blobUrl);
                        // 下载完成，关闭loading提示
                        pubLoading("close");
                        ElMessage.success('文件下载成功!');
                        return;
                    }
                    // 更新接收长度并显示进度
                    receivedLength += value.length;
                    chunks.push(value);
                    const progress = (receivedLength / contentLength) * 100;
                    // 更新loading提示中的进度
                    pubLoading("update", {
                        text: `下载中... (${Math.round(progress)}%)`
                    });
                    // 继续读取流
                    pump();
                });
            };

            pump();
        })
        .catch(error => {
            // 下载失败，关闭loading并弹出错误提示
            pubLoading("close");
            ElMessage.error('下载失败: ' + error.message);
        });
};


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

/**
 *loading动画
 * @param type   start|update|close
 * @param cfg
 */
export function pubLoading(type = "start", cfg = {}) {
    let loading;
    switch (type) {
        case "start":
            // 如果已经存在，强行关闭之前的loading
            if (window["yns_loading"]) {
                window["yns_loading"].close();
            }
            loading = ElLoading.service({
                text: cfg.text || '执行中，请稍候...',
                spinner: cfg.spinner || '<el-icon><loading /></el-icon>',
                background: cfg.background || 'rgba(243,243,243,0.7)',
            });
            window["yns_loading"] = loading;
            break;
        case "update":
            if (!window["yns_loading"]) return false;

            let cfgInfoArr = Object.keys(cfg);
            cfgInfoArr.forEach(info => {
                if (window["yns_loading"][info])
                    window["yns_loading"][info].value = cfg[info];
            })
            break;
        case "close":
            if (window["yns_loading"]) {
                window["yns_loading"].close();
            }
            break;
    }
    return window["yns_loading"];
}

// 动态加载外部脚本
export function loadScript(src) {
    return new Promise((resolve, reject) => {
        const script = document.createElement("script");
        script.src = src;
        script.async = true;
        script.onload = resolve;
        script.onerror = reject;
        document.head.appendChild(script);
    });
}

/**
 * 判断字符串是否“看起来像” Base64 密文
 * 条件：
 *   1. 仅包含 A-Z / a-z / 0-9 / + / /
 *   2. 末尾可选 1‒2 个 =
 *   3. 长度是 4 的倍数
 *   4. 字符数不少于 24（≈16 字节），避免把普通字段误判成密文
 */
export function isProbablyCipher(str) {
    return typeof str === 'string'
        && /^[A-Za-z0-9+/]+={0,2}$/.test(str)
        && (str.length % 4 === 0)
        && str.length >= 24
}

//删除html里的图片等内容
export function stripImages(htmlContent) {
    const tempDiv = document.createElement('div');
    tempDiv.innerHTML = htmlContent;
    const images = tempDiv.querySelectorAll('img');
    images.forEach(img => img.remove());
    const ads = tempDiv.querySelectorAll('iframe, .advertisement');
    ads.forEach(ad => ad.remove());
    const result = tempDiv.innerHTML;
    tempDiv.remove();
    return result;
};

//提取html中的第一个图片
export function extractFirstImage(htmlContent) {
    const tempDiv = document.createElement('div');
    tempDiv.innerHTML = htmlContent;

    // 查找第一张图片
    const img = tempDiv.querySelector('img');
    return img ? img.src : '';
}

//提取html中的纯文本
export function extractPlainTextFromHTML(html) {
    const div = document.createElement('div')
    div.innerHTML = html
    const text = div.textContent || div.innerText || ''
    return text;
}


/**
 * 系统通用发送消息
 * @param sendUserCode  发送用户账号
 * @param receiverUserCode  接收用户账号
 * @param type  类型  comment:评论
 * @param execute   操作页面路径
 * @param remark    通知文本
 * @returns {Promise<boolean>}
 */
export async function sendNotifications(sendUserCode, receiverUserCode, type, execute, remark) {
    if (!sendUserCode || !receiverUserCode || !type) {
        console.error("发送消息失败,参数传递不完整!");
        return false;
    }
    if (!execute) {
        execute = window.location.href;
    }
    if (!remark) {
        remark = "未知操作";
    }
    let result = await sendAxiosRequest("/pub-api/notice/addNotice", {
        sendUserCode, receiverUserCode, type, execute, remark
    });
    if (result && !result.isError) {
        return true;
    } else {
        console.error("通知用户失败");
        return false;
    }
}
