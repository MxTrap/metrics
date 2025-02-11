package main

import (
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/server/app"
)

func main() {
	cfg := config.NewServerConfig()
	application := app.NewApp(cfg)

	application.Run()
}
