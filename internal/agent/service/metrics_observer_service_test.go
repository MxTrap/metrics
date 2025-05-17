package service

import (
	"context"
	"github.com/MxTrap/metrics/internal/agent/models"
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
	mockStorage := &MockMetricsStorage{}
	s := NewMetricsObserverService(mockStorage, 2)

	// Ожидаемые метрики
	expectedMetrics := models.Metrics{
		// Пример структуры, зависит от реализации models.Metrics
	}

	// Настройка мока
	mockStorage.On("GetMetrics").Return(expectedMetrics)

	// Вызов метода
	result := s.GetMetrics()

	// Проверка
	assert.Equal(t, expectedMetrics, result)
	mockStorage.AssertCalled(t, "GetMetrics")
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
