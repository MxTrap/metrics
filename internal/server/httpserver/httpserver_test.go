package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MxTrap/metrics/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLogger struct{}

func (m *mockLogger) LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func TestNewRouter(t *testing.T) {
	cfg := config.AddrConfig{
		Host: "localhost",
		Port: 8080,
	}
	log := &mockLogger{}
	key := "testkey"

	server := NewRouter(cfg, log, key, "", "")
	require.NotNil(t, server, "server should not be nil")
	assert.NotNil(t, server.Router, "router should not be nil")
	assert.NotNil(t, server.server, "http server should not be nil")
	assert.Equal(t, "localhost:8080", server.server.Addr, "server address should match config")

	server.Router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	server.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "request should succeed")
}

func TestStop(t *testing.T) {
	cfg := config.AddrConfig{
		Host: "localhost",
		Port: 0,
	}
	log := &mockLogger{}
	key := "testkey"

	server := NewRouter(cfg, log, key, "", "")

	go func() {
		_ = server.Run()
	}()

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := server.Stop(ctx)
	assert.NoError(t, err, "server should stop without error")

	resp, err := http.Get("http://localhost:" + fmt.Sprint(server.server.Addr[strings.Index(server.server.Addr, ":")+1:]) + "/test")
	assert.Error(t, err, "request to stopped server should fail")
	if resp != nil {
		_ = resp.Body.Close()
	}
}

func TestNewRouterInvalidTemplatesPath(t *testing.T) {
	cfg := config.AddrConfig{
		Host: "localhost",
		Port: 8080,
	}
	log := &mockLogger{}
	key := "testkey"

	server := NewRouter(cfg, log, key, "", "")
	assert.NotNil(t, server, "server should be created even with invalid templates path")

	server.Router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	server.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "request should succeed despite invalid templates")
}
