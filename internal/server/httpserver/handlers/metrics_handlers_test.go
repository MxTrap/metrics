package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	commonmodels "github.com/MxTrap/metrics/internal/common/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMetricService для мока интерфейса MetricService
type MockMetricService struct {
	mock.Mock
}

func (m *MockMetricService) Save(ctx context.Context, metric commonmodels.Metric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockMetricService) SaveAll(ctx context.Context, metrics []commonmodels.Metric) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func (m *MockMetricService) Find(ctx context.Context, metric commonmodels.Metric) (commonmodels.Metric, error) {
	args := m.Called(ctx, metric)
	return args.Get(0).(commonmodels.Metric), args.Error(1)
}

func (m *MockMetricService) GetAll(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockMetricService) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func setupRouter(h *MetricsHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h.router = r
	h.RegisterRoutes()
	return r
}

func TestParseMetric(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected commonmodels.Metric
		wantErr  bool
	}{
		{
			name:  "Valid gauge metric",
			input: []byte(`{"id":"testGauge","type":"gauge","value":42.5}`),
			expected: commonmodels.Metric{
				ID:    "testGauge",
				MType: commonmodels.Gauge,
				Value: float64Ptr(42.5),
			},
			wantErr: false,
		},
		{
			name:  "Valid counter metric",
			input: []byte(`{"id":"testCounter","type":"counter","delta":100}`),
			expected: commonmodels.Metric{
				ID:    "testCounter",
				MType: commonmodels.Counter,
				Delta: int64Ptr(100),
			},
			wantErr: false,
		},
		{
			name:     "Invalid JSON",
			input:    []byte(`{"id":"test","type":"gauge",value:42.5}`),
			expected: commonmodels.Metric{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := MetricsHandler{}
			result, err := h.parseMetric(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		search   string
		expected commonmodels.Metric
		wantErr  bool
	}{
		{
			name:   "Valid gauge URL",
			url:    "/update/gauge/testGauge/42.5",
			search: "update",
			expected: commonmodels.Metric{
				ID:    "testGauge",
				MType: commonmodels.Gauge,
				Value: float64Ptr(42.5),
			},
			wantErr: false,
		},
		{
			name:   "Valid counter URL",
			url:    "/update/counter/testCounter/100",
			search: "update",
			expected: commonmodels.Metric{
				ID:    "testCounter",
				MType: commonmodels.Counter,
				Delta: int64Ptr(100),
			},
			wantErr: false,
		},
		{
			name:     "Invalid gauge value",
			url:      "/update/gauge/testGauge/invalid",
			search:   "update",
			expected: commonmodels.Metric{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := MetricsHandler{}
			result, err := h.parseURL(tt.url, tt.search)
			fmt.Println(result)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetMetricValue(t *testing.T) {
	tests := []struct {
		name     string
		metric   commonmodels.Metric
		expected interface{}
	}{
		{
			name: "Gauge metric",
			metric: commonmodels.Metric{
				MType: commonmodels.Gauge,
				Value: float64Ptr(42.5),
			},
			expected: 42.5,
		},
		{
			name: "Counter metric",
			metric: commonmodels.Metric{
				MType: commonmodels.Counter,
				Delta: int64Ptr(100),
			},
			expected: int64(100),
		},
		{
			name:     "Unknown metric type",
			metric:   commonmodels.Metric{MType: "unknown"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := MetricsHandler{}
			result := h.getMetricValue(tt.metric)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSave(t *testing.T) {
	mockService := &MockMetricService{}
	h := NewMetricHandler(mockService, nil)
	router := setupRouter(h)

	tests := []struct {
		name           string
		url            string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "Successful save",
			url:  "/update/gauge/testGauge/42.5",
			mockSetup: func() {
				metric := commonmodels.Metric{ID: "testGauge", MType: commonmodels.Gauge, Value: float64Ptr(42.5)}
				mockService.On("Save", mock.Anything, metric).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req, _ := http.NewRequest("POST", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSaveJSON(t *testing.T) {
	mockService := &MockMetricService{}
	h := NewMetricHandler(mockService, nil)
	router := setupRouter(h)

	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "Successful save JSON",
			body: commonmodels.Metric{ID: "testGauge", MType: commonmodels.Gauge, Value: float64Ptr(42.5)},
			mockSetup: func() {
				metric := commonmodels.Metric{ID: "testGauge", MType: commonmodels.Gauge, Value: float64Ptr(42.5)}
				mockService.On("Save", mock.Anything, metric).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/update/", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}
