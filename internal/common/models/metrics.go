//go:generate easyjson -all -snake_case metrics.go

// Package models определяет структуры данных для представления метрик.
// Предоставляет типы Metric и Metrics для работы с метриками в формате JSON.
package models

// Metric представляет метрику с идентификатором, типом и значением.
// Поддерживает типы gauge (Value) и counter (Delta).
type Metric struct {
	ID    string   `json:"id"`              // Идентификатор метрики.
	MType string   `json:"type"`            // Тип метрики (gauge или counter).
	Delta *int64   `json:"delta,omitempty"` // Значение для метрик типа counter.
	Value *float64 `json:"value,omitempty"` // Значение для метрик типа gauge.
}

//easyjson:json
type Metrics []Metric
