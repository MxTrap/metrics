package config

import (
	"flag"
	"os"
	"strconv"
)

type AgentConfig struct {
	ServerConfig   HTTPConfig
	ReportInterval int
	PollInterval   int
}

func LoadAgentConfig() *AgentConfig {
	rInterval := flag.Int("r", 10, "interval of sending data to server")
	pInterval := flag.Int("p", 2, "interval of data collecting from runtime")
	httpConfig := GetDefaultHTTPConfig()
	_ = flag.Value(httpConfig)
	flag.Var(httpConfig, "a", "server host:port")
	flag.Parse()

	if interval := os.Getenv("REPORT_INTERVAL"); interval != "" {
		iInterval, err := strconv.Atoi(interval)
		if err == nil {
			rInterval = &iInterval
		}
	}

	if interval := os.Getenv("POLL_INTERVAL"); interval != "" {
		iInterval, err := strconv.Atoi(interval)
		if err == nil {
			pInterval = &iInterval
		}
	}

	if addr := os.Getenv("ADDRESS"); addr != "" {
		_ = httpConfig.Set(addr)
	}

	return &AgentConfig{
		ServerConfig:   *httpConfig,
		ReportInterval: *rInterval,
		PollInterval:   *pInterval,
	}
}
