package httpclient

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
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

func (HTTPClient) compress(data []byte) (*bytes.Buffer, error) {
	var b bytes.Buffer

	gz, err := gzip.NewWriterLevel(&b, gzip.BestSpeed)
	defer func(gz *gzip.Writer) {
		err := gz.Close()
		if err != nil {
			fmt.Println("failed to close gzip writer")
		}
	}(gz)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %v", err)
	}

	_, err = gz.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}

	err = gz.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}

	return &b, nil
}

func (h *HTTPClient) postMetric(metric common_moodels.Metrics) error {
	body, err := easyjson.Marshal(metric)

	if err != nil {
		return err
	}

	compressed, err := h.compress(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/update/", h.serverURL), compressed)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := h.client.Do(req)
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
			ID:    key,
			MType: common_moodels.Gauge,
			Value: &val,
		})
		if err != nil {
			return
		}
	}

	err := h.postMetric(common_moodels.Metrics{
		ID:    "PollCount",
		MType: common_moodels.Counter,
		Delta: &metrics.Counter.PollCount,
	})
	if err != nil {
		return
	}
}
