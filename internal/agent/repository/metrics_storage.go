package repository

import (
	"github.com/MxTrap/metrics/internal/agent/models"
	"maps"
	"math/rand"
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
	s.storage.Gauge["RandomValue"] = rand.Float64()
	s.storage.Counter.PollCount += 1
}

func (s *MetricsStorage) GetMetrics() models.Metrics {
	return s.storage
}
