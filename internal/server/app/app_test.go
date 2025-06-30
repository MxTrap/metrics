package app

import (
	"context"
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/config/serverconfig"
	"github.com/MxTrap/metrics/internal/server/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

// mockMetricsService мокает service.MetricsService.
type mockMetricsService struct {
	mock.Mock
}

func (m *mockMetricsService) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockMetricsService) Stop() {
	m.Called()
}

// mockHTTPServer мокает httpserver.HTTPServer.
type mockHTTPServer struct {
	mock.Mock
}

func (m *mockHTTPServer) Run() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockHTTPServer) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// mockPgxPool мокает pgxpool.Pool.
type mockPgxPool struct {
	mock.Mock
}

// mockLogger мокает logger.Logger.
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Logger() *logger.Logger {
	args := m.Called()
	return args.Get(0).(*logger.Logger)
}

func (m *mockLogger) Info(msg string) {
	m.Called(msg)
}

func (m *mockLogger) Error(msg string) {
	m.Called(msg)
}

func TestNewApp(t *testing.T) {

	// Конфигурация без PostgreSQL
	cfg := &serverconfig.ServerConfig{
		HTTPAddr:        config.AddrConfig{Host: "localhost", Port: 8080},
		FileStoragePath: "",
		StoreInterval:   300,
		Restore:         true,
		Key:             "test_key",
		CryptoKey:       "",
	}

	// Создаём App
	app, err := NewApp(cfg, context.Background())
	require.NoError(t, err, "NewApp should succeed")
	assert.NotNil(t, app, "app should not be nil")
	assert.NotNil(t, app.httpServer, "httpServer should not be nil")
	assert.NotNil(t, app.metricsService, "metricsService should not be nil")
	assert.NotNil(t, app.logger, "logger should not be nil")
}
