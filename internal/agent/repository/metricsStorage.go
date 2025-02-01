package repository

import (
	"github.com/MxTrap/metrics/internal/agent/models"
	"maps"
	"math/rand/v2"
)

type MetricsStorage struct {
	storage models.Metrics
}

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		storage: models.Metrics{
			Gauge:   models.GaugeMetrics{},
			Counter: models.CounterMetrics{},
		},
	}
}

func (s *MetricsStorage) SaveMetrics(m models.GaugeMetrics) {
	maps.Copy(s.storage.Gauge, m)
	s.storage.Counter.PollCount += 1
	s.storage.Counter.RandomValue = int64(rand.Int())
}

func (s *MetricsStorage) GetMetrics() models.Metrics {
	return s.storage
}
