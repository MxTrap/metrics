package config

type HTTPConfig struct {
	Host string
	Port int16
}

type Config struct {
	HTTP HTTPConfig
}

func LoadConfig() *Config {
	return &Config{
		HTTP: HTTPConfig{
			Host: "localhost",
			Port: 8080,
		},
	}
}
