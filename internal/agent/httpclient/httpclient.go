package httpclient

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/MxTrap/metrics/internal/agent/service"
	common_models "github.com/MxTrap/metrics/internal/common/models"
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

func (h *HTTPClient) postMetric(metric common_models.Metrics) error {
	body, err := easyjson.Marshal(metric)

	if err != nil {
		return err
	}

	compressed, err := h.compress(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/updates/", h.serverURL), compressed)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	go func() {
		var response *http.Response
		for i := 0; i < 4; i++ {
			response, err = h.client.Do(req)
			if err == nil {
				err := response.Body.Close()
				if err != nil {
					return
				}
				break
			}
			if i < 3 {
				time.Sleep(time.Duration(1+2*i) * time.Second)
			}
		}
	}()

	return nil
}

func (h *HTTPClient) sendMetrics() {
	metrics := h.service.GetMetrics()

	m := make([]common_models.Metric, 20)

	metrics.Gauge.Range(func(key string, value float64) {
		m = append(m, common_models.Metric{
			ID:    key,
			MType: common_models.Gauge,
			Value: &value,
		})
	})

	m = append(m, common_models.Metric{
		ID:    "PollCount",
		MType: common_models.Counter,
		Delta: &metrics.Counter.PollCount,
	})

	err := h.postMetric(m)
	if err != nil {
		return
	}
}
