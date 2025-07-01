package serverconfig

import (
	"encoding/json"
	"flag"
	"github.com/MxTrap/metrics/config"
	"github.com/caarlos0/env/v11"
	"os"
	"reflect"
	"time"
)

type ServerConfig struct {
	HTTPAddr        config.AddrConfig `env:"ADDRESS"`
	GRPCAddr        config.AddrConfig `env:"GRPC_ADDRESS"`
	StoreInterval   int               `env:"STORE_INTERVAL"`
	FileStoragePath string            `env:"FILE_STORAGE_PATH"`
	Restore         bool              `env:"RESTORE"`
	DatabaseDSN     string            `env:"DATABASE_DSN"`
	Key             string            `env:"KEY"`
	CryptoKey       string            `env:"CRYPTO_KEY"`
	TrustedSubnet   string            `env:"TRUSTED_SUBNET"`
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
	cryptoKey := flag.String("crypto-key", "", "secret key")
	databaseDSN := flag.String("d", "", "database DSN")
	trustedSubnet := flag.String("t", "", "trusted subnet")

	httpAddr := config.NewDefaultHTTPAddr()
	flag.Var(&httpAddr, "a", "server host:port")
	flag.Parse()
	grpcAddr := config.NewDefaultGRPCAddr()
	flag.Var(&grpcAddr, "g", "server host:port")
	flag.Parse()

	cfg.HTTPAddr = httpAddr
	cfg.GRPCAddr = grpcAddr
	cfg.StoreInterval = *sInterval
	cfg.FileStoragePath = *sPath
	cfg.Restore = *restore
	cfg.DatabaseDSN = *databaseDSN
	cfg.Key = *key
	cfg.CryptoKey = *cryptoKey
	cfg.TrustedSubnet = *trustedSubnet
}

func (cfg *ServerConfig) parseFromEnv() error {
	return env.ParseWithOptions(cfg, env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(config.AddrConfig{}): func(v string) (interface{}, error) {
				httpConfig := config.AddrConfig{}
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
	cfgPath := flag.String("c", "", "path to config file")

	type path struct {
		Path string `env:"CONFIG"`
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
		HTTPAddress   string `json:"address"`
		GRPCAddress   string `json:"grpc_address"`
		Restore       bool   `json:"restore"`
		StoreInterval string `json:"store_interval"`
		StoreFile     string `json:"store_file"`
		DatabaseDsn   string `json:"database_dsn"`
		CryptoKey     string `json:"crypto_key"`
		TrustedSubnet string `json:"trusted_subnet"`
	}
	tmp := tmpConfig{}
	err = json.Unmarshal(fileBytes, &tmp)
	if err != nil {
		return err
	}

	if tmp.HTTPAddress != "" {
		httpConfig := config.NewDefaultHTTPAddr()
		err = httpConfig.Set(tmp.HTTPAddress)
		if err != nil {
			return err
		}
		cfg.HTTPAddr = httpConfig
	}

	if tmp.GRPCAddress != "" {
		addrConfig := config.NewDefaultGRPCAddr()
		err = addrConfig.Set(tmp.GRPCAddress)
		if err != nil {
			return err
		}
		cfg.GRPCAddr = addrConfig
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
	cfg.TrustedSubnet = tmp.TrustedSubnet

	return nil
}
