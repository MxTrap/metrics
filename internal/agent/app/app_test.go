package app

import (
	"context"
	"github.com/MxTrap/metrics/config"
	"sync"
	"testing"
	"time"

	"github.com/MxTrap/metrics/config/agentconfig"
	"github.com/MxTrap/metrics/internal/agent/httpclient"
	"github.com/MxTrap/metrics/internal/agent/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRunner struct {
	mock.Mock
}

func (m *mockRunner) Run(ctx context.Context) {
	m.Called(ctx)
}

func TestNewApp(t *testing.T) {
	// Конфигурация без шифрования
	cfg := &agentconfig.AgentConfig{
		ServerConfig:   config.HTTPConfig{Host: "localhost", Port: 8080},
		ReportInterval: 10,
		PollInterval:   2,
		Key:            "test_key",
		RateLimit:      1,
		CryptoKey:      "",
	}

	// Создаём App
	app := NewApp(cfg)
	assert.NotNil(t, app, "app should not be nil")
	assert.NotNil(t, app.service, "service should not be nil")
	assert.NotNil(t, app.client, "client should not be nil")

	// Проверяем типы
	_, ok := app.service.(*service.MetricsObserverService)
	assert.True(t, ok, "service should be MetricsObserverService")
	_, ok = app.client.(*httpclient.HTTPClient)
	assert.True(t, ok, "client should be HTTPClient")
}

func TestNewAppWithEncryption(t *testing.T) {
	// Конфигурация с шифрованием
	cfg := &agentconfig.AgentConfig{
		ServerConfig:   config.HTTPConfig{Host: "localhost", Port: 8080},
		ReportInterval: 10,
		PollInterval:   2,
		Key:            "test_key",
		RateLimit:      1,
		CryptoKey:      "",
	}

	// Создаём App
	app := NewApp(cfg)
	assert.NotNil(t, app, "app should not be nil")
	assert.NotNil(t, app.service, "service should not be nil")
	assert.NotNil(t, app.client, "client should not be nil")
}

func TestRun(t *testing.T) {
	// Мокаем service и client
	serviceRunner := &mockRunner{}
	clientRunner := &mockRunner{}

	// Создаём каналы для отслеживания вызовов
	serviceStarted := make(chan struct{})
	clientStarted := make(chan struct{})
	serviceRunner.On("Run", mock.Anything).Run(func(args mock.Arguments) {
		close(serviceStarted)
	}).Return()
	clientRunner.On("Run", mock.Anything).Run(func(args mock.Arguments) {
		close(clientStarted)
	}).Return()

	app := &App{
		service: serviceRunner,
		client:  clientRunner,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := app.Run(ctx)
		assert.NoError(t, err, "Run should succeed")
	}()

	select {
	case <-serviceStarted:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("service did not start")
	}
	select {
	case <-clientStarted:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("client did not start")
	}

	cancel()
	wg.Wait()

	serviceRunner.AssertExpectations(t)
	clientRunner.AssertExpectations(t)
}
