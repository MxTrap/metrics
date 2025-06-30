package http

import (
	"compress/gzip"
	"context"
	commonmodels "github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockMetricsObserver struct {
	mock.Mock
}

func (m *mockMetricsObserver) GetMetrics() commonmodels.Metrics {
	args := m.Called()
	return args.Get(0).(commonmodels.Metrics)
}

type mockEncrypter struct {
	mock.Mock
}

func (m *mockEncrypter) Encrypt(plaintext []byte) ([]byte, error) {
	args := m.Called(plaintext)
	return args.Get(0).([]byte), args.Error(1)
}

func TestNewHTTPClient(t *testing.T) {
	observer := &mockMetricsObserver{}
	client := NewClient(observer, "localhost:8080", 2, "testkey", 1)

	assert.NotNil(t, client, "client should not be nil")
	assert.NotNil(t, client.client, "http client should not be nil")
	assert.Equal(t, observer, client.service, "service should match")
	assert.Equal(t, "localhost:8080", client.serverURL, "serverURL should match")
	assert.Equal(t, 2, client.reportInterval, "reportInterval should match")
	assert.Equal(t, "testkey", client.key, "key should match")
	assert.Equal(t, 1, client.rateLimit, "rateLimit should match")
}

func TestCompress(t *testing.T) {
	client := &HTTPClient{}
	data := []byte(`{"test":"data"}`)

	compressed, err := client.compress(data)
	require.NoError(t, err, "compress should succeed")

	reader, err := gzip.NewReader(compressed)
	require.NoError(t, err, "failed to create gzip reader")
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err, "failed to decompress")
	assert.Equal(t, data, decompressed, "decompressed data should match original")
}

func TestRegisterEncrypter(t *testing.T) {
	getter := &mockMetricsObserver{}
	encrypter := &mockEncrypter{}
	client := NewClient(getter, "localhost:8080", 2, "secret", 3)
	client.RegisterEncrypter(encrypter)
	assert.Equal(t, encrypter, client.encrypter)
}

func TestPostMetric(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type should be application/json")
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"), "Content-Encoding should be gzip")

		reader, err := gzip.NewReader(r.Body)
		require.NoError(t, err, "failed to create gzip reader")
		defer reader.Close()
		require.NoError(t, err, "failed to read request body")

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	metrics := commonmodels.Metrics{
		{ID: "testGauge", MType: commonmodels.Gauge, Value: utils.MakePointer(42.0)},
	}

	key := "testkey"
	observer := &mockMetricsObserver{}
	observer.On("GetMetrics").Return(metrics)
	client := NewClient(observer, server.URL[7:], 2, key, 1)

	ctx := context.Background()
	err := client.postMetric(ctx)
	require.NoError(t, err, "postMetric should succeed")

}

func TestPostMetricWithRetries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer := &mockMetricsObserver{}
	observer.On("GetMetrics").Return(commonmodels.Metrics{})
	client := NewClient(observer, server.URL[7:], 2, "", 1)
	ctx := context.Background()

	err := client.postMetric(ctx)
	require.NoError(t, err, "postMetric should succeed after retries")
}

func TestRun(t *testing.T) {
	observer := &mockMetricsObserver{}
	metrics := commonmodels.Metrics{
		{ID: "PollCount", MType: commonmodels.Counter, Value: utils.MakePointer[float64](100)},
	}
	observer.On("GetMetrics").Return(metrics).Times(2) // Ожидаем минимум 2 вызова

	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(observer, server.URL[7:], 1, "", 1)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go client.Run(ctx)

	time.Sleep(2500 * time.Millisecond)

	assert.GreaterOrEqual(t, requestCount, 2, "should send at least 2 requests")
	observer.AssertExpectations(t)
}
