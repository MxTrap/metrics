package models

type CounterMetrics struct {
	PollCount int64
}

type Metrics struct {
	Counter CounterMetrics
	Gauge   GaugeMetrics
}
