package httpserver

import (
	"fmt"
	"net/http"

	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/server/httpserver/handlers"
)

type HTTPServer struct {
	mux  *http.ServeMux
	host string
}

func New(cfg config.HTTPConfig, service handlers.MetricsSaver) *HTTPServer {
	mux := http.NewServeMux()
	handler := handlers.NewHandler(service)
	mux.HandleFunc("/", handler.Save)

	return &HTTPServer{
		mux:  mux,
		host: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}
}

func (h HTTPServer) Run() {
	http.ListenAndServe(h.host, h.mux)
}
