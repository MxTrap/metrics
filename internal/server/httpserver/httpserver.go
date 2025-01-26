package httpserver

import (
	"fmt"
	"net/http"

	"github.com/MxTrap/metrics/config"
	"github.com/MxTrap/metrics/internal/server/httpserver/handlers"
)

type HttpServer struct {
	mux  *http.ServeMux
	host string
}

func New(cfg config.HttpConfig, service handlers.MetricsSaver) *HttpServer {
	mux := http.NewServeMux()
	handler := handlers.NewHandler(service)
	mux.HandleFunc("/", handler.Save)

	return &HttpServer{
		mux:  mux,
		host: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}
}

func (h HttpServer) Run() {
	http.ListenAndServe(h.host, h.mux)
}
