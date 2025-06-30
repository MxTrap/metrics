package repository

import (
	"encoding/json"
	"github.com/MxTrap/metrics/internal/utils"
	"os"
	"path/filepath"
	"testing"

	"github.com/MxTrap/metrics/internal/common/models"
	"github.com/stretchr/testify/assert"
)

func setupTestStorage(t *testing.T) (*MetricsFileStorage, string) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "metrics_test_*.json")
	assert.NoError(t, err, "Failed to create temp file")

	// Инициализируем MetricsFileStorage
	storage := &MetricsFileStorage{
		filePath: tmpFile.Name(),
		file:     tmpFile,
	}

	return storage, tmpFile.Name()
}

func TestNewMetricsFileStorage(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "metrics.json")

	storage := NewMetricsFileStorage(filePath)
	assert.NotNil(t, storage, "NewMetricsFileStorage should return non-nil storage")
	assert.Equal(t, filePath, storage.filePath, "File path should match")
	assert.NotNil(t, storage.file, "File should be initialized")

	err := storage.Close()
	assert.NoError(t, err, "Close should not return error")

	invalidPath := tmpDir
	storage = NewMetricsFileStorage(invalidPath)
	assert.Nil(t, storage, "NewMetricsFileStorage should return nil for invalid path")
}

func TestMetricsFileStorage_Save(t *testing.T) {
	storage, tmpFilePath := setupTestStorage(t)
	defer os.Remove(tmpFilePath)

	metrics := map[string]models.Metric{
		"gauge1":   {ID: "gauge1", MType: models.Gauge, Value: utils.MakePointer(42.5)},
		"counter1": {ID: "counter1", MType: models.Counter, Delta: utils.MakePointer(int64(100))},
	}

	err := storage.Save(metrics)
	assert.NoError(t, err, "Save should not return error")

	data, err := os.ReadFile(tmpFilePath)
	assert.NoError(t, err, "Failed to read temp file")
	var savedMetrics map[string]models.Metric
	err = json.Unmarshal(data, &savedMetrics)
	assert.NoError(t, err, "Failed to unmarshal saved data")
	assert.Equal(t, metrics, savedMetrics, "Saved metrics should match input")

	invalidMetrics := map[string]models.Metric{
		"invalid": {ID: "invalid", MType: "unknown", Value: nil},
	}

	err = storage.Save(invalidMetrics)
	assert.NoError(t, err, "Save should handle invalid metrics gracefully")
}

func TestMetricsFileStorage_Read(t *testing.T) {
	storage, tmpFilePath := setupTestStorage(t)
	defer os.Remove(tmpFilePath)

	result, err := storage.Read()
	assert.NoError(t, err, "Read should not return error for empty file")
	assert.Equal(t, map[string]models.Metric{}, result, "Read should return empty map for empty file")

	metrics := map[string]models.Metric{
		"gauge1":   {ID: "gauge1", MType: models.Gauge, Value: utils.MakePointer(42.5)},
		"counter1": {ID: "counter1", MType: models.Counter, Delta: utils.MakePointer(int64(100))},
	}
	data, err := json.Marshal(metrics)
	assert.NoError(t, err, "Failed to marshal test metrics")
	err = os.WriteFile(tmpFilePath, data, os.ModePerm)
	assert.NoError(t, err, "Failed to write test data to file")

	storage.file.Close()
	storage.file, err = os.OpenFile(tmpFilePath, os.O_RDWR, os.ModePerm)
	assert.NoError(t, err, "Failed to reopen file")

	result, err = storage.Read()
	assert.NoError(t, err, "Read should not return error")
	assert.Equal(t, metrics, result, "Read should return correct metrics")

	storage.file.Close()
	err = os.WriteFile(tmpFilePath, []byte("invalid json"), os.ModePerm)
	assert.NoError(t, err, "Failed to write invalid data")
	storage.file, err = os.OpenFile(tmpFilePath, os.O_RDWR, os.ModePerm)
	assert.NoError(t, err, "Failed to reopen file")

	result, err = storage.Read()
	assert.Error(t, err, "Read should return error for invalid JSON")
	assert.Nil(t, result, "Read should return nil for invalid JSON")
}

func TestMetricsFileStorage_Close(t *testing.T) {
	storage, tmpFilePath := setupTestStorage(t)
	defer os.Remove(tmpFilePath)

	err := storage.Close()
	assert.NoError(t, err, "Close should not return error")

	_, err = storage.file.Write([]byte("test"))
	assert.Error(t, err, "Write to closed file should return error")
}
