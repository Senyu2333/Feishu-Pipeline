package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"feishu-pipeline/apps/api-go/internal/agent"
	"feishu-pipeline/apps/api-go/internal/config"
	"feishu-pipeline/apps/api-go/internal/feishu"
	"feishu-pipeline/apps/api-go/internal/httpapi"
	"feishu-pipeline/apps/api-go/internal/job"
	"feishu-pipeline/apps/api-go/internal/service"
	"feishu-pipeline/apps/api-go/internal/store"
)

type App struct {
	Config config.Config
	Server *http.Server
	store  *store.Store
}

func New(ctx context.Context, version string) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger := log.New(os.Stdout, "[api-go] ", log.LstdFlags|log.Lshortfile)
	repository, err := store.New(ctx, cfg.DatabasePath)
	if err != nil {
		return nil, err
	}

	svc := service.New(repository, agent.NewEngine(), feishu.NewClient(cfg), version)
	runner := job.NewRunner(logger, svc)
	svc.SetQueue(runner)
	runner.Start(ctx)

	handler := httpapi.NewRouter(logger, svc, cfg.SessionCookieName)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: handler,
	}

	return &App{
		Config: cfg,
		Server: server,
		store:  repository,
	}, nil
}

func (a *App) Close() error {
	if a.store != nil {
		return a.store.Close()
	}
	return nil
}
