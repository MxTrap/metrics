package config

import (
	"github.com/caarlos0/env"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAgentConfig(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		envVars     map[string]string
		expected    *AgentConfig
		expectError bool
	}{
		{
			name: "Environment variables override",
			args: []string{"-a", "server:9090"},
			envVars: map[string]string{
				"ADDRESS":         "env:8080",
				"REPORT_INTERVAL": "15",
				"POLL_INTERVAL":   "3",
				"KEY":             "envkey",
				"RATE_LIMIT":      "5",
			},
			expected: &AgentConfig{
				ServerConfig: HTTPConfig{
					Host: "env",
					Port: 8080,
				},
				ReportInterval: 15,
				PollInterval:   3,
				Key:            "envkey",
				RateLimit:      5,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				err := os.Setenv(k, v)
				assert.NoError(t, err, "Failed to set env var")
				defer os.Unsetenv(k)
			}

			config, err := NewAgentConfig()

			if tt.expectError {
				assert.Error(t, err, "Expected error")
				assert.Nil(t, config, "Config should be nil on error")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.NotNil(t, config, "Config should not be nil")
				assert.Equal(t, tt.expected.ServerConfig, config.ServerConfig, "ServerConfig should match")
				assert.Equal(t, tt.expected.ReportInterval, config.ReportInterval, "ReportInterval should match")
				assert.Equal(t, tt.expected.PollInterval, config.PollInterval, "PollInterval should match")
				assert.Equal(t, tt.expected.Key, config.Key, "Key should match")
				assert.Equal(t, tt.expected.RateLimit, config.RateLimit, "RateLimit should match")
			}
		})
	}
}

func TestNewAgentConfig_EnvParsing(t *testing.T) {
	config := &AgentConfig{
		ServerConfig: HTTPConfig{
			Host: "localhost",
			Port: 8080,
		},
		ReportInterval: 10,
		PollInterval:   2,
		Key:            "",
		RateLimit:      1,
	}

	envVars := map[string]string{
		"ADDRESS":         "env:8080",
		"REPORT_INTERVAL": "15",
		"POLL_INTERVAL":   "3",
		"KEY":             "envkey",
		"RATE_LIMIT":      "5",
	}

	// Устанавливаем переменные окружения
	for k, v := range envVars {
		err := os.Setenv(k, v)
		assert.NoError(t, err, "Failed to set env var")
		defer os.Unsetenv(k)
	}

	// Вызываем env.ParseWithFuncs с реальным парсером
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
	assert.Equal(t, HTTPConfig{Host: "env", Port: 8080}, config.ServerConfig, "ServerConfig should be overridden")
	assert.Equal(t, 15, config.ReportInterval, "ReportInterval should be overridden")
	assert.Equal(t, 3, config.PollInterval, "PollInterval should be overridden")
	assert.Equal(t, "envkey", config.Key, "Key should be overridden")
	assert.Equal(t, 5, config.RateLimit, "RateLimit should be overridden")
}
