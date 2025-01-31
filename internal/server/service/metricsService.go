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
}

type StorageService struct {
	metricTypes MetricTypes
}

type MetricTypeFunc struct {
	ParseFunc func(str string) (any, error)
	SaveFunc  func(metric string, value any) error
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
			},
		},
	}
}

func (s *StorageService) Save(url string) error {
	searchWord := "update"
	idx := strings.Index(url, searchWord)
	if idx == -1 {
		return models.ErrNotFoundMetric
	}
	splitedURL := strings.Split(url[idx+len(searchWord)+1:], "/")

	if len(splitedURL) < 3 {
		return models.ErrNotFoundMetric
	}

	metricType, metric, metricValue := splitedURL[0], splitedURL[1], splitedURL[2]

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
