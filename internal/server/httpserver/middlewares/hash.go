package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func HashDecodeMiddleware(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if key != "" && c.Request.Method != http.MethodGet {
			hashHeaderStr := c.Request.Header.Get("HashSHA256")
			if hashHeaderStr == "" {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			hashHeader, err := hex.DecodeString(hashHeaderStr)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			var bodyBytes []byte
			if c.Request.Body != nil {
				bodyBytes, _ = io.ReadAll(c.Request.Body)
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			h := hmac.New(sha256.New, []byte(key))

			h.Write(bodyBytes)

			if !hmac.Equal(hashHeader, h.Sum(nil)) {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}
	}
}

type responseWriter struct {
	body *bytes.Buffer
	gin.ResponseWriter
	key string
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)

	h := hmac.New(sha256.New, []byte(w.key))
	h.Write(b)
	w.Header().Set("HashSHA256", hex.EncodeToString(h.Sum(nil)))

	return w.ResponseWriter.Write(b)
}

func HashEncodeMiddleware(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if key != "" {
			w := &responseWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer, key: key}
			c.Writer = w
		}

		c.Next()
	}
}
