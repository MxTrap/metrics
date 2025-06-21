package serverconfig

import (
	"flag"
	"github.com/MxTrap/metrics/config"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func beforeEach() {
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
}

func TestNewServerConfig(t *testing.T) {
	beforeEach()
	tempDir, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "config.json")
	configContent := []byte(
		`{
		"address": "127.0.1:9090",
		"restore": true,
		"store_interval": "600s",
		"store_file": "/tmp/metrics.txt",
		"database_dsn": "postgres://user:pass@localhost:5432/db",
		"crypto_key": "/tmp/keys/private.pem"
	}`,
	)
	err = os.WriteFile(configFile, configContent, 0644)
	require.NoError(t, err, "failed to write config file")

	os.Setenv("CONFIG", configFile)
	os.Setenv("ADDRESS", "192.168.1.1:8081")
	os.Setenv("STORE_INTERVAL", "900")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/env_metrics.txt")
	os.Setenv("RESTORE", "false")
	os.Setenv("DATABASE_DSN", "postgres://env_user:pass@localhost:5432/env_db")
	os.Setenv("KEY", "env_key")
	os.Setenv("CRYPTO_KEY", "/env/path/private.pem")
	defer func() {
		os.Unsetenv("CONFIG")
		os.Unsetenv("ADDRESS")
		os.Unsetenv("STORE_INTERVAL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("RESTORE")
		os.Unsetenv("DATABASE_DSN")
		os.Unsetenv("KEY")
		os.Unsetenv("CRYPTO_KEY")
	}()

	os.Args = []string{
		"test",
		"-a", "localhost:8082",
		"-i", "1200",
		"-f", "/tmp/flag_metrics.txt",
		"-r",
		"-d", "postgres://flag_user:pass@localhost:5432/flag_db",
		"-k", "flag_key",
		"-c", "/flag/path/private.pem",
	}

	cfg, err := NewServerConfig()
	require.NoError(t, err, "NewServerConfig should succeed")

	// Проверяем, что значения из переменных окружения имеют приоритет
	assert.Equal(t, config.HTTPConfig{Host: "192.168.1.1", Port: 8081}, cfg.HTTP, "HTTP should match env")
	assert.Equal(t, 900, cfg.StoreInterval, "StoreInterval should match env")
	assert.Equal(t, "/tmp/env_metrics.txt", cfg.FileStoragePath, "FileStoragePath should match env")
	assert.False(t, cfg.Restore, "Restore should match env")
	assert.Equal(t, "postgres://env_user:pass@localhost:5432/env_db", cfg.DatabaseDSN, "DatabaseDSN should match env")
	assert.Equal(t, "env_key", cfg.Key, "Key should match env")
	assert.Equal(t, "/env/path/private.pem", cfg.CryptoKey, "CryptoKey should match env")
}

func TestParseFromFile(t *testing.T) {
	beforeEach()
	tempDir, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err, "failed to create temp dir")

	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "config.json")
	configContent := []byte(
		`{
			"address": "127.0.0.1:9090",
			"restore": true,
			"store_interval": "600s",
			"store_file": "/tmp/metrics.txt",
			"database_dsn": "postgres://user:pass@localhost:5432/db",
			"crypto_key": "/tmp/keys/private.pem"
		}
		`,
	)

	err = os.WriteFile(configFile, configContent, 0644)
	require.NoError(t, err, "failed to write config file")

	// Устанавливаем переменную окружения CONFIG
	os.Setenv("CONFIG", configFile)
	defer os.Unsetenv("CONFIG")

	cfg := &ServerConfig{}
	err = cfg.parseFromFile()
	require.NoError(t, err, "parseFromFile should succeed")

	assert.Equal(t, config.HTTPConfig{Host: "127.0.0.1", Port: 9090}, cfg.HTTP, "HTTP should match file")
	assert.Equal(t, 600, cfg.StoreInterval, "StoreInterval should match file")
	assert.Equal(t, "/tmp/metrics.txt", cfg.FileStoragePath, "FileStoragePath should match file")
	assert.True(t, cfg.Restore, "Restore should match file")
	assert.Equal(t, "postgres://user:pass@localhost:5432/db", cfg.DatabaseDSN, "DatabaseDSN should match file")
	assert.Equal(t, "/tmp/keys/private.pem", cfg.CryptoKey, "CryptoKey should match file")
}

func TestParseFromFileInvalidPath(t *testing.T) {
	beforeEach()
	cfg := &ServerConfig{}
	os.Setenv("CONFIG", "/invalid/path")
	defer os.Unsetenv("CONFIG")

	err := cfg.parseFromFile()
	assert.Error(t, err, "parseFromFile should fail with invalid path")
}

func TestParseFromFlags(t *testing.T) {
	beforeEach()
	os.Args = []string{"test", "-a", "localhost:8082", "-i", "1200", "-f", "/tmp/flag_metrics.txt", "-r", "-d", "postgres://flag_user:pass@localhost:5432/flag_db", "-k", "flag_key", "-crypto-key", "/flag/path/private.pem"}

	cfg := &ServerConfig{}
	cfg.parseFromFlags()

	assert.Equal(t, config.HTTPConfig{Host: "localhost", Port: 8082}, cfg.HTTP, "HTTP should match flags")
	assert.Equal(t, 1200, cfg.StoreInterval, "StoreInterval should match flags")
	assert.Equal(t, "/tmp/flag_metrics.txt", cfg.FileStoragePath, "FileStoragePath should match flags")
	assert.True(t, cfg.Restore, "Restore should match flags")
	assert.Equal(t, "postgres://flag_user:pass@localhost:5432/flag_db", cfg.DatabaseDSN, "DatabaseDSN should match flags")
	assert.Equal(t, "flag_key", cfg.Key, "Key should match flags")
	assert.Equal(t, "/flag/path/private.pem", cfg.CryptoKey, "CryptoKey should match flags")
}

func TestParseFromEnv(t *testing.T) {
	beforeEach()
	// Устанавливаем переменные окружения
	os.Setenv("ADDRESS", "192.168.1.1:8081")
	os.Setenv("STORE_INTERVAL", "900")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/env_metrics.txt")
	os.Setenv("RESTORE", "false")
	os.Setenv("DATABASE_DSN", "postgres://env_user:pass@localhost:5432/env_db")
	os.Setenv("KEY", "env_key")
	os.Setenv("CRYPTO_KEY", "/env/path/private.pem")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("STORE_INTERVAL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("RESTORE")
		os.Unsetenv("DATABASE_DSN")
		os.Unsetenv("KEY")
		os.Unsetenv("CRYPTO_KEY")
	}()

	cfg := &ServerConfig{}
	err := cfg.parseFromEnv()
	require.NoError(t, err, "parseFromEnv should succeed")

	assert.Equal(t, config.HTTPConfig{Host: "192.168.1.1", Port: 8081}, cfg.HTTP, "HTTP should match env")
	assert.Equal(t, 900, cfg.StoreInterval, "StoreInterval should match env")
	assert.Equal(t, "/tmp/env_metrics.txt", cfg.FileStoragePath, "FileStoragePath should match env")
	assert.False(t, cfg.Restore, "Restore should match env")
	assert.Equal(t, "postgres://env_user:pass@localhost:5432/env_db", cfg.DatabaseDSN, "DatabaseDSN should match env")
	assert.Equal(t, "env_key", cfg.Key, "Key should match env")
	assert.Equal(t, "/env/path/private.pem", cfg.CryptoKey, "CryptoKey should match env")
}

func TestParseFromFileEmptyPath(t *testing.T) {
	beforeEach()
	cfg := &ServerConfig{}
	err := cfg.parseFromFile()
	assert.NoError(t, err, "parseFromFile should succeed with empty path")
}
