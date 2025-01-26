package service

import (
	"fmt"
	"strings"

	"github.com/MxTrap/metrics/internal/server/models"
)

type MemStorage struct {
	metrics map[string]map[string]any
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: map[string]map[string]any{},
	}
}

func (s *MemStorage) Save(url string) error {
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
	acceptedMetricType, ok := models.AcceptableMetricTypes[metricType]
	if !ok {
		return models.ErrUnknownMetricType
	}

	parsedValue, err := acceptedMetricType.ParseFunc(metricValue)
	if err != nil {
		return models.ErrWrongMetricValue
	}

	if _, ok := s.metrics[metricType]; !ok {
		s.metrics[metricType] = map[string]any{}
	}
	store := s.metrics[metricType]

	if err := acceptedMetricType.SaveFunc(store, metric, parsedValue); err != nil {
		return err
	}

	fmt.Println(s.metrics)

	return nil
}
