package service

import (
	"context"
	"errors"
	common_models "github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/common/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

var gVal1 = utils.MakePointer[float64](1.1)
var cVal1 = utils.MakePointer[int64](1)

type mockStorage struct {
	metrics map[string]common_models.Metric
}

func (s *mockStorage) Save(_ context.Context, metric common_models.Metric) error {
	s.metrics[metric.ID] = metric
	return nil
}

func (s *mockStorage) SaveAll(_ context.Context, metrics map[string]common_models.Metric) error {
	s.metrics = metrics
	return nil
}

func (s *mockStorage) Find(_ context.Context, metric string) (common_models.Metric, error) {
	value, ok := s.metrics[metric]
	if !ok {
		return common_models.Metric{}, errors.New("not found")
	}
	return value, nil
}

func (s *mockStorage) GetAll(_ context.Context) (map[string]any, error) {
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
	return dst, nil
}

func (s mockStorage) Ping(_ context.Context) error {
	return nil
}

func newMockStorage() MetricStorageService {
	return &mockStorage{map[string]common_models.Metric{
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
		metric  common_models.Metric
		want    any
		wantErr bool
	}{
		{
			"test find gauge value 1",
			common_models.Metric{
				ID:    "gauge1",
				MType: "gauge",
			},
			common_models.Metric{
				ID:    "gauge1",
				MType: "gauge",
				Value: gVal1,
			},
			false,
		},
		{
			"test find unknown metric type",
			common_models.Metric{
				ID:    "gauge1",
				MType: "unknown",
			},
			common_models.Metric{},
			true,
		},
		{
			"test find unknown metric",
			common_models.Metric{
				ID:    "gau",
				MType: "gauge",
			},
			common_models.Metric{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newMockService()
			ctx := context.Background()
			got, err := s.Find(ctx, tt.metric)
			if (err != nil) != tt.wantErr {
				assert.Errorf(t, err, "find() error = %v, wantErr %v", err, tt.wantErr)
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
				storageService: &mockStorage{map[string]common_models.Metric{}},
			},
			want: map[string]any{},
		},
		{
			name:    "test get all data from mocked storageService",
			service: &MetricsService{storageService: newMockStorage()},
			want: map[string]any{
				"gauge1":   1.1,
				"counter1": int64(1),
			},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			all, err := tt.service.GetAll(ctx)
			if err != nil {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, all)
		})
	}
}

func TestMetricsService_Save(t *testing.T) {
	tests := []struct {
		name    string
		metric  common_models.Metric
		wantErr bool
	}{
		{
			name: "test save valid data",
			metric: common_models.Metric{
				ID:    "gau",
				MType: "gauge",
				Delta: cVal1,
			},
			wantErr: false,
		},
		{
			name: "test save valid data with invalid metric type",
			metric: common_models.Metric{
				ID:    "gau",
				MType: "unknown",
				Delta: cVal1,
			},
			wantErr: true,
		},
		{
			name: "test save without data",
			metric: common_models.Metric{
				ID:    "gau",
				MType: "gauge",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newMockService()
			ctx := context.Background()
			if err := s.Save(ctx, tt.metric); (err != nil) != tt.wantErr {
				assert.Errorf(t, err, "save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
