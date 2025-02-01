package config

import (
	"flag"
)

type ServerConfig struct {
	HTTP HTTPConfig
}

func LoadServerConfig() *ServerConfig {
	httpConfig := GetDefaultHTTPConfig()
	_ = flag.Value(httpConfig)

	flag.Var(httpConfig, "a", "")
	flag.Parse()

	return &ServerConfig{
		HTTP: *httpConfig,
	}
}
