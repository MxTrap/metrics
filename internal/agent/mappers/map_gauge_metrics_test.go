package mappers

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestMapGaugeMetrics(t *testing.T) {
	tests := []struct {
		name     string
		input    runtime.MemStats
		expected map[string]float64
	}{
		{
			name: "Typical memory stats",
			input: runtime.MemStats{
				Alloc:         1000,
				BuckHashSys:   2000,
				Frees:         300,
				GCCPUFraction: 0.25,
				GCSys:         4000,
				HeapAlloc:     5000,
				HeapIdle:      6000,
				HeapInuse:     7000,
				HeapObjects:   800,
				HeapReleased:  9000,
				HeapSys:       10000,
				LastGC:        123456789,
				Lookups:       10,
				MCacheInuse:   11000,
				MCacheSys:     12000,
				MSpanInuse:    13000,
				MSpanSys:      14000,
				Mallocs:       1500,
				NextGC:        16000,
				NumForcedGC:   2,
				NumGC:         3,
				OtherSys:      17000,
				PauseTotalNs:  18000,
				StackInuse:    19000,
				StackSys:      20000,
				Sys:           21000,
				TotalAlloc:    22000,
			},
			expected: map[string]float64{
				"Alloc":         1000,
				"BuckHashSys":   2000,
				"Frees":         300,
				"GCCPUFraction": 0.25,
				"GCSys":         4000,
				"HeapAlloc":     5000,
				"HeapIdle":      6000,
				"HeapInuse":     7000,
				"HeapObjects":   800,
				"HeapReleased":  9000,
				"HeapSys":       10000,
				"LastGC":        123456789,
				"Lookups":       10,
				"MCacheInuse":   11000,
				"MCacheSys":     12000,
				"MSpanInuse":    13000,
				"MSpanSys":      14000,
				"Mallocs":       1500,
				"NextGC":        16000,
				"NumForcedGC":   2,
				"NumGC":         3,
				"OtherSys":      17000,
				"PauseTotalNs":  18000,
				"StackInuse":    19000,
				"StackSys":      20000,
				"Sys":           21000,
				"TotalAlloc":    22000,
			},
		},
		{
			name:  "Zero values",
			input: runtime.MemStats{},
			expected: map[string]float64{
				"Alloc":         0,
				"BuckHashSys":   0,
				"Frees":         0,
				"GCCPUFraction": 0,
				"GCSys":         0,
				"HeapAlloc":     0,
				"HeapIdle":      0,
				"HeapInuse":     0,
				"HeapObjects":   0,
				"HeapReleased":  0,
				"HeapSys":       0,
				"LastGC":        0,
				"Lookups":       0,
				"MCacheInuse":   0,
				"MCacheSys":     0,
				"MSpanInuse":    0,
				"MSpanSys":      0,
				"Mallocs":       0,
				"NextGC":        0,
				"NumForcedGC":   0,
				"NumGC":         0,
				"OtherSys":      0,
				"PauseTotalNs":  0,
				"StackInuse":    0,
				"StackSys":      0,
				"Sys":           0,
				"TotalAlloc":    0,
			},
		},
		{
			name: "Partial values",
			input: runtime.MemStats{
				Alloc:      500,
				HeapAlloc:  1000,
				NumGC:      5,
				TotalAlloc: 2000,
			},
			expected: map[string]float64{
				"Alloc":         500,
				"BuckHashSys":   0,
				"Frees":         0,
				"GCCPUFraction": 0,
				"GCSys":         0,
				"HeapAlloc":     1000,
				"HeapIdle":      0,
				"HeapInuse":     0,
				"HeapObjects":   0,
				"HeapReleased":  0,
				"HeapSys":       0,
				"LastGC":        0,
				"Lookups":       0,
				"MCacheInuse":   0,
				"MCacheSys":     0,
				"MSpanInuse":    0,
				"MSpanSys":      0,
				"Mallocs":       0,
				"NextGC":        0,
				"NumForcedGC":   0,
				"NumGC":         5,
				"OtherSys":      0,
				"PauseTotalNs":  0,
				"StackInuse":    0,
				"StackSys":      0,
				"Sys":           0,
				"TotalAlloc":    2000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapGaugeMetrics(tt.input)
			assert.Equal(t, tt.expected, result, "MapGaugeMetrics did not return expected values")
		})
	}
}
