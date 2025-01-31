package config

type HTTPConfig struct {
	Host string
	Port int16
}

type ServerConfig struct {
	HTTP HTTPConfig
}

type AgentConfig struct {
	ServerConfig HTTPConfig
	ClientConfig HTTPConfig
}

func LoadServerConfig() *ServerConfig {
	return &ServerConfig{
		HTTP: HTTPConfig{
			Host: "localhost",
			Port: 8080,
		},
	}
}
