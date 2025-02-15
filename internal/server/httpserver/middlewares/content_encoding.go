package middlewares

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"slices"
	"strings"
)

func ContentEncodingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.Header.Get("Content-Encoding"), "gzip") {
			c.Next()
			return
		}

		gz, err := gzip.NewReader(c.Request.Body)

		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Request.Body = gz
	}
}

func AcceptEncodingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			return
		}

		contentTypes := []string{
			"application/json",
			"text/html",
		}

		if slices.Contains(contentTypes, c.Request.Header.Get("Content-Type")) {
			type writer struct {
				gin.ResponseWriter
				Writer io.Writer
			}
			gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
			defer gz.Close()
			if err != nil {
				c.Status(http.StatusBadRequest)
				return
			}

			c.Header("Content-Encoding", "gzip")

			c.Writer = writer{
				c.Writer,
				gz,
			}
		}
	}
}
