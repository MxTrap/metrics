package service

import (
	common_models "github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/server/models"
)

type MetricStorageService interface {
	Save(metrics common_models.Metrics)
	Find(metric string) (common_models.Metrics, bool)
	GetAll() map[string]any
}

type MetricsService struct {
	storageService MetricStorageService
}

func NewMetricsService(sService MetricStorageService) *MetricsService {

	return &MetricsService{
		storageService: sService,
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

	s.storageService.Save(metric)

	return nil
}

func (s *MetricsService) Find(metric common_models.Metrics) (common_models.Metrics, error) {
	if !s.validateMetric(metric.MType) {
		return common_models.Metrics{}, models.ErrUnknownMetricType
	}
	val, ok := s.storageService.Find(metric.ID)
	if !ok {
		return common_models.Metrics{}, models.ErrNotFoundMetric
	}

	return val, nil
}

func (s *MetricsService) GetAll() map[string]any {
	return s.storageService.GetAll()
}
