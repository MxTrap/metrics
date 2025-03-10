package repository

import (
	"context"
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
		args []common_models.Metric
		want map[string]common_models.Metric
	}{
		{
			name: "save counter metric",
			args: []common_models.Metric{
				{
					ID:    "metric",
					MType: "counter",
					Delta: p1,
				},
			},
			want: map[string]common_models.Metric{
				"metric": {
					ID:    "metric",
					MType: "counter",
					Delta: p1,
				},
			},
		},
		{
			name: "save some counter metric",
			args: []common_models.Metric{
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
			want: map[string]common_models.Metric{
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
			name: "save same counter metrics",
			args: []common_models.Metric{
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
			want: map[string]common_models.Metric{
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
				metrics: map[string]common_models.Metric{},
			}
			ctx := context.TODO()
			for _, arg := range tt.args {
				err := s.Save(ctx, arg)
				assert.NoError(t, err)
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
				metrics: map[string]common_models.Metric{},
			},
		},
	}
	for _, tt := range tests {
		storage, err := NewMemStorage()
		assert.NoError(t, err)
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, storage, tt.want)
		})
	}
}
