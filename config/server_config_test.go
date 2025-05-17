package config

import (
	"github.com/caarlos0/env"
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
)

func TestNewServerConfig_EnvParsing(t *testing.T) {
	config := &ServerConfig{
		HTTP: HTTPConfig{
			Host: "localhost",
			Port: 8080,
		},
		StoreInterval:   300,
		FileStoragePath: "./temp.txt",
		Restore:         false,
		DatabaseDSN:     "",
		Key:             "",
	}

	envVars := map[string]string{
		"ADDRESS":           "env:8080",
		"STORE_INTERVAL":    "900",
		"FILE_STORAGE_PATH": "/env/metrics.txt",
		"RESTORE":           "true",
		"DATABASE_DSN":      "postgres://env:env@localhost:5432/envdb",
		"KEY":               "envkey",
	}

	// Устанавливаем переменные окружения
	for k, v := range envVars {
		err := os.Setenv(k, v)
		assert.NoError(t, err, "Failed to set env var")
		defer os.Unsetenv(k)
	}

	// Вызываем env.ParseWithFuncs
	err := env.ParseWithFuncs(config, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(HTTPConfig{}): func(v string) (interface{}, error) {
			cfg := HTTPConfig{}
			err := cfg.Set(v)
			if err != nil {
				return nil, err
			}
			return cfg, nil
		},
	})
	assert.NoError(t, err, "Env parsing should succeed")
	assert.Equal(t, HTTPConfig{Host: "env", Port: 8080}, config.HTTP, "HTTP should be overridden")
	assert.Equal(t, 900, config.StoreInterval, "StoreInterval should be overridden")
	assert.Equal(t, "/env/metrics.txt", config.FileStoragePath, "FileStoragePath should be overridden")
	assert.Equal(t, true, config.Restore, "Restore should be overridden")
	assert.Equal(t, "postgres://env:env@localhost:5432/envdb", config.DatabaseDSN, "DatabaseDSN should be overridden")
	assert.Equal(t, "envkey", config.Key, "Key should be overridden")
}
