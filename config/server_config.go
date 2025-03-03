package config

import (
	"flag"
	"github.com/caarlos0/env"
	"reflect"
)

type ServerConfig struct {
	HTTP            HTTPConfig `env:"ADDRESS"`
	StoreInterval   int        `env:"STORE_INTERVAL"`
	FileStoragePath string     `env:"FILE_STORAGE_PATH"`
	Restore         bool       `env:"RESTORE"`
	DatabaseDSN     string     `env:"DATABASE_DSN"`
}

func NewServerConfig() (*ServerConfig, error) {
	sInterval := flag.Int("i", 300, "interval of saving data to file")
	sPath := flag.String("f", "./temp.txt", "path to file")
	restore := flag.Bool("r", true, "restore data")
	//postgres://postgres:admin@localhost:5432/metrics?sslmode=disable
	databaseDSN := flag.String("d", "", "database DSN")
	httpConfig := NewDefaultConfig()
	flag.Var(&httpConfig, "a", "server host:port")
	flag.Parse()

	cfg := &ServerConfig{
		HTTP:            httpConfig,
		StoreInterval:   *sInterval,
		FileStoragePath: *sPath,
		Restore:         *restore,
		DatabaseDSN:     *databaseDSN,
	}

	err := env.ParseWithFuncs(cfg, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(HTTPConfig{}): func(v string) (interface{}, error) {
			cfg := HTTPConfig{}
			err := cfg.Set(v)
			if err != nil {
				return nil, err
			}
			return cfg, nil
		},
	})
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
