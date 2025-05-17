package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()
	expected := HTTPConfig{
		Host: "localhost",
		Port: 8080,
	}
	assert.Equal(t, expected, config, "NewDefaultConfig should return default HTTPConfig")
}

func TestHTTPConfig_String(t *testing.T) {
	tests := []struct {
		name     string
		config   HTTPConfig
		expected string
	}{
		{
			name:     "Default config",
			config:   HTTPConfig{Host: "localhost", Port: 8080},
			expected: "localhost:8080",
		},
		{
			name:     "Custom host and port",
			config:   HTTPConfig{Host: "server", Port: 9090},
			expected: "server:9090",
		},
		{
			name:     "Empty host",
			config:   HTTPConfig{Host: "", Port: 80},
			expected: ":80",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.String()
			assert.Equal(t, tt.expected, result, "String should return correct format")
		})
	}
}

func TestHTTPConfig_Set(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    HTTPConfig
		expectError bool
	}{
		{
			name:  "Valid address",
			input: "localhost:8080",
			expected: HTTPConfig{
				Host: "localhost",
				Port: 8080,
			},
			expectError: false,
		},
		{
			name:  "Custom host and port",
			input: "server:9090",
			expected: HTTPConfig{
				Host: "server",
				Port: 9090,
			},
			expectError: false,
		},
		{
			name:        "Invalid format - no port",
			input:       "localhost",
			expected:    HTTPConfig{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &HTTPConfig{}
			err := config.Set(tt.input)
			if tt.expectError {
				assert.Error(t, err, "Set should return error")
				assert.Equal(t, HTTPConfig{}, *config, "Config should remain unchanged on error")
			} else {
				assert.NoError(t, err, "Set should not return error")
				assert.Equal(t, tt.expected, *config, "Set should correctly parse input")
			}
		})
	}
}
