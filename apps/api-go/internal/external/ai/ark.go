package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	arkmodel "github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
)

// Client AI 客户端接口
type Client interface {
	Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, error)
	// GenerateStream 流式生成，每个 token 写入 ch，结束后关闭 ch
	GenerateStream(ctx context.Context, systemPrompt string, userPrompt string, ch chan<- string) error
}

type ArkConfig struct {
	BaseURL     string
	Model       string
	APIKey      string
	Temperature float32
	MaxTokens   int
	Timeout     time.Duration
}

type ArkClient struct {
	model *arkmodel.ChatModel
}

func NewArkClient(ctx context.Context, cfg ArkConfig) (*ArkClient, error) {
	timeout := cfg.Timeout
	maxTokens := cfg.MaxTokens
	temperature := cfg.Temperature

	model, err := arkmodel.NewChatModel(ctx, &arkmodel.ChatModelConfig{
		BaseURL:     cfg.BaseURL,
		Model:       cfg.Model,
		APIKey:      cfg.APIKey,
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
		Timeout:     &timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("create ark chat model: %w", err)
	}

	return &ArkClient{model: model}, nil
}

func (c *ArkClient) Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	message, err := c.model.Generate(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	})
	if err != nil {
		return "", fmt.Errorf("generate with ark: %w", err)
	}
	return strings.TrimSpace(message.Content), nil
}

// GenerateStream 调用 Ark Stream 接口，逐 token 发送到 ch
func (c *ArkClient) GenerateStream(ctx context.Context, systemPrompt string, userPrompt string, ch chan<- string) error {
	reader, err := c.model.Stream(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	})
	if err != nil {
		return fmt.Errorf("stream with ark: %w", err)
	}
	defer reader.Close()

	for {
		msg, err := reader.Recv()
		if err != nil {
			// io.EOF 表示流结束，不是错误
			break
		}
		if msg != nil && msg.Content != "" {
			select {
			case ch <- msg.Content:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return nil
}
