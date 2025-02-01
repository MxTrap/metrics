package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/MxTrap/metrics/internal/server/models"
)

type Storage interface {
	SaveGaugeMetric(metric string, value float64)
	SaveCounterMetric(metric string, value int64)
	FindGaugeMetric(metric string) (float64, bool)
	FindCounterMetric(metric string) (int64, bool)
	GetAll() map[string]any
}

type MetricsService struct {
	metricTypes metricTypes
	storage     Storage
}

type metricTypes map[string]metricTypeFunc

type metricTypeFunc struct {
	ParseFunc func(str string) (any, error)
	SaveFunc  func(metric string, value any) error
	FindFunc  func(metric string) (any, bool)
}

func (t metricTypes) GetMetricFunctions(metricType string) (metricTypeFunc, bool) {
	val, ok := t[metricType]
	return val, ok
}

func NewMetricsService(storage Storage) *MetricsService {

	return &MetricsService{
		storage: storage,
		metricTypes: metricTypes{
			"gauge": {
				ParseFunc: func(str string) (any, error) {
					return strconv.ParseFloat(str, 64)
				},
				SaveFunc: func(metric string, value any) error {
					cVal, ok := value.(float64)
					if !ok {
						return errors.New("value is not a float64")
					}
					storage.SaveGaugeMetric(metric, cVal)
					return nil
				},
				FindFunc: func(metric string) (any, bool) {
					return storage.FindGaugeMetric(metric)
				},
			},

			"counter": {
				ParseFunc: func(str string) (any, error) {
					return strconv.ParseInt(str, 10, 64)
				},
				SaveFunc: func(metric string, value any) error {
					cVal, ok := value.(int64)
					if !ok {
						return errors.New("value is not a int64")
					}
					storage.SaveCounterMetric(metric, cVal)
					return nil
				},
				FindFunc: func(metric string) (any, bool) {
					return storage.FindCounterMetric(metric)
				},
			},
		},
	}
}

func (_ *MetricsService) parseURL(url string, searchWord string) ([]string, error) {
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

	acceptedMetricType, ok := s.metricTypes.GetMetricFunctions(metricType)
	if !ok {
		return models.ErrUnknownMetricType
	}

	parsedValue, err := acceptedMetricType.ParseFunc(metricValue)
	if err != nil || parsedValue == nil {
		return models.ErrWrongMetricValue
	}

	err = acceptedMetricType.SaveFunc(metric, parsedValue)
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
	metricFunc, ok := s.metricTypes.GetMetricFunctions(metricType)
	if !ok {
		return nil, models.ErrUnknownMetricType
	}
	val, ok := metricFunc.FindFunc(metric)
	if !ok {
		return nil, models.ErrNotFoundMetric
	}
	return val, nil
}

func (s *MetricsService) GetAll() map[string]any {
	return s.storage.GetAll()
}
