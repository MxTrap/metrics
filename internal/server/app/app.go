package app

import (
	"context"
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/server/httpserver"
	"github.com/MxTrap/metrics/internal/server/httpserver/handlers"
	"github.com/MxTrap/metrics/internal/server/logger"
	"github.com/MxTrap/metrics/internal/server/repository"
	"github.com/MxTrap/metrics/internal/server/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type App struct {
	httpServer     *httpserver.HTTPServer
	storageService *service.StorageService
}

func NewApp(cfg *config.ServerConfig) *App {
	ctx := context.Background()
	log := logger.NewLogger()

	storage := repository.NewMemStorage()
	fileStorage := repository.NewMetricsFileStorage(cfg.FileStoragePath)
	pgStorage, err := repository.NewPostgresStorage(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Logger.Error(err)
	}

	sService := service.NewStorageService(fileStorage, storage, cfg.StoreInterval, cfg.Restore)
	metricsService := service.NewMetricsService(sService)
	httpRouter := httpserver.NewRouter(cfg.HTTP, log)
	httpRouter.Router.GET("/ping", func(c *gin.Context) {
		err := pgStorage.Ping()
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}
		c.Status(http.StatusOK)
	})
	_ = handlers.NewMetricHandler(metricsService, httpRouter.Router)

	return &App{
		httpServer:     httpRouter,
		storageService: sService,
	}
}

func (a App) Run() {
	err := a.storageService.Start()
	if err != nil {
		return
	}
	a.httpServer.Run()
}

func (a App) Stop() {
	a.storageService.Stop()
}
