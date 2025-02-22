package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env"
	"reflect"
)

type ServerConfig struct {
	HTTP            HTTPConfig `env:"ADDRESS"`
	StoreInterval   int        `env:"STORE_INTERVAL" envDefault:"300"`
	FileStoragePath string     `env:"FILE_STORAGE_PATH" envDefault:"./data"`
	Restore         bool       `env:"RESTORE" envDefault:"false"`
}

func NewServerConfig() (*ServerConfig, error) {
	sInterval := flag.Int("i", 300, "interval of saving data to file")
	sPath := flag.String("f", "", "path to file")
	restore := flag.Bool("r", false, "restore data")
	httpConfig := NewDefaultConfig()
	flag.Var(&httpConfig, "a", "server host:port")
	flag.Parse()

	cfg := &ServerConfig{
		HTTP:            httpConfig,
		StoreInterval:   *sInterval,
		FileStoragePath: *sPath,
		Restore:         *restore,
	}

	fmt.Println(cfg)

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
