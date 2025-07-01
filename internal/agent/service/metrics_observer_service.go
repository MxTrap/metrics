// Package service предоставляет сервис для сбора и хранения системных метрик.
// Реализует MetricsObserverService, который периодически собирает метрики памяти и CPU, сохраняя их в хранилище.
package service

import (
	"context"
	"github.com/MxTrap/metrics/internal/agent/mappers"
	"github.com/MxTrap/metrics/internal/agent/models"
	common "github.com/MxTrap/metrics/internal/common/models"
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

// NewMetricsObserverService создаёт новый MetricsObserverService с указанным хранилищем и интервалом опроса.
// Возвращает указатель на инициализированный MetricsObserverService.
func NewMetricsObserverService(service MetricsStorage, pollInterval int) *MetricsObserverService {
	return &MetricsObserverService{
		storage:      service,
		pollInterval: pollInterval,
	}
}

// Run запускает сервис, периодически собирая метрики памяти и CPU.
// Выполняется до отмены контекста, после чего останавливает сбор метрик.
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

// collectMemStatMetrics собирает метрики памяти с использованием runtime.MemStats.
// Сохраняет метрики в хранилище через SaveMetrics.
func (s *MetricsObserverService) collectMemStatMetrics() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	s.storage.SaveMetrics(mappers.MapGaugeMetrics(ms))
}

// collectGopsutilMetrics собирает метрики памяти и CPU с использованием gopsutil.
// Сохраняет метрики общей и свободной памяти, а также использования CPU в хранилище.
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

// GetMetrics возвращает все метрики из хранилища.
// Возвращает массив models.Metrics, содержащую сохранённые метрики.
func (s *MetricsObserverService) GetMetrics() common.Metrics {
	metrics := s.storage.GetMetrics()
	m := make([]common.Metric, 0, len(metrics.Gauge.Metrics)+1)

	metrics.Gauge.Range(func(key string, value float64) {
		m = append(m, common.Metric{
			ID:    key,
			MType: common.Gauge,
			Value: &value,
		})
	})

	m = append(m, common.Metric{
		ID:    "PollCount",
		MType: common.Counter,
		Delta: &metrics.Counter.PollCount,
	})
	return m
}
