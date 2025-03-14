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
	ctx          context.Context
}

func NewMetricsObserverService(ctx context.Context, service MetricsStorage, pollInterval int) *MetricsObserverService {
	return &MetricsObserverService{
		ctx:          ctx,
		storage:      service,
		pollInterval: pollInterval,
	}
}

func (s *MetricsObserverService) Run() {

	go func(service *MetricsObserverService) {
		for s.ctx != nil {
			s.CollectMetrics()
			time.Sleep(time.Duration(s.pollInterval) * time.Second)
		}
	}(s)
}

func (s *MetricsObserverService) Stop() {
	s.ctx = nil
}

func (s *MetricsObserverService) CollectMetrics() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	s.storage.SaveMetrics(mappers.MapGaugeMetrics(ms))
}

func (s *MetricsObserverService) GetMetrics() models.Metrics {
	return s.storage.GetMetrics()
}
