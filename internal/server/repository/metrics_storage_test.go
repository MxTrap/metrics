package repository

import (
	common_models "github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/common/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemStorage_Save(t *testing.T) {
	p1 := utils.MakePointer[int64](1)
	p2 := utils.MakePointer[int64](2)
	p3 := utils.MakePointer[int64](3)

	tests := []struct {
		name string
		args []common_models.Metrics
		want map[string]common_models.Metrics
	}{
		{
			name: "Save counter metric",
			args: []common_models.Metrics{
				{
					ID:    "metric",
					MType: "counter",
					Delta: p1,
				},
			},
			want: map[string]common_models.Metrics{
				"metric": {
					ID:    "metric",
					MType: "counter",
					Delta: p1,
				},
			},
		},
		{
			name: "Save some counter metric",
			args: []common_models.Metrics{
				{
					ID:    "metric1",
					MType: "counter",
					Delta: p1,
				},
				{
					ID:    "metric2",
					MType: "counter",
					Delta: p2,
				},
				{
					ID:    "metric3",
					MType: "counter",
					Delta: p3,
				},
			},
			want: map[string]common_models.Metrics{
				"metric1": {
					ID:    "metric1",
					MType: "counter",
					Delta: p1,
				},
				"metric2": {
					ID:    "metric2",
					MType: "counter",
					Delta: p2,
				},
				"metric3": {
					ID:    "metric3",
					MType: "counter",
					Delta: p3,
				},
			},
		},
		{
			name: "Save same counter metrics",
			args: []common_models.Metrics{
				{
					ID:    "metric1",
					MType: "counter",
					Delta: p1,
				},
				{
					ID:    "metric1",
					MType: "counter",
					Delta: p2,
				},
			},
			want: map[string]common_models.Metrics{
				"metric1": {
					ID:    "metric1",
					MType: "counter",
					Delta: p2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				metrics: map[string]common_models.Metrics{},
			}
			for _, arg := range tt.args {
				err := s.Save(arg)
				if err != nil {
					return
				}
			}
			assert.Equal(t, tt.want, s.metrics)
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
				metrics: map[string]common_models.Metrics{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, NewMemStorage(), tt.want)
		})
	}
}
