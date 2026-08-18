package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

const aiHTMLSystemPrompt = `输出内容必须兼容 wangEditor v5。只允许 p、br、strong、em、a、ul、ol、li、h2、h3、h4、h5、pre、code、blockquote 标签；只允许 http/https 的 href 属性。普通正文放在完整的 p 标签内，代码放在完整的 pre/code 标签内，不使用 Markdown，不输出 style、class、id 或事件属性，所有标签必须正确闭合。不要在回答中复述这些约束。`

type aiRequest struct {
	Model       string      `json:"model"`
	Messages    []aiMessage `json:"messages"`
	Temperature float64     `json:"temperature"`
	Stream      bool        `json:"stream,omitempty"`
}

type aiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (s *Service) AskAI(ctx context.Context, message string) model.Result {
	response, err := s.openAIRequest(ctx, message, false)
	if err != nil {
		return model.Failure("AI 请求失败:" + err.Error())
	}
	defer response.Body.Close()
	var payload any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return model.Failure("解析 AI 响应失败:" + err.Error())
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return model.Failure(fmt.Sprintf("AI 服务返回 %s: %v", response.Status, payload))
	}
	return model.Success(payload)
}

func (s *Service) StreamAI(c *gin.Context, message string) error {
	response, err := s.openAIRequest(contextOf(c), message, true)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 64*1024))
		return fmt.Errorf("AI 服务返回 %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if json.Unmarshal([]byte(data), &chunk) != nil || len(chunk.Choices) == 0 || chunk.Choices[0].Delta.Content == "" {
			continue
		}
		writeSSE(c, chunk.Choices[0].Delta.Content)
		c.Writer.Flush()
	}
	return scanner.Err()
}

func (s *Service) openAIRequest(ctx context.Context, message string, stream bool) (*http.Response, error) {
	messages := []aiMessage{{Role: "user", Content: message}}
	if stream {
		messages = append([]aiMessage{{Role: "system", Content: aiHTMLSystemPrompt}}, messages...)
	}
	temperature := s.Config.AI.Temperature
	if !stream {
		temperature = 2
	}
	payload, err := json.Marshal(aiRequest{
		Model: s.Config.AI.Model, Messages: messages, Temperature: temperature, Stream: stream,
	})
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.Config.AI.APIURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+s.Config.AI.APIKey)
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: time.Duration(s.Config.AI.TimeoutSec) * time.Second}
	if stream {
		client.Timeout = 0
	}
	return client.Do(request)
}

func writeSSE(c *gin.Context, content string) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	for _, line := range lines {
		_, _ = fmt.Fprintf(c.Writer, "data: %s\n", line)
	}
	_, _ = io.WriteString(c.Writer, "\n")
}
