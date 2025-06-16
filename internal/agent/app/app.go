package app

import (
	"context"
	"fmt"
	"github.com/MxTrap/metrics/config/agentconfig"
	"github.com/MxTrap/metrics/internal/agent/httpclient"
	"github.com/MxTrap/metrics/internal/agent/repository"
	"github.com/MxTrap/metrics/internal/agent/service"
	"os"
)

type runner interface {
	Run(ctx context.Context)
}

type App struct {
	service runner
	client  runner
}

func NewApp(cfg *agentconfig.AgentConfig) *App {
	storage := repository.NewMetricsStorage()
	mService := service.NewMetricsObserverService(storage, cfg.PollInterval)

	client := httpclient.NewHTTPClient(
		mService,
		fmt.Sprintf("%s:%d", cfg.ServerConfig.Host, cfg.ServerConfig.Port),
		cfg.ReportInterval,
		cfg.Key,
		cfg.RateLimit,
	)

	if cfg.CryptoKey != "" {
		encrypter, err := service.NewEncrypterSvc(cfg.CryptoKey)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		client.RegisterEncrypter(encrypter)
	}

	return &App{
		service: mService,
		client:  client,
	}
}

func (a App) Run(ctx context.Context) error {
	fmt.Println("starting metrics observer")
	go a.service.Run(ctx)
	go a.client.Run(ctx)
	return nil
}
