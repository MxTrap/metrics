package app

import (
	"context"
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/server/httpserver"
	"github.com/MxTrap/metrics/internal/server/httpserver/handlers"
	"github.com/MxTrap/metrics/internal/server/logger"
	"github.com/MxTrap/metrics/internal/server/migrator"
	"github.com/MxTrap/metrics/internal/server/repository"
	"github.com/MxTrap/metrics/internal/server/service"
)

type App struct {
	httpServer     *httpserver.HTTPServer
	storageService *service.StorageService
	ctx            context.Context
}

func NewApp(cfg *config.ServerConfig) *App {
	ctx := context.Background()
	log := logger.NewLogger()

	fileStorage := repository.NewMetricsFileStorage(cfg.FileStoragePath)
	var storage service.Storage
	var storageErr error
	storage, storageErr = repository.NewMemStorage()
	if cfg.DatabaseDSN != "" {
		m, err := migrator.NewMigrator(cfg.DatabaseDSN)
		if err != nil {
			return nil
		}
		err = m.InitializeDB()
		if err != nil {
			return nil
		}
		storage, storageErr = repository.NewPostgresStorage(ctx, cfg.DatabaseDSN)
	}
	if storageErr != nil {
		log.Logger.Error(storageErr)
	}

	sService := service.NewStorageService(fileStorage, storage, cfg.StoreInterval, cfg.Restore)
	metricsService := service.NewMetricsService(sService)
	httpRouter := httpserver.NewRouter(cfg.HTTP, log)
	metricHandler := handlers.NewMetricHandler(metricsService, httpRouter.Router)
	metricHandler.RegisterRoutes()

	return &App{
		httpServer:     httpRouter,
		storageService: sService,
		ctx:            ctx,
	}
}

func (a App) Run() {
	err := a.storageService.Start(a.ctx)
	if err != nil {
		return
	}
	a.httpServer.Run()
}

func (a App) Stop() {
	a.storageService.Stop()
}
