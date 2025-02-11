package service

import (
	"github.com/MxTrap/metrics/internal/server/models"
	"strconv"
)

type gaugeMetricsStorage interface {
	SaveGaugeMetric(metric string, value float64)
	FindGaugeMetric(metric string) (float64, bool)
}

type GaugeMetricService struct {
	storage gaugeMetricsStorage
}

func NewGaugeMetricService(storage gaugeMetricsStorage) *GaugeMetricService {
	return &GaugeMetricService{
		storage: storage,
	}
}

func (s *GaugeMetricService) parse(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}
func (s *GaugeMetricService) Save(metric string, value string) error {
	parsed, err := s.parse(value)
	if err != nil {
		return models.ErrWrongMetricValue
	}
	s.storage.SaveGaugeMetric(metric, parsed)
	return nil
}
func (s *GaugeMetricService) Find(metric string) (any, bool) {
	return s.storage.FindGaugeMetric(metric)
}
