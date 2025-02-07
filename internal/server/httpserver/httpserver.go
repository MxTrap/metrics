package httpserver

import (
	"fmt"
	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/server/httpserver/handlers"
	"github.com/MxTrap/metrics/internal/utils"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	router *gin.Engine
	host   string
}

func NewRouter(cfg config.HTTPConfig, service handlers.MetricService) *HTTPServer {
	router := gin.Default()
	router.HandleMethodNotAllowed = true
	router.LoadHTMLGlob(utils.GetProjectPath() + "/internal/server/templates/*")
	handler := handlers.NewHandler(service)
	uri := "/:metricType/:metricName"
	router.GET("/value"+uri, handler.Find)
	router.POST(fmt.Sprintf("/update/%s/:metricValue", uri), handler.Save)
	router.GET("/", handler.GetAll)

	return &HTTPServer{
		router: router,
		host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}
}

func (h HTTPServer) Run() {
	err := h.router.Run(h.host)
	if err != nil {
		return
	}
}
