package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port              int
	DatabasePath      string
	BaseURL           string
	FeishuAppID       string
	FeishuAppSecret   string
	FeishuRedirectURL string
	FeishuBotName     string
	SessionCookieName string
}

func Load() (Config, error) {
	port := 8080
	if raw := os.Getenv("PORT"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, fmt.Errorf("invalid PORT: %w", err)
		}
		port = parsed
	}

	cfg := Config{
		Port:              port,
		DatabasePath:      getenv("DATABASE_PATH", "./data/requirement-delivery.db"),
		BaseURL:           getenv("BASE_URL", fmt.Sprintf("http://localhost:%d", port)),
		FeishuAppID:       os.Getenv("FEISHU_APP_ID"),
		FeishuAppSecret:   os.Getenv("FEISHU_APP_SECRET"),
		FeishuRedirectURL: getenv("FEISHU_REDIRECT_URL", fmt.Sprintf("http://localhost:%d/api/auth/feishu/callback", port)),
		FeishuBotName:     getenv("FEISHU_BOT_NAME", "需求交付机器人"),
		SessionCookieName: getenv("SESSION_COOKIE_NAME", "feishu_pipeline_session"),
	}

	return cfg, nil
}

func (c Config) FeishuEnabled() bool {
	return c.FeishuAppID != "" && c.FeishuAppSecret != ""
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
