package main

import (
	"context"
	"github.com/MxTrap/metrics/config/serverconfig"
	"github.com/MxTrap/metrics/internal/server/app"
	"github.com/MxTrap/metrics/internal/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	BuildDate    string
	BuildCommit  string
	BuildVersion string
)

func main() {
	utils.PrintBuildFlags(BuildDate, BuildCommit, BuildVersion)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg, err := serverconfig.NewServerConfig()
	if err != nil {
		log.Fatal(err)
	}

	application, err := app.NewApp(cfg, ctx)
	if err != nil {
		log.Fatal("Application initialization failed: ", err)
	}

	err = application.Run(ctx)
	if err != nil {
		log.Fatal("Application run failed: ", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	<-stop

	err = application.GracefulShutdown(ctx)
	if err != nil {
		log.Fatal("Application graceful shutdown failed: ", err)
		return
	}
}
