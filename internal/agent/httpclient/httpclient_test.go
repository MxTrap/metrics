package httpclient

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/MxTrap/metrics/internal/agent/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestHTTPClient_compress(t *testing.T) {
	client := &HTTPClient{}
	data := []byte(`{"test":"data"}`)

	compressed, err := client.compress(data)
	assert.NoError(t, err, "Compress should not return error")
	assert.NotNil(t, compressed, "Compressed buffer should not be nil")

	// Проверяем, что данные сжаты
	reader, err := gzip.NewReader(compressed)
	assert.NoError(t, err, "Failed to create gzip reader")
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	assert.NoError(t, err, "Failed to decompress data")
	assert.Equal(t, data, decompressed, "Decompressed data should match original")

	client.compress([]byte{}) // Пустые данные
	compressed, err = client.compress([]byte{})
	assert.NoError(t, err, "Compress should handle empty data")
	reader, err = gzip.NewReader(compressed)
	assert.NoError(t, err, "Gzip reader should handle empty compressed data")
	decompressed, err = io.ReadAll(reader)
	assert.NoError(t, err, "Decompress should succeed for empty data")
	assert.Empty(t, decompressed, "Decompressed empty data should be empty")
}

type MockMetricsObserverService struct {
	mock.Mock
}

func (m *MockMetricsObserverService) GetMetrics() models.Metrics {
	args := m.Called()
	return args.Get(0).(models.Metrics)
}

// Бенчмарк для httpclient.compress
func BenchmarkHTTPClientCompress(b *testing.B) {
	client := &HTTPClient{}
	sizes := []int{100, 1000, 10000, 100000} // Размеры данных в байтах

	for _, size := range sizes {
		data := bytes.Repeat([]byte("a"), size)
		b.Run("Size"+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := client.compress(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Бенчмарк для httpclient.sendMetrics
func BenchmarkHTTPClientSendMetrics(b *testing.B) {
	service := &MockMetricsObserverService{}
	client := NewHTTPClient(service, "localhost:8080", 10, "", 1)

	metricCounts := []int{10, 100, 1000}
	for _, count := range metricCounts {
		metrics := models.Metrics{
			Gauge: *models.NewGaugeMetrics(),
		}
		for i := 0; i < count; i++ {
			metrics.Gauge.Set("gauge"+strconv.Itoa(i), float64(i))
		}
		metrics.Counter.PollCount = 100
		service.On("GetMetrics").Return(metrics).Once()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		client.client = server.Client()

		b.Run("Metrics"+strconv.Itoa(count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := client.sendMetrics(context.Background())
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
