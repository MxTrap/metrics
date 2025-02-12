package app

import (
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/server/httpserver"
	"github.com/MxTrap/metrics/internal/server/logger"
	"github.com/MxTrap/metrics/internal/server/repository"
	"github.com/MxTrap/metrics/internal/server/service"
)

type App struct {
	httpServer *httpserver.HTTPServer
}

func NewApp(cfg *config.ServerConfig) *App {
	log := logger.NewLogger()

	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	http := httpserver.NewRouter(cfg.HTTP, metricsService, log)

	return &App{
		httpServer: http,
	}
}

func (a App) Run() {
	a.httpServer.Run()
}
