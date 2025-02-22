package repository

import (
	"github.com/MxTrap/metrics/internal/common/models"
)

type MemStorage struct {
	metrics map[string]models.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: map[string]models.Metrics{},
	}
}

func (s *MemStorage) Save(metric models.Metrics) {
	if val, ok := s.metrics[metric.ID]; ok && metric.MType == models.Counter {
		*(metric.Delta) = *(metric.Delta) + *(val.Delta)
	}
	s.metrics[metric.ID] = metric
}

func (s *MemStorage) Find(metric string) (models.Metrics, bool) {
	value, ok := s.metrics[metric]
	return value, ok
}

func (s *MemStorage) GetAll() map[string]models.Metrics {
	return s.metrics
}

func (s *MemStorage) SaveAll(metrics map[string]models.Metrics) {
	s.metrics = metrics
}
