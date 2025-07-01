package main

import (
	"context"
	"github.com/MxTrap/metrics/config/agentconfig"
	"github.com/MxTrap/metrics/internal/agent/app"
	"github.com/MxTrap/metrics/internal/utils"
	"log"
	"net/http"
	_ "net/http/pprof"
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

	cfg, err := agentconfig.NewAgentConfig()
	ctx, cancel := context.WithCancel(context.Background())

	if err != nil {
		log.Fatal(err)
	}
	clientApp := app.NewApp(cfg)

	clientApp.Run(ctx)
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	<-stop

	cancel()
}
