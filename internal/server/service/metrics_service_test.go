package service

import (
	"context"
	"errors"
	"github.com/MxTrap/metrics/internal/common/models"
	servermodels "github.com/MxTrap/metrics/internal/server/models"
	"github.com/MxTrap/metrics/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strconv"
	"sync"
	"testing"
	"time"
)

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) Save(ctx context.Context, metric models.Metric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *mockStorage) SaveAll(ctx context.Context, metrics map[string]models.Metric) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func (m *mockStorage) Find(ctx context.Context, metric string) (models.Metric, error) {
	args := m.Called(ctx, metric)
	return args.Get(0).(models.Metric), args.Error(1)
}

func (m *mockStorage) GetAll(ctx context.Context) (map[string]models.Metric, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]models.Metric), args.Error(1)
}

func (m *mockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type mockFileStorage struct {
	mock.Mock
}

func (m *mockFileStorage) Save(metrics map[string]models.Metric) error {
	args := m.Called(metrics)
	return args.Error(0)
}

func (m *mockFileStorage) Read() (map[string]models.Metric, error) {
	args := m.Called()
	return args.Get(0).(map[string]models.Metric), args.Error(1)
}

func (m *mockFileStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewMetricsService(t *testing.T) {
	storage := &mockStorage{}
	fileStorage := &mockFileStorage{}
	service := NewMetricsService(fileStorage, storage, 5, true)
	assert.NotNil(t, service)
	assert.Equal(t, fileStorage, service.fileStorage)
	assert.Equal(t, storage, service.storage)
	assert.Equal(t, 5, service.saveInterval)
	assert.True(t, service.restore)
	assert.Nil(t, service.ticker)
}

func TestValidateMetric(t *testing.T) {
	service := &MetricsService{}

	assert.True(t, service.validateMetric("gauge"))
	assert.True(t, service.validateMetric("counter"))
	assert.False(t, service.validateMetric("unknown"))
}

func TestSaveAll(t *testing.T) {
	storage := &mockStorage{}
	fileStorage := &mockFileStorage{}

	service := &MetricsService{storage: storage, fileStorage: fileStorage, saveInterval: 0}
	metrics := []models.Metric{
		{ID: "gauge1", MType: "gauge", Value: ptr(42.5)},
		{ID: "counter1", MType: "counter", Delta: ptr(int64(100))},
		{ID: "counter1", MType: "counter", Delta: ptr(int64(50))},
		{ID: "invalid", MType: "unknown"},
	}

	expectedMap := map[string]models.Metric{
		"gauge1":   {ID: "gauge1", MType: "gauge", Value: ptr(42.5)},
		"counter1": {ID: "counter1", MType: "counter", Delta: ptr(int64(150))},
	}

	storage.On("GetAll", mock.Anything).Return(map[string]models.Metric{
		"gauge1":   {ID: "gauge1", MType: "gauge", Value: ptr(42.5)},
		"counter1": {ID: "counter1", MType: "counter", Delta: ptr(int64(150))},
	}, nil)
	storage.On("SaveAll", mock.Anything, expectedMap).Return(nil)
	fileStorage.On("Save", expectedMap).Return(nil)

	err := service.SaveAll(context.Background(), metrics)
	assert.NoError(t, err)
	storage.AssertExpectations(t)
	fileStorage.AssertExpectations(t)
}

func TestSaveAllAsync(t *testing.T) {
	storage := &mockStorage{}
	fileStorage := &mockFileStorage{}
	service := &MetricsService{storage: storage, fileStorage: fileStorage, saveInterval: 5}
	metrics := []models.Metric{
		{ID: "gauge1", MType: "gauge", Value: ptr(42.5)},
	}
	expectedMap := map[string]models.Metric{
		"gauge1": {ID: "gauge1", MType: "gauge", Value: ptr(42.5)},
	}

	storage.On("SaveAll", mock.Anything, expectedMap).Return(nil)

	err := service.SaveAll(context.Background(), metrics)
	assert.NoError(t, err)
	storage.AssertExpectations(t)
}

func TestSaveAllError(t *testing.T) {
	storage := &mockStorage{}
	fileStorage := &mockFileStorage{}
	service := &MetricsService{storage: storage, fileStorage: fileStorage, saveInterval: 0}
	metrics := []models.Metric{
		{ID: "gauge1", MType: "gauge", Value: ptr(42.5)},
	}
	expectedMap := map[string]models.Metric{
		"gauge1": {ID: "gauge1", MType: "gauge", Value: ptr(42.5)},
	}

	storage.On("SaveAll", mock.Anything, expectedMap).Return(errors.New("storage error"))

	err := service.SaveAll(context.Background(), metrics)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage error")
	storage.AssertExpectations(t)
}

func TestSave(t *testing.T) {
	storage := &mockStorage{}
	fileStorage := &mockFileStorage{}
	service := &MetricsService{storage: storage, fileStorage: fileStorage, saveInterval: 0}
	metric := models.Metric{ID: "gauge1", MType: "gauge", Value: ptr(42.5)}
	metricsMap := map[string]models.Metric{"gauge1": metric}

	storage.On("Save", mock.Anything, metric).Return(nil)
	fileStorage.On("Save", metricsMap).Return(nil)
	storage.On("GetAll", mock.Anything).Return(metricsMap, nil)

	err := service.Save(context.Background(), metric)
	assert.NoError(t, err)
	storage.AssertExpectations(t)
	fileStorage.AssertExpectations(t)
}

func TestSaveInvalidType(t *testing.T) {
	service := &MetricsService{}
	metric := models.Metric{ID: "invalid", MType: "unknown"}
	err := service.Save(context.Background(), metric)
	assert.Error(t, err)
	assert.Equal(t, servermodels.ErrUnknownMetricType, err)
}

func TestSaveNoValue(t *testing.T) {
	service := &MetricsService{}
	metric := models.Metric{ID: "gauge1", MType: "gauge"}
	err := service.Save(context.Background(), metric)
	assert.Error(t, err)
	assert.Equal(t, servermodels.ErrWrongMetricValue, err)
}

func TestFind(t *testing.T) {
	storage := &mockStorage{}
	service := &MetricsService{storage: storage}
	metric := models.Metric{ID: "gauge1", MType: "gauge"}
	expected := models.Metric{ID: "gauge1", MType: "gauge", Value: ptr(42.5)}

	storage.On("Find", mock.Anything, "gauge1").Return(expected, nil)

	result, err := service.Find(context.Background(), metric)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	storage.AssertExpectations(t)
}

func TestFindInvalidType(t *testing.T) {
	service := &MetricsService{}
	metric := models.Metric{ID: "gauge1", MType: "unknown"}
	_, err := service.Find(context.Background(), metric)
	assert.Error(t, err)
	assert.Equal(t, servermodels.ErrUnknownMetricType, err)
}

func TestFindNotFound(t *testing.T) {
	storage := &mockStorage{}
	service := &MetricsService{storage: storage}
	metric := models.Metric{ID: "gauge1", MType: "gauge"}

	storage.On("Find", mock.Anything, "gauge1").Return(models.Metric{}, errors.New("not found"))

	_, err := service.Find(context.Background(), metric)
	assert.Error(t, err)
	assert.Equal(t, servermodels.ErrNotFoundMetric, err)
	storage.AssertExpectations(t)
}

func TestGetAll(t *testing.T) {
	storage := &mockStorage{}
	service := &MetricsService{storage: storage}
	metrics := map[string]models.Metric{
		"gauge1":   {ID: "gauge1", MType: "gauge", Value: ptr(42.5)},
		"counter1": {ID: "counter1", MType: "counter", Delta: ptr(int64(100))},
	}
	expected := map[string]interface{}{
		"gauge1":   42.5,
		"counter1": int64(100),
	}

	storage.On("GetAll", mock.Anything).Return(metrics, nil)

	result, err := service.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	storage.AssertExpectations(t)
}

func TestPing(t *testing.T) {
	storage := &mockStorage{}
	service := &MetricsService{storage: storage}

	storage.On("Ping", mock.Anything).Return(nil)

	err := service.Ping(context.Background())
	assert.NoError(t, err)
	storage.AssertExpectations(t)
}

func TestStartRestore(t *testing.T) {
	storage := &mockStorage{}
	fileStorage := &mockFileStorage{}
	service := &MetricsService{storage: storage, fileStorage: fileStorage, restore: true}
	metrics := map[string]models.Metric{
		"gauge1": {ID: "gauge1", MType: "gauge", Value: ptr(42.5)},
	}

	fileStorage.On("Read").Return(metrics, nil)
	storage.On("SaveAll", mock.Anything, metrics).Return(nil)

	err := service.Start(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, service.ticker)
	fileStorage.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestStartWithTicker(t *testing.T) {
	storage := &mockStorage{}
	fileStorage := &mockFileStorage{}
	service := &MetricsService{storage: storage, fileStorage: fileStorage, saveInterval: 1}
	metrics := map[string]models.Metric{
		"gauge1": {ID: "gauge1", MType: "gauge", Value: ptr(42.5)},
	}

	storage.On("GetAll", mock.Anything).Return(metrics, nil)
	fileStorage.On("Save", metrics).Return(nil)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := service.Start(ctx)
		assert.NoError(t, err)
	}()

	time.Sleep(2 * time.Second) // Даём время для срабатывания ticker
	cancel()
	wg.Wait()

	assert.NotNil(t, service.ticker)
	fileStorage.AssertCalled(t, "Save", metrics)
	storage.AssertCalled(t, "GetAll", mock.Anything)
}

func TestStop(t *testing.T) {
	storage := &mockStorage{}
	fileStorage := &mockFileStorage{}
	service := &MetricsService{storage: storage, fileStorage: fileStorage, ticker: time.NewTicker(time.Second)}
	metrics := map[string]models.Metric{
		"gauge1": {ID: "gauge1", MType: "gauge", Value: ptr(42.5)},
	}

	storage.On("GetAll", mock.Anything).Return(metrics, nil)
	fileStorage.On("Save", metrics).Return(nil)
	fileStorage.On("Close").Return(nil)

	err := service.Stop(context.Background())
	assert.NoError(t, err)
	fileStorage.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func ptr[T any](v T) *T {
	return &v
}

func BenchmarkServiceSaveAll(b *testing.B) {
	fileStorage := &mockFileStorage{}
	storage := &mockStorage{}
	service := NewMetricsService(fileStorage, storage, 0, false)

	metricCounts := []int{10, 100, 1000}
	for _, count := range metricCounts {
		metrics := make([]models.Metric, count)
		for i := 0; i < count; i++ {
			if i%2 == 0 {
				metrics[i] = models.Metric{
					ID:    "gauge" + strconv.Itoa(i),
					MType: models.Gauge,
					Value: utils.MakePointer(float64(i)),
				}
			} else {
				metrics[i] = models.Metric{
					ID:    "counter" + strconv.Itoa(i/2),
					MType: models.Counter,
					Delta: utils.MakePointer(int64(i)),
				}
			}
		}

		storage.On("SaveAll", mock.Anything, mock.Anything).Return(nil)
		storage.On("GetAll", mock.Anything).Return(map[string]models.Metric{}, nil)
		fileStorage.On("Save", mock.Anything).Return(nil)

		b.Run("Metrics"+strconv.Itoa(count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := service.SaveAll(context.Background(), metrics)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
