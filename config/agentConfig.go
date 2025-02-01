package config

import "flag"

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

	return &AgentConfig{
		ServerConfig:   *httpConfig,
		ReportInterval: *rInterval,
		PollInterval:   *pInterval,
	}
}
