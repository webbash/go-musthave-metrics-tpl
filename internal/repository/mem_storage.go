package repository

import "sync"

type MemStorage struct {
	mu      sync.RWMutex
	counter map[string]int64
	gauge   map[string]float64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counter: make(map[string]int64),
		gauge:   make(map[string]float64),
	}
}

func (m *MemStorage) GetAllGauges() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]float64, len(m.gauge))
	for name, val := range m.gauge {
		result[name] = val
	}
	return result
}

func (m *MemStorage) GetAllCounters() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]int64, len(m.counter))
	for name, val := range m.counter {
		result[name] = val
	}
	return result
}

func (m *MemStorage) GetCounter(metricName string) (int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.counter[metricName]
	return value, ok
}

func (m *MemStorage) GetGauge(metricName string) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.gauge[metricName]
	return value, ok
}

func (m *MemStorage) IncrementCounter(metricName string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter[metricName] += value
}

func (m *MemStorage) UpdateGauge(metricName string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauge[metricName] = value
}
