// Package repository предоставляет хранилище для метрик агента.
// Реализует MetricsStorage для сохранения и получения метрик, включая счетчики и значения типа gauge.
package repository

import (
	"github.com/MxTrap/metrics/internal/agent/models"
	"math/rand"
)

type MetricsStorage struct {
	storage models.Metrics
}

// NewMetricsStorage создаёт новое хранилище метрик с инициализированными структурами gauge и счетчиков.
// Возвращает указатель на инициализированный MetricsStorage.
func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		storage: models.Metrics{
			Gauge:   *models.NewGaugeMetrics(),
			Counter: models.CounterMetrics{},
		},
	}
}

// SaveMetrics сохраняет метрики в хранилище.
// Принимает карту значений gauge, добавляет случайное значение RandomValue и инкрементирует счетчик PollCount.
func (s *MetricsStorage) SaveMetrics(m map[string]float64) {
	s.storage.Gauge.Load(m)
	s.storage.Gauge.Set("RandomValue", rand.Float64())
	s.storage.Counter.PollCount += 1
}

// GetMetrics возвращает все метрики из хранилища.
// Возвращает структуру models.Metrics, содержащую значения gauge и счетчики.
func (s *MetricsStorage) GetMetrics() models.Metrics {
	return s.storage
}
