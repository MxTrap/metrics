package config

import (
	"flag"
	"os"
)

type ServerConfig struct {
	HTTP HTTPConfig
}

func LoadServerConfig() *ServerConfig {
	httpConfig := GetDefaultHTTPConfig()
	_ = flag.Value(httpConfig)

	flag.Var(httpConfig, "a", "")
	flag.Parse()

	if addr := os.Getenv("ADDRESS"); addr != "" {
		_ = httpConfig.Set(addr)
	}

	return &ServerConfig{
		HTTP: *httpConfig,
	}
}
