package repository

import "fmt"

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
	}
}

func (s *MemStorage) SaveGaugeMetric(metric string, value float64) {
	s.gauge[metric] = value
	fmt.Println(s.gauge)
}

func (s *MemStorage) SaveCounterMetric(metric string, value int64) {
	storedVal, ok := s.counter[metric]
	if !ok {
		storedVal = 0
	}
	s.counter[metric] = storedVal + value

	fmt.Println(s.counter)
}
