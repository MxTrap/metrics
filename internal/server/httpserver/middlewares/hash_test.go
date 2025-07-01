package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupGinContext(method, path, body string, headers map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
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

func TestHashDecodeMiddleware(t *testing.T) {
	key := "secret"
	body := `{"data":"test"}`
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(body))
	validHash := hex.EncodeToString(h.Sum(nil))

	tests := []struct {
		name           string
		key            string
		url            string
		body           string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name: "Valid HMAC with /updates",
			key:  key,
			url:  "/updates",
			body: body,
			headers: map[string]string{
				"HashSHA256": validHash,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No key, skip middleware",
			key:            "",
			url:            "/updates",
			body:           body,
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-/updates URL, skip HMAC check",
			key:            key,
			url:            "/other",
			body:           body,
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing HashSHA256 header",
			key:            key,
			url:            "/updates",
			body:           body,
			headers:        map[string]string{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid HashSHA256 header",
			key:  key,
			url:  "/updates",
			body: body,
			headers: map[string]string{
				"HashSHA256": "invalid_hex",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Incorrect HMAC",
			key:  key,
			url:  "/updates",
			body: body,
			headers: map[string]string{
				"HashSHA256": hex.EncodeToString([]byte("wrong_hash")),
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupGinContext(http.MethodPost, tt.url, tt.body, tt.headers)
			middleware := HashDecodeMiddleware(tt.key)

			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code should match")

			if tt.expectedStatus == http.StatusOK {
				// Проверяем, что тело запроса осталось доступным
				body, err := io.ReadAll(c.Request.Body)
				assert.NoError(t, err, "Should read body without error")
				assert.Equal(t, tt.body, string(body), "Body should remain unchanged")
			}
		})
	}
}

func TestHashDecodeMiddleware_BodyReadError(t *testing.T) {
	key := "secret"
	body := `{"data":"test"}`
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(body))
	validHash := hex.EncodeToString(h.Sum(nil))

	// Создаем запрос с телом, которое нельзя прочитать
	c, w := setupGinContext(http.MethodPost, "/updates", "", map[string]string{
		"HashSHA256": validHash,
	})
	c.Request.Body = &errorReader{err: assert.AnError}

	middleware := HashDecodeMiddleware(key)
	middleware(c)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Should return BadRequest on body read error")
}

type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

func (r *errorReader) Close() error {
	return nil
}

func TestHashEncodeMiddlewareWithKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	key := "secret"
	middleware := HashEncodeMiddleware(key)
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test response")
	})

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())

	expectedHMAC := hmac.New(sha256.New, []byte(key))
	expectedHMAC.Write([]byte("test response"))
	expectedHash := hex.EncodeToString(expectedHMAC.Sum(nil))
	assert.Equal(t, expectedHash, w.Header().Get("HashSHA256"))
}

func TestHashEncodeMiddlewareEmptyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	key := ""
	middleware := HashEncodeMiddleware(key)
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test response")
	})

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Проверяем HTTP-ответ
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())

	// Проверяем отсутствие заголовка HashSHA256
	assert.Empty(t, w.Header().Get("HashSHA256"))
}

func TestResponseWriter(t *testing.T) {
	key := "secret"
	body := &bytes.Buffer{}
	writer := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(writer)

	responseWriter := &responseWriter{
		ResponseWriter: ctx.Writer,
		body:           body,
		key:            key,
	}

	// Пишем тестовые данные
	data := []byte("test data")
	n, err := responseWriter.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)

	// Проверяем тело ответа
	assert.Equal(t, "test data", body.String())
	assert.Equal(t, "test data", writer.Body.String())

	// Проверяем HMAC-подпись
	expectedHMAC := hmac.New(sha256.New, []byte(key))
	expectedHMAC.Write(data)
	expectedHash := hex.EncodeToString(expectedHMAC.Sum(nil))
	assert.Equal(t, expectedHash, writer.Header().Get("HashSHA256"))
}

func TestResponseWriterEmptyData(t *testing.T) {
	key := "secret"
	body := &bytes.Buffer{}
	writer := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(writer)
	responseWriter := &responseWriter{
		ResponseWriter: ctx.Writer,
		body:           body,
		key:            key,
	}

	// Пишем пустые данные
	data := []byte{}
	n, err := responseWriter.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)

	// Проверяем тело ответа
	assert.Empty(t, body.String())
	assert.Empty(t, writer.Body.String())

	// Проверяем HMAC-подпись для пустых данных
	expectedHMAC := hmac.New(sha256.New, []byte(key))
	expectedHMAC.Write(data)
	expectedHash := hex.EncodeToString(expectedHMAC.Sum(nil))
	assert.Equal(t, expectedHash, writer.Header().Get("HashSHA256"))
}
