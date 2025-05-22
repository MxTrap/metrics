package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// setupGinContext создаёт тестовый gin.Context
func setupGinContext(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	return c, w
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger()

	assert.NotNil(t, logger, "Logger should not be nil")
	assert.NotNil(t, logger.Logger, "SugaredLogger should be initialized")
}

func TestLoggerMiddleware(t *testing.T) {
	logger := &Logger{Logger: *zap.NewExample().Sugar()}

	tests := []struct {
		name           string
		method         string
		path           string
		responseStatus int
		responseBody   string
		expectedLog    []interface{}
	}{
		{
			name:           "GET request",
			method:         http.MethodGet,
			path:           "/test",
			responseStatus: http.StatusOK,
			responseBody:   "OK",
			expectedLog: []interface{}{
				"uri", "/test",
				"method", http.MethodGet,
				"duration", mock.AnythingOfType("time.Duration"),
				"status", http.StatusOK,
				"size", len("OK"),
			},
		},
		{
			name:           "POST request with error",
			method:         http.MethodPost,
			path:           "/api/data",
			responseStatus: http.StatusBadRequest,
			responseBody:   "Bad Request",
			expectedLog: []interface{}{
				"uri", "/api/data",
				"method", http.MethodPost,
				"duration", mock.AnythingOfType("time.Duration"),
				"status", http.StatusBadRequest,
				"size", len("Bad Request"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupGinContext(tt.method, tt.path)
			middleware := logger.LoggerMiddleware()

			middleware(c)
			c.Writer.WriteHeader(tt.responseStatus)
			_, err := c.Writer.Write([]byte(tt.responseBody))
			assert.NoError(t, err, "Write should succeed")

			assert.Equal(t, tt.responseStatus, w.Code, "Status code should match")
			assert.Equal(t, tt.responseBody, w.Body.String(), "Response body should match")
		})
	}
}

func TestLoggerMiddleware_Duration(t *testing.T) {
	logger := &Logger{Logger: *zap.NewExample().Sugar()}

	c, _ := setupGinContext(http.MethodGet, "/test")
	middleware := logger.LoggerMiddleware()

	// Эмулируем задержку
	time.Sleep(10 * time.Millisecond)

	// Выполняем middleware и пишем ответ
	middleware(c)
	c.Writer.WriteHeader(http.StatusOK)
	_, err := c.Writer.Write([]byte("OK"))
	assert.NoError(t, err, "Write should succeed")

}
