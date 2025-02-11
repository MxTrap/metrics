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
	ctx     context.Context
	service *service.MetricsObserverService
	client  *httpclient.HTTPClient
}

func NewApp(cfg *config.AgentConfig) *App {
	ctx := context.Background()
	storage := repository.NewMetricsStorage()
	mService := service.NewMetricsObserverService(ctx, storage, cfg.PollInterval)
	client := httpclient.NewHTTPClient(
		ctx,
		mService,
		fmt.Sprintf("%s:%d", cfg.ServerConfig.Host, cfg.ServerConfig.Port),
		cfg.ReportInterval,
	)

	return &App{
		ctx:     ctx,
		service: mService,
		client:  client,
	}
}

func (a App) Run() {
	a.service.Run()
	a.client.Run()
}

func (a App) Shutdown() {
	a.ctx.Done()
}
