package middlewares

import (
	"compress/gzip"
	"fmt"
	"github.com/gin-gonic/gin"
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

type writer struct {
	gin.ResponseWriter
	Writer *gzip.Writer
}

func (w writer) Write(b []byte) (int, error) {

	contentTypes := []string{
		"application/json; charset=utf-8",
		"text/html; charset=utf-8",
	}

	if slices.Contains(contentTypes, w.ResponseWriter.Header().Get("Content-Type")) {
		return w.Writer.Write(b)
	}

	return w.ResponseWriter.Write(b)
}

func AcceptEncodingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			return
		}

		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				fmt.Println(err)
			}
		}(gz)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		c.Writer = writer{
			c.Writer,
			gz,
		}
		c.Next()
		if c.Errors.Last() == nil {
			c.Writer.Header().Add("Content-Encoding", "gzip")
		}
	}
}
