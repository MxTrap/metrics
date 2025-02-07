package repository

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
}

func (s *MemStorage) SaveCounterMetric(metric string, value int64) {
	storedVal, ok := s.counter[metric]
	if !ok {
		storedVal = 0
	}
	s.counter[metric] = storedVal + value
}

func (s *MemStorage) FindGaugeMetric(metric string) (float64, bool) {
	value, ok := s.gauge[metric]
	return value, ok
}

func (s *MemStorage) FindCounterMetric(metric string) (int64, bool) {
	value, ok := s.counter[metric]
	return value, ok
}

func (s *MemStorage) GetAll() map[string]any {
	dst := map[string]any{}
	for k, v := range s.gauge {
		dst[k] = v
	}
	for k, v := range s.counter {
		dst[k] = v
	}
	return dst
}
