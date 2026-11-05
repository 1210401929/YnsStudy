/**
 * 文章编辑器的“一键整理”
 * 只做稳妥的几件事：
 *   1. 删除段落开头的空格、全角空格；
 *   2. 中文后面紧跟的英文标点改成中文标点，括号里有中文的英文括号改成中文括号；
 *   3. 删除文章开头、结尾的空段落，连续多个空段落只留一个。
 * 代码块、行内代码、链接里的文字不动。
 * 全部通过编辑器操作完成，同一次整理是一个撤销步骤，按 Ctrl+Z 可以恢复。
 */
import {SlateEditor, SlateNode, SlateText, SlateTransforms} from '@wangeditor/editor'

const CJK = '\\u3400-\\u4dbf\\u4e00-\\u9fff\\uf900-\\ufaff'

// 这些元素里的文字不整理：代码块、链接
const SKIP_TYPES = new Set(['pre', 'code', 'link'])

const BLANK_TEXT = /^[\s　 ​]*$/
const LEADING_BLANK = /^[ \t　 ]+/

const PUNCTUATION_RULES = [
    [new RegExp(`([${CJK}])[ \\t]*,`, 'g'), '$1，'],
    [new RegExp(`([${CJK}])[ \\t]*;`, 'g'), '$1；'],
    [new RegExp(`([${CJK}])[ \\t]*:`, 'g'), '$1：'],
    [new RegExp(`([${CJK}])[ \\t]*\\?`, 'g'), '$1？'],
    [new RegExp(`([${CJK}])[ \\t]*!`, 'g'), '$1！'],
    // 句号：后面跟着点、字母、数字时不改，避免改坏省略号、文件名（如“文件.txt”）
    [new RegExp(`([${CJK}])\\.(?![.\\w])`, 'g'), '$1。'],
    // 括号：里面有中文才改，像 (EdgeOne) 这样的英文括号保持不变
    [new RegExp(`\\(([^()]*[${CJK}][^()]*)\\)`, 'g'), '（$1）'],
]

function fixPunctuation(text) {
    let count = 0
    let result = text
    for (const [pattern, replacement] of PUNCTUATION_RULES) {
        result = result.replace(pattern, (...args) => {
            count++
            return replacement.replace(/\$1/, args[1])
        })
    }
    return {text: result, count}
}

function insideSkippedElement(editor, path) {
    for (const [node] of SlateNode.ancestors(editor, path)) {
        if (SKIP_TYPES.has(node.type)) return true
    }
    return false
}

const isBlankParagraph = (node) => node.type === 'paragraph' &&
    node.children.every(child => SlateText.isText(child) && BLANK_TEXT.test(child.text))

/**
 * 整理编辑器里的正文，返回整理了多少处
 */
export function tidyArticle(editor) {
    if (!editor) return 0
    let count = 0

    // 每段第一个文字片段：要去掉段首空白
    const paragraphStarts = new Set()
    editor.children.forEach((node, index) => {
        if (node.type !== 'paragraph') return
        const [, leafPath] = SlateNode.first(editor, [index])
        paragraphStarts.add(leafPath.join(','))
    })

    const edits = []
    for (const [leaf, path] of SlateNode.texts(editor)) {
        if (!leaf.text || leaf.code || insideSkippedElement(editor, path)) continue
        let {text, count: fixed} = fixPunctuation(leaf.text)
        if (paragraphStarts.has(path.join(',')) && LEADING_BLANK.test(text)) {
            text = text.replace(LEADING_BLANK, '')
            fixed++
        }
        if (fixed) {
            edits.push({path, oldLength: leaf.text.length, text})
            count += fixed
        }
    }

    // 连续空段落只留第一个，开头和结尾的全部删除；至少保留一段
    const blank = editor.children.map(isBlankParagraph)
    const firstContent = blank.indexOf(false)
    const lastContent = blank.lastIndexOf(false)
    const removeIndexes = []
    blank.forEach((isBlank, index) => {
        if (!isBlank || firstContent === -1) return
        if (index < firstContent || index > lastContent || blank[index - 1]) removeIndexes.push(index)
    })

    if (!edits.length && !removeIndexes.length) return 0

    // 选区可能落在要删除的段落里，先取消选区
    SlateTransforms.deselect(editor)
    SlateEditor.withoutNormalizing(editor, () => {
        // 改文字不改变段落位置，从后往前改，前面的路径不受影响
        for (const {path, oldLength, text} of edits.reverse()) {
            const at = {anchor: {path, offset: 0}, focus: {path, offset: oldLength}}
            if (text) SlateTransforms.insertText(editor, text, {at})
            else SlateTransforms.delete(editor, {at})
        }
        for (const index of removeIndexes.reverse()) {
            SlateTransforms.removeNodes(editor, {at: [index]})
        }
    })
    return count + removeIndexes.length
}
