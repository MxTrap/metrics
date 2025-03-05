package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/MxTrap/metrics/internal/common/models"
)

type MemStorage struct {
	metrics map[string]models.Metric
}

func NewMemStorage() (*MemStorage, error) {
	return &MemStorage{
		metrics: map[string]models.Metric{},
	}, nil
}

func (s *MemStorage) Ping(_ context.Context) error {
	return errors.New("not implemented")
}

func (s *MemStorage) Save(_ context.Context, metric models.Metric) error {
	if val, ok := s.metrics[metric.ID]; ok && metric.MType == models.Counter {
		*(metric.Delta) = *(metric.Delta) + *(val.Delta)
	}
	s.metrics[metric.ID] = metric
	return nil
}

func (s *MemStorage) Find(_ context.Context, metric string) (models.Metric, error) {
	fmt.Println("find ", s.metrics)
	value, ok := s.metrics[metric]
	if !ok {
		return models.Metric{}, errors.New("not found")
	}
	return value, nil
}

func (s *MemStorage) GetAll(_ context.Context) (map[string]models.Metric, error) {
	return s.metrics, nil
}

func (s *MemStorage) SaveAll(_ context.Context, metrics map[string]models.Metric) error {
	s.metrics = metrics
	fmt.Println("save all", s.metrics)
	return nil
}
