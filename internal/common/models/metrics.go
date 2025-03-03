//go:generate easyjson -all -snake_case metrics.go
package models

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type Metrics struct {
	Data map[string]Metric `json:"data"`
}
