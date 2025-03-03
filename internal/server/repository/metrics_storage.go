package repository

import (
	"context"
	"errors"
	"github.com/MxTrap/metrics/internal/common/models"
)

type MemStorage struct {
	metrics map[string]models.Metrics
}

func NewMemStorage() (*MemStorage, error) {
	return &MemStorage{
		metrics: map[string]models.Metrics{},
	}, nil
}

func (s *MemStorage) Ping(_ context.Context) error {
	return errors.New("not implemented")
}

func (s *MemStorage) Save(_ context.Context, metric models.Metrics) error {
	if val, ok := s.metrics[metric.ID]; ok && metric.MType == models.Counter {
		*(metric.Delta) = *(metric.Delta) + *(val.Delta)
	}
	s.metrics[metric.ID] = metric
	return nil
}

func (s *MemStorage) Find(_ context.Context, metric string) (models.Metrics, error) {
	value, ok := s.metrics[metric]
	if !ok {
		return models.Metrics{}, errors.New("not found")
	}
	return value, nil
}

func (s *MemStorage) GetAll(_ context.Context) (map[string]models.Metrics, error) {
	return s.metrics, nil
}

func (s *MemStorage) SaveAll(_ context.Context, metrics map[string]models.Metrics) error {
	s.metrics = metrics
	return nil
}
