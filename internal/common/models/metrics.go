//go:generate easyjson -all -snake_case metrics.go
package models

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

//easyjson:json
type Metrics []Metric
