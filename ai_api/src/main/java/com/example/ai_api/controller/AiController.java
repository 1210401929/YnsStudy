package com.example.ai_api.controller;

import com.alibaba.dashscope.exception.InputRequiredException;
import com.alibaba.dashscope.exception.NoApiKeyException;
import com.example.common_api.bean.ResultBody;
import com.example.ai_api.service.AlibabaAiService;
import com.example.common_api.bean.UserBean;
import com.example.common_api.util.AesUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.server.ResponseStatusException;
import reactor.core.publisher.Flux;

import javax.servlet.http.HttpSession;
import java.time.Duration;
import java.util.Map;

@Controller
@RequestMapping("/ai-api/ai")
public class AiController {
    @Autowired
    AlibabaAiService alibabaAiService;

    //一次性返回所有结果
    @RequestMapping("/qWen")
    @ResponseBody
    public ResultBody qWen(@RequestBody Map<String, String> params, HttpSession session) throws NoApiKeyException, InputRequiredException {
        String message = params.get("message");
        return alibabaAiService.sendMessages(message);
    }
    //流式输出  前端也必须流式接收
    @RequestMapping(value = "/qWenStream", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    @ResponseBody
    public Flux<String> qWenStream(@RequestParam String message,HttpSession session) throws NoApiKeyException, InputRequiredException {
        UserBean userInfo = (UserBean)session.getAttribute("userInfo");
        if(userInfo==null){
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "用户未登录");
        }
        // 规范化 message，去掉空格/换行，并把空格还原为 '+'
        String normalizedMessage = message.trim()
                .replace(' ', '+')
                .replaceAll("\\r|\\n|\\t", "");
        String finalMessage = AesUtil.decrypt(normalizedMessage);
        return alibabaAiService.callWithMessageStream(finalMessage)
                .map(content ->  content + "\n\n"); // SSE 格式
    }

}
