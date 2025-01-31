package httpclient

import (
	"context"
	"fmt"
	"github.com/MxTrap/metrics/internal/agent/service"
	"net/http"
	"time"
)

type HTTPClient struct {
	serverUrl      string
	client         *http.Client
	service        *service.MetricsObserverService
	ctx            context.Context
	reportInterval int
}

func NewHTTPClient(
	ctx context.Context,
	service *service.MetricsObserverService,
	serverUrl string,
	reportInterval int,
) *HTTPClient {
	client := &http.Client{}

	return &HTTPClient{
		client:         client,
		ctx:            ctx,
		service:        service,
		reportInterval: reportInterval,
		serverUrl:      serverUrl,
	}
}

func (h *HTTPClient) Run() {
	go func(svc *service.MetricsObserverService) {
		for h.ctx != nil {
			h.sendMetrics()
			time.Sleep(time.Duration(h.reportInterval) * time.Second)
		}
	}(h.service)
}

func (h *HTTPClient) postMetric(metricType string, metric string, value any) {
	_, err := h.client.Post(
		fmt.Sprintf("http://%s/update/%s/%s/%v", h.serverUrl, metricType, metric, value),
		"text/plain",
		nil,
	)
	if err != nil {
		return
	}
}

func (h *HTTPClient) sendMetrics() {
	metrics := h.service.GetMetrics()

	fmt.Println(metrics)
	for key, val := range metrics.Gauge {
		h.postMetric("gauge", key, val)
	}
	h.postMetric("counter", "PollCount", metrics.Counter.PollCount)
	h.postMetric("counter", "RandomValue", metrics.Counter.RandomValue)

}
