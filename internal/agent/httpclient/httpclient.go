package httpclient

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/MxTrap/metrics/internal/agent/models"
	commonmodels "github.com/MxTrap/metrics/internal/common/models"
	"github.com/mailru/easyjson"
	"io"
	"net/http"
	"sync"
	"time"
)

type metricsObserver interface {
	GetMetrics() models.Metrics
}

type HTTPClient struct {
	serverURL      string
	client         *http.Client
	service        metricsObserver
	reportInterval int
	key            string
	rateLimit      int
}

func NewHTTPClient(
	service metricsObserver,
	serverURL string,
	reportInterval int,
	key string,
	rateLimit int,
) *HTTPClient {
	client := &http.Client{}

	return &HTTPClient{
		client:         client,
		service:        service,
		reportInterval: reportInterval,
		serverURL:      serverURL,
		key:            key,
		rateLimit:      rateLimit,
	}
}

func (c *HTTPClient) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second * time.Duration(c.reportInterval))
	inCh := make(chan struct{})
	resCh := make(chan error)
	wg := &sync.WaitGroup{}
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(inCh)
				return
			case <-ticker.C:
				inCh <- struct{}{}
			}
		}
	}()
	go func() {
		for i := 0; i < c.rateLimit; i++ {
			wg.Add(1)
			go func(i int) {
				for {
					select {
					case <-ctx.Done():
						wg.Done()
						return
					case <-inCh:
						err := c.sendMetrics(ctx)
						if err != nil {
							resCh <- fmt.Errorf("error from gorutine %d: %w", i, err)
						}
					}
				}
			}(i)
			wg.Wait()
			close(resCh)
		}
	}()

	for res := range resCh {
		fmt.Println(res)
	}
	ticker.Stop()
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

func (c *HTTPClient) postMetric(ctx context.Context, metric commonmodels.Metrics) error {
	body, err := easyjson.Marshal(metric)

	if err != nil {
		return err
	}

	compressed, err := c.compress(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("http://%s/updates/", c.serverURL),
		compressed,
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	if c.key != "" {
		h := hmac.New(sha256.New, []byte(c.key))
		cBody, err := req.GetBody()
		if err != nil {
			return err
		}
		all, err := io.ReadAll(cBody)
		if err != nil {
			return err
		}
		h.Write(all)
		dst := h.Sum(nil)
		req.Header.Set("HashSHA256", hex.EncodeToString(dst))
	}

	const maxRetryAmount = 3
	var response *http.Response
	for i := 0; i <= maxRetryAmount; i++ {
		response, err = c.client.Do(req)
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

func (c *HTTPClient) sendMetrics(ctx context.Context) error {
	metrics := c.service.GetMetrics()

	m := make([]commonmodels.Metric, 0, len(metrics.Gauge.Metrics)+1)

	metrics.Gauge.Range(func(key string, value float64) {
		m = append(m, commonmodels.Metric{
			ID:    key,
			MType: commonmodels.Gauge,
			Value: &value,
		})
	})

	m = append(m, commonmodels.Metric{
		ID:    "PollCount",
		MType: commonmodels.Counter,
		Delta: &metrics.Counter.PollCount,
	})

	err := c.postMetric(ctx, m)
	if err != nil {
		return err
	}
	return nil
}
