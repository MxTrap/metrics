package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"github.com/MxTrap/metrics/internal/agent/models"
	"github.com/MxTrap/metrics/internal/agent/service"
	common_moodels "github.com/MxTrap/metrics/internal/common/models"
	"github.com/mailru/easyjson"
	"net/http"
	"time"
)

type HTTPClient struct {
	serverURL      string
	client         *http.Client
	service        *service.MetricsObserverService
	ctx            context.Context
	reportInterval int
}

func NewHTTPClient(
	ctx context.Context,
	service *service.MetricsObserverService,
	serverURL string,
	reportInterval int,
) *HTTPClient {
	client := &http.Client{}

	return &HTTPClient{
		client:         client,
		ctx:            ctx,
		service:        service,
		reportInterval: reportInterval,
		serverURL:      serverURL,
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

func (h *HTTPClient) postMetric(metric common_moodels.Metrics) error {

	body, err := easyjson.Marshal(metric)
	if err != nil {
		return err
	}
	resp, err := h.client.Post(
		fmt.Sprintf("http://%s/update", h.serverURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}
	return nil
}

func (h *HTTPClient) sendMetrics() {
	metrics := h.service.GetMetrics()

	for key, val := range metrics.Gauge {
		err := h.postMetric(common_moodels.Metrics{
			ID:    models.Gauge,
			MType: key,
			Value: &val,
		})
		if err != nil {
			return
		}
	}

	err := h.postMetric(common_moodels.Metrics{
		ID:    models.Counter,
		MType: "PollCount",
		Delta: &metrics.Counter.PollCount,
	})
	if err != nil {
		return
	}
	err = h.postMetric(common_moodels.Metrics{
		ID:    models.Counter,
		MType: "RandomValue",
		Delta: &metrics.Counter.RandomValue,
	})
	if err != nil {
		return
	}

}
