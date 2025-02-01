package config

import (
	"flag"
	"fmt"
)

type ServerConfig struct {
	HTTP HTTPConfig
}

func LoadServerConfig() *ServerConfig {
	httpConfig := GetDefaultHTTPConfig()
	_ = flag.Value(httpConfig)

	flag.Var(httpConfig, "a", "")
	flag.Parse()

	fmt.Println(httpConfig, *httpConfig)

	return &ServerConfig{
		HTTP: *httpConfig,
	}
}
