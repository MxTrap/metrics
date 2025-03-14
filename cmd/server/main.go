package main

import (
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/server/app"
	"log"
)

func main() {
	cfg, err := config.NewServerConfig()
	if err != nil {
		log.Fatal(err)
	}
	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatal("Application initialization failed: ", err)
	}

	err = application.Run()
	if err != nil {
		log.Fatal("Application run failed: ", err)
	}

	defer application.Shutdown()
}
