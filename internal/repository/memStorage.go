package repository

type MemStorage struct {
	counter map[string]int64
	gauge   map[string]float64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counter: make(map[string]int64),
		gauge:   make(map[string]float64),
	}
}

func (m *MemStorage) IncrementCounter(metricName string, value int64) {
	m.counter[metricName] += value
}

func (m *MemStorage) UpdateGauge(metricName string, value float64) {
	m.gauge[metricName] = value
}
