package repository

import (
	"github.com/MxTrap/metrics/internal/agent/models"
	"math/rand"
)

type MetricsStorage struct {
	storage models.Metrics
}

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		storage: models.Metrics{
			Gauge:   *models.NewGaugeMetrics(),
			Counter: models.CounterMetrics{},
		},
	}
}

func (s *MetricsStorage) SaveMetrics(m map[string]float64) {
	s.storage.Gauge.Load(m)
	s.storage.Gauge.Set("RandomValue", rand.Float64())
	s.storage.Counter.PollCount += 1
}

func (s *MetricsStorage) GetMetrics() models.Metrics {
	return s.storage
}
