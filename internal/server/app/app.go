package app

import (
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/server/httpserver"
	"github.com/MxTrap/metrics/internal/server/service"
)

type App struct {
	httpServer *httpserver.HttpServer
}

func NewApp(cfg *config.Config) *App {
	service := service.NewMemStorage()
	http := httpserver.New(cfg.Http, service)

	return &App{
		httpServer: http,
	}
}

func (a App) Run() {
	a.httpServer.Run()
}
