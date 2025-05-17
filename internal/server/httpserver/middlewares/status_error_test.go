package middlewares

import (
	"errors"
	"github.com/MxTrap/metrics/internal/server/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupGinContextStatusError() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, w
}

func TestStatusErrorMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "No error",
			err:            nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ErrNotFoundMetric",
			err:            models.ErrNotFoundMetric,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "ErrUnknownMetricType",
			err:            models.ErrUnknownMetricType,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ErrWrongMetricValue",
			err:            models.ErrWrongMetricValue,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Generic error",
			err:            errors.New("unexpected error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupGinContextStatusError()
			middleware := StatusErrorMiddleware()

			// Добавляем ошибку в c.Errors, если она есть
			if tt.err != nil {
				c.Error(tt.err)
			}

			// Вызываем middleware
			middleware(c)

			// Проверяем статус ответа
			assert.Equal(t, tt.expectedStatus, w.Code, "Status code should match")
		})
	}
}

func TestStatusErrorMiddleware_MultipleErrors(t *testing.T) {
	c, w := setupGinContextStatusError()
	middleware := StatusErrorMiddleware()

	// Добавляем несколько ошибок, последняя должна определять статус
	c.Error(errors.New("first error"))
	c.Error(models.ErrNotFoundMetric)

	middleware(c)

	assert.Equal(t, http.StatusNotFound, w.Code, "Status code should match last error (ErrNotFoundMetric)")
}
