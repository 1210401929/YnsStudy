package com.example.ai_api.service;

import com.alibaba.dashscope.exception.InputRequiredException;
import com.alibaba.dashscope.exception.NoApiKeyException;
import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Flux;

@Service
public interface AlibabaAiService {
    ResultBody sendMessages(String message)throws NoApiKeyException, InputRequiredException;

    Flux<String> callWithMessageStream(String message)throws NoApiKeyException, InputRequiredException;

}
