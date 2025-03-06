package repository

import (
	"github.com/MxTrap/metrics/internal/agent/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetricsStorage_GetMetrics(t *testing.T) {
	gaugeMetrics := *models.NewGaugeMetrics()
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
			name: "test get metrics",
			fields: fields{
				storage: models.Metrics{
					Gauge: gaugeMetrics,
					Counter: models.CounterMetrics{
						PollCount: 1,
					},
				},
			},
			want: models.Metrics{
				Gauge: gaugeMetrics,
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
		args []map[string]float64
		want models.Metrics
	}{
		{
			name: "test save metrics",
			args: []map[string]float64{
				{
					"gauge1": 1,
					"gauge2": 2.2,
				},
			},
			want: models.Metrics{
				Gauge: models.GaugeMetrics{
					Metrics: map[string]float64{
						"gauge1":      1,
						"gauge2":      2.2,
						"RandomValue": 1,
					},
				},
				Counter: models.CounterMetrics{
					PollCount: 1,
				},
			},
		},

		{
			name: "test save 2 metrics",
			args: []map[string]float64{
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
				Gauge: models.GaugeMetrics{
					Metrics: map[string]float64{
						"gauge2":      2,
						"gauge3":      3.3,
						"RandomValue": 1,
					},
				},
				Counter: models.CounterMetrics{
					PollCount: 2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := *NewMetricsStorage()
			for _, val := range tt.args {
				s.SaveMetrics(val)
				s.storage.Gauge.Metrics["RandomValue"] = 1
			}
			assert.Equal(t, tt.want.Gauge.Metrics, s.storage.Gauge.Metrics)
			assert.Equal(t, tt.want.Counter, s.storage.Counter)
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
					Gauge:   *models.NewGaugeMetrics(),
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
