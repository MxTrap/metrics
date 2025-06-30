package agentconfig

import (
	"encoding/json"
	"flag"
	"github.com/MxTrap/metrics/config"
	"github.com/caarlos0/env/v11"
	"os"
	"reflect"
	"time"
)

type AgentConfig struct {
	HTTPServerAddr config.AddrConfig `env:"ADDRESS"`
	GRPCServerAddr config.AddrConfig `env:"GRPC_ADDRESS"`
	ReportInterval int               `env:"REPORT_INTERVAL"`
	PollInterval   int               `env:"POLL_INTERVAL"`
	Key            string            `env:"KEY"`
	RateLimit      int               `env:"RATE_LIMIT"`
	CryptoKey      string            `env:"CRYPTO_KEY"`
}

func NewAgentConfig() (*AgentConfig, error) {
	agentConfig := &AgentConfig{}
	err := agentConfig.parseFromFile()
	if err != nil {
		return nil, err
	}

	agentConfig.parseFromFlags()
	err = agentConfig.parseFromEnv()
	if err != nil {
		return nil, err
	}

	return agentConfig, nil
}

func (cfg *AgentConfig) parseFromFlags() {
	rInterval := flag.Int("r", 10, "interval of sending data to server")
	pInterval := flag.Int("p", 2, "interval of data collecting from runtime")
	key := flag.String("k", "", "secret key")
	rateLimit := flag.Int("l", 1, "rate limit")
	cryptoKey := flag.String("crypto-key", "", "crypto key")

	httpAddr := config.NewDefaultHTTPAddr()
	flag.Var(&httpAddr, "a", "server host:port")
	flag.Parse()

	grpcAddr := config.NewDefaultGRPCAddr()
	flag.Var(&grpcAddr, "g", "server host:port")
	flag.Parse()

	cfg.HTTPServerAddr = httpAddr
	cfg.GRPCServerAddr = grpcAddr
	cfg.ReportInterval = *rInterval
	cfg.PollInterval = *pInterval
	cfg.Key = *key
	cfg.RateLimit = *rateLimit
	cfg.CryptoKey = *cryptoKey

}

func (cfg *AgentConfig) parseFromEnv() error {
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

func (cfg *AgentConfig) parseFromFile() error {
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
		Address        string `json:"address"`
		GRPCAddress    string `json:"grpc_address"`
		ReportInterval string `json:"report_interval"`
		PollInterval   string `json:"poll_interval"`
		CryptoKey      string `json:"crypto_key"`
	}
	tmp := &tmpConfig{}

	err = json.Unmarshal(fileBytes, tmp)
	if err != nil {
		return err
	}

	if tmp.Address != "" {
		httpConfig := config.NewDefaultHTTPAddr()
		err = httpConfig.Set(tmp.Address)
		if err != nil {
			return err
		}
		cfg.HTTPServerAddr = httpConfig
	}
	if tmp.GRPCAddress != "" {
		grpcConfig := config.AddrConfig{
			Host: "localhost",
			Port: 9090,
		}
		err = grpcConfig.Set(tmp.GRPCAddress)
		if err != nil {
			return err
		}
		cfg.GRPCServerAddr = grpcConfig
	}
	if tmp.ReportInterval != "" {
		dReportInterval, err := time.ParseDuration(tmp.ReportInterval)
		if err != nil {
			return err
		}
		cfg.ReportInterval = int(dReportInterval.Seconds())
	}

	if tmp.PollInterval != "" {
		dPollInterval, err := time.ParseDuration(tmp.PollInterval)
		if err != nil {
			return err
		}
		cfg.PollInterval = int(dPollInterval.Seconds())
	}

	cfg.CryptoKey = tmp.CryptoKey
	return nil
}
