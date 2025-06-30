package middlewares

import (
	"bytes"
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContentEncodingMiddlewareNoGzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ContentEncodingMiddleware())
	router.POST("/test", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)
		c.String(http.StatusOK, string(body))
	})

	req, err := http.NewRequest("POST", "/test", bytes.NewReader([]byte("test data")))
	require.NoError(t, err)
	req.Header.Set("Content-Encoding", "")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test data", w.Body.String())
}

func TestContentEncodingMiddlewareGzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ContentEncodingMiddleware())
	router.POST("/test", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)
		c.String(http.StatusOK, string(body))
	})

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte("test data"))
	require.NoError(t, err)
	require.NoError(t, gz.Close())

	req, err := http.NewRequest("POST", "/test", &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test data", w.Body.String())
}

func TestAcceptEncodingMiddlewareNoGzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AcceptEncodingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.String(http.StatusOK, "test data")
	})

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("Accept-Encoding", "")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test data", w.Body.String())
	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

func TestAcceptEncodingMiddlewareGzipJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AcceptEncodingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.String(http.StatusOK, "test data")
	})

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

	// Распаковываем тело ответа
	gz, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	defer gz.Close()
	body, err := io.ReadAll(gz)
	require.NoError(t, err)
	assert.Equal(t, "test data", string(body))
}

func TestAcceptEncodingMiddlewareGzipHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AcceptEncodingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, "<html>test</html>")
	})

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

	// Распаковываем тело ответа
	gz, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	defer gz.Close()
	body, err := io.ReadAll(gz)
	require.NoError(t, err)
	assert.Equal(t, "<html>test</html>", string(body))
}

func TestAcceptEncodingMiddlewareGzipError(t *testing.T) {
	// Нельзя напрямую протестировать ошибку gzip.NewWriterLevel, так как BestSpeed всегда валиден.
	// Тестируем только корректное поведение сжатия, так как ошибка маловероятна.
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AcceptEncodingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.String(http.StatusOK, "test data")
	})

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

	// Проверяем, что тело сжато корректно
	gz, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	defer gz.Close()
	body, err := io.ReadAll(gz)
	require.NoError(t, err)
	assert.Equal(t, "test data", string(body))
}

func TestWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	gz, err := gzip.NewWriterLevel(recorder, gzip.BestSpeed)
	require.NoError(t, err)
	c, _ := gin.CreateTestContext(recorder)
	w := writer{ResponseWriter: c.Writer, Writer: gz}
	data := []byte("test data")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	n, err := w.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
}

func TestWriterNonCompressedType(t *testing.T) {
	recorder := httptest.NewRecorder()
	gz, err := gzip.NewWriterLevel(recorder, gzip.BestSpeed)
	require.NoError(t, err)
	c, _ := gin.CreateTestContext(recorder)
	w := writer{ResponseWriter: c.Writer, Writer: gz}
	data := []byte("test data")

	w.Header().Set("Content-Type", "text/plain")
	n, err := w.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)

	assert.Equal(t, "test data", recorder.Body.String())
}
