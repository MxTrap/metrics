package service

import (
	"context"
	common_models "github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/server/models"
)

type MetricStorageService interface {
	Save(ctx context.Context, metrics common_models.Metric) error
	SaveAll(ctx context.Context, metrics map[string]common_models.Metric) (err error)
	Find(ctx context.Context, metric string) (common_models.Metric, error)
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

func (s *MetricsService) Save(ctx context.Context, metric common_models.Metric) error {
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

func (s *MetricsService) SaveAll(ctx context.Context, metrics []common_models.Metric) error {
	m := make(map[string]common_models.Metric)
	for _, metric := range metrics {
		if s.validateMetric(metric.MType) {
			if metric.MType == common_models.Counter {
				val, ok := m[metric.ID]
				if ok {
					*val.Delta = *metric.Delta + *val.Delta
					m[metric.ID] = val
					continue
				}
			}
			m[metric.ID] = metric
		}
	}
	err := s.storageService.SaveAll(ctx, m)
	if err != nil {
		return err
	}

	return nil
}

func (s *MetricsService) Find(ctx context.Context, metric common_models.Metric) (common_models.Metric, error) {
	if !s.validateMetric(metric.MType) {
		return common_models.Metric{}, models.ErrUnknownMetricType
	}
	val, err := s.storageService.Find(ctx, metric.ID)
	if err != nil {
		return common_models.Metric{}, models.ErrNotFoundMetric
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
