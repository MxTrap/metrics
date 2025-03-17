package app

import (
	"context"
	"fmt"
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/agent/httpclient"
	"github.com/MxTrap/metrics/internal/agent/repository"
	"github.com/MxTrap/metrics/internal/agent/service"
)

type App struct {
	service *service.MetricsObserverService
	client  *httpclient.HTTPClient
}

func NewApp(cfg *config.AgentConfig) *App {
	storage := repository.NewMetricsStorage()
	mService := service.NewMetricsObserverService(storage, cfg.PollInterval)
	client := httpclient.NewHTTPClient(
		mService,
		fmt.Sprintf("%s:%d", cfg.ServerConfig.Host, cfg.ServerConfig.Port),
		cfg.ReportInterval,
		cfg.Key,
		cfg.RateLimit,
	)

	return &App{
		service: mService,
		client:  client,
	}
}

func (a App) Run(ctx context.Context) error {
	a.service.Run(ctx)
	a.client.Run(ctx)
	return nil
}
