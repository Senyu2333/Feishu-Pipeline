package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	arkmodel "github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
)

type Client interface {
	Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, error)
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
