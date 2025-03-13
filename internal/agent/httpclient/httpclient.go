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
	reportInterval int
}

func NewHTTPClient(
	service *service.MetricsObserverService,
	serverURL string,
	reportInterval int,
) *HTTPClient {
	client := &http.Client{}

	return &HTTPClient{
		client:         client,
		service:        service,
		reportInterval: reportInterval,
		serverURL:      serverURL,
	}
}

func (h *HTTPClient) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second * time.Duration(h.reportInterval))
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				err := h.sendMetrics(ctx)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}()
}

func (*HTTPClient) compress(data []byte) (*bytes.Buffer, error) {
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

func (h *HTTPClient) postMetric(ctx context.Context, metric common_models.Metrics) error {
	body, err := easyjson.Marshal(metric)

	if err != nil {
		return err
	}

	compressed, err := h.compress(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("http://%s/updates/", h.serverURL),
		compressed,
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	const maxRetryAmount = 3
	var response *http.Response
	for i := 0; i <= maxRetryAmount; i++ {
		response, err = h.client.Do(req)
		if err == nil {
			err := response.Body.Close()
			if err != nil {
				return err
			}
			break
		}
		if i < maxRetryAmount {
			time.Sleep(time.Duration(1+2*i) * time.Second)
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (h *HTTPClient) sendMetrics(ctx context.Context) error {
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

	err := h.postMetric(ctx, m)
	if err != nil {
		return err
	}
	return nil
}
