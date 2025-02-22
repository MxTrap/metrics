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

func (s *MemStorage) Save(metric models.Metrics) error {
	if val, ok := s.metrics[metric.ID]; ok && metric.MType == models.Counter {
		*(val.Delta) = *(val.Delta) + *(metric.Delta)
	}
	s.metrics[metric.ID] = metric
	return nil
}

func (s *MemStorage) Find(metric string) (models.Metrics, bool) {
	value, ok := s.metrics[metric]
	return value, ok
}

func (s *MemStorage) GetAll() map[string]any {
	dst := map[string]any{}
	for k, v := range s.metrics {
		var val any
		if v.Delta != nil {
			val = *v.Delta
		}
		if v.Value != nil {
			val = *v.Value
		}
		dst[k] = val
	}
	return dst
}
