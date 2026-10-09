/**
 * 上传前在浏览器里压缩图片：限制最大尺寸，并优先转成 WebP（浏览器不支持时用 JPEG）。
 *
 * - GIF（可能是动图）和 SVG（矢量图）原样上传。
 * - 带透明通道的 PNG 在浏览器不支持 WebP 编码时原样上传，避免透明背景变黑。
 * - 压缩后反而更大、且没有缩小尺寸时，保留原文件。
 * - 任何一步失败都返回原文件，不影响上传。
 */

const SKIP_TYPES = ['image/gif', 'image/svg+xml']

export const IMAGE_PRESETS = {
    // 文章配图：正文最宽约 1000px，2 倍屏下 1920px 已足够清晰
    article: { maxWidth: 1920, maxHeight: 1920, quality: 0.82 },
    // 头像：显示尺寸很小，512px 足够
    avatar: { maxWidth: 512, maxHeight: 512, quality: 0.85 }
}

async function decodeImage(file) {
    if (typeof createImageBitmap === 'function') {
        try {
            // 按照 EXIF 方向摆正手机拍摄的照片
            return await createImageBitmap(file, { imageOrientation: 'from-image' })
        } catch {
            // 个别浏览器不支持参数，继续使用 <img> 解码
        }
    }
    const url = URL.createObjectURL(file)
    try {
        const image = new Image()
        image.decoding = 'async'
        image.src = url
        await image.decode()
        return image
    } finally {
        URL.revokeObjectURL(url)
    }
}

function canvasToBlob(canvas, type, quality) {
    return new Promise(resolve => canvas.toBlob(resolve, type, quality))
}

function renameFile(name, extension) {
    const base = (name || 'image').replace(/\.[^.]+$/, '')
    return `${base}.${extension}`
}

/**
 * @param {File} file 用户选择或粘贴的图片
 * @param {{maxWidth:number,maxHeight:number,quality:number}} options
 * @returns {Promise<File>} 压缩后的文件；无法或无需压缩时返回原文件
 */
export async function compressImage(file, options = IMAGE_PRESETS.article) {
    if (!file || !file.type || !file.type.startsWith('image/') || SKIP_TYPES.includes(file.type)) {
        return file
    }
    const { maxWidth, maxHeight, quality } = { ...IMAGE_PRESETS.article, ...options }
    try {
        const image = await decodeImage(file)
        const width = image.width
        const height = image.height
        if (!width || !height) return file

        const scale = Math.min(1, maxWidth / width, maxHeight / height)
        const targetWidth = Math.max(1, Math.round(width * scale))
        const targetHeight = Math.max(1, Math.round(height * scale))

        const canvas = document.createElement('canvas')
        canvas.width = targetWidth
        canvas.height = targetHeight
        const context = canvas.getContext('2d')
        context.imageSmoothingQuality = 'high'
        context.drawImage(image, 0, 0, targetWidth, targetHeight)
        if (typeof image.close === 'function') image.close()

        let blob = await canvasToBlob(canvas, 'image/webp', quality)
        let extension = 'webp'
        if (!blob || blob.type !== 'image/webp') {
            // 不支持 WebP 编码时，PNG 可能带透明通道，转 JPEG 会丢失透明度，保留原图
            if (file.type === 'image/png') return file
            blob = await canvasToBlob(canvas, 'image/jpeg', quality)
            extension = 'jpg'
        }
        if (!blob) return file
        if (blob.size >= file.size && scale === 1) return file

        return new File([blob], renameFile(file.name, extension), {
            type: blob.type,
            lastModified: Date.now()
        })
    } catch (error) {
        console.warn('图片压缩失败，使用原图上传', error)
        return file
    }
}
