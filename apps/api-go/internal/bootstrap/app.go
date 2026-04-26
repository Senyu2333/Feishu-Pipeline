package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"feishu-pipeline/apps/api-go/internal/agent"
	"feishu-pipeline/apps/api-go/internal/controller"
	"feishu-pipeline/apps/api-go/internal/external/ai"
	"feishu-pipeline/apps/api-go/internal/external/feishu"
	"feishu-pipeline/apps/api-go/internal/job"
	"feishu-pipeline/apps/api-go/internal/pipeline"
	"feishu-pipeline/apps/api-go/internal/repo"
	"feishu-pipeline/apps/api-go/internal/router"
	"feishu-pipeline/apps/api-go/internal/service"
)

type Application struct {
	Config     *Config
	HTTPServer *http.Server
	repository *repo.Repository
}

func NewApplication(ctx context.Context, configPath string, version string) (*Application, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	if version != "" {
		cfg.App.Version = version
	}

	repository, err := repo.NewSQLiteRepository(cfg.Database.Path)
	if err != nil {
		return nil, err
	}

	var aiClient ai.Client
	if cfg.AI.Provider == "ark" && cfg.AI.Ark.APIKey != "" {
		arkClient, err := ai.NewArkClient(ctx, ai.ArkConfig{
			BaseURL:     cfg.AI.Ark.BaseURL,
			Model:       cfg.AI.Ark.Model,
			APIKey:      cfg.AI.Ark.APIKey,
			Temperature: cfg.AI.Ark.Temperature,
			MaxTokens:   cfg.AI.Ark.MaxTokens,
			Timeout:     time.Duration(cfg.AI.Ark.TimeoutSec) * time.Second,
		})
		if err != nil {
			return nil, err
		}
		aiClient = arkClient
	} else {
		log.Printf("ark ai disabled: provider=%s api_key_configured=%t", cfg.AI.Provider, cfg.AI.Ark.APIKey != "")
	}

	agentEngine, err := agent.NewEngine(aiClient)
	if err != nil {
		return nil, err
	}
	feishuClient := feishu.NewClient(feishu.Config{
		Enabled:              cfg.Feishu.Enabled,
		AppID:                cfg.Feishu.AppID,
		AppSecret:            cfg.Feishu.AppSecret,
		RedirectURL:          cfg.Feishu.RedirectURL,
		OpenBaseURL:          cfg.Feishu.OpenBaseURL,
		OAuthScope:           cfg.Feishu.OAuthScope,
		BotName:              cfg.Feishu.BotName,
		ReceiveIDType:        cfg.Feishu.ReceiveIDType,
		DocFolderToken:       cfg.Feishu.DocFolderToken,
		BitableName:          cfg.Feishu.BitableName,
		BitableFolderToken:   cfg.Feishu.BitableFolderToken,
		BitableAppToken:      cfg.Feishu.BitableAppToken,
		BitableTableID:       cfg.Feishu.BitableTableID,
		BitableTemplateToken: cfg.Feishu.BitableTemplateToken,
		BaseURL:              cfg.App.BaseURL,
	})

	authService := service.NewAuthService(repository, feishuClient, time.Duration(cfg.App.SessionTTLHours)*time.Hour)
	healthService := service.NewHealthService(cfg.App.Name, cfg.App.Version)
	sessionService := service.NewSessionService(repository, authService, aiClient)
	taskService := service.NewTaskService(repository, feishuClient)
	adminService := service.NewAdminService(repository)
	pipelineProvider := pipeline.NewTextGenerationProvider(cfg.AI.Provider, cfg.AI.Ark.Model, aiClient)
	pipelineExecutor := pipeline.NewSequentialExecutor(pipeline.WithAgentRunner(pipeline.NewAgentRunner(pipelineProvider, pipeline.DefaultPromptRegistry())))
	pipelineService := service.NewPipelineService(repository, service.WithPipelineExecutor(pipelineExecutor))
	publishService := service.NewPublishService(repository, authService, agentEngine, feishuClient, pipelineService)
	sessionService.SetPublisher(publishService)

	runner := job.NewRunner(nil, publishService, pipelineService)
	publishService.SetQueue(runner)
	pipelineService.SetQueue(runner)
	runner.Start(ctx)

	engine := router.New(router.Dependencies{
		CookieName:         cfg.App.SessionCookieName,
		HealthController:   controller.NewHealthController(healthService),
		AuthController:     controller.NewAuthController(authService, cfg.App.SessionCookieName, time.Duration(cfg.App.SessionTTLHours)*time.Hour, cfg.App.CookieSecure, cfg.App.CookieSameSite),
		SessionController:  controller.NewSessionController(sessionService, publishService),
		TaskController:     controller.NewTaskController(taskService),
		AdminController:    controller.NewAdminController(adminService),
		PipelineController: controller.NewPipelineController(pipelineService),
		AuthService:        authService,
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: engine,
	}

	return &Application{
		Config:     cfg,
		HTTPServer: server,
		repository: repository,
	}, nil
}

func (a *Application) Close() error {
	if a.repository != nil {
		return a.repository.Close()
	}
	return nil
}
