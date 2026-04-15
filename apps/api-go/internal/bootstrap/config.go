package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Feishu   FeishuConfig   `mapstructure:"feishu"`
}

type AppConfig struct {
	Name              string `mapstructure:"name"`
	Version           string `mapstructure:"version"`
	Port              int    `mapstructure:"port"`
	Mode              string `mapstructure:"mode"`
	BaseURL           string `mapstructure:"base_url"`
	SessionCookieName string `mapstructure:"session_cookie_name"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type FeishuConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	AppID           string `mapstructure:"app_id"`
	AppSecret       string `mapstructure:"app_secret"`
	RedirectURL     string `mapstructure:"redirect_url"`
	BotName         string `mapstructure:"bot_name"`
	ReceiveIDType   string `mapstructure:"receive_id_type"`
	BitableAppToken string `mapstructure:"bitable_app_token"`
	BitableTableID  string `mapstructure:"bitable_table_id"`
}

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath("./config")

	if configPath != "" {
		v.SetConfigFile(configPath)
	}

	v.SetEnvPrefix("FEISHU_PIPELINE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.App.Port == 0 {
		cfg.App.Port = 8080
	}
	if cfg.App.Name == "" {
		cfg.App.Name = "requirement-delivery-api"
	}
	if cfg.App.SessionCookieName == "" {
		cfg.App.SessionCookieName = "feishu_pipeline_session"
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "./data/requirement-delivery.db"
	}
	if cfg.Feishu.BotName == "" {
		cfg.Feishu.BotName = "需求交付机器人"
	}
	if cfg.Feishu.ReceiveIDType == "" {
		cfg.Feishu.ReceiveIDType = "open_id"
	}

	cfg.Database.Path = resolvePath(cfg.Database.Path)
	return &cfg, nil
}

func resolvePath(path string) string {
	if path == "" {
		return path
	}
	if filepath.IsAbs(path) {
		return path
	}

	workspace, err := os.Getwd()
	if err != nil {
		return path
	}
	return filepath.Join(workspace, path)
}
