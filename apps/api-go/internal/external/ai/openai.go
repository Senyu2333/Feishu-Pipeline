package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIClient OpenAI兼容API客户端
type OpenAIClient struct {
	baseURL    string
	model      string
	apiKey     string
	maxTokens  int
	timeout    time.Duration
	httpClient *http.Client
}

// OpenAIConfig OpenAI客户端配置
type OpenAIConfig struct {
	BaseURL     string
	Model       string
	APIKey      string
	Temperature float32
	MaxTokens   int
	Timeout     time.Duration
}

// OpenAICompletionRequest OpenAI completion请求结构
type OpenAICompletionRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float32         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

// OpenAIMessage OpenAI消息结构
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAICompletionResponse OpenAI completion响应结构
type OpenAICompletionResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
	Error   *OpenAIError   `json:"error,omitempty"`
}

// OpenAIChoice OpenAI响应选项
type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// OpenAIUsage Token使用统计
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIError OpenAI错误结构
type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}

// NewOpenAIClient 创建OpenAI兼容客户端
func NewOpenAIClient(cfg OpenAIConfig) *OpenAIClient {
	return &OpenAIClient{
		baseURL:   cfg.BaseURL,
		model:     cfg.Model,
		apiKey:    cfg.APIKey,
		maxTokens: cfg.MaxTokens,
		timeout:   cfg.Timeout,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Generate 生成文本
func (c *OpenAIClient) Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, TokenUsage, error) {
	reqBody := OpenAICompletionRequest{
		Model: c.model,
		Messages: []OpenAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens: c.maxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", TokenUsage{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(c.baseURL, "/")), bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", TokenUsage{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", TokenUsage{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", TokenUsage{}, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp OpenAICompletionResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
			return "", TokenUsage{}, fmt.Errorf("api error: %s (code: %s, type: %s)", errResp.Error.Message, errResp.Error.Code, errResp.Error.Type)
		}
		return "", TokenUsage{}, fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var respBody OpenAICompletionResponse
	if err := json.Unmarshal(body, &respBody); err != nil {
		return "", TokenUsage{}, fmt.Errorf("unmarshal response: %w, body: %s", err, string(body))
	}

	if len(respBody.Choices) == 0 {
		return "", TokenUsage{}, fmt.Errorf("no choices in response")
	}

	content := strings.TrimSpace(respBody.Choices[0].Message.Content)
	if content == "" {
		return "", TokenUsage{}, fmt.Errorf("empty content in response")
	}

	usage := TokenUsage{
		InputTokens:  respBody.Usage.PromptTokens,
		OutputTokens: respBody.Usage.CompletionTokens,
		TotalTokens:  respBody.Usage.TotalTokens,
		Raw: map[string]any{
			"prompt_tokens":     respBody.Usage.PromptTokens,
			"completion_tokens": respBody.Usage.CompletionTokens,
			"total_tokens":      respBody.Usage.TotalTokens,
		},
	}

	return content, usage, nil
}

// GenerateStream 流式生成（暂未实现）
func (c *OpenAIClient) GenerateStream(ctx context.Context, systemPrompt string, userPrompt string, ch chan<- string) error {
	return fmt.Errorf("stream generation not implemented for OpenAI client")
}
