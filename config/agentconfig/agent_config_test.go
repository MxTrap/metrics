package agentconfig

import (
	"flag"
	"github.com/MxTrap/metrics/config"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func beforeEach() {
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
}

func TestNewAgentConfig(t *testing.T) {
	beforeEach()
	tempDir, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "agent_config.json")
	configContent := []byte(`{
		"address": "127.0.0.1:9090",
		"report_interval": "15s",
		"poll_interval": "3s",
		"crypto_key": "/tmp/keys/private.pem"
	}`)
	err = os.WriteFile(configFile, configContent, 0644)
	require.NoError(t, err, "failed to write config file")

	// Устанавливаем переменные окружения
	os.Setenv("CONFIG", configFile)
	os.Setenv("ADDRESS", "192.168.1.1:8081")
	os.Setenv("REPORT_INTERVAL", "20")
	os.Setenv("POLL_INTERVAL", "4")
	os.Setenv("KEY", "env_key")
	os.Setenv("RATE_LIMIT", "2")
	os.Setenv("CRYPTO_KEY", "/env/path/private.pem")
	defer func() {
		os.Unsetenv("CONFIG")
		os.Unsetenv("ADDRESS")
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("POLL_INTERVAL")
		os.Unsetenv("KEY")
		os.Unsetenv("RATE_LIMIT")
		os.Unsetenv("CRYPTO_KEY")
	}()

	os.Args = []string{
		"test",
		"-a", "localhost:8082",
		"-r", "25",
		"-p", "5",
		"-k", "flag_key",
		"-l", "3",
		"-crypto-key", "/flag/path/private.pem",
	}

	cfg, err := NewAgentConfig()
	require.NoError(t, err, "NewAgentConfig should succeed")

	assert.Equal(t, config.AddrConfig{Host: "192.168.1.1", Port: 8081}, cfg.HTTPServerAddr, "HTTPServerAddr should match env")
	assert.Equal(t, 20, cfg.ReportInterval, "ReportInterval should match env")
	assert.Equal(t, 4, cfg.PollInterval, "PollInterval should match env")
	assert.Equal(t, "env_key", cfg.Key, "Key should match env")
	assert.Equal(t, 2, cfg.RateLimit, "RateLimit should match env")
	assert.Equal(t, "/env/path/private.pem", cfg.CryptoKey, "CryptoKey should match env")
}

func TestParseFromFile(t *testing.T) {
	beforeEach()
	tempDir, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "agent_config.json")
	configContent := []byte(`{
		"address": "127.0.0.1:9090",
		"report_interval": "15s",
		"poll_interval": "3s",
		"crypto_key": "/tmp/keys/private.pem"
	}`)
	err = os.WriteFile(configFile, configContent, 0644)
	require.NoError(t, err, "failed to write config file")

	os.Setenv("CONFIG", configFile)
	defer os.Unsetenv("CONFIG")

	cfg := &AgentConfig{}
	err = cfg.parseFromFile()
	require.NoError(t, err, "parseFromFile should succeed")

	assert.Equal(t, config.AddrConfig{Host: "127.0.0.1", Port: 9090}, cfg.HTTPServerAddr, "HTTPServerAddr should match file")
	assert.Equal(t, 15, cfg.ReportInterval, "ReportInterval should match file")
	assert.Equal(t, 3, cfg.PollInterval, "PollInterval should match file")
	assert.Equal(t, "/tmp/keys/private.pem", cfg.CryptoKey, "CryptoKey should match file")
	assert.Empty(t, cfg.Key, "Key should be empty")
	assert.Equal(t, 0, cfg.RateLimit, "RateLimit should be zero")
}

func TestParseFromFileInvalidPath(t *testing.T) {
	beforeEach()

	cfg := &AgentConfig{}
	os.Setenv("CONFIG", "/invalid/path")
	defer os.Unsetenv("CONFIG")

	err := cfg.parseFromFile()
	assert.Error(t, err, "parseFromFile should fail with invalid path")
}

func TestParseFromFileInvalidJSON(t *testing.T) {
	beforeEach()

	tempDir, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "agent_config.json")
	configContent := []byte(`{invalid json}`)
	err = os.WriteFile(configFile, configContent, 0644)
	require.NoError(t, err, "failed to write config file")

	os.Setenv("CONFIG", configFile)
	defer os.Unsetenv("CONFIG")

	cfg := &AgentConfig{}
	err = cfg.parseFromFile()
	assert.Error(t, err, "parseFromFile should fail with invalid JSON")
}

func TestParseFromFileInvalidDuration(t *testing.T) {
	beforeEach()

	tempDir, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "agent_config.json")
	configContent := []byte(`{
		"report_interval": "invalid_duration"
	}`)
	err = os.WriteFile(configFile, configContent, 0644)
	require.NoError(t, err, "failed to write config file")

	os.Setenv("CONFIG", configFile)
	defer os.Unsetenv("CONFIG")

	cfg := &AgentConfig{}
	err = cfg.parseFromFile()
	assert.Error(t, err, "parseFromFile should fail with invalid duration")
}

func TestParseFromFlags(t *testing.T) {
	beforeEach()

	os.Args = []string{
		"test",
		"-a", "localhost:8082",
		"-r", "25",
		"-p", "5",
		"-k", "flag_key",
		"-l", "3",
		"-crypto-key", "/flag/path/private.pem",
	}

	cfg := &AgentConfig{}
	cfg.parseFromFlags()

	assert.Equal(t, config.AddrConfig{Host: "localhost", Port: 8082}, cfg.HTTPServerAddr, "HTTPServerAddr should match flags")
	assert.Equal(t, 25, cfg.ReportInterval, "ReportInterval should match flags")
	assert.Equal(t, 5, cfg.PollInterval, "PollInterval should match flags")
	assert.Equal(t, "flag_key", cfg.Key, "Key should match flags")
	assert.Equal(t, 3, cfg.RateLimit, "RateLimit should match flags")
	assert.Equal(t, "/flag/path/private.pem", cfg.CryptoKey, "CryptoKey should match flags")
}

func TestParseFromEnv(t *testing.T) {
	beforeEach()

	os.Setenv("ADDRESS", "192.168.1.1:8081")
	os.Setenv("REPORT_INTERVAL", "20")
	os.Setenv("POLL_INTERVAL", "4")
	os.Setenv("KEY", "env_key")
	os.Setenv("RATE_LIMIT", "2")
	os.Setenv("CRYPTO_KEY", "/env/path/private.pem")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("POLL_INTERVAL")
		os.Unsetenv("KEY")
		os.Unsetenv("RATE_LIMIT")
		os.Unsetenv("CRYPTO_KEY")
	}()

	cfg := &AgentConfig{}
	err := cfg.parseFromEnv()
	require.NoError(t, err, "parseFromEnv should succeed")

	assert.Equal(t, config.AddrConfig{Host: "192.168.1.1", Port: 8081}, cfg.HTTPServerAddr, "HTTPServerAddr should match env")
	assert.Equal(t, 20, cfg.ReportInterval, "ReportInterval should match env")
	assert.Equal(t, 4, cfg.PollInterval, "PollInterval should match env")
	assert.Equal(t, "env_key", cfg.Key, "Key should match env")
	assert.Equal(t, 2, cfg.RateLimit, "RateLimit should match env")
	assert.Equal(t, "/env/path/private.pem", cfg.CryptoKey, "CryptoKey should match env")
}

func TestParseFromFileEmptyPath(t *testing.T) {
	beforeEach()
	cfg := &AgentConfig{}
	err := cfg.parseFromFile()
	assert.NoError(t, err, "parseFromFile should succeed with empty path")
}
