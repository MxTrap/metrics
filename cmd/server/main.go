package main

import (
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/server/app"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.NewServerConfig()
	if err != nil {
		log.Fatal(err)
	}
	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatal("Application initialization failed")
	}

	application.Run()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	application.Shutdown()
}
