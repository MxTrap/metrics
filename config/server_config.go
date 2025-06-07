package config

import (
	"encoding/json"
	"flag"
	"github.com/MxTrap/metrics/internal/utils"
	"github.com/caarlos0/env/v11"
	"os"
	"reflect"
	"time"
)

type ServerConfig struct {
	HTTP            HTTPConfig `env:"ADDRESS"`
	StoreInterval   int        `env:"STORE_INTERVAL"`
	FileStoragePath string     `env:"FILE_STORAGE_PATH"`
	Restore         bool       `env:"RESTORE"`
	DatabaseDSN     string     `env:"DATABASE_DSN"`
	Key             string     `env:"KEY"`
	CryptoKey       string     `env:"CRYPTO_KEY"`
}

func NewServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}

	err := cfg.parseFromFile()
	if err != nil {
		return nil, err
	}
	cfg.parseFromFlags()

	err = cfg.parseFromEnv()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *ServerConfig) parseFromFlags() {
	sInterval := flag.Int("i", 300, "interval of saving data to file")
	sPath := flag.String("f", "./temp.txt", "path to file")
	restore := flag.Bool("r", false, "restore data")
	key := flag.String("k", "", "secret key")
	cryptoKey := flag.String("crypto-key", utils.GetProjectPath()+"/keys/private.pem", "secret key")
	databaseDSN := flag.String("d", "", "database DSN")
	httpConfig := NewDefaultConfig()
	flag.Var(&httpConfig, "a", "server host:port")
	flag.Parse()

	cfg.HTTP = httpConfig
	cfg.StoreInterval = *sInterval
	cfg.FileStoragePath = *sPath
	cfg.Restore = *restore
	cfg.DatabaseDSN = *databaseDSN
	cfg.Key = *key
	cfg.CryptoKey = *cryptoKey
}

func (cfg *ServerConfig) parseFromEnv() error {
	return env.ParseWithOptions(cfg, env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(HTTPConfig{}): func(v string) (interface{}, error) {
				httpConfig := HTTPConfig{}
				err := httpConfig.Set(v)
				if err != nil {
					return nil, err
				}
				return httpConfig, nil
			},
		},
	})
}

func (cfg *ServerConfig) parseFromFile() error {
	cfgPath := flag.String("c", utils.GetProjectPath()+"/config/server_config.json", "path to config file")

	type path struct {
		Path string `env:"Config"`
	}
	envPath, err := env.ParseAs[path]()
	if err != nil {
		return err
	}
	if envPath.Path != "" {
		cfgPath = &envPath.Path
	}
	if *cfgPath == "" {
		return nil
	}
	fileBytes, err := os.ReadFile(*cfgPath)
	if err != nil {
		return err
	}

	type tmpConfig struct {
		Address       string `json:"address"`
		Restore       bool   `json:"restore"`
		StoreInterval string `json:"store_interval"`
		StoreFile     string `json:"store_file"`
		DatabaseDsn   string `json:"database_dsn"`
		CryptoKey     string `json:"crypto_key"`
	}
	tmp := tmpConfig{}
	err = json.Unmarshal(fileBytes, &tmp)
	if err != nil {
		return err
	}

	if tmp.Address != "" {
		httpConfig := NewDefaultConfig()
		err = httpConfig.Set(tmp.Address)
		if err != nil {
			return err
		}
		cfg.HTTP = httpConfig
	}

	if tmp.StoreInterval != "" {
		dStoreInterval, err := time.ParseDuration(tmp.StoreInterval)
		if err != nil {
			return err
		}
		cfg.StoreInterval = int(dStoreInterval.Seconds())
	}

	cfg.FileStoragePath = tmp.StoreFile
	cfg.Restore = tmp.Restore
	cfg.DatabaseDSN = tmp.DatabaseDsn
	cfg.CryptoKey = tmp.CryptoKey

	return nil
}
