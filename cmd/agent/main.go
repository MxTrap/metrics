package main

import (
	"context"
	"fmt"
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/agent/app"
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

func formatFlagValue(val string) string {
	if val == "" {
		return "N/A"
	}
	return val
}

func printBuildFlags() {
	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		formatFlagValue(BuildVersion),
		formatFlagValue(BuildDate),
		formatFlagValue(BuildCommit),
	)
}

func main() {
	printBuildFlags()

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
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	<-stop

	cancel()
}
