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
	storageService *service.MetricsService
	ctx            context.Context
	logger         *logger.Logger
}

func NewApp(cfg *config.ServerConfig) (*App, error) {
	ctx := context.Background()
	log := logger.NewLogger()

	fileStorage := repository.NewMetricsFileStorage(cfg.FileStoragePath)
	var storage service.Storage
	var storageErr error
	storage, storageErr = repository.NewMemStorage()
	if cfg.DatabaseDSN != "" {
		m, err := migrator.NewMigrator(cfg.DatabaseDSN)
		if err != nil {
			log.Logger.Error("could not create migrator ", err)
			return nil, err
		}
		err = m.InitializeDB()

		if err != nil {
			log.Logger.Error("could not initialize database ", err)
			return nil, err
		}
		storage, storageErr = repository.NewPostgresStorage(ctx, cfg.DatabaseDSN, log)
	}
	if storageErr != nil {
		log.Logger.Error(storageErr)
		return nil, storageErr
	}

	metricsService := service.NewMetricsService(fileStorage, storage, cfg.StoreInterval, cfg.Restore)
	httpRouter := httpserver.NewRouter(cfg.HTTP, log)
	metricHandler := handlers.NewMetricHandler(metricsService, httpRouter.Router)
	metricHandler.RegisterRoutes()

	return &App{
		httpServer: httpRouter,
		ctx:        ctx,
		logger:     log,
	}, nil
}

func (a App) Run() error {
	a.logger.Logger.Info("starting server")
	err := a.storageService.Start(a.ctx)
	if err != nil {
		a.logger.Logger.Error(err.Error())
		return err
	}
	err = a.httpServer.Run()
	if err != nil {
		a.logger.Logger.Error(err)
		return err
	}
	return nil
}

func (a App) Shutdown() {
	a.logger.Logger.Info("shutting down server")
	a.storageService.Stop()
	err := a.httpServer.Stop(a.ctx)
	if err != nil {
		a.logger.Logger.Error(err.Error())
		return
	}
}
