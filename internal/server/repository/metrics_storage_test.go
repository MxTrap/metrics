package repository

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemStorage_SaveCounterMetric(t *testing.T) {
	type args struct {
		metric string
		value  int64
	}
	tests := []struct {
		name string
		args []args
		want map[string]int64
	}{
		{
			name: "Save counter metric",
			args: []args{
				{"metric", 1},
			},
			want: map[string]int64{
				"metric": 1,
			},
		},
		{
			name: "Save some counter metric",
			args: []args{
				{"metric1", 1},
				{"metric2", 2},
				{"metric3", 3},
			},
			want: map[string]int64{
				"metric1": 1,
				"metric2": 2,
				"metric3": 3,
			},
		},
		{
			name: "Save same counter metrics",
			args: []args{
				{"metric1", 1},
				{"metric1", 1},
			},
			want: map[string]int64{
				"metric1": 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				counter: map[string]int64{},
			}
			for _, arg := range tt.args {
				s.SaveCounterMetric(arg.metric, arg.value)
			}
			assert.Equal(t, tt.want, s.counter)
		})
	}
}

func TestMemStorage_SaveGaugeMetric(t *testing.T) {
	type args struct {
		metric string
		value  float64
	}
	tests := []struct {
		name string
		args []args
		want map[string]float64
	}{
		{
			name: "Save counter metric",
			args: []args{
				{"metric", 1.1},
			},
			want: map[string]float64{
				"metric": 1.1,
			},
		},
		{
			name: "Save some counter metric",
			args: []args{
				{"metric1", 1.1},
				{"metric2", 2.3},
				{"metric3", 3},
			},
			want: map[string]float64{
				"metric1": 1.1,
				"metric2": 2.3,
				"metric3": 3,
			},
		},
		{
			name: "Save same counter metrics",
			args: []args{
				{"metric", 1},
				{"metric", 2.4},
			},
			want: map[string]float64{
				"metric": 2.4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				gauge: map[string]float64{},
			}
			for _, arg := range tt.args {
				s.SaveGaugeMetric(arg.metric, arg.value)
			}
			assert.Equal(t, tt.want, s.gauge)
		})
	}

}

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want *MemStorage
	}{
		{
			name: "test storage creation",
			want: &MemStorage{
				counter: map[string]int64{},
				gauge:   map[string]float64{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, NewMemStorage(), tt.want)
		})
	}
}
