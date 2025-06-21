package httpserver

import (
	"context"
	"fmt"
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/server/httpserver/middlewares"
	"github.com/MxTrap/metrics/internal/utils"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HTTPServer struct {
	server *http.Server
	Router *gin.Engine
}

type logger interface {
	LoggerMiddleware() gin.HandlerFunc
}

func NewRouter(cfg config.HTTPConfig, log logger, key string, cryptoKey string, cidr string) *HTTPServer {
	router := gin.New()
	router.Use(
		log.LoggerMiddleware(),
		gin.Recovery(),
		middlewares.IpValidator(cidr),
		middlewares.HashDecodeMiddleware(key),
		middlewares.ContentEncodingMiddleware(),
		middlewares.AcceptEncodingMiddleware(),
		middlewares.HashEncodeMiddleware(key),
		middlewares.StatusErrorMiddleware(),
	)
	if cryptoKey != "" {
		decrypter, err := middlewares.NewDecrypter(cryptoKey)
		if err != nil {
			panic(err)
		}
		router.Use(decrypter.DecrypterMiddleware())
	}
	router.HandleMethodNotAllowed = true
	router.LoadHTMLGlob(utils.GetProjectPath() + "/internal/server/templates/*")

	return &HTTPServer{
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler: router.Handler(),
		},
		Router: router,
	}
}

func (h HTTPServer) Run() error {
	pprof.Register(h.Router)
	return h.server.ListenAndServe()
}

func (h HTTPServer) Stop(ctx context.Context) error {
	err := h.server.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}
