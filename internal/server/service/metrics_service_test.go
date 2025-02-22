package service

import (
	common_models "github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/common/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

var gVal1 = utils.MakePointer[float64](1.1)
var cVal1 = utils.MakePointer[int64](1)

type mockStorage struct {
	metrics map[string]common_models.Metrics
}

func (s *mockStorage) Save(metric common_models.Metrics) {
	s.metrics[metric.ID] = metric
}

func (s *mockStorage) Find(metric string) (common_models.Metrics, bool) {
	value, ok := s.metrics[metric]
	return value, ok
}

func (s *mockStorage) GetAll() map[string]any {
	dst := map[string]any{}
	for k, v := range s.metrics {
		var val any
		if v.Delta != nil {
			val = *v.Delta
		}
		if v.Value != nil {
			val = *v.Value
		}
		dst[k] = val
	}
	return dst
}
func newMockStorage() MetricStorageService {
	return &mockStorage{map[string]common_models.Metrics{
		"gauge1": {
			ID:    "gauge1",
			MType: "gauge",
			Value: gVal1,
		},
		"counter1": {
			ID:    "counter1",
			MType: "counter",
			Delta: cVal1,
		},
	}}
}

func newMockService() *MetricsService {
	storage := newMockStorage()
	return &MetricsService{
		storageService: storage,
	}

}

func TestMetricsService_Find(t *testing.T) {
	tests := []struct {
		name    string
		metric  common_models.Metrics
		want    any
		wantErr bool
	}{
		{
			"test find gauge value 1",
			common_models.Metrics{
				ID:    "gauge1",
				MType: "gauge",
			},
			common_models.Metrics{
				ID:    "gauge1",
				MType: "gauge",
				Value: gVal1,
			},
			false,
		},
		{
			"test find unknown metric type",
			common_models.Metrics{
				ID:    "gauge1",
				MType: "unknown",
			},
			common_models.Metrics{},
			true,
		},
		{
			"test find unknown metric",
			common_models.Metrics{
				ID:    "gau",
				MType: "gauge",
			},
			common_models.Metrics{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newMockService()
			got, err := s.Find(tt.metric)
			if (err != nil) != tt.wantErr {
				assert.Errorf(t, err, "Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMetricsService_GetAll(t *testing.T) {
	tests := []struct {
		name    string
		service *MetricsService
		want    map[string]any
	}{
		{
			name: "test get all data from empty storageService",
			service: &MetricsService{
				storageService: &mockStorage{map[string]common_models.Metrics{}},
			},
			want: map[string]any{},
		},
		{
			name:    "test get all data from mocked storageService",
			service: &MetricsService{storageService: newMockStorage()},
			want: map[string]any{
				"gauge1":   gVal1,
				"counter1": cVal1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.service.GetAll())
		})
	}
}

func TestMetricsService_Save(t *testing.T) {
	tests := []struct {
		name    string
		metric  common_models.Metrics
		wantErr bool
	}{
		{
			name: "test save valid data",
			metric: common_models.Metrics{
				ID:    "gau",
				MType: "gauge",
				Delta: cVal1,
			},
			wantErr: false,
		},
		{
			name: "test save valid data with invalid metric type",
			metric: common_models.Metrics{
				ID:    "gau",
				MType: "unknown",
				Delta: cVal1,
			},
			wantErr: true,
		},
		{
			name: "test save without data",
			metric: common_models.Metrics{
				ID:    "gau",
				MType: "gauge",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newMockService()
			if err := s.Save(tt.metric); (err != nil) != tt.wantErr {
				assert.Errorf(t, err, "Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
