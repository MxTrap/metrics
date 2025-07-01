package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/MxTrap/metrics/internal/common/models"
	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockMetricSvc struct {
	mock.Mock
}

func (m *mockMetricSvc) Save(ctx context.Context, metric models.Metric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *mockMetricSvc) SaveAll(ctx context.Context, metrics []models.Metric) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func (m *mockMetricSvc) Find(ctx context.Context, metric models.Metric) (models.Metric, error) {
	args := m.Called(ctx, metric)
	return args.Get(0).(models.Metric), args.Error(1)
}

func (m *mockMetricSvc) GetAll(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockMetricSvc) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewMetricHandler(t *testing.T) {
	service := &mockMetricSvc{}
	router := gin.New()
	handler := NewMetricHandler(service, router)
	assert.NotNil(t, handler)
	assert.Equal(t, service, handler.service)
	assert.Equal(t, router, handler.router)
}

func TestRegisterRoutes(t *testing.T) {
	service := &mockMetricSvc{}
	router := gin.New()
	handler := NewMetricHandler(service, router)
	handler.RegisterRoutes()

	routes := router.Routes()
	routePaths := make([]string, len(routes))
	for i, r := range routes {
		routePaths[i] = r.Path
	}

	expectedPaths := []string{
		"/value/:metricType/:metricName",
		"/update/",
		"/update/:metricType/:metricName/:metricValue",
		"/updates/",
		"/value/",
		"/",
		"/ping",
	}
	assert.ElementsMatch(t, expectedPaths, routePaths)
}

func TestParseMetric(t *testing.T) {
	handler := MetricsHandler{}
	metric := models.Metric{ID: "testGauge", MType: models.Gauge, Value: ptr(42.5)}
	data, err := easyjson.Marshal(metric)
	require.NoError(t, err)

	parsed, err := handler.parseMetric(data)
	require.NoError(t, err)
	assert.Equal(t, metric, parsed)

	invalidData := []byte("invalid json")
	_, err = handler.parseMetric(invalidData)
	assert.Error(t, err)
}

func TestParseURL(t *testing.T) {
	handler := MetricsHandler{}
	tests := []struct {
		name        string
		url         string
		searchWord  string
		expected    models.Metric
		expectError bool
	}{
		{
			name:       "ValidGauge",
			url:        "/update/gauge/testGauge/42.5",
			searchWord: "update",
			expected:   models.Metric{ID: "testGauge", MType: models.Gauge, Value: ptr(42.5)},
		},
		{
			name:       "ValidCounter",
			url:        "/update/counter/testCounter/100",
			searchWord: "update",
			expected:   models.Metric{ID: "testCounter", MType: models.Counter, Delta: ptr(int64(100))},
		},
		{
			name:        "InvalidURL",
			url:         "/update/invalid",
			searchWord:  "update",
			expectError: true,
		},
		{
			name:        "InvalidGaugeValue",
			url:         "/update/gauge/testGauge/invalid",
			searchWord:  "update",
			expectError: true,
		},
		{
			name:        "InvalidCounterValue",
			url:         "/update/counter/testCounter/invalid",
			searchWord:  "update",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := handler.parseURL(tt.url, tt.searchWord)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, parsed)
		})
	}
}

func TestGetMetricValue(t *testing.T) {
	handler := MetricsHandler{}
	tests := []struct {
		name     string
		metric   models.Metric
		expected interface{}
	}{
		{
			name:     "Gauge",
			metric:   models.Metric{MType: models.Gauge, Value: ptr(42.5)},
			expected: 42.5,
		},
		{
			name:     "Counter",
			metric:   models.Metric{MType: models.Counter, Delta: ptr(int64(100))},
			expected: int64(100),
		},
		{
			name:     "UnknownType",
			metric:   models.Metric{MType: "unknown"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.getMetricValue(tt.metric)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSaveAll(t *testing.T) {
	service := &mockMetricSvc{}
	router := gin.New()
	handler := NewMetricHandler(service, router)
	gin.SetMode(gin.TestMode)

	metrics := []models.Metric{
		{ID: "gauge1", MType: models.Gauge, Value: ptr(42.5)},
		{ID: "counter1", MType: models.Counter, Delta: ptr(int64(100))},
	}
	data, err := json.Marshal(metrics)
	require.NoError(t, err)

	service.On("SaveAll", mock.Anything, metrics).Return(nil)

	req, _ := http.NewRequest("POST", "/updates/", bytes.NewReader(data))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.saveAll(c)
	assert.Equal(t, http.StatusOK, w.Code)
	service.AssertExpectations(t)
}

func TestSaveJSON(t *testing.T) {
	service := &mockMetricSvc{}
	router := gin.New()
	handler := NewMetricHandler(service, router)
	gin.SetMode(gin.TestMode)

	metric := models.Metric{ID: "gauge1", MType: models.Gauge, Value: ptr(42.5)}
	data, err := easyjson.Marshal(metric)
	require.NoError(t, err)

	service.On("Save", mock.Anything, metric).Return(nil)

	req, _ := http.NewRequest("POST", "/update/", bytes.NewReader(data))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.saveJSON(c)
	assert.Equal(t, http.StatusOK, w.Code)
	service.AssertExpectations(t)
}

func TestSave(t *testing.T) {
	service := &mockMetricSvc{}
	router := gin.New()
	handler := NewMetricHandler(service, router)
	gin.SetMode(gin.TestMode)

	metric := models.Metric{ID: "gauge1", MType: models.Gauge, Value: ptr(42.5)}
	service.On("Save", mock.Anything, metric).Return(nil)

	req, _ := http.NewRequest("POST", "/update/gauge/gauge1/42.5", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.save(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFind(t *testing.T) {
	service := &mockMetricSvc{}
	router := gin.New()
	handler := NewMetricHandler(service, router)
	gin.SetMode(gin.TestMode)

	metric := models.Metric{ID: "gauge1", MType: models.Gauge, Value: ptr(42.5)}
	service.On("Find", mock.Anything, models.Metric{ID: "gauge1", MType: models.Gauge}).Return(metric, nil)

	req, _ := http.NewRequest("GET", "/value/gauge/gauge1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.find(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFindJSON(t *testing.T) {
	service := &mockMetricSvc{}
	router := gin.New()
	handler := NewMetricHandler(service, router)
	gin.SetMode(gin.TestMode)

	metric := models.Metric{ID: "gauge1", MType: models.Gauge, Value: ptr(42.5)}
	data, err := easyjson.Marshal(metric)
	require.NoError(t, err)

	service.On("Find", mock.Anything, metric).Return(metric, nil)

	req, _ := http.NewRequest("POST", "/value/", bytes.NewReader(data))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.findJSON(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var response models.Metric
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, metric, response)
	service.AssertExpectations(t)
}

func TestGetAll(t *testing.T) {
	service := &mockMetricSvc{}
	gin.SetMode(gin.TestMode)

	metrics := map[string]any{
		"gauge1":   models.Metric{ID: "gauge1", MType: models.Gauge, Value: ptr(42.5)},
		"counter1": models.Metric{ID: "counter1", MType: models.Counter, Delta: ptr(int64(100))},
	}
	service.On("GetAll", mock.Anything).Return(metrics, nil)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

}

func TestPing(t *testing.T) {
	service := &mockMetricSvc{}
	router := gin.New()
	handler := NewMetricHandler(service, router)
	gin.SetMode(gin.TestMode)

	service.On("Ping", mock.Anything).Return(nil)

	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.ping(c)
	assert.Equal(t, http.StatusOK, w.Code)
	service.AssertExpectations(t)
}

func ptr[T any](v T) *T {
	return &v
}
