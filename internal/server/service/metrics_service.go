package service

import (
	common_models "github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/server/models"
)

type Storage interface {
	Save(metrics common_models.Metrics) error
	Find(metric string) (common_models.Metrics, bool)
	GetAll() map[string]any
}

type FileStorage interface {
	Save(metrics map[string]any) error
	Read() (map[string]any, error)
}

type MetricsService struct {
	storage     Storage
	fileStorage FileStorage
}

func NewMetricsService(storage Storage, fileStorage FileStorage) *MetricsService {

	return &MetricsService{
		storage:     storage,
		fileStorage: fileStorage,
	}
}

func (MetricsService) validateMetric(metricType string) bool {
	_, ok := models.MetricTypes[metricType]
	return ok
}

func (s *MetricsService) Save(metric common_models.Metrics) error {
	if !s.validateMetric(metric.MType) {
		return models.ErrUnknownMetricType
	}

	if metric.Delta == nil && metric.Value == nil {
		return models.ErrWrongMetricValue
	}

	err := s.storage.Save(metric)
	if err != nil {
		return err
	}

	return nil
}

func (s *MetricsService) Find(metric common_models.Metrics) (common_models.Metrics, error) {
	if !s.validateMetric(metric.MType) {
		return common_models.Metrics{}, models.ErrUnknownMetricType
	}
	val, ok := s.storage.Find(metric.ID)
	if !ok {
		return common_models.Metrics{}, models.ErrNotFoundMetric
	}

	return val, nil
}

func (s *MetricsService) GetAll() map[string]any {
	return s.storage.GetAll()
}

func (s *MetricsService) saveToFile() error {
	s.storage.GetAll()
	return nil
}
