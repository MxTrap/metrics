package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockStorage struct {
	storage map[string]any
}

func (s *mockStorage) SaveGaugeMetric(metric string, value float64) {
	s.storage[metric] = value
}

func (s *mockStorage) SaveCounterMetric(metric string, value int64) {
	s.storage[metric] = value
}

func (s *mockStorage) FindGaugeMetric(metric string) (float64, bool) {
	value, ok := s.storage[metric]
	return value.(float64), ok
}

func (s *mockStorage) FindCounterMetric(metric string) (int64, bool) {
	value, ok := s.storage[metric]
	return value.(int64), ok
}

func (s *mockStorage) GetAll() map[string]any {
	return s.storage
}
func newMockStorage() Storage {
	return &mockStorage{map[string]any{
		"gauge1":   1.1,
		"counter1": 1,
	}}
}

func newMockService() *MetricsService {
	storage := newMockStorage()
	return &MetricsService{
		storage: storage,
		metricTypes: metricTypes{
			"gauge": NewGaugeMetricService(storage),
		},
	}

}

func TestMetricsService_Find(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    any
		wantErr bool
	}{
		{
			"test find gauge value 1",
			"value/gauge/gauge1",
			1.1,
			false,
		},
		{
			"test wrong url",
			"valu/gauge/gauge1",
			nil,
			true,
		},
		{
			"test find unknown metric type",
			"value/gauge1/gauge1",
			nil,
			true,
		},
		{
			"test find unknown metric",
			"value/gauge1/gau",
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newMockService()
			got, err := s.Find(tt.url)
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
			name: "test get all data from empty storage",
			service: &MetricsService{
				storage: &mockStorage{make(map[string]any)},
			},
			want: map[string]any{},
		},
		{
			name:    "test get all data from mocked storage",
			service: &MetricsService{storage: newMockStorage()},
			want: map[string]any{
				"gauge1":   1.1,
				"counter1": 1,
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
		url     string
		wantErr bool
	}{
		{
			name:    "test save valid data",
			url:     "update/gauge/gauge1/1",
			wantErr: false,
		},
		{
			name:    "test save valid data with invalid url",
			url:     "upd/gauge/gauge1/1",
			wantErr: true,
		},
		{
			name:    "test save without data",
			url:     "update/gauge/gauge1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newMockService()
			if err := s.Save(tt.url); (err != nil) != tt.wantErr {
				assert.Errorf(t, err, "Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricsService_parseURL(t *testing.T) {
	type args struct {
		url        string
		searchWord string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "test parse valid url",
			args: args{
				url:        "/value/gauge/gauge1",
				searchWord: "value",
			},
			want:    []string{"gauge", "gauge1"},
			wantErr: false,
		},
		{
			name: "test parse invalid url",
			args: args{
				url:        "/value/gauge/gauge1",
				searchWord: "random",
			},
			want:    []string(nil),
			wantErr: true,
		},
		{
			name: "test parse too short url",
			args: args{
				url:        "/value",
				searchWord: "value",
			},
			want:    []string(nil),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			me := &MetricsService{}
			got, err := me.parseURL(tt.args.url, tt.args.searchWord)
			if (err != nil) != tt.wantErr {
				assert.Errorf(t, err, "parseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
