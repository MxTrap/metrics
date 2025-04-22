package service

import (
	"context"
	"github.com/MxTrap/metrics/internal/agent/mappers"
	"github.com/MxTrap/metrics/internal/agent/models"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
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
				return
			case <-ticker.C:
				service.collectMemStatMetrics()
			}
		}
	}(s)
	go func(service *MetricsObserverService) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				service.collectGopsutilMetrics()
			}
		}
	}(s)

	<-ctx.Done()
	ticker.Stop()
}

func (s *MetricsObserverService) collectMemStatMetrics() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	s.storage.SaveMetrics(mappers.MapGaugeMetrics(ms))
}

func (s *MetricsObserverService) collectGopsutilMetrics() {
	v, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	info, err := cpu.Percent(0, false)
	if err != nil || len(info) == 0 {
		return
	}
	s.storage.SaveMetrics(map[string]float64{
		"TotalMemory":     float64(v.Total),
		"FreeMemory":      float64(v.Free),
		"CPUutilization1": info[0],
	})
}

func (s *MetricsObserverService) GetMetrics() models.Metrics {
	return s.storage.GetMetrics()
}
