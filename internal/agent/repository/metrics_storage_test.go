package repository

import (
	"github.com/MxTrap/metrics/internal/agent/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetricsStorage_GetMetrics(t *testing.T) {
	type fields struct {
		storage models.Metrics
	}
	tests := []struct {
		name   string
		fields fields
		want   models.Metrics
	}{
		{
			name:   "test get empty metrics",
			fields: fields{},
			want:   models.Metrics{},
		},
		{
			name: "test get empty metrics",
			fields: fields{
				storage: models.Metrics{
					Gauge: map[string]float64{
						"gauge1": 1,
						"gauge2": 2.2,
					},
					Counter: models.CounterMetrics{
						PollCount: 1,
					},
				},
			},
			want: models.Metrics{
				Gauge: map[string]float64{
					"gauge1": 1,
					"gauge2": 2.2,
				},
				Counter: models.CounterMetrics{
					PollCount: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MetricsStorage{
				storage: tt.fields.storage,
			}
			assert.Equal(t, tt.want, s.GetMetrics())
		})
	}
}

func TestMetricsStorage_SaveMetrics(t *testing.T) {
	tests := []struct {
		name string
		args []models.GaugeMetrics
		want models.Metrics
	}{
		{
			name: "test save metrics",
			args: []models.GaugeMetrics{
				{
					"gauge1": 1,
					"gauge2": 2.2,
				},
			},
			want: models.Metrics{
				Gauge: map[string]float64{
					"gauge1":      1,
					"gauge2":      2.2,
					"RandomValue": 1,
				},
				Counter: models.CounterMetrics{
					PollCount: 1,
				},
			},
		},

		{
			name: "test save 2 metrics",
			args: []models.GaugeMetrics{
				{
					"gauge1": 1,
					"gauge2": 2.2,
				},
				{
					"gauge2": 2,
					"gauge3": 3.3,
				},
			},
			want: models.Metrics{
				Gauge: map[string]float64{
					"gauge1":      1,
					"gauge2":      2,
					"gauge3":      3.3,
					"RandomValue": 1,
				},
				Counter: models.CounterMetrics{
					PollCount: 2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MetricsStorage{
				storage: models.Metrics{
					Gauge: models.GaugeMetrics{},
				},
			}
			for _, val := range tt.args {
				s.SaveMetrics(val)
				s.storage.Gauge["RandomValue"] = 1
			}
			assert.Equal(t, tt.want, s.storage)
		})
	}
}

func TestNewMetricsStorage(t *testing.T) {
	tests := []struct {
		name string
		want *MetricsStorage
	}{
		{
			name: "test storage creation",
			want: &MetricsStorage{
				storage: models.Metrics{
					Gauge:   map[string]float64{},
					Counter: models.CounterMetrics{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMetricsStorage())
		})
	}
}
