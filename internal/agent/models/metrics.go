package models

import (
	"sync"
)

type GaugeMetrics struct {
	mx      *sync.RWMutex
	metrics map[string]float64
}

func NewGaugeMetrics() *GaugeMetrics {
	return &GaugeMetrics{
		metrics: make(map[string]float64),
		mx:      &sync.RWMutex{},
	}
}

func (c *GaugeMetrics) Load(m map[string]float64) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.metrics = m
}

func (c *GaugeMetrics) Get(key string) (float64, bool) {
	c.mx.RLock()
	defer c.mx.RUnlock()
	val, ok := c.metrics[key]
	return val, ok
}

func (c *GaugeMetrics) Set(key string, value float64) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.metrics[key] = value
}

func (c *GaugeMetrics) Range(callback func(key string, value float64)) {
	c.mx.Lock()
	defer c.mx.Unlock()
	for k, v := range c.metrics {
		callback(k, v)
	}
}

type CounterMetrics struct {
	PollCount int64
}

type Metrics struct {
	Counter CounterMetrics
	Gauge   GaugeMetrics
}
