package main

import (
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/agent/app"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.LoadServerConfig()
	clientApp := app.NewApp(cfg)

	clientApp.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

}
