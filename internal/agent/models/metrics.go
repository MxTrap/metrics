package models

type GaugeMetrics map[string]any

type CounterMetrics struct {
	PollCount   int64
	RandomValue int64
}

type Metrics struct {
	Counter CounterMetrics
	Gauge   GaugeMetrics
}
