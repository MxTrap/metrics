package models

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGaugeMetrics(t *testing.T) {
	gm := NewGaugeMetrics()
	assert.NotNil(t, gm.Metrics, "Metrics map should be initialized")
	assert.NotNil(t, gm.mx, "Mutex should be initialized")
	assert.Empty(t, gm.Metrics, "Metrics map should be empty")
}

func TestGaugeMetrics_Load(t *testing.T) {
	gm := NewGaugeMetrics()
	input := map[string]float64{
		"metric1": 42.5,
		"metric2": 100.0,
	}

	gm.Load(input)

	assert.Equal(t, input, gm.Metrics, "Load should copy input map to Metrics")
}

func TestGaugeMetrics_Get(t *testing.T) {
	gm := NewGaugeMetrics()
	gm.Metrics["metric1"] = 42.5

	tests := []struct {
		name     string
		key      string
		expected float64
		ok       bool
	}{
		{
			name:     "Existing key",
			key:      "metric1",
			expected: 42.5,
			ok:       true,
		},
		{
			name:     "Non-existing key",
			key:      "metric2",
			expected: 0.0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := gm.Get(tt.key)
			assert.Equal(t, tt.expected, val)
			assert.Equal(t, tt.ok, ok)
		})
	}
}

func TestGaugeMetrics_Set(t *testing.T) {
	gm := NewGaugeMetrics()

	gm.Set("metric1", 42.5)
	val, ok := gm.Get("metric1")
	assert.True(t, ok, "Set key should exist")
	assert.Equal(t, 42.5, val, "Set should store correct value")

	gm.Set("metric1", 100.0)
	val, ok = gm.Get("metric1")
	assert.True(t, ok, "Set key should still exist")
	assert.Equal(t, 100.0, val, "Set should update value")
}

func TestGaugeMetrics_Range(t *testing.T) {
	gm := NewGaugeMetrics()
	input := map[string]float64{
		"metric1": 42.5,
		"metric2": 100.0,
	}
	gm.Load(input)

	collected := make(map[string]float64)
	gm.Range(func(key string, value float64) {
		collected[key] = value
	})

	assert.Equal(t, input, collected, "Range should iterate over all metrics")
}

func TestGaugeMetrics_ConcurrentAccess(t *testing.T) {
	gm := NewGaugeMetrics()
	var wg sync.WaitGroup
	numGoroutines := 100
	keys := []string{"metric1", "metric2", "metric3"}

	// Запускаем горутины для конкурентной записи
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := keys[i%len(keys)]
			gm.Set(key, float64(i))
		}(i)
	}

	// Запускаем горутины для конкурентного чтения
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, key := range keys {
				gm.Get(key)
			}
		}()
	}

	// Запускаем горутины для конкурентного Range
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gm.Range(func(key string, value float64) {
				// Просто проходим по метрикам
			})
		}()
	}

	wg.Wait()

	// Проверяем, что данные не повреждены
	for _, key := range keys {
		_, ok := gm.Get(key)
		assert.True(t, ok, "Key %s should exist after concurrent access", key)
	}
}

func TestGaugeMetrics_ConcurrentLoad(t *testing.T) {
	gm := NewGaugeMetrics()
	var wg sync.WaitGroup
	numGoroutines := 100
	input := map[string]float64{
		"metric1": 42.5,
		"metric2": 100.0,
	}

	// Запускаем горутины для конкурентного вызова Load
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gm.Load(input)
		}()
	}

	wg.Wait()

	// Проверяем, что данные корректны
	for k, v := range input {
		val, ok := gm.Get(k)
		assert.True(t, ok, "Key %s should exist", k)
		assert.Equal(t, v, val, "Value for key %s should match", k)
	}
}
