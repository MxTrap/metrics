// Package repository предоставляет хранилище для метрик в памяти.
// Реализует MemStorage для сохранения, получения и управления метриками.
package repository

import (
	"context"
	"errors"
	"github.com/MxTrap/metrics/internal/common/models"
	"maps"
)

type MemStorage struct {
	metrics map[string]models.Metric
}

// NewMemStorage создаёт новое хранилище метрик в памяти.
// Возвращает указатель на инициализированный MemStorage или ошибку.
func NewMemStorage() (*MemStorage, error) {
	return &MemStorage{
		metrics: map[string]models.Metric{},
	}, nil
}

func (s *MemStorage) Ping(_ context.Context) error {
	return errors.New("not implemented")
}

// Save сохраняет метрику в хранилище.
// Для метрик типа Counter агрегирует значение Delta с существующей метрикой.
// Возвращает ошибку при неудаче.
func (s *MemStorage) Save(_ context.Context, metric models.Metric) error {
	if val, ok := s.metrics[metric.ID]; ok && metric.MType == models.Counter {
		*(metric.Delta) = *(metric.Delta) + *(val.Delta)
	}
	s.metrics[metric.ID] = metric
	return nil
}

// Find получает метрику по её идентификатору.
// Возвращает метрику или ошибку, если метрика не найдена.
func (s *MemStorage) Find(_ context.Context, metric string) (models.Metric, error) {
	value, ok := s.metrics[metric]
	if !ok {
		return models.Metric{}, errors.New("not found")
	}
	return value, nil
}

// GetAll возвращает все метрики из хранилища.
// Возвращает карту метрик или ошибку при неудаче.
func (s *MemStorage) GetAll(_ context.Context) (map[string]models.Metric, error) {
	return s.metrics, nil
}

// SaveAll сохраняет набор метрик в хранилище.
// Копирует переданные метрики в хранилище, перезаписывая существующие.
// Возвращает ошибку при неудаче.
func (s *MemStorage) SaveAll(_ context.Context, metrics map[string]models.Metric) error {
	maps.Copy(s.metrics, metrics)
	return nil
}
