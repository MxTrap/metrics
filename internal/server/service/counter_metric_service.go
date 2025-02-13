package service

import (
	"github.com/MxTrap/metrics/internal/server/models"
	"strconv"
)

type counterMetricsStorage interface {
	SaveCounterMetric(metric string, value int64)
	FindCounterMetric(metric string) (int64, bool)
}

type CounterMetricService struct {
	storage counterMetricsStorage
}

func NewCounterMetricService(storage counterMetricsStorage) *CounterMetricService {
	return &CounterMetricService{
		storage: storage,
	}
}

func (*CounterMetricService) parse(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

func (s *CounterMetricService) SaveJSON(metric string, value any) error {
	parsed, ok := value.(*int64)
	if !ok {
		return models.ErrWrongMetricValue
	}
	s.storage.SaveCounterMetric(metric, *parsed)
	return nil
}

func (s *CounterMetricService) Save(metric string, value string) error {
	parsedValue, err := s.parse(value)
	if err != nil {
		return models.ErrWrongMetricValue
	}
	s.storage.SaveCounterMetric(metric, parsedValue)
	return nil
}
func (s *CounterMetricService) Find(metric string) (any, bool) {
	return s.storage.FindCounterMetric(metric)
}
