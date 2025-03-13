package service

import (
	"context"
	"github.com/MxTrap/metrics/internal/agent/mappers"
	"github.com/MxTrap/metrics/internal/agent/models"
	"runtime"
	"time"
)

type MetricsStorage interface {
	SaveMetrics(map[string]float64)
	GetMetrics() models.Metrics
}

type MetricsObserverService struct {
	storage      MetricsStorage
	pollInterval int
}

func NewMetricsObserverService(service MetricsStorage, pollInterval int) *MetricsObserverService {
	return &MetricsObserverService{
		storage:      service,
		pollInterval: pollInterval,
	}
}

func (s *MetricsObserverService) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second * time.Duration(s.pollInterval))
	go func(service *MetricsObserverService) {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				service.CollectMetrics()
			}
		}
	}(s)
}

func (s *MetricsObserverService) CollectMetrics() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	s.storage.SaveMetrics(mappers.MapGaugeMetrics(ms))
}

func (s *MetricsObserverService) GetMetrics() models.Metrics {
	return s.storage.GetMetrics()
}
