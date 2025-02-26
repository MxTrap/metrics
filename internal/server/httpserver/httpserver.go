package httpserver

import (
	"fmt"
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/server/httpserver/middlewares"
	"github.com/MxTrap/metrics/internal/utils"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	Router *gin.Engine
	host   string
}

type logger interface {
	LoggerMiddleware() gin.HandlerFunc
}

func NewRouter(cfg config.HTTPConfig, log logger) *HTTPServer {
	router := gin.New()
	router.Use(
		log.LoggerMiddleware(),
		gin.Recovery(),
		middlewares.ContentEncodingMiddleware(),
		middlewares.AcceptEncodingMiddleware(),
		middlewares.StatusErrorMiddleware(),
	)
	router.HandleMethodNotAllowed = true
	router.LoadHTMLGlob(utils.GetProjectPath() + "/internal/server/templates/*")

	return &HTTPServer{
		Router: router,
		host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}
}

func (h HTTPServer) Run() {
	err := h.Router.Run(h.host)
	if err != nil {
		return
	}
}
