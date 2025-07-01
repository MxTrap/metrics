package grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/protos/gen"
	"github.com/MxTrap/metrics/internal/utils"
	"github.com/gogo/protobuf/proto"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	"sync"
	"time"
)

type metricsGetter interface {
	GetMetrics() models.Metrics
}

type Client struct {
	conn           *grpclib.ClientConn
	client         gen.MetricServiceClient
	service        metricsGetter
	reportInterval int
	key            string
	rateLimit      int
}

func NewClient(
	serverAddr string,
	service metricsGetter,
	reportInterval int,
	key string,
	rateLimit int,
) (*Client, error) {
	conn, err := grpclib.NewClient(serverAddr, grpclib.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := gen.NewMetricServiceClient(conn)

	return &Client{
		conn:           conn,
		client:         client,
		service:        service,
		reportInterval: reportInterval,
		key:            key,
		rateLimit:      rateLimit,
	}, nil
}

func (c *Client) postMetrics(ctx context.Context) error {
	metrics := c.service.GetMetrics()
	rMetrics := make([]*gen.Metric, len(metrics))
	for i, m := range metrics {
		rMetrics[i] = &gen.Metric{
			Id:    m.ID,
			Type:  m.MType,
			Value: m.Value,
			Delta: m.Delta,
		}
	}
	reqBody := &gen.SaveAllRequest{
		Metrics: rMetrics,
	}

	md := metadata.New(map[string]string{})
	md.Set("X-Real-IP", utils.GetLocalIP())

	if c.key != "" {
		h := hmac.New(sha256.New, []byte(c.key))
		marshal, err := proto.Marshal(reqBody)
		if err != nil {
			return err
		}

		h.Write(marshal)
		dst := h.Sum(nil)

		md.Set("HashSHA256", hex.EncodeToString(dst))
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	err := utils.Retry(func() error {
		_, err := c.client.SaveAll(ctx, reqBody)
		if err != nil {
			return err
		}
		return nil
	}, 3)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second * time.Duration(c.reportInterval))
	inCh := make(chan struct{})
	errCh := make(chan error)
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
						err := c.postMetrics(ctx)
						if err != nil {
							errCh <- fmt.Errorf("error from gorutine %d: %w", i, err)
						}
					}
				}
			}(i)
			wg.Wait()
			close(errCh)
		}
	}()

	for res := range errCh {
		fmt.Println(res)
	}
	ticker.Stop()
	err := c.conn.Close()
	if err != nil {
		log.Printf("error closing grpc client: %v", err)
		return
	}
}
