package grpc

import (
	"github.com/MxTrap/metrics/internal/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type mockMetricsGetter struct {
	mock.Mock
}

func (m *mockMetricsGetter) GetMetrics() models.Metrics {
	args := m.Called()
	return args.Get(0).(models.Metrics)
}

func TestNewClient(t *testing.T) {
	service := &mockMetricsGetter{}
	client, err := NewClient("test-addr", service, 5, "secret", 2)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.Equal(t, service, client.service)
	assert.Equal(t, 5, client.reportInterval)
	assert.Equal(t, "secret", client.key)
	assert.Equal(t, 2, client.rateLimit)
}
