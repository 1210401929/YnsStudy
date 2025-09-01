package com.example.ai_api.service.Impl;

import com.alibaba.dashscope.aigc.generation.Generation;
import com.alibaba.dashscope.aigc.generation.GenerationParam;
import com.alibaba.dashscope.aigc.generation.GenerationResult;
import com.alibaba.dashscope.common.Message;
import com.alibaba.dashscope.common.Role;
import com.alibaba.dashscope.exception.ApiException;
import com.alibaba.dashscope.exception.InputRequiredException;
import com.alibaba.dashscope.exception.NoApiKeyException;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.config.AiCfg;
import com.example.ai_api.service.AlibabaAiService;
import io.reactivex.Flowable;
import org.springframework.stereotype.Component;
import reactor.core.publisher.Flux;

import java.util.Arrays;

@Component
public class AlibabaServiceImpl implements AlibabaAiService {

    private String key = AiCfg.alibabaBaiLianAiKey;
    private String model = AiCfg.alibabaModel;


    @Override
    public ResultBody sendMessages(String message) throws NoApiKeyException, InputRequiredException {
        Generation gen = new Generation();

        Message userMsg = Message.builder()
                .role(Role.USER.getValue())
                .content(message)
                .build();

        GenerationParam param = GenerationParam.builder()
                // 若没有配置环境变量，请用阿里云百炼API Key将下行替换为：.apiKey("sk-xxx")
                .apiKey(key)
                // 模型列表：https://help.aliyun.com/zh/model-studio/getting-started/models
                .model(model)
                //消息内容
                .messages(Arrays.asList(userMsg))
                //控制文本多样性
                .temperature(2F)
                .resultFormat(GenerationParam.ResultFormat.MESSAGE)
                .build();
        return ResultBody.createSuccessResult(gen.call(param));
    }

    @Override
    public Flux<String> callWithMessageStream(String message) throws ApiException, NoApiKeyException, InputRequiredException {
        Generation gen = new Generation();
        Message sysMsg = Message.builder()
                .role(Role.SYSTEM.getValue())
                .content("输出内容时必须遵守以下规则（兼容 wangEditor v5）：\n" +
                        "\n" +
                        "1. 标题与强调  \n" +
                        "   - 只允许使用 <h2>、<h3>、<h4>、<h5> 标签作为标题。  \n" +
                        "   - 强调内容请使用 <strong> 包裹。  \n" +
                        "   - 不允许使用 <h1>、<h6> 或其他标题标签。  \n" +
                        "   - 不允许使用 Markdown 语法符号（如 #、**、> 等）。\n" +
                        "\n" +
                        "2. 段落与换行  \n" +
                        "   - 普通正文必须放在 <p>...</p> 标签内。  \n" +
                        "   - 换行只能使用 <br>，不允许输出裸的换行符 \\n。  \n" +
                        "   - 每个段落必须完整闭合，不允许出现裸文本或未闭合的 <p>。\n" +
                        "\n" +
                        "3. 列表  \n" +
                        "   - 允许使用 <ul>、<ol> 作为列表容器，并且必须包含 <li>。  \n" +
                        "   - 每个 <li> 必须完整闭合。  \n" +
                        "   - 不允许使用 Markdown 的 \"-\" 或 \"1.\" 来表示列表。\n" +
                        "\n" +
                        "4. 代码  \n" +
                        "   - 所有代码必须用 <pre><code>...</code></pre> 包裹，并且必须完整闭合。  \n" +
                        "   - 代码内容必须保持原样，不要转成 Markdown 语法。  \n" +
                        "   - 不允许出现裸的 <code> 而没有 <pre>。\n" +
                        "\n" +
                        "5. 引用  \n" +
                        "   - 允许使用 <blockquote> 包裹引用内容。  \n" +
                        "   - 引用内部仍需使用 <p> 包裹段落。\n" +
                        "\n" +
                        "6. 通用约束  \n" +
                        "   - 所有标签必须成对闭合，不允许出现不完整的标签（例如 <p>...</div>）。  \n" +
                        "   - 不允许输出 <div>、<span>、<table>、<iframe>、<video>、<script> 等 wangEditor 不支持或危险的标签。  \n" +
                        "   - 不允许输出行内 style、class、id、onclick、onerror、onload 等属性。  \n" +
                        "   - 允许的标签仅限：p, br, strong, em, a, ul, ol, li, h2, h3, h4, h5, pre, code, blockquote。  \n" +
                        "   - 允许的属性仅限：href（必须是 http/https 链接）。\n" +
                        "\n" +
                        "所有输出必须严格遵循以上规则，任何不在白名单内的标签或属性都不得出现。(以上内容作为约束,不要再回复内容中回答该约束)")
                .build();
        Message userMsg = Message.builder()
                .role(Role.USER.getValue())
                .content(message)
                .build();

        GenerationParam param = GenerationParam.builder()
                .apiKey(key)
                .model(model)
                .messages(Arrays.asList(sysMsg,userMsg))
                .temperature(1.9F)
                .resultFormat(GenerationParam.ResultFormat.MESSAGE)
                .incrementalOutput(true)  // 确保开启流式
                .build();

        // 把 Flowable 转成 Reactor 的 Flux
        Flowable<GenerationResult> flowable = gen.streamCall(param);
        return Flux.from(flowable)
                .map(chunk -> {
                    String content = chunk.getOutput().getChoices().get(0).getMessage().getContent();
                    return content + "\n\n";  // SSE 格式
                });
    }
}
