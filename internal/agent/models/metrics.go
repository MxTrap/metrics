package models

import (
	"sync"
)

type GaugeMetrics struct {
	mx      *sync.RWMutex
	Metrics map[string]float64
}

func NewGaugeMetrics() *GaugeMetrics {
	return &GaugeMetrics{
		Metrics: make(map[string]float64),
		mx:      &sync.RWMutex{},
	}
}

func (c *GaugeMetrics) Load(m map[string]float64) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.Metrics = m
}

func (c *GaugeMetrics) Get(key string) (float64, bool) {
	c.mx.RLock()
	defer c.mx.RUnlock()
	val, ok := c.Metrics[key]
	return val, ok
}

func (c *GaugeMetrics) Set(key string, value float64) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.Metrics[key] = value
}

func (c *GaugeMetrics) Range(callback func(key string, value float64)) {
	c.mx.Lock()
	defer c.mx.Unlock()
	for k, v := range c.Metrics {
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
