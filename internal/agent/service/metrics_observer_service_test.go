package service

import (
	"context"
	"github.com/MxTrap/metrics/internal/agent/models"
	common "github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type MockMetricsStorage struct {
	mock.Mock
}

func (m *MockMetricsStorage) SaveMetrics(metrics map[string]float64) {
	m.Called(metrics)
}

func (m *MockMetricsStorage) GetMetrics() models.Metrics {
	args := m.Called()
	return args.Get(0).(models.Metrics)
}

func TestGetMetrics(t *testing.T) {
	storage := &MockMetricsStorage{}
	service := NewMetricsObserverService(storage, 2)
	gaugeMetrics := models.NewGaugeMetrics()
	gaugeMetrics.Set("Alloc", 1000.0)
	gaugeMetrics.Set("Free", 500.0)
	gaugeMetrics.Set("TotalMemory", 500.0)

	metrics := models.Metrics{
		Gauge: *gaugeMetrics,
		Counter: models.CounterMetrics{
			PollCount: 5,
		},
	}
	storage.On("GetMetrics").Return(metrics)

	result := service.GetMetrics()

	expected := []common.Metric{
		{ID: "Alloc", MType: common.Gauge, Value: utils.MakePointer(1000.0)},
		{ID: "Free", MType: common.Gauge, Value: utils.MakePointer(500.0)},
		{ID: "TotalMemory", MType: common.Gauge, Value: utils.MakePointer(500.0)},
		{ID: "PollCount", MType: common.Counter, Delta: utils.MakePointer[int64](5)},
	}
	assert.ElementsMatch(t, expected, result)
	storage.AssertExpectations(t)
}

func TestRun(t *testing.T) {
	mockStorage := &MockMetricsStorage{}
	s := NewMetricsObserverService(mockStorage, 1) // Маленький интервал для быстрого теста

	// Подготовка контекста с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Ожидаем вызов SaveMetrics хотя бы один раз
	mockStorage.On("SaveMetrics", mock.Anything).Return()

	// Запуск Run в отдельной горутине
	go s.Run(ctx)

	// Ждем немного, чтобы горутины выполнили хотя бы один цикл
	time.Sleep(1500 * time.Millisecond)

	// Отменяем контекст
	cancel()

	// Даем время на завершение
	time.Sleep(100 * time.Millisecond)

	// Проверка, что SaveMetrics был вызван
	mockStorage.AssertCalled(t, "SaveMetrics", mock.Anything)
}

func TestRun_ContextCancellation(t *testing.T) {
	mockStorage := &MockMetricsStorage{}
	s := NewMetricsObserverService(mockStorage, 1)

	// Контекст с немедленной отменой
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Запуск Run
	s.Run(ctx)

	// Проверка, что SaveMetrics не был вызван
	mockStorage.AssertNotCalled(t, "SaveMetrics")
}
