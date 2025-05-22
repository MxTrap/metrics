package main

import (
	"context"
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/agent/app"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.NewAgentConfig()
	ctx, cancel := context.WithCancel(context.Background())

	if err != nil {
		log.Fatal(err)
	}
	clientApp := app.NewApp(cfg)

	err = clientApp.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	cancel()
}
