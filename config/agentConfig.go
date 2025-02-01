package config

import "flag"

type AgentConfig struct {
	ServerConfig   HTTPConfig
	ReportInterval int
	PollInterval   int
}

func LoadAgentConfig() *AgentConfig {
	rInterval := flag.Int("dest", 10, "interval of sending data to server")
	pInterval := flag.Int("w", 2, "interval of data collecting from runtime")
	httpConfig := GetDefaultHTTPConfig()
	_ = flag.Value(httpConfig)
	flag.Var(httpConfig, "a", "")
	flag.Parse()

	return &AgentConfig{
		ServerConfig:   *httpConfig,
		ReportInterval: *rInterval,
		PollInterval:   *pInterval,
	}
}
