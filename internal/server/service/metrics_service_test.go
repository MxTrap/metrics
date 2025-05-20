package service

import (
	"context"
	commonmodels "github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/common/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strconv"
	"testing"
	"time"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetAll(ctx context.Context) (map[string]commonmodels.Metric, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]commonmodels.Metric), args.Error(1)
}

func (m *MockStorage) Find(ctx context.Context, metric string) (commonmodels.Metric, error) {
	args := m.Called(ctx, metric)
	return args.Get(0).(commonmodels.Metric), args.Error(1)
}

func (m *MockStorage) Save(ctx context.Context, metric commonmodels.Metric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockStorage) SaveAll(ctx context.Context, metrics map[string]commonmodels.Metric) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) Save(metrics map[string]commonmodels.Metric) error {
	args := m.Called(metrics)
	return args.Error(0)
}

func (m *MockFileStorage) Read() (map[string]commonmodels.Metric, error) {
	args := m.Called()
	return args.Get(0).(map[string]commonmodels.Metric), args.Error(1)
}

func (m *MockFileStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func BenchmarkServiceSaveAll(b *testing.B) {
	fileStorage := &MockFileStorage{}
	storage := &MockStorage{}
	service := NewMetricsService(fileStorage, storage, 0, false)

	metricCounts := []int{10, 100, 1000}
	for _, count := range metricCounts {
		metrics := make([]commonmodels.Metric, count)
		for i := 0; i < count; i++ {
			if i%2 == 0 {
				metrics[i] = commonmodels.Metric{
					ID:    "gauge" + strconv.Itoa(i),
					MType: commonmodels.Gauge,
					Value: utils.MakePointer(float64(i)),
				}
			} else {
				metrics[i] = commonmodels.Metric{
					ID:    "counter" + strconv.Itoa(i/2),
					MType: commonmodels.Counter,
					Delta: utils.MakePointer(int64(i)),
				}
			}
		}

		storage.On("SaveAll", mock.Anything, mock.Anything).Return(nil)
		storage.On("GetAll", mock.Anything).Return(map[string]commonmodels.Metric{}, nil)
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

func TestNewMetricsService(t *testing.T) {
	fileStorage := &MockFileStorage{}
	storage := &MockStorage{}
	saveInterval := 10
	restore := true

	service := NewMetricsService(fileStorage, storage, saveInterval, restore)

	assert.NotNil(t, service, "Service should not be nil")
	assert.Equal(t, fileStorage, service.fileStorage, "fileStorage should match")
	assert.Equal(t, storage, service.storage, "Storage should match")
	assert.Equal(t, saveInterval, service.saveInterval, "SaveInterval should match")
	assert.Equal(t, restore, service.restore, "Restore should match")
	assert.Nil(t, service.ticker, "Ticker should be nil initially")
}

func TestMetricsService_validateMetric(t *testing.T) {
	service := &MetricsService{}

	tests := []struct {
		name       string
		metricType string
		expected   bool
	}{
		{
			name:       "Valid Gauge",
			metricType: commonmodels.Gauge,
			expected:   true,
		},
		{
			name:       "Valid Counter",
			metricType: commonmodels.Counter,
			expected:   true,
		},
		{
			name:       "Invalid Type",
			metricType: "unknown",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.validateMetric(tt.metricType)
			assert.Equal(t, tt.expected, result, "ValidateMetric should return correct result")
		})
	}
}

func TestMetricsService_Save(t *testing.T) {
	fileStorage := &MockFileStorage{}
	storage := &MockStorage{}
	service := NewMetricsService(fileStorage, storage, 0, false)

	tests := []struct {
		name           string
		metric         commonmodels.Metric
		storageErr     error
		fileStorageErr error
		expectedErr    error
	}{
		{
			name: "Valid Gauge",
			metric: commonmodels.Metric{
				ID:    "test",
				MType: commonmodels.Gauge,
				Value: utils.MakePointer(42.0),
			},
			storageErr:     nil,
			fileStorageErr: nil,
			expectedErr:    nil,
		},
		{
			name: "Valid Counter",
			metric: commonmodels.Metric{
				ID:    "test",
				MType: commonmodels.Counter,
				Delta: utils.MakePointer(int64(100)),
			},
			storageErr:     nil,
			fileStorageErr: nil,
			expectedErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.On("Save", mock.Anything, tt.metric).Return(tt.storageErr).Once()
			if tt.storageErr == nil && tt.expectedErr == nil {
				storage.On("GetAll", mock.Anything).Return(map[string]commonmodels.Metric{}, nil).Once()
				fileStorage.On("Save", mock.Anything).Return(tt.fileStorageErr).Once()
			}

			err := service.Save(context.Background(), tt.metric)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr, "Error should match")
			} else {
				assert.NoError(t, err, "Should not return error")
			}

			storage.AssertExpectations(t)
			fileStorage.AssertExpectations(t)
		})
	}
}

func TestMetricsService_SaveAll(t *testing.T) {
	fileStorage := &MockFileStorage{}
	storage := &MockStorage{}
	service := NewMetricsService(fileStorage, storage, 0, false)

	tests := []struct {
		name           string
		metrics        []commonmodels.Metric
		storageErr     error
		fileStorageErr error
		expectedErr    error
		expectedMap    map[string]commonmodels.Metric
	}{
		{
			name: "Valid Metrics",
			metrics: []commonmodels.Metric{
				{ID: "gauge1", MType: commonmodels.Gauge, Value: utils.MakePointer(42.0)},
				{ID: "counter1", MType: commonmodels.Counter, Delta: utils.MakePointer(int64(100))},
			},
			storageErr:     nil,
			fileStorageErr: nil,
			expectedErr:    nil,
			expectedMap: map[string]commonmodels.Metric{
				"gauge1":   {ID: "gauge1", MType: commonmodels.Gauge, Value: utils.MakePointer(42.0)},
				"counter1": {ID: "counter1", MType: commonmodels.Counter, Delta: utils.MakePointer(int64(100))},
			},
		},
		{
			name: "Counter Aggregation",
			metrics: []commonmodels.Metric{
				{ID: "counter1", MType: commonmodels.Counter, Delta: utils.MakePointer(int64(100))},
				{ID: "counter1", MType: commonmodels.Counter, Delta: utils.MakePointer(int64(50))},
			},
			storageErr:     nil,
			fileStorageErr: nil,
			expectedErr:    nil,
			expectedMap: map[string]commonmodels.Metric{
				"counter1": {ID: "counter1", MType: commonmodels.Counter, Delta: utils.MakePointer(int64(150))},
			},
		},
		{
			name: "Invalid Metric Type",
			metrics: []commonmodels.Metric{
				{ID: "invalid", MType: "unknown", Value: utils.MakePointer(42.0)},
				{ID: "gauge1", MType: commonmodels.Gauge, Value: utils.MakePointer(42.0)},
			},
			storageErr:     nil,
			fileStorageErr: nil,
			expectedErr:    nil,
			expectedMap: map[string]commonmodels.Metric{
				"gauge1": {ID: "gauge1", MType: commonmodels.Gauge, Value: utils.MakePointer(42.0)},
			},
		},
		{
			name: "Storage Error",
			metrics: []commonmodels.Metric{
				{ID: "gauge1", MType: commonmodels.Gauge, Value: utils.MakePointer(42.0)},
			},
			storageErr:     assert.AnError,
			fileStorageErr: nil,
			expectedErr:    assert.AnError,
			expectedMap:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.On("SaveAll", mock.Anything, mock.Anything).Return(tt.storageErr).Once()
			if tt.storageErr == nil && tt.expectedErr == nil {
				storage.On("GetAll", mock.Anything).Return(map[string]commonmodels.Metric{}, nil).Once()
				fileStorage.On("Save", mock.Anything).Return(tt.fileStorageErr).Once()
			}

			err := service.SaveAll(context.Background(), tt.metrics)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr, "Error should match")
			} else {
				assert.NoError(t, err, "Should not return error")
				storage.AssertCalled(t, "SaveAll", mock.Anything, tt.expectedMap)
			}

			storage.AssertExpectations(t)
			fileStorage.AssertExpectations(t)
		})
	}
}

func TestMetricsService_Find(t *testing.T) {
	storage := &MockStorage{}
	service := NewMetricsService(nil, storage, 0, false)

	tests := []struct {
		name        string
		metric      commonmodels.Metric
		storageVal  commonmodels.Metric
		storageErr  error
		expected    commonmodels.Metric
		expectedErr error
	}{
		{
			name: "Found Gauge",
			metric: commonmodels.Metric{
				ID:    "test",
				MType: commonmodels.Gauge,
			},
			storageVal: commonmodels.Metric{
				ID:    "test",
				MType: commonmodels.Gauge,
				Value: utils.MakePointer(42.0),
			},
			storageErr: nil,
			expected: commonmodels.Metric{
				ID:    "test",
				MType: commonmodels.Gauge,
				Value: utils.MakePointer(42.0),
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.On("Find", mock.Anything, tt.metric.ID).Return(tt.storageVal, tt.storageErr).Once()

			result, err := service.Find(context.Background(), tt.metric)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr, "Error should match")
			} else {
				assert.NoError(t, err, "Should not return error")
				assert.Equal(t, tt.expected, result, "Metric should match")
			}

			storage.AssertExpectations(t)
		})
	}
}

func TestMetricsService_GetAll(t *testing.T) {
	storage := &MockStorage{}
	service := NewMetricsService(nil, storage, 0, false)

	tests := []struct {
		name        string
		storageData map[string]commonmodels.Metric
		storageErr  error
		expected    map[string]interface{}
		expectedErr error
	}{
		{
			name: "Valid Metrics",
			storageData: map[string]commonmodels.Metric{
				"gauge1":   {ID: "gauge1", MType: commonmodels.Gauge, Value: utils.MakePointer(42.0)},
				"counter1": {ID: "counter1", MType: commonmodels.Counter, Delta: utils.MakePointer(int64(100))},
			},
			storageErr: nil,
			expected: map[string]interface{}{
				"gauge1":   42.0,
				"counter1": int64(100),
			},
			expectedErr: nil,
		},
		{
			name:        "Empty Storage",
			storageData: map[string]commonmodels.Metric{},
			storageErr:  nil,
			expected:    map[string]interface{}{},
			expectedErr: nil,
		},
		{
			name:        "Storage Error",
			storageData: nil,
			storageErr:  assert.AnError,
			expected:    map[string]interface{}{},
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.On("GetAll", mock.Anything).Return(tt.storageData, tt.storageErr).Once()

			result, err := service.GetAll(context.Background())
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr, "Error should match")
			} else {
				assert.NoError(t, err, "Should not return error")
				assert.Equal(t, tt.expected, result, "Result should match")
			}

			storage.AssertExpectations(t)
		})
	}
}

func TestMetricsService_Ping(t *testing.T) {
	storage := &MockStorage{}
	service := NewMetricsService(nil, storage, 0, false)

	tests := []struct {
		name        string
		storageErr  error
		expectedErr error
	}{
		{
			name:        "Successful Ping",
			storageErr:  nil,
			expectedErr: nil,
		},
		{
			name:        "Ping Error",
			storageErr:  assert.AnError,
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.On("Ping", mock.Anything).Return(tt.storageErr).Once()

			err := service.Ping(context.Background())
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr, "Error should match")
			} else {
				assert.NoError(t, err, "Should not return error")
			}

			storage.AssertExpectations(t)
		})
	}
}

func TestMetricsService_Start(t *testing.T) {
	fileStorage := &MockFileStorage{}
	storage := &MockStorage{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name           string
		saveInterval   int
		restore        bool
		fileReadData   map[string]commonmodels.Metric
		fileReadErr    error
		storageSaveErr error
		expectedErr    error
	}{
		{
			name:         "Restore with sync save",
			saveInterval: 0,
			restore:      true,
			fileReadData: map[string]commonmodels.Metric{
				"gauge1": {ID: "gauge1", MType: commonmodels.Gauge, Value: utils.MakePointer(42.0)},
			},
			fileReadErr:    nil,
			storageSaveErr: nil,
			expectedErr:    nil,
		},
		{
			name:           "Async save",
			saveInterval:   1,
			restore:        false,
			fileReadData:   nil,
			fileReadErr:    nil,
			storageSaveErr: nil,
			expectedErr:    nil,
		},
		{
			name:           "File read error",
			saveInterval:   0,
			restore:        true,
			fileReadData:   nil,
			fileReadErr:    assert.AnError,
			storageSaveErr: nil,
			expectedErr:    assert.AnError,
		},
		{
			name:         "Storage save error",
			saveInterval: 0,
			restore:      true,
			fileReadData: map[string]commonmodels.Metric{
				"gauge1": {ID: "gauge1", MType: commonmodels.Gauge, Value: utils.MakePointer(42.0)},
			},
			fileReadErr:    nil,
			storageSaveErr: assert.AnError,
			expectedErr:    assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewMetricsService(fileStorage, storage, tt.saveInterval, tt.restore)
			if tt.restore {
				fileStorage.On("Read").Return(tt.fileReadData, tt.fileReadErr).Once()
				if tt.fileReadErr == nil {
					storage.On("SaveAll", mock.Anything, tt.fileReadData).Return(tt.storageSaveErr).Once()
				}
			}
			if tt.saveInterval > 0 && tt.expectedErr == nil {
				storage.On("GetAll", mock.Anything).Return(map[string]commonmodels.Metric{}, nil).Once()
				fileStorage.On("Save", mock.Anything).Return(nil).Once()
			}

			err := service.Start(ctx)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr, "Error should match")
			} else {
				assert.NoError(t, err, "Should not return error")
				if tt.saveInterval > 0 {
					// Даем время тикеру
					time.Sleep(1500 * time.Millisecond)
					cancel()
					// Ждем завершения горутины
					time.Sleep(100 * time.Millisecond)
					assert.NotNil(t, service.ticker, "Ticker should be initialized")
				}
			}

			fileStorage.AssertExpectations(t)
			storage.AssertExpectations(t)
		})
	}
}
