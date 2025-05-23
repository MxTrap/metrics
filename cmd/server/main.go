package main

import (
	"fmt"
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/server/app"
	"log"
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
