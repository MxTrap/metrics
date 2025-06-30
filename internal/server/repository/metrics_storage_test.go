package repository

import (
	"context"
	"github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	storage, err := NewMemStorage()
	require.NoError(t, err)
	assert.NotNil(t, storage)
	assert.NotNil(t, storage.metrics)
	assert.Empty(t, storage.metrics)
}

func TestPing(t *testing.T) {
	storage, err := NewMemStorage()
	require.NoError(t, err)

	err = storage.Ping(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
}

func TestSaveGauge(t *testing.T) {
	storage, err := NewMemStorage()
	require.NoError(t, err)

	metric := models.Metric{
		ID:    "testGauge",
		MType: models.Gauge,
		Value: utils.MakePointer(42.5),
	}
	err = storage.Save(context.Background(), metric)
	require.NoError(t, err)

	saved, err := storage.Find(context.Background(), "testGauge")
	require.NoError(t, err)
	assert.Equal(t, metric, saved)
}

func TestSaveCounter(t *testing.T) {
	storage, err := NewMemStorage()
	require.NoError(t, err)

	metric1 := models.Metric{
		ID:    "testCounter",
		MType: models.Counter,
		Delta: utils.MakePointer[int64](10),
	}
	err = storage.Save(context.Background(), metric1)
	require.NoError(t, err)

	metric2 := models.Metric{
		ID:    "testCounter",
		MType: models.Counter,
		Delta: utils.MakePointer[int64](20),
	}
	err = storage.Save(context.Background(), metric2)
	require.NoError(t, err)

	saved, err := storage.Find(context.Background(), "testCounter")
	require.NoError(t, err)
	assert.Equal(t, models.Counter, saved.MType)
	assert.Equal(t, int64(30), *saved.Delta)
}

func TestFindNotFound(t *testing.T) {
	storage, err := NewMemStorage()
	require.NoError(t, err)

	_, err = storage.Find(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "not found", err.Error())
}

func TestGetAll(t *testing.T) {
	storage, err := NewMemStorage()
	require.NoError(t, err)

	metrics := map[string]models.Metric{
		"gauge1":   {ID: "gauge1", MType: models.Gauge, Value: utils.MakePointer(42.5)},
		"counter1": {ID: "counter1", MType: models.Counter, Delta: utils.MakePointer(int64(100))},
	}
	for _, m := range metrics {
		err = storage.Save(context.Background(), m)
		require.NoError(t, err)
	}

	result, err := storage.GetAll(context.Background())
	require.NoError(t, err)
	assert.Equal(t, metrics, result)
}

func TestGetAllEmpty(t *testing.T) {
	storage, err := NewMemStorage()
	require.NoError(t, err)

	result, err := storage.GetAll(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestSaveAll(t *testing.T) {
	storage, err := NewMemStorage()
	require.NoError(t, err)

	existing := map[string]models.Metric{
		"gauge1": {ID: "gauge1", MType: models.Gauge, Value: utils.MakePointer(42.5)},
	}
	for _, m := range existing {
		err = storage.Save(context.Background(), m)
		require.NoError(t, err)
	}

	newMetrics := map[string]models.Metric{
		"gauge1":   {ID: "gauge1", MType: models.Gauge, Value: utils.MakePointer(99.9)},
		"counter1": {ID: "counter1", MType: models.Counter, Delta: utils.MakePointer[int64](100)},
	}
	err = storage.SaveAll(context.Background(), newMetrics)
	require.NoError(t, err)

	result, err := storage.GetAll(context.Background())
	require.NoError(t, err)
	assert.Equal(t, newMetrics, result)
}
