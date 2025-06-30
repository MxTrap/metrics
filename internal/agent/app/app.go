package app

import (
	"context"
	"fmt"
	"github.com/MxTrap/metrics/config/agentconfig"
	"github.com/MxTrap/metrics/internal/agent/grpc"
	"github.com/MxTrap/metrics/internal/agent/http"
	"github.com/MxTrap/metrics/internal/agent/repository"
	"github.com/MxTrap/metrics/internal/agent/service"
	"log"
)

type runner interface {
	Run(ctx context.Context)
}

type App struct {
	service    runner
	httpClient runner
	grpcClient runner
}

func NewApp(cfg *agentconfig.AgentConfig) *App {
	storage := repository.NewMetricsStorage()
	mService := service.NewMetricsObserverService(storage, cfg.PollInterval)

	httpClient := http.NewClient(
		mService,
		fmt.Sprintf("%s:%d", cfg.HTTPServerAddr.Host, cfg.HTTPServerAddr.Port),
		cfg.ReportInterval,
		cfg.Key,
		cfg.RateLimit,
	)
	grpcClient, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.GRPCServerAddr.Host, cfg.GRPCServerAddr.Port),
		mService,
		cfg.ReportInterval,
		cfg.Key,
		cfg.RateLimit,
	)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.CryptoKey != "" {
		encrypter, err := service.NewEncrypterSvc(cfg.CryptoKey)
		if err != nil {
			log.Fatal(err)
		}
		httpClient.RegisterEncrypter(encrypter)
	}

	return &App{
		service:    mService,
		httpClient: httpClient,
		grpcClient: grpcClient,
	}
}

func (a *App) Run(ctx context.Context) error {
	fmt.Println("starting metrics observer")
	go a.service.Run(ctx)
	go a.httpClient.Run(ctx)
	go a.grpcClient.Run(ctx)
	return nil
}
