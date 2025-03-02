package service

import (
	"context"
	common_models "github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/server/models"
)

type MetricStorageService interface {
	Save(ctx context.Context, metrics common_models.Metrics) error
	Find(ctx context.Context, metric string) (common_models.Metrics, error)
	GetAll(ctx context.Context) (map[string]any, error)
	Ping(ctx context.Context) error
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

func (s *MetricsService) Save(ctx context.Context, metric common_models.Metrics) error {
	if !s.validateMetric(metric.MType) {
		return models.ErrUnknownMetricType
	}

	if metric.Delta == nil && metric.Value == nil {
		return models.ErrWrongMetricValue
	}

	err := s.storageService.Save(ctx, metric)
	if err != nil {
		return err
	}

	return nil
}

func (s *MetricsService) Find(ctx context.Context, metric common_models.Metrics) (common_models.Metrics, error) {
	if !s.validateMetric(metric.MType) {
		return common_models.Metrics{}, models.ErrUnknownMetricType
	}
	val, err := s.storageService.Find(ctx, metric.ID)
	if err != nil {
		return common_models.Metrics{}, models.ErrNotFoundMetric
	}

	return val, nil
}

func (s *MetricsService) GetAll(ctx context.Context) (map[string]any, error) {
	return s.storageService.GetAll(ctx)
}

func (s *MetricsService) Ping(ctx context.Context) error {
	err := s.storageService.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}
