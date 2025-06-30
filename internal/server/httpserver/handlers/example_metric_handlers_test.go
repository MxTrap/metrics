package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/MxTrap/metrics/internal/common/models"
	"github.com/gin-gonic/gin"
)

// mockMetricSvc реализует MetricService для тестовых примеров.
type mockMetricService struct {
	metrics map[string]models.Metric
}

func (m *mockMetricService) Save(_ context.Context, metric models.Metric) error {
	m.metrics[metric.ID] = metric
	return nil
}

func (m *mockMetricService) SaveAll(_ context.Context, metrics []models.Metric) error {
	for _, metric := range metrics {
		m.metrics[metric.ID] = metric
	}
	return nil
}

func (m *mockMetricService) Find(_ context.Context, metric models.Metric) (models.Metric, error) {
	if val, ok := m.metrics[metric.ID]; ok && val.MType == metric.MType {
		return val, nil
	}
	return models.Metric{}, fmt.Errorf("metric not found")
}

func (m *mockMetricService) GetAll(_ context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range m.metrics {
		if v.MType == models.Gauge && v.Value != nil {
			result[k] = *v.Value
		} else if v.MType == models.Counter && v.Delta != nil {
			result[k] = *v.Delta
		}
	}
	return result, nil
}

func (m *mockMetricService) Ping(_ context.Context) error {
	return nil
}

// ExampleMetricsHandler_saveJSON демонстрирует сохранение метрики через JSON-запрос.
func ExampleMetricsHandler_saveJSON() {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	service := &mockMetricService{metrics: make(map[string]models.Metric)}
	handler := NewMetricHandler(service, router)
	handler.RegisterRoutes()

	metric := models.Metric{
		ID:    "testGauge",
		MType: models.Gauge,
		Value: func() *float64 { v := 42.0; return &v }(),
	}
	body, _ := json.Marshal(metric)
	req, _ := http.NewRequest("POST", "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 200
}

// ExampleMetricsHandler_find демонстрирует получение значения метрики через URL.
func ExampleMetricsHandler_find() {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	service := &mockMetricService{metrics: make(map[string]models.Metric)}
	handler := NewMetricHandler(service, router)
	handler.RegisterRoutes()

	metric := models.Metric{
		ID:    "testGauge",
		MType: models.Gauge,
		Value: func() *float64 { v := 42.0; return &v }(),
	}
	service.Save(context.Background(), metric)

	req, _ := http.NewRequest("GET", "/value/gauge/testGauge", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code, w.Body.String())
	// Output: 200
}

// ExampleMetricsHandler_findJSON демонстрирует получение метрики через JSON-запрос.
func ExampleMetricsHandler_findJSON() {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	service := &mockMetricService{metrics: make(map[string]models.Metric)}
	handler := NewMetricHandler(service, router)
	handler.RegisterRoutes()

	metric := models.Metric{
		ID:    "testCounter",
		MType: models.Counter,
		Delta: func() *int64 { v := int64(100); return &v }(),
	}
	service.Save(context.Background(), metric)

	query := models.Metric{
		ID:    "testCounter",
		MType: models.Counter,
	}
	body, _ := json.Marshal(query)
	req, _ := http.NewRequest("POST", "/value/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var result models.Metric
	json.Unmarshal(w.Body.Bytes(), &result)
	fmt.Println(w.Code, result.ID, result.MType, *result.Delta)
	// Output: 200 testCounter counter 100
}

// ExampleMetricsHandler_saveAll демонстрирует сохранение нескольких метрик через JSON.
func ExampleMetricsHandler_saveAll() {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	service := &mockMetricService{metrics: make(map[string]models.Metric)}
	handler := NewMetricHandler(service, router)
	handler.RegisterRoutes()

	metrics := models.Metrics{
		{
			ID:    "testGauge",
			MType: models.Gauge,
			Value: func() *float64 { v := 42.0; return &v }(),
		},
		{
			ID:    "testCounter",
			MType: models.Counter,
			Delta: func() *int64 { v := int64(100); return &v }(),
		},
	}
	body, _ := json.Marshal(metrics)
	req, _ := http.NewRequest("POST", "/updates/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 200
}

// ExampleMetricsHandler_ping демонстрирует проверку доступности хранилища.
func ExampleMetricsHandler_ping() {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	service := &mockMetricService{metrics: make(map[string]models.Metric)}
	handler := NewMetricHandler(service, router)
	handler.RegisterRoutes()

	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
}
