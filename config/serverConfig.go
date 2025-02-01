package config

import (
	"flag"
)

type ServerConfig struct {
	HTTP HTTPConfig
}

func LoadServerConfig() *ServerConfig {
	httpConfig := new(HTTPConfig)
	_ = flag.Value(httpConfig)

	flag.Var(httpConfig, "a", "")
	flag.Parse()

	if httpConfig == nil {
		dConfig := GetDefaultHTTPConfig()
		httpConfig = &dConfig
	}

	return &ServerConfig{
		HTTP: *httpConfig,
	}
}
