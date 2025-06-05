package config

import (
	"flag"
	"github.com/caarlos0/env"
	"reflect"
)

type AgentConfig struct {
	ServerConfig   HTTPConfig `env:"ADDRESS"`
	ReportInterval int        `env:"REPORT_INTERVAL"`
	PollInterval   int        `env:"POLL_INTERVAL"`
	Key            string     `env:"KEY"`
	RateLimit      int        `env:"RATE_LIMIT"`
	CryptoKey      string     `env:"CRYPTO_KEY"`
}

func NewAgentConfig() (*AgentConfig, error) {
	rInterval := flag.Int("r", 10, "interval of sending data to server")
	pInterval := flag.Int("p", 2, "interval of data collecting from runtime")
	key := flag.String("k", "", "secret key")
	rateLimit := flag.Int("l", 1, "rate limit")
	cryptoKey := flag.String("crypto-key", "", "crypto key")
	httpConfig := NewDefaultConfig()
	flag.Var(&httpConfig, "a", "server host:port")
	flag.Parse()

	agentConfig := &AgentConfig{
		ServerConfig:   httpConfig,
		ReportInterval: *rInterval,
		PollInterval:   *pInterval,
		Key:            *key,
		RateLimit:      *rateLimit,
		CryptoKey:      *cryptoKey,
	}

	err := env.ParseWithFuncs(agentConfig, map[reflect.Type]env.ParserFunc{
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

	return agentConfig, nil
}
