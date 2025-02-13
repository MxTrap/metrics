package models

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type GaugeMetrics map[string]float64

type CounterMetrics struct {
	PollCount   int64
	RandomValue int64
}

type Metrics struct {
	Counter CounterMetrics
	Gauge   GaugeMetrics
}
