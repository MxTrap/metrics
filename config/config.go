package config

type HttpConfig struct {
	Host string
	Port int16
}

type Config struct {
	Http HttpConfig
}

func LoadConfig() *Config {
	return &Config{
		Http: HttpConfig{
			Host: "localhost",
			Port: 8080,
		},
	}
}
