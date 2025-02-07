package config

import (
	"flag"
	"os"
)

type ServerConfig struct {
	HTTP HTTPConfig
}

func NewServerConfig() *ServerConfig {
	httpConfig := NewDefaultHTTPConfig()
	flag.Var(httpConfig, "a", "")
	flag.Parse()

	if addr := os.Getenv("ADDRESS"); addr != "" {
		_ = httpConfig.Set(addr)
	}

	return &ServerConfig{
		HTTP: *httpConfig,
	}
}
