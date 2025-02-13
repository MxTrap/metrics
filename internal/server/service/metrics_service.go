package service

import (
	common_moodels "github.com/MxTrap/metrics/internal/common/models"
	"strings"

	"github.com/MxTrap/metrics/internal/server/models"
)

type Storage interface {
	gaugeMetricsStorage
	counterMetricsStorage
	GetAll() map[string]any
}

type metricTypes map[string]metricTypeService

type metricTypeService interface {
	Save(metric string, value string) error
	SaveJSON(metric string, value any) error
	Find(metric string) (any, bool)
}

func (t metricTypes) GetMetricTypeService(metricType string) (metricTypeService, bool) {
	val, ok := t[metricType]
	return val, ok
}

type MetricsService struct {
	metricTypes metricTypes
	storage     Storage
}

func NewMetricsService(storage Storage) *MetricsService {

	return &MetricsService{
		storage: storage,
		metricTypes: metricTypes{
			"gauge":   NewGaugeMetricService(storage),
			"counter": NewCounterMetricService(storage),
		},
	}
}

func (*MetricsService) parseURL(url string, searchWord string) ([]string, error) {
	idx := strings.Index(url, searchWord+"/")
	if idx == -1 {
		return nil, models.ErrNotFoundMetric
	}
	splitedURL := strings.Split(url[idx+len(searchWord)+1:], "/")

	if len(splitedURL) < 2 {
		return nil, models.ErrNotFoundMetric
	}
	return splitedURL, nil
}

func (s *MetricsService) Save(url string) error {
	parsedMetric, err := s.parseURL(url, "update")
	if err != nil {
		return err
	}
	if len(parsedMetric) < 3 {
		return models.ErrWrongMetricValue
	}
	metricType, metric, metricValue := parsedMetric[0], parsedMetric[1], parsedMetric[2]

	acceptedMetricType, ok := s.metricTypes.GetMetricTypeService(metricType)
	if !ok {
		return models.ErrUnknownMetricType
	}

	err = acceptedMetricType.Save(metric, metricValue)
	if err != nil {
		return err
	}

	return nil
}

func (s *MetricsService) SaveJSON(metric common_moodels.Metrics) error {

	acceptedMetricType, ok := s.metricTypes.GetMetricTypeService(metric.MType)
	if !ok {
		return models.ErrUnknownMetricType
	}

	var value any

	if metric.Delta != nil {
		value = metric.Delta
	}

	if metric.Value != nil {
		value = metric.Value
	}

	if value == nil {
		return models.ErrWrongMetricValue
	}

	err := acceptedMetricType.SaveJSON(metric.ID, value)
	if err != nil {
		return err
	}

	return nil
}

func (s *MetricsService) Find(url string) (any, error) {
	parsedMetric, err := s.parseURL(url, "value")
	if err != nil {
		return nil, err
	}
	metricType, metric := parsedMetric[0], parsedMetric[1]
	metricFunc, ok := s.metricTypes.GetMetricTypeService(metricType)
	if !ok {
		return nil, models.ErrUnknownMetricType
	}
	val, ok := metricFunc.Find(metric)
	if !ok {
		return nil, models.ErrNotFoundMetric
	}
	return val, nil
}

func (s *MetricsService) FindJSON(metric common_moodels.Metrics) (common_moodels.Metrics, error) {
	metricFunc, ok := s.metricTypes.GetMetricTypeService(metric.MType)
	if !ok {
		return common_moodels.Metrics{}, models.ErrUnknownMetricType
	}
	val, ok := metricFunc.Find(metric.ID)
	if !ok {
		return common_moodels.Metrics{}, models.ErrNotFoundMetric
	}
	if metric.MType == common_moodels.Gauge {
		cVal, ok := val.(float64)
		if !ok {
			return common_moodels.Metrics{}, models.ErrNotFoundMetric
		}
		metric.Value = &cVal
	}

	if metric.MType == common_moodels.Counter {
		cVal, ok := val.(int64)
		if !ok {
			return common_moodels.Metrics{}, models.ErrNotFoundMetric
		}
		metric.Delta = &cVal
	}
	return metric, nil
}

func (s *MetricsService) GetAll() map[string]any {
	return s.storage.GetAll()
}
