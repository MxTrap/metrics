package middlewares

import (
	"bytes"
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupGinContextContentEncoding(method, path, body string, headers map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	c.Request = req
	return c, w
}

func createGzipBody(data string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	_, err = gz.Write([]byte(data))
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return &buf, nil
}

func TestContentEncodingMiddleware(t *testing.T) {
	body := `{"data":"test"}`
	gzipBody, err := createGzipBody(body)
	assert.NoError(t, err, "Failed to create gzip body")

	tests := []struct {
		name           string
		headers        map[string]string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "No Content-Encoding",
			headers:        map[string]string{},
			body:           body,
			expectedStatus: http.StatusOK,
			expectedBody:   body,
		},
		{
			name: "Valid gzip Content-Encoding",
			headers: map[string]string{
				"Content-Encoding": "gzip",
			},
			body:           gzipBody.String(),
			expectedStatus: http.StatusOK,
			expectedBody:   body,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupGinContextContentEncoding(http.MethodPost, "/", tt.body, tt.headers)
			middleware := ContentEncodingMiddleware()

			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code should match")

			if tt.expectedStatus == http.StatusOK {
				// Проверяем, что тело запроса доступно и корректно
				bodyBytes, err := io.ReadAll(c.Request.Body)
				assert.NoError(t, err, "Should read body without error")
				assert.Equal(t, tt.expectedBody, string(bodyBytes), "Request body should match")
			}
		})
	}
}

func TestAcceptEncodingMiddleware(t *testing.T) {
	body := `{"response":"ok"}`
	contentTypes := []string{
		"application/json; charset=utf-8",
		"text/html; charset=utf-8",
		"application/xml",
	}

	tests := []struct {
		name           string
		headers        map[string]string
		contentType    string
		body           string
		expectedStatus int
		expectGzip     bool
	}{
		{
			name:           "No Accept-Encoding",
			headers:        map[string]string{},
			contentType:    contentTypes[0],
			body:           body,
			expectedStatus: http.StatusOK,
			expectGzip:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupGinContextContentEncoding(http.MethodGet, "/", "", tt.headers)
			middleware := AcceptEncodingMiddleware()

			// Устанавливаем Content-Type
			c.Header("Content-Type", tt.contentType)

			// Эмулируем обработчик, который пишет ответ
			middleware(c)
			if w.Code != http.StatusBadRequest {
				_, err := c.Writer.Write([]byte(tt.body))
				assert.NoError(t, err, "Write should succeed")
			}

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code should match")

			if tt.expectGzip {
				assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"), "Content-Encoding should be gzip")

				// Проверяем, что тело сжато
				reader, err := gzip.NewReader(w.Body)
				assert.NoError(t, err, "Failed to create gzip reader")
				defer reader.Close()
				decompressed, err := io.ReadAll(reader)
				assert.NoError(t, err, "Failed to decompress body")
				assert.Equal(t, tt.body, string(decompressed), "Decompressed body should match")
			} else {
				assert.Empty(t, w.Header().Get("Content-Encoding"), "Content-Encoding should not be set")
				assert.Equal(t, tt.body, w.Body.String(), "Body should be unchanged")
			}
		})
	}
}
