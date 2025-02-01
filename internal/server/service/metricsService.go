package service

import (
	"errors"
	"fmt"
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

type StorageService struct {
	metricTypes MetricTypes
	storage     Storage
}

type MetricTypeFunc struct {
	ParseFunc func(str string) (any, error)
	SaveFunc  func(metric string, value any) error
	FindFunc  func(metric string) (any, bool)
}

type MetricTypes map[string]any

func (t *MetricTypes) GetMetricFunctions(metricType string) *MetricTypeFunc {
	if val, ok := (*t)[metricType]; ok {
		c, ok := val.(MetricTypeFunc)
		if !ok {
			return nil
		}
		return &c
	}

	return nil
}

func NewMetricsService(storage Storage) *StorageService {

	return &StorageService{
		storage: storage,
		metricTypes: MetricTypes{
			"gauge": MetricTypeFunc{
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

			"counter": MetricTypeFunc{
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

func (s StorageService) parseUrl(url string, searchWord string) ([]string, error) {
	idx := strings.Index(url, searchWord)
	if idx == -1 {
		return nil, models.ErrNotFoundMetric
	}
	splitedURL := strings.Split(url[idx+len(searchWord)+1:], "/")

	if len(splitedURL) < 2 {
		return nil, models.ErrNotFoundMetric
	}
	return splitedURL, nil
}

func (s *StorageService) Save(url string) error {
	parsedMetric, err := s.parseUrl(url, "update")
	if err != nil {
		return err
	}
	metricType, metric, metricValue := parsedMetric[0], parsedMetric[1], parsedMetric[2]

	acceptedMetricType := s.metricTypes.GetMetricFunctions(metricType)
	if acceptedMetricType == nil {
		return models.ErrUnknownMetricType
	}

	parsedValue, err := (*acceptedMetricType).ParseFunc(metricValue)
	if err != nil || parsedValue == nil {
		return models.ErrWrongMetricValue
	}

	err = acceptedMetricType.SaveFunc(metric, parsedValue)
	if err != nil {
		return err
	}

	return nil
}

func (s *StorageService) Find(url string) (any, error) {
	parsedMetric, err := s.parseUrl(url, "value")
	if err != nil {
		return nil, err
	}
	metricType, metric := parsedMetric[0], parsedMetric[1]
	fmt.Println(metricType, metric)
	metricFunc := s.metricTypes.GetMetricFunctions(metricType)
	if metricFunc == nil {
		return nil, models.ErrUnknownMetricType
	}
	val, ok := metricFunc.FindFunc(metric)
	if !ok {
		return nil, models.ErrNotFoundMetric
	}
	return val, nil
}

func (s *StorageService) GetAll() map[string]any {
	return s.storage.GetAll()
}
