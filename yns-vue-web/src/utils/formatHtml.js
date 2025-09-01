// /src/utils/normalizeAiHtml.ts
import DOMPurify from 'dompurify';
export let built_in_preface =
`输出内容时必须遵守以下规则（兼容 wangEditor v5）：

1. 标题与强调  
   - 只允许使用 <h2>、<h3>、<h4>、<h5> 标签作为标题。  
   - 强调内容请使用 <strong> 包裹。  
   - 不允许使用 <h1>、<h6> 或其他标题标签。  
   - 不允许使用 Markdown 语法符号（如 #、**、> 等）。

2. 段落与换行  
   - 普通正文必须放在 <p>...</p> 标签内。  
   - 换行只能使用 <br>，不允许输出裸的换行符 \\n。  
   - 每个段落必须完整闭合，不允许出现裸文本或未闭合的 <p>。

3. 列表  
   - 允许使用 <ul>、<ol> 作为列表容器，并且必须包含 <li>。  
   - 每个 <li> 必须完整闭合。  
   - 不允许使用 Markdown 的 "-" 或 "1." 来表示列表。

4. 代码  
   - 所有代码必须用 <pre><code>...</code></pre> 包裹，并且必须完整闭合。  
   - 代码内容必须保持原样，不要转成 Markdown 语法。  
   - 不允许出现裸的 <code> 而没有 <pre>。

5. 引用  
   - 允许使用 <blockquote> 包裹引用内容。  
   - 引用内部仍需使用 <p> 包裹段落。

6. 通用约束  
   - 所有标签必须成对闭合，不允许出现不完整的标签（例如 <p>...</div>）。  
   - 不允许输出 <div>、<span>、<table>、<iframe>、<video>、<script> 等 wangEditor 不支持或危险的标签。  
   - 不允许输出行内 style、class、id、onclick、onerror、onload 等属性。  
   - 允许的标签仅限：p, br, strong, em, a, ul, ol, li, h2, h3, h4, h5, pre, code, blockquote。  
   - 允许的属性仅限：href（必须是 http/https 链接）。

所有输出必须严格遵循以上规则，任何不在白名单内的标签或属性都不得出现。(以上内容作为约束,不要再回复内容中回答该约束)`;
const ALLOWED_TAGS = [
    'p', 'br', 'strong', 'em', 'a', 'code', 'pre',
    'ul', 'ol', 'li',
    'h1', 'h2', 'h3', 'h4', 'h5', 'h6', // ✅ 保留 h1~h6
    'blockquote'
];

const ALLOWED_ATTR = ['href'];

const NL = '\n';

function escapeHtml(s = '') {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function isHttpUrl(u = '') {
    return /^https?:\/\//i.test(u.trim());
}

/**
 * 第 0 步：预处理明显坏片段（“急救绷带”）
 * - 把裸的 "pre>" => "<pre>"，"code>" => "<code>"，"</p>，" => "</p>"
 * - 修复常见的反引号/符号导致的错误收尾，比如 "8mb4'code>" -> "8mb4</code>"
 * - 把中文逗号造成的坏闭合修正
 */
export function preFixGrossErrors(raw = '') {
    let s = raw.replace(/[\u200B-\u200D\uFEFF]/g, '');

    // 把裸的 "code>" / "pre>" 修成开标签（可保留，通用规则也能覆盖）
    s = s.replace(/(^|[^<])code>/gi, '$1<code>');
    s = s.replace(/(^|[^<])pre>/gi,  '$1<pre>');

    // 常见错误收尾修复："</code" -> "</code>", "</p" -> "</p>"
    s = s.replace(/<\/(code|pre|p)(?=\s|$)/gi, '</$1>');

    // “中文逗号”跟在闭合标签后
    s = s.replace(/<\/(code|pre|p)>，/g, '</$1>');

    // ---- 通用：补齐或清理孤立的 xxx> 形式 ----
    // A. 若以字母开头（合法标签名），则补成 <xxx>
    //   1) 行首：  "xxx>..."
    s = s.replace(/^([A-Za-z]\w*)>(?=[^\w<])/gm, '<$1>');
    //   2) 非 '<' 或字母数字后："... xxx>..."
    s = s.replace(/([^<\w])([A-Za-z]\w*)>(?=[^\w<])/g, '$1<$2>');

    // B. 若非字母开头（比如 1>、3a> 等），则清理掉，避免生成非法标签
    s = s.replace(/(^|[^<\w])(\d\w*)>(?=[^\w<])/g, '$1');

    return s;
}





/**
 * 第 1 步：把 Markdown-ish 转换成受控 HTML
 * - ```fenced code``` -> <pre><code>…</code></pre>
 * - `inline code` -> <code>…</code>
 * - > 引用行 -> <blockquote><p>…</p></blockquote>
 * - 其余文本 -> 分段成 <p>…</p>（单换行 -> <br>）
 */
function markdownishToHtml(raw) {
    let s = raw.replace(/\r\n?/g, NL);

    // ``` 代码块
    s = s.replace(/```([a-z0-9_\-]+)?\s*([\s\S]*?)```/gi, (_, _lang, body) => {
        const codeEsc = escapeHtml((body || '').replace(/\n+$/, ''));
        return `<pre><code>${codeEsc}</code></pre>`;
    });

    // 行内代码
    s = s.replace(/`([^`]+?)`/g, (_, m) => `<code>${escapeHtml(m)}</code>`);

    // 处理 blockquote
    const lines = s.split(NL);
    const out = [];
    let buf = [];

    const flush = () => {
        if (!buf.length) return;
        const inner = buf
            .map(l => l.replace(/^\s*>\s?/, ''))
            .map(l => l.trim() ? `<p>${l}</p>` : '')
            .join('');
        if (inner) out.push(`<blockquote>${inner}</blockquote>`);
        buf = [];
    };

    for (const l of lines) {
        if (/^\s*>\s?/.test(l)) buf.push(l);
        else {
            flush();
            out.push(l);
        }
    }
    flush();

    const joined = out.join(NL);
    const parts = joined.split(/\n{2,}/);

    return parts.map(seg => {
        if (/<pre><code>|<\/code><\/pre>|<blockquote>/.test(seg)) return seg;
        const t = seg.trim();
        if (!t) return '';
        return `<p>${t.replace(/\n/g, '<br>')}</p>`;
    }).filter(Boolean).join('');
}

/**
 * 第 2 步：DOM 结构纠偏
 * - div/section/article/span -> p
 * - h1 -> h2
 * - a[href] 仅允许 http/https
 */
function normalizeUnsupportedTags(root) {
    root.querySelectorAll('div,section,article,span').forEach(el => {
        const p = document.createElement('p');
        while (el.firstChild) p.appendChild(el.firstChild);
        el.replaceWith(p);
    });

    root.querySelectorAll('h1').forEach(el => {
        const h2 = document.createElement('h2');
        while (el.firstChild) h2.appendChild(el.firstChild);
        el.replaceWith(h2);
    });

    root.querySelectorAll('a').forEach(a => {
        const href = a.getAttribute('href') || '';
        if (!isHttpUrl(href)) a.removeAttribute('href');
    });
}

/** 第 3 步：两轮净化（先净化→纠偏→再净化） */
function sanitizeAndNormalize(html) {
    const clean1 = DOMPurify.sanitize(html, {
        ALLOWED_TAGS, ALLOWED_ATTR,
        USE_PROFILES: {html: true},
        FORBID_ATTR: ['style', 'class', 'id', 'onclick', 'onerror', 'onload']
    });

    const tmp = document.createElement('div');
    tmp.innerHTML = clean1;
    normalizeUnsupportedTags(tmp);

    const clean2 = DOMPurify.sanitize(tmp.innerHTML, {
        ALLOWED_TAGS, ALLOWED_ATTR,
        USE_PROFILES: {html: true},
        FORBID_ATTR: ['style', 'class', 'id', 'onclick', 'onerror', 'onload']
    });

    return clean2;
}

/** 对外导出：给 AI 的“半残 HTML/Markdown 混合”做兜底修复 */
export function normalizeAiHtml(raw) {
    if (!raw) return '';
    const pre = preFixGrossErrors(raw);
    const asHtml = markdownishToHtml(pre);
    return sanitizeAndNormalize(asHtml);
}
